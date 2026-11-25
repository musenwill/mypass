package manager

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/musenwill/mypass/data"
	"github.com/musenwill/mypass/errs"
	"github.com/musenwill/mypass/util"
)

type SrvApi interface {
	Init(gitUrl string) error
	All() ([]*data.Record, error)
	Olds(before time.Time) ([]*data.Record, error)
	Groups() ([]string, error)
	Titles() ([]string, error)
	Filter(groupLike, titleLike string) ([]*data.Record, error)
	// Delete removes a group, an account, or both. It fails when there is
	// nothing to remove.
	Delete(group, title string) error
	Put(group, title, password, describe string) (*data.Record, error)
	Get(title string) (*data.Record, error)
	GetBySn(sn int) (*data.Record, error)
	History(title string) ([]*data.Record, error)

	// State tells what the store file is, no pincode is needed.
	State() (State, error)
	// Load opens the store with the pincode. A blank file is a fresh store, the
	// pincode becomes its key on the first save.
	Load(pincode []byte) error
	// Empty tells whether the store has no record.
	Empty() (bool, error)
	// TokenSet tells whether the store already has a token.
	TokenSet() bool
	// UnlockToken verifies the token, or makes it the token of a fresh store.
	// It must be done before any password is read or written.
	UnlockToken(token []byte) error
	// ResetKey re-encrypts everything with the new factors. The token must be
	// unlocked first, the change lands on disk with the next Save.
	ResetKey(newPincode, newToken []byte) error
	// Upgrade turns a v1 store into v2, keeping the pincode and the token. It
	// returns where the v1 store is kept as a backup.
	Upgrade(pincode, token []byte) (string, error)
	Save() error
}

type impl struct {
	passPath string
	kdf      util.KdfParams

	store    data.Store
	storeHdr *storeHeader
	storeKey []byte
	tokenHdr *tokenHeader
	tokenKey []byte
}

func New() SrvApi {
	return &impl{passPath: passfile(), kdf: util.DefaultKdfParams}
}

// Init lays out the store directory, for its owner alone. A store that is
// already there is left alone.
func (p *impl) Init(gitUrl string) error {
	if !pathExists(passdir()) {
		if err := os.MkdirAll(passdir(), 0700); err != nil {
			return err
		}
	}
	if !pathExists(p.passPath) {
		if err := os.WriteFile(p.passPath, nil, 0600); err != nil {
			return err
		}
	}

	return saveConf(&config{Git: gitUrl}, configfile())
}

func (p *impl) All() ([]*data.Record, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	return p.store.ListAll()
}

func (p *impl) Olds(before time.Time) ([]*data.Record, error) {
	records, err := p.All()
	if err != nil {
		return nil, err
	}

	var result []*data.Record
	for _, r := range records {
		if r.Ct.Before(before) {
			result = append(result, r)
		}
	}

	return result, nil
}

func (p *impl) Groups() ([]string, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	return p.store.ListGroups()
}

func (p *impl) Titles() ([]string, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	return p.store.ListTitles()
}

func (p *impl) Filter(groupLike, titleLike string) ([]*data.Record, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	return p.store.Filter(groupLike, titleLike)
}

// Delete removes the group, the account, or both. It fails when nothing is
// there to remove, so a mistyped name never looks like a success.
func (p *impl) Delete(group, title string) error {
	if err := p.ensureLoaded(); err != nil {
		return err
	}

	removed := 0
	if strings.TrimSpace(group) != "" {
		removed += p.store.DeleteGroup(group)
	}
	if strings.TrimSpace(title) != "" {
		removed += p.store.DeleteTitle(title)
	}
	if removed == 0 {
		return errs.NothingToDelete
	}

	return nil
}

func (p *impl) Put(group, title, password, describe string) (*data.Record, error) {
	if err := p.ensureToken(); err != nil {
		return nil, err
	}

	encoded, err := sealPassword(p.tokenKey, title, password)
	if err != nil {
		return nil, err
	}
	return p.store.Put(group, title, encoded, describe)
}

func (p *impl) Get(title string) (*data.Record, error) {
	if err := p.ensureToken(); err != nil {
		return nil, err
	}

	record, err := p.store.Get(title)
	if err != nil {
		return nil, err
	}
	return p.decryptRecord(record)
}

