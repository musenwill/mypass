package manager

import (
	"bytes"
	"encoding/base64"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/musenwill/mypass/errs"
	"github.com/musenwill/mypass/util"
)

var (
	testPincode = []byte("pin")
	testToken   = []byte("token")

	// cheap enough for tests, the params live in the store file anyway
	fastKdf = util.KdfParams{Name: "argon2id", Memory: 8, Time: 1, Threads: 1}
)

// legacyStoreV1 is a store written by the v1 code itself, with testPincode and
// testToken. Its times are the wall-clock of a utc+8 machine.
const legacyStoreV1 = "EGsXVjddOopkN4GSnB+Gse0WmCptGWOcVEV1Zq8YEWl+N6RkwThxnMq1bgMWBb+oFoAK6ekA2ddGXp9Kvo/MXC17e9mwy3aGPUNaGn1RW1ROc/miIRdMSsa08XCrBHPWA+wRn1sg0Ruwk0ZMarNROEYVbQs7ZqRHEK/V7Tuxo13Y9rYrO/wuj+Kx2ots0oSxzTM72PpMWFDEiXAZZXgGcU4/O6X0v5kJLcWRI3zxh4HpU5I50ir59yA4zSRCopfBEYQLv9/W4gB2jN1tvk/QIYDv6HNrlEuE4OQdQvqlM1VWOsiIkxTjrIxYaQEL/picjR4FHkUOz87OQXwq2mOxF5JPrAArVOzzbWDKOaT8zZBSxmKeEFoFiUZsPTf/ZXbtqEePWCTENf6BnmGh930rOA+um+lRRcSZ/ea0zsUxlGlc6OpyIOaIGh56/Q2yv+Ff399PgIID+XFMXeHCCyOMOflHVRdlQ82eQrdz7o7qoEO3vYvO/sm6HTeDeg2OvgOeBqvfnpYjp2CqtTY+RoiS/Sebvqxe+6CYxc/rMwr8TdbHJJTe8Flk+LcUCWpCMf6loEO4pfpPqVgsiTgPx+WsGpHu+RxvSbHbQaa/Bq2heAYD+rN2W2fC4k6gyQCFd2QKOy2OPiFUeKF6PHDQ5QYlmdr1yF+Ms4RNZojOHhx+9VXN5cS6Ln2Fhv8naJ+P9TiRi8Mb7nXlecQsvAdi8JQGeg=="

// legacyPasswords is by title, newest first.
var legacyPasswords = map[string][]string{
	"qq":          {"admin2", "admin1", "admin0"},
	"wx":          {"admin"},
	"mail.qq.com": {"admin"},
	"支付宝":         {"password"},
	"淘宝":          {"password"},
}

var cst = time.FixedZone("CST", 8*3600)

func TestState(t *testing.T) {
	srv := &impl{passPath: filepath.Join(t.TempDir(), "pass"), kdf: fastKdf}
	assertState(t, srv, Missing)

	must(t, os.WriteFile(srv.passPath, nil, 0600))
	assertState(t, srv, Blank)

	must(t, os.WriteFile(srv.passPath, []byte(legacyStoreV1), 0600))
	assertState(t, srv, Legacy)

	// a v2 file with a damaged first line is no v1 store
	must(t, os.WriteFile(srv.passPath, []byte("mypass v3\n{}\n\n"), 0600))
	assertState(t, srv, Unknown)
}

// A store checked out with CRLF line endings, by git on Windows, reads like
// one with LF.
func TestCRLFStore(t *testing.T) {
	srv := newTestSrv(t)
	must(t, srv.Load(testPincode))
	must(t, srv.UnlockToken(testToken))
	put(t, srv, "tencent", "qq", "admin")

	content := read(t, srv.passPath)
	must(t, os.WriteFile(srv.passPath, bytes.ReplaceAll(content, []byte("\n"), []byte("\r\n")), 0600))

	again := &impl{passPath: srv.passPath, kdf: fastKdf}
	assertState(t, again, Current)
	if err := again.Load(testPincode); err != nil {
		t.Fatalf("got %v expected the store to load with CRLF line endings", err)
	}
}

