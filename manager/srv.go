package manager

import (
	"strings"

	"github.com/musenwill/mypass/data"
	"github.com/musenwill/mypass/util"
)

type SrvApi interface {
	Init(gitUrl string) error
	Groups() ([]string, error)
	Titles() ([]string, error)
	Filter(groupLike, titleLike string) ([][]string, error)
	Delete(group, title string) error
	Put(group, title, password, describe string) error
	Get(title string) ([]string, error)
	History(title string) ([][]string, error)
	Load(pincode, token []byte) error
	Save() error
	Empty() (bool, error)
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
	return p.store.ListGroups()
}

func (p *impl) Titles() ([]string, error) {
	return p.store.ListTitles()
}

func (p *impl) Filter(grouplike, titleLike string) ([][]string, error) {
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
		err := p.store.DeleteTitle(group)
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

func (p *impl) Get(title string) ([]string, error) {
	return p.store.Get(title)
}

func (p *impl) History(title string) ([][]string, error) {
	return p.store.GetHistory(title)
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
			return err
		}
	}

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

func (p *impl) Empty() (bool, error) {
	content, err := util.ReadFromFile(passfile())
	if err != nil {
		return true, err
	}
	return strings.TrimSpace(string(content)) == "", nil
}
