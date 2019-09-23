package manager

import (
	"strings"
	"time"

	"github.com/musenwill/mypass/data"
	"github.com/musenwill/mypass/errs"
	"github.com/musenwill/mypass/util"
)

type SrvApi interface {
	Init(gitUrl string) error
	All() ([]*data.Record, error)
	Olds(time time.Time) ([]*data.Record, error)
	Groups() ([]string, error)
	Titles() ([]string, error)
	Filter(groupLike, titleLike string) ([]*data.Record, error)
	Delete(group, title string) error
	Put(group, title, password, describe string) error
	Get(title string) (*data.Record, error)
	History(title string) ([]*data.Record, error)
	LoadOld(Oldpincode, token []byte) error
	Load() error
	Save() error
	Empty() (bool, error)
	SetStoreCrypto(crypto util.CryptoApi)
	SetRecordKey(key []byte)
	Migrate() error
}

type impl struct {
	store       data.Store
	storeCrypto util.CryptoApi
	recordkey   []byte
}

func New() SrvApi {
	return &impl{}
}

func (p *impl) Init(gitUrl string) error {
	if !pathExists(passdir()) {
		err := createDir(passdir())
		if err != nil {
			return err
		}
	}
	if !pathExists(passfile()) {
		err := createFile(passfile())
		if err != nil {
			return err
		}
	}
	if !pathExists(configfile()) {
		err := createFile(configfile())
		if err != nil {
			return err
		}
	}

	conf := &config{Git: gitUrl}
	return saveConf(conf, configfile())
}

func (p *impl) All() ([]*data.Record, error) {
	return p.store.ListAll()
}

func (p *impl) Olds(time time.Time) ([]*data.Record, error) {
	records, err := p.All()
	if err != nil {
		return nil, err
	}

	var result []*data.Record
	for _, r := range records {
		if r.Ct.Before(time) {
			result = append(result, r)
		}
	}

	return result, nil
}

func (p *impl) Groups() ([]string, error) {
	return p.store.ListGroups()
}

func (p *impl) Titles() ([]string, error) {
	return p.store.ListTitles()
}

func (p *impl) Filter(grouplike, titleLike string) ([]*data.Record, error) {
	return p.store.Filter(grouplike, titleLike)
}

func (p *impl) Delete(group, title string) error {
	if strings.TrimSpace(group) != "" {
		err := p.store.DeleteGroup(group)
		if err != nil {
			return err
		}
	}
	if strings.TrimSpace(title) != "" {
		err := p.store.DeleteTitle(title)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *impl) Put(group, title, password, describe string) error {
	crypto := util.NewHMacCrypto([]byte(title), p.recordkey)
	encoded, err := crypt2base64(crypto, []byte(password))
	if err != nil {
		return err
	}
	_, err = p.store.Put(group, title, encoded, describe)
	return err
}

func (p *impl) Get(title string) (*data.Record, error) {
	crypto := util.NewHMacCrypto([]byte(title), p.recordkey)
	r, err := p.store.Get(title)
	if err != nil {
		return nil, err
	}

	decoded, err := decryptFromBase64(crypto, r.Password)
	if err != nil {
		return nil, err
	}
	r.Password = string(decoded)
	return r, nil
}

func (p *impl) History(title string) ([]*data.Record, error) {
	lst, err := p.store.GetHistory(title)
	if err != nil {
		return nil, err
	}

	for _, r := range lst {
		crypto := util.NewHMacCrypto([]byte(r.Title), p.recordkey)
		decoded, err := decryptFromBase64(crypto, r.Password)
		if err != nil {
			return nil, err
		}
		r.Password = string(decoded)
	}

	return lst, nil
}

func (p *impl) SetStoreCrypto(crypto util.CryptoApi) {
	p.storeCrypto = crypto
}

func (p *impl) SetRecordKey(key []byte) {
	p.recordkey = key
}

func (p *impl) Load() error {
	empty, err := p.Empty()
	if err != nil {
		return err
	}

	content := make([]byte, 0)
	if !empty {
		content, err = read(p.storeCrypto, passfile())
		if err != nil {
			if err == errs.DecryptError {
				err = errs.InvalidKey
			} else if err == errs.NoSuchFile {
				err = errs.Uninited
			}
			return err
		}
	}

	store, err := data.New(string(content))
	if err != nil {
		if err == errs.InvalidCsvRecord {
			err = errs.InvalidKey
		}
		return err
	}
	p.store = store

	return nil
}

func (p *impl) LoadOld(pincode, token []byte) error {
	crypto := util.NewHMacCrypto(pincode, token)
	p.SetStoreCrypto(crypto)

	empty, err := p.Empty()
	if err != nil {
		return err
	}

	content := make([]byte, 0)
	if !empty {
		content, err = read(crypto, passfile())
		if err != nil {
			if err == errs.DecryptError {
				err = errs.InvalidKey
			} else if err == errs.NoSuchFile {
				err = errs.Uninited
			}
			return err
		}
	}

	store, err := data.New(string(content))
	if err != nil {
		if err == errs.InvalidCsvRecord {
			err = errs.InvalidKey
		}
		return err
	}
	p.store = store

	return nil
}

func (p *impl) Save() error {
	content, err := p.store.Save()
	if err != nil {
		return err
	}

	return write(p.storeCrypto, []byte(content), passfile())
}

func (p *impl) Empty() (bool, error) {
	content, err := util.ReadFromFile(passfile())
	if err != nil {
		if err == errs.NoSuchFile {
			err = errs.Uninited
		}
		return true, err
	}
	return strings.TrimSpace(string(content)) == "", nil
}

func (p *impl) Migrate() error {
	for _, r := range p.store.GetRecords() {
		crypto := util.NewHMacCrypto([]byte(r.Title), p.recordkey)
		encoded, err := crypt2base64(crypto, []byte(r.Password))
		if err != nil {
			return err
		}
		r.Password = encoded
	}
	return nil
}