func TestFreshStore(t *testing.T) {
	srv := newTestSrv(t)
	assertState(t, srv, Blank)

	must(t, srv.Load(testPincode))
	if srv.TokenSet() {
		t.Error("got a token expected a fresh store to have no token")
	}
	if _, err := srv.Put("tencent", "qq", "admin", ""); err != errs.TokenUnset {
		t.Errorf("got %v expected no password to be stored before the token is set", err)
	}

	// a fresh store takes the token, nothing is there to verify it against
	must(t, srv.UnlockToken(testToken))
	if !srv.TokenSet() {
		t.Error("got no token expected the token to be set")
	}

	put(t, srv, "tencent", "qq", "admin")
	put(t, srv, "tencent", "wx", "admin")
	put(t, srv, "ali", "支付宝", "password")

	// sns follow the group + title order
	for sn, title := range map[int]string{1: "支付宝", 2: "qq", 3: "wx"} {
		if r, err := srv.GetBySn(sn); err != nil || r.Title != title {
			t.Errorf("got %v, %v expected %v on sn %v", r, err, title, sn)
		}
	}
	if _, err := srv.GetBySn(4); err != errs.SnNotFound {
		t.Errorf("got %v expected no such sn", err)
	}

	// the file is a v2 one, readable by its owner only
	assertState(t, srv, Current)
	content := read(t, srv.passPath)
	if lines := strings.SplitN(string(content), "\n", 3); len(lines) != 3 || lines[0] != formatMagic {
		t.Errorf("got %q expected the store to start with %q", firstLine(content), formatMagic)
	} else if !strings.Contains(lines[1], `"kdf":"argon2id"`) {
		t.Errorf("got %v expected the kdf params in the header", lines[1])
	}
	if mode := stat(t, srv.passPath).Mode().Perm(); mode != 0600 {
		t.Errorf("got mode %v expected 0600", mode)
	}

	// the next command finds everything back
	again, err := reopen(srv.passPath, testPincode, testToken)
	must(t, err)
	if !again.TokenSet() {
		t.Error("got no token expected the token to be kept")
	}
	if r, err := again.GetBySn(2); err != nil || r.Password != "admin" {
		t.Errorf("got %v, %v expected the password of qq", r, err)
	}

	// sns move with the accounts, on every save
	must(t, again.Delete("", "qq"))
	must(t, again.Save())
	onceMore, err := reopen(srv.passPath, testPincode, testToken)
	must(t, err)
	if r, err := onceMore.GetBySn(2); err != nil || r.Title != "wx" {
		t.Errorf("got %v, %v expected wx to take sn 2", r, err)
	}
}

func TestWrongPincode(t *testing.T) {
	srv := newTestSrv(t)
	must(t, srv.Load(testPincode))
	must(t, srv.UnlockToken(testToken))
	put(t, srv, "tencent", "qq", "admin")

	wrong := &impl{passPath: srv.passPath, kdf: fastKdf}
	if err := wrong.Load([]byte("another pin")); err != errs.InvalidPincode {
		t.Errorf("got %v expected an invalid pincode error", err)
	}
}

func TestWrongTokenLeavesFileAlone(t *testing.T) {
	srv := newTestSrv(t)
	must(t, srv.Load(testPincode))
	must(t, srv.UnlockToken(testToken))
	put(t, srv, "tencent", "qq", "admin")
	before := read(t, srv.passPath)

	again, err := reopen(srv.passPath, testPincode, nil)
	must(t, err)
	if err := again.UnlockToken([]byte("another token")); err != errs.InvalidToken {
		t.Errorf("got %v expected an invalid token error", err)
	}

	// no password is read or written with a wrong token
	if _, err := again.Put("tencent", "wx", "admin", ""); err != errs.TokenUnset {
		t.Errorf("got %v expected no password to be stored", err)
	}
	if _, err := again.Get("qq"); err != errs.TokenUnset {
		t.Errorf("got %v expected no password to be read", err)
	}
	if _, err := again.History("qq"); err != errs.TokenUnset {
		t.Errorf("got %v expected no password to be read", err)
	}
	if err := again.ResetKey(testPincode, testToken); err != errs.TokenUnset {
		t.Errorf("got %v expected the key not to be reset", err)
	}

	if after := read(t, srv.passPath); !bytes.Equal(before, after) {
		t.Error("got a changed store expected the file to be left alone")
	}
}

