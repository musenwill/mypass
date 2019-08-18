package manager

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type config struct {
	Git string `yaml:"git"`
}

func loadConf(configFile string) (*config, error) {
	conf := new(config)
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func saveConf(conf *config, configFile string) error {
	content, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(configFile, content, os.ModePerm)
	return err
}
