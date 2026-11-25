package manager

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/musenwill/mypass/errs"
	"github.com/musenwill/mypass/util"
)

// A store file is three lines:
//
//	mypass v2
//	{"kdf":"argon2id","m":65536,"t":3,"p":1,"salt":"…"}
//	<base64 of nonce || ciphertext || tag>
//
// The first two lines are the store header, in plaintext. The third is the
// container, AES-256-GCM under the pincode key. Inside the container is the token
// header on the first line, then the csv of the records.
//
// The lines are written with LF, and are read with CRLF as well. The store is
// text, git on Windows usually checks it out with CRLF, `core.autocrlf`.
const (
	// formatMagic is the first line of a store file. A file without it is a v1
	// store when it is one line of base64, see stateOf.
	formatMagic = "mypass v2"

	saltLen = 16

	// tokenCheck is sealed with the token key. Opening it proves a token is the right one.
	tokenCheck = "mypass token check"
)

// State tells what the store file is, no pincode is needed to know it.
type State int

const (
	Missing State = iota // no store yet, run `mypass init`
	Blank                // an empty file, the store of a fresh `mypass init`
	Legacy               // the v1 format, run `mypass upgrade`
	Current              // the v2 format
	Unknown              // neither one, damaged or written by a newer mypass
)

func (p State) String() string {
	switch p {
	case Missing:
		return "missing"
	case Blank:
		return "blank"
	case Legacy:
		return "v1"
	case Unknown:
		return "unknown"
	}
	return "v2"
}

// storeHeader tells how the container key comes from the pincode. The kdf params
// live here, so they can be tuned without another file format.
type storeHeader struct {
	util.KdfParams
	Salt string `json:"salt"`
}

// tokenHeader is the first line inside the container. It does the same as
// storeHeader for the token key, and keeps the check sealed with that key.
type tokenHeader struct {
	util.KdfParams
	Salt  string `json:"salt"`
	Check string `json:"check"`
}

// stateOf tells the state of a store file. A v1 store is one line of base64.
func stateOf(content []byte) State {
	body := bytes.TrimSpace(content)
	switch {
	case len(body) == 0:
		return Blank
	case firstLine(content) == formatMagic:
		return Current
	case !bytes.ContainsAny(body, "\r\n"):
		if _, err := base64.StdEncoding.DecodeString(string(body)); err == nil {
			return Legacy
		}
	}
	return Unknown
}

// firstLine is the first line of a store file, without its line ending.
func firstLine(content []byte) string {
	line := content
	if i := bytes.IndexByte(content, '\n'); i >= 0 {
		line = content[:i]
	}
	return strings.TrimRight(string(line), "\r")
}

func newStoreHeader(pincode []byte, kdf util.KdfParams) (*storeHeader, []byte, error) {
	salt, key, err := newKey(pincode, kdf)
	if err != nil {
		return nil, nil, err
	}
	return &storeHeader{KdfParams: kdf, Salt: base64.StdEncoding.EncodeToString(salt)}, key, nil
}

// key derives the container key of a loaded header.
func (p *storeHeader) key(pincode []byte) ([]byte, error) {
	salt, err := base64.StdEncoding.DecodeString(p.Salt)
	if err != nil {
		return nil, errs.InvalidFormat
	}
	return p.KdfParams.Derive(pincode, salt)
}

// file is the whole content of a store file, with a fresh nonce on every call.
func (p *storeHeader) file(key, inner []byte) ([]byte, error) {
	line, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	body, err := crypt2base64(util.NewCrypto(key, aadOf(line)), inner)
	if err != nil {
		return nil, err
	}

	return []byte(formatMagic + "\n" + string(line) + "\n" + body + "\n"), nil
}

// open decrypts the container of a store file.
func (p *storeHeader) open(key []byte, body string) ([]byte, error) {
	line, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return decryptFromBase64(util.NewCrypto(key, aadOf(line)), body)
}

// aadOf is what the container is bound to: the magic line and the header line
// as written, so nobody can change the kdf params or the salt unnoticed.
func aadOf(header []byte) []byte {
	return append([]byte(formatMagic+"\n"), header...)
}

func newTokenHeader(token []byte, kdf util.KdfParams) (*tokenHeader, []byte, error) {
	salt, key, err := newKey(token, kdf)
	if err != nil {
		return nil, nil, err
	}

	check, err := crypt2base64(util.NewCrypto(key, nil), []byte(tokenCheck))
	if err != nil {
		return nil, nil, err
	}
	return &tokenHeader{KdfParams: kdf, Salt: base64.StdEncoding.EncodeToString(salt), Check: check}, key, nil
}

// verify derives the token key, and fails unless the token is the right one.
func (p *tokenHeader) verify(token []byte) ([]byte, error) {
	salt, err := base64.StdEncoding.DecodeString(p.Salt)
	if err != nil {
		return nil, errs.InvalidToken
	}

	key, err := p.KdfParams.Derive(token, salt)
	if err != nil {
		return nil, errs.InvalidToken
	}

	if _, err := decryptFromBase64(util.NewCrypto(key, nil), p.Check); err != nil {
		return nil, errs.InvalidToken
	}
	return key, nil
}

func newKey(factor []byte, kdf util.KdfParams) (salt, key []byte, err error) {
	salt, err = util.RandomBytes(saltLen)
	if err != nil {
		return nil, nil, err
	}

	key, err = kdf.Derive(factor, salt)
	if err != nil {
		return nil, nil, err
	}
	return salt, key, nil
}

// packInner is the content of the container, the token header on the first line,
// then the csv of the records.
func packInner(token *tokenHeader, csv string) ([]byte, error) {
	line, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	return append(append(line, '\n'), csv...), nil
}

func unpackInner(inner []byte) (*tokenHeader, string, error) {
	line := inner
	csv := ""
	if i := bytes.IndexByte(inner, '\n'); i >= 0 {
		line, csv = inner[:i], string(inner[i+1:])
	}

	token := &tokenHeader{}
	if err := json.Unmarshal(line, token); err != nil {
		return nil, "", errs.InvalidFormat
	}
	return token, csv, nil
}