// Deleting needs the pincode alone, like every command that only lists. Saving
// touches no password, so it works without the token.
func TestDeleteNeedsNoToken(t *testing.T) {
	srv := newTestSrv(t)
	must(t, srv.Load(testPincode))
	must(t, srv.UnlockToken(testToken))
	put(t, srv, "tencent", "qq", "admin")
	put(t, srv, "tencent", "wx", "admin")

	again, err := reopen(srv.passPath, testPincode, nil)
	must(t, err)
	must(t, again.Delete("", "qq"))
	must(t, again.Save())

	onceMore, err := reopen(srv.passPath, testPincode, testToken)
	must(t, err)
	if _, err := onceMore.Get("qq"); err != errs.DataNotFound {
		t.Errorf("got %v expected qq to be deleted", err)
	}
	if r, err := onceMore.GetBySn(1); err != nil || r.Title != "wx" || r.Password != "admin" {
		t.Errorf("got %v, %v expected wx to be left, with sn 1", r, err)
	}

	// deleting what is not there is no success
	if err := onceMore.Delete("", "qq"); err != errs.NothingToDelete {
		t.Errorf("got %v expected nothing to delete", err)
	}
	if err := onceMore.Delete("", ""); err != errs.NothingToDelete {
		t.Errorf("got %v expected nothing to delete without a group or a title", err)
	}
}

func TestTamperedFile(t *testing.T) {
	srv := newTestSrv(t)
	must(t, srv.Load(testPincode))
	must(t, srv.UnlockToken(testToken))
	put(t, srv, "tencent", "qq", "admin")
	lines := strings.SplitN(string(read(t, srv.passPath)), "\n", 3)

	cases := map[string][]byte{
		// a changed container, or a changed header, both break the authentication
		"changed container": rewriteLine(lines, 2, func(line string) string {
			return flip(line, len(line)/2)
		}),
		"changed kdf": rewriteLine(lines, 1, func(line string) string {
			return strings.Replace(line, `"m":8`, `"m":16`, 1)
		}),
		"changed salt": rewriteLine(lines, 1, func(line string) string {
			return flip(line, len(line)-3)
		}),
		"bad header":   rewriteLine(lines, 1, func(string) string { return "not json" }),
		"truncated":    []byte(formatMagic + "\n" + lines[1] + "\n"),
		"no container": []byte(formatMagic + "\n" + lines[1] + "\n\n"),
	}

	for name, tampered := range cases {
		must(t, os.WriteFile(srv.passPath, tampered, 0600))

		again := &impl{passPath: srv.passPath, kdf: fastKdf}
		if err := again.Load(testPincode); err == nil {
			t.Errorf("got no err expected the store to be rejected on %v", name)
		}
	}
}

func TestSaveFailureKeepsOldFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root writes into a read-only directory")
	}

	srv := newTestSrv(t)
	must(t, srv.Load(testPincode))
	must(t, srv.UnlockToken(testToken))
	put(t, srv, "tencent", "qq", "admin")
	before := read(t, srv.passPath)

	dir := filepath.Dir(srv.passPath)
	must(t, os.Chmod(dir, 0500))
	defer os.Chmod(dir, 0700)

	if _, err := srv.Put("tencent", "wx", "admin", ""); err != nil {
		t.Fatal(err)
	}
	if err := srv.Save(); err == nil {
		t.Error("got no err expected the save to fail")
	}

	must(t, os.Chmod(dir, 0700))
	if after := read(t, srv.passPath); !bytes.Equal(before, after) {
		t.Error("got a changed store expected the old file to survive the failure")
	}
	// the temp file of the failed save is gone too
	if left, _ := filepath.Glob(filepath.Join(dir, ".*.tmp")); len(left) != 0 {
		t.Errorf("got %v expected no temp file left behind", left)
	}
}

func TestResetKey(t *testing.T) {
	srv := newTestSrv(t)
	must(t, srv.Load(testPincode))
	must(t, srv.UnlockToken(testToken))
	put(t, srv, "tencent", "qq", "admin")
	put(t, srv, "tencent", "qq", "admin2")
	before := read(t, srv.passPath)

	must(t, srv.ResetKey([]byte("new pin"), []byte("new token")))
	must(t, srv.Save())
	if after := read(t, srv.passPath); bytes.Equal(before, after) {
		t.Error("got the same store expected new salts in the header")
	}

	// both factors are needed with their new values
	again, err := reopen(srv.passPath, []byte("new pin"), []byte("new token"))
	must(t, err)
	history, err := again.History("qq")
	if err != nil || len(history) != 2 || history[0].Password != "admin2" || history[1].Password != "admin" {
		t.Errorf("got %v, %v expected the history of qq to be kept", history, err)
	}

	old := &impl{passPath: srv.passPath, kdf: fastKdf}
	if err := old.Load(testPincode); err != errs.InvalidPincode {
		t.Errorf("got %v expected the old pincode to be rejected", err)
	}
	if _, err := reopen(srv.passPath, []byte("new pin"), testToken); err != errs.InvalidToken {
		t.Errorf("got %v expected the old token to be rejected", err)
	}
}