func (p *impl) GetBySn(sn int) (*data.Record, error) {
	if err := p.ensureToken(); err != nil {
		return nil, err
	}

	record, err := p.store.GetBySn(sn)
	if err != nil {
		return nil, err
	}
	return p.decryptRecord(record)
}

func (p *impl) History(title string) ([]*data.Record, error) {
	if err := p.ensureToken(); err != nil {
		return nil, err
	}

	records, err := p.store.GetHistory(title)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errs.DataNotFound
	}

	var history []*data.Record
	for _, r := range records {
		record, err := p.decryptRecord(r)
		if err != nil {
			return nil, err
		}
		history = append(history, record)
	}

	return history, nil
}

func (p *impl) State() (State, error) {
	content, err := os.ReadFile(p.passPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Missing, nil
		}
		return Missing, err
	}

	return stateOf(content), nil
}

func (p *impl) Load(pincode []byte) error {
	content, err := os.ReadFile(p.passPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errs.Uninited
		}
		return err
	}

	switch stateOf(content) {
	case Blank:
		return p.loadFresh(pincode)
	case Current:
		return p.loadCurrent(content, pincode)
	case Legacy:
		return errs.LegacyStore
	}
	return errs.InvalidFormat
}

func (p *impl) loadFresh(pincode []byte) error {
	header, key, err := newStoreHeader(pincode, p.kdf)
	if err != nil {
		return err
	}

	store, err := data.New("")
	if err != nil {
		return err
	}

	p.storeHdr, p.storeKey, p.tokenHdr, p.tokenKey, p.store = header, key, nil, nil, store
	return nil
}

func (p *impl) loadCurrent(content, pincode []byte) error {
	lines := strings.SplitN(string(content), "\n", 3)
	if len(lines) != 3 || firstLine(content) != formatMagic {
		return errs.InvalidFormat
	}

	header := &storeHeader{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(lines[1])), header); err != nil {
		return errs.InvalidFormat
	}
	key, err := header.key(pincode)
	if err != nil {
		return errs.InvalidFormat
	}

	// a wrong pincode and a damaged file look the same here, the container can't tell
	inner, err := header.open(key, strings.TrimSpace(lines[2]))
	if err != nil {
		return errs.InvalidPincode
	}

	token, csv, err := unpackInner(inner)
	if err != nil {
		return err
	}
	store, err := data.New(csv)
	if err != nil {
		return errs.InvalidFormat
	}

	p.storeHdr, p.storeKey, p.tokenHdr, p.tokenKey, p.store = header, key, token, nil, store
	return nil
}

func (p *impl) Empty() (bool, error) {
	if err := p.ensureLoaded(); err != nil {
		return true, err
	}
	return len(p.store.GetRecords()) == 0, nil
}

func (p *impl) TokenSet() bool {
	return p.tokenHdr != nil
}

func (p *impl) UnlockToken(token []byte) error {
	if err := p.ensureLoaded(); err != nil {
		return err
	}

	if p.tokenHdr != nil {
		key, err := p.tokenHdr.verify(token)
		if err != nil {
			return err
		}
		p.tokenKey = key
		return nil
	}

	// a fresh store takes the token, nothing can be verified against yet
	header, key, err := newTokenHeader(token, p.kdf)
	if err != nil {
		return err
	}
	p.tokenHdr, p.tokenKey = header, key
	return nil
}

func (p *impl) ResetKey(newPincode, newToken []byte) error {
	if err := p.ensureToken(); err != nil {
		return err
	}

	// every password is opened with the old token key first, one that fails
	// leaves the store as it is
	records := p.store.GetRecords()
	plain := make([]*data.Record, 0, len(records))
	for _, r := range records {
		password, err := openPassword(p.tokenKey, r)
		if err != nil {
			return err
		}
		record := *r
		record.Password = string(password)
		plain = append(plain, &record)
	}

	return p.reseal(plain, newPincode, newToken)
}

func (p *impl) Upgrade(pincode, token []byte) (string, error) {
	content, err := os.ReadFile(p.passPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errs.Uninited
		}
		return "", err
	}
	if stateOf(content) != Legacy {
		return "", errs.InvalidFormat
	}

	records, err := readLegacy(content, pincode, token)
	if err != nil {
		return "", err
	}
	if err := p.reseal(records, pincode, token); err != nil {
		return "", err
	}

	// keep the v1 file before the v2 one takes its place
	backup := p.passPath + legacyBackupSuffix
	if err := writeFileAtomic(backup, content); err != nil {
		return "", err
	}

	return backup, p.Save()
}

