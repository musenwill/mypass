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
	Load(pincode, token []byte) error
	Save() error
	Empty() (bool, error)
	SetCrypto(crypto util.CryptoApi)
}

type impl struct {
	store  data.Store
	crypto util.CryptoApi
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
	_, err := p.store.Put(group, title, password, describe)
	return err
}

func (p *impl) Get(title string) (*data.Record, error) {
	return p.store.Get(title)
}

func (p *impl) History(title string) ([]*data.Record, error) {
	return p.store.GetHistory(title)
}

func (p *impl) SetCrypto(crypto util.CryptoApi) {
	p.crypto = crypto
}

func (p *impl) Load(pincode, token []byte) error {
	crypto := util.NewCrypto(pincode, token)
	p.crypto = crypto

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

	return write(p.crypto, []byte(content), passfile())
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