func TestUpgrade(t *testing.T) {
	original := time.Local
	defer func() { time.Local = original }()
	time.Local = cst

	srv := &impl{passPath: filepath.Join(t.TempDir(), "pass"), kdf: fastKdf}
	must(t, os.WriteFile(srv.passPath, []byte(legacyStoreV1), 0600))
	assertState(t, srv, Legacy)

	// other commands don't touch a v1 store
	if err := srv.Load(testPincode); err != errs.LegacyStore {
		t.Errorf("got %v expected a legacy store error", err)
	}

	// a wrong factor changes nothing
	if _, err := srv.Upgrade([]byte("another pin"), testToken); err != errs.InvalidPincode {
		t.Errorf("got %v expected an invalid pincode error", err)
	}
	if _, err := srv.Upgrade(testPincode, []byte("another token")); err != errs.InvalidToken {
		t.Errorf("got %v expected an invalid token error", err)
	}
	if content := read(t, srv.passPath); string(content) != legacyStoreV1 {
		t.Error("got a changed store expected the v1 store to be left alone")
	}
	if _, err := os.Stat(srv.passPath + legacyBackupSuffix); !os.IsNotExist(err) {
		t.Error("got a backup expected none on failure")
	}

	backup, err := srv.Upgrade(testPincode, testToken)
	must(t, err)
	if act, exp := backup, srv.passPath+legacyBackupSuffix; act != exp {
		t.Errorf("got %v expected the backup at %v", act, exp)
	}
	if content := read(t, backup); string(content) != legacyStoreV1 {
		t.Error("got a changed backup expected the v1 store as it was")
	}
	assertState(t, srv, Current)

	// every password, its history and its time made it through
	again, err := reopen(srv.passPath, testPincode, testToken)
	must(t, err)
	for title, exp := range legacyPasswords {
		history, err := again.History(title)
		if err != nil {
			t.Errorf("got %v expected the history of %v", err, title)
			continue
		}
		var got []string
		for _, r := range history {
			got = append(got, r.Password)
		}
		if strings.Join(got, ",") != strings.Join(exp, ",") {
			t.Errorf("got %v expected %v on %v", got, exp, title)
		}
	}

	wx, err := again.Get("wx")
	must(t, err)
	// v1 kept 12:31:47, the wall-clock of a utc+8 machine
	if act, exp := wx.Ct.UTC().Format(time.RFC3339), "2019-08-18T04:31:47Z"; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}
	if act, exp := wx.Describe, "for wx"; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}
	if act, exp := len(again.store.GetRecords()), 7; act != exp {
		t.Errorf("got %v records expected %v", act, exp)
	}

	// sns are given in the group + title order
	exp := map[int]string{1: "支付宝", 2: "淘宝", 3: "mail.qq.com", 4: "qq", 5: "wx"}
	for sn, title := range exp {
		if r, err := again.GetBySn(sn); err != nil || r.Title != title {
			t.Errorf("got %v, %v expected %v on sn %v", r, err, title, sn)
		}
	}

	// there is nothing left to upgrade
	if _, err := srv.Upgrade(testPincode, testToken); err == nil {
		t.Error("got no err expected nothing to upgrade on a v2 store")
	}
}

func TestPrintable(t *testing.T) {
	cases := map[string]bool{
		"":            true,
		"p@ssw0rd 中文": true,
		"\x01":        false,
		"\x7f":        false,
		"\xff":        false,
	}
	for content, exp := range cases {
		if act := printable([]byte(content)); act != exp {
			t.Errorf("got %v expected %v on %q", act, exp, content)
		}
	}
}

// MYPASS_HOME points at another store, the default is ~/.mypass.
func TestHomeEnv(t *testing.T) {
	t.Setenv(homeEnv, "")
	usr, err := user.Current()
	must(t, err)
	if act, exp := passfile(), usr.HomeDir+"/.mypass/pass"; act != exp {
		t.Errorf("got %v expected the store in ~/.mypass", act)
	}

	dir := filepath.Join(t.TempDir(), "store")
	t.Setenv(homeEnv, dir)
	if act, exp := passfile(), dir+"/pass"; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}

	// `mypass init` lays out the store there
	must(t, New().Init("git@example.com:musenwill/mypass.git"))
	if content := read(t, passfile()); len(content) != 0 {
		t.Errorf("got %v expected an empty store", content)
	}
	if conf := read(t, configfile()); !strings.Contains(string(conf), "git: git@example.com:musenwill/mypass.git") {
		t.Errorf("got %v expected the git url in the config", string(conf))
	}
}

