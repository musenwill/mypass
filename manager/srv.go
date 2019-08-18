package manager

import (
	"github.com/musenwill/mypass/data"
	"github.com/musenwill/mypass/util"
)

type SrvApi interface {
	Init(gitUrl string) error
	Groups() ([]string, error)
	Titles() ([]string, error)
	Filter(groupLike, titleLike string) ([][]string, error)
	Delete(group, title string) error
	Put(group, title, describe string) error
	Get(title string, print bool) ([]string, error)
	History(title string) ([][]string, error)
	Load(pincode, token []byte) error
	Save() error
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

func (p *impl) Groups() ([]string, error) {
	return nil, nil
}

func (p *impl) Titles() ([]string, error) {
	return nil, nil
}

func (p *impl) Filter(grouplike, titleLike string) ([][]string, error) {
	return nil, nil
}

func (p *impl) Delete(group, title string) error {
	return nil
}

func (p *impl) Put(group, title, describe string) error {
	return nil
}

func (p *impl) Get(title string, print bool) ([]string, error) {

	return nil, nil
}

func (p *impl) History(title string) ([][]string, error) {
	return nil, nil
}

func (p *impl) Load(pincode, token []byte) error {
	crypto := util.NewCrypto(pincode, token)
	content, err := read(crypto, passfile())
	if err != nil {
		return err
	}

	p.crypto = crypto
	store, err := data.New(string(content))
	if err != nil {
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