// reseal seals the records, given with plaintext passwords, into a new store
// under new keys, and makes it the loaded one. The store loaded before is left
// alone when anything fails on the way.
func (p *impl) reseal(records []*data.Record, pincode, token []byte) error {
	storeHdr, storeKey, err := newStoreHeader(pincode, p.kdf)
	if err != nil {
		return err
	}
	tokenHdr, tokenKey, err := newTokenHeader(token, p.kdf)
	if err != nil {
		return err
	}

	store, err := data.New("")
	if err != nil {
		return err
	}
	for _, r := range records {
		password, err := sealPassword(tokenKey, r.Title, r.Password)
		if err != nil {
			return err
		}
		record := *r
		record.Password = password
		store.Add(&record)
	}

	p.storeHdr, p.storeKey, p.tokenHdr, p.tokenKey, p.store = storeHdr, storeKey, tokenHdr, tokenKey, store
	return nil
}

// Save writes the store to disk atomically, the old file survives any failure.
// It needs the pincode alone, records are kept encrypted in memory. The store
// has to have a token, the token key itself is not needed, saving touches no password.
func (p *impl) Save() error {
	if err := p.ensureLoaded(); err != nil {
		return err
	}
	if p.tokenHdr == nil {
		return errs.TokenUnset
	}

	csv, err := p.store.Save()
	if err != nil {
		return err
	}
	inner, err := packInner(p.tokenHdr, csv)
	if err != nil {
		return err
	}

	content, err := p.storeHdr.file(p.storeKey, inner)
	if err != nil {
		return err
	}
	return writeFileAtomic(p.passPath, content)
}

// ensureLoaded fails until the store is opened.
func (p *impl) ensureLoaded() error {
	if p.store == nil || p.storeHdr == nil {
		return errors.New("store is not loaded")
	}
	return nil
}

// ensureToken fails until the token is verified, every password goes through its key.
func (p *impl) ensureToken() error {
	if p.tokenKey == nil {
		return errs.TokenUnset
	}
	return nil
}

// decryptRecord gives a copy of the record, with the plaintext password. The
// loaded store keeps the encrypted one.
func (p *impl) decryptRecord(record *data.Record) (*data.Record, error) {
	password, err := openPassword(p.tokenKey, record)
	if err != nil {
		return nil, err
	}

	plain := *record
	plain.Password = string(password)
	return &plain, nil
}

// sealPassword encrypts a password with the key of its title.
func sealPassword(tokenKey []byte, title, password string) (string, error) {
	return crypt2base64(util.NewCrypto(util.RecordKey(tokenKey, title), []byte(title)), []byte(password))
}

func openPassword(tokenKey []byte, record *data.Record) ([]byte, error) {
	crypto := util.NewCrypto(util.RecordKey(tokenKey, record.Title), []byte(record.Title))
	password, err := decryptFromBase64(crypto, record.Password)
	if err != nil {
		return nil, errs.InvalidToken
	}
	return password, nil
}

// readLegacy decrypts a v1 store, and gives its records with plaintext passwords.
func readLegacy(content, pincode, token []byte) ([]*data.Record, error) {
	plain, err := decryptFromBase64(util.NewLegacyCrypto(pincode), strings.TrimSpace(string(content)))
	if err != nil {
		return nil, errs.InvalidPincode
	}

	store, err := data.NewLegacy(string(plain))
	if err != nil {
		return nil, errs.InvalidPincode
	}

	records := store.GetRecords()
	for _, r := range records {
		password, err := decryptFromBase64(util.NewLegacyHMacCrypto([]byte(r.Title), token), r.Password)
		if err != nil || !printable(password) {
			// v1 has no integrity check, a wrong token may decrypt into garbage
			return nil, errs.InvalidToken
		}
		r.Password = string(password)
	}

	return records, nil
}

// printable tells whether content can be a password typed by a user.
func printable(content []byte) bool {
	if !utf8.Valid(content) {
		return false
	}
	for _, c := range string(content) {
		if unicode.IsControl(c) {
			return false
		}
	}
	return true
}