func TestFormat(t *testing.T) {
	salt, key, err := newKey(testPincode, fastKdf)
	must(t, err)
	header := &storeHeader{KdfParams: fastKdf, Salt: base64.StdEncoding.EncodeToString(salt)}

	// a file reads back with the same key
	content, err := header.file(key, []byte("inner"))
	must(t, err)
	lines := strings.Split(string(content), "\n")
	if len(lines) != 4 || lines[0] != formatMagic || lines[3] != "" {
		t.Errorf("got %v lines expected magic, header and container", len(lines))
	}
	inner, err := header.open(key, lines[2])
	if err != nil || string(inner) != "inner" {
		t.Errorf("got %v, %v expected the inner content back", string(inner), err)
	}

	// and not with a changed one, nor with a damaged container
	if _, err := header.open(flipBytes(key), lines[2]); err != errs.DecryptError {
		t.Errorf("got %v expected a decrypt error on another key", err)
	}
	if _, err := header.open(key, lines[2][:len(lines[2])-1]); err == nil {
		t.Error("got no err expected a failure on a truncated container")
	}

	// a fresh nonce on every file
	other, err := header.file(key, []byte("inner"))
	if err != nil || bytes.Equal(content, other) {
		t.Error("got the same file expected a fresh nonce on every save")
	}

	// the token check only opens with the token
	tokenHeader, tokenKey, err := newTokenHeader(testToken, fastKdf)
	must(t, err)
	if got, err := tokenHeader.verify(testToken); err != nil || !bytes.Equal(got, tokenKey) {
		t.Errorf("got %v expected the token key back, %v", got, err)
	}
	if _, err := tokenHeader.verify([]byte("another token")); err != errs.InvalidToken {
		t.Errorf("got %v expected an invalid token error", err)
	}

	// the inner content survives the round trip, with and without records
	for _, csv := range []string{"", "1,g,t,p,d,2019-08-18T12:31:47Z\n"} {
		inner, err := packInner(tokenHeader, csv)
		must(t, err)
		got, gotCsv, err := unpackInner(inner)
		if err != nil || gotCsv != csv || got.Check != tokenHeader.Check {
			t.Errorf("got %v, %v, %v expected the csv %q back", got, gotCsv, err, csv)
		}
	}
	if _, _, err := unpackInner([]byte("not json")); err != errs.InvalidFormat {
		t.Errorf("got %v expected an invalid format error", err)
	}
}

// newTestSrv gives the blank store of a fresh `mypass init`.
func newTestSrv(t *testing.T) *impl {
	t.Helper()

	srv := &impl{passPath: filepath.Join(t.TempDir(), "pass"), kdf: fastKdf}
	must(t, os.WriteFile(srv.passPath, nil, 0600))
	return srv
}

// reopen loads the store the way the next command does, without the token when
// token is nil.
func reopen(path string, pincode, token []byte) (*impl, error) {
	srv := &impl{passPath: path, kdf: fastKdf}
	if err := srv.Load(pincode); err != nil {
		return nil, err
	}
	if token == nil {
		return srv, nil
	}
	if err := srv.UnlockToken(token); err != nil {
		return nil, err
	}
	return srv, nil
}

func put(t *testing.T, srv SrvApi, group, title, password string) {
	t.Helper()

	if _, err := srv.Put(group, title, password, "for "+title); err != nil {
		t.Fatal(err)
	}
	must(t, srv.Save())
}

func assertState(t *testing.T, srv *impl, exp State) {
	t.Helper()

	if act, err := srv.State(); err != nil || act != exp {
		t.Errorf("got %v, %v expected the store to be %v", act, err, exp)
	}
}

func rewriteLine(lines []string, i int, rewrite func(string) string) []byte {
	copied := append([]string(nil), lines...)
	copied[i] = rewrite(lines[i])
	return []byte(strings.Join(copied, "\n"))
}

// flip swaps a character of the base64 text, so the decoded bytes change.
func flip(text string, i int) string {
	b := []byte(text)
	if b[i] == 'A' {
		b[i] = 'B'
	} else {
		b[i] = 'A'
	}
	return string(b)
}

func flipBytes(b []byte) []byte {
	flipped := append([]byte(nil), b...)
	flipped[0] ^= 0x01
	return flipped
}

func must(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) []byte {
	t.Helper()

	content, err := os.ReadFile(path)
	must(t, err)
	return content
}

func stat(t *testing.T, path string) os.FileInfo {
	t.Helper()

	info, err := os.Stat(path)
	must(t, err)
	return info
}
