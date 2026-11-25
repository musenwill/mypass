package manager

import (
	"os"

	"gopkg.in/yaml.v2"
)

// config is config.yaml. It keeps the git address the store is synced by, the
// synchronization itself is done by the user.
type config struct {
	Git string `yaml:"git"`
}

func saveConf(conf *config, configFile string) error {
	content, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, content, 0600)
}
