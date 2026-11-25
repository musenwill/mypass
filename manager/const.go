package manager

import (
	"log"
	"os"
	"os/user"
	"strings"
)

// homeEnv points mypass at another data directory than ~/.mypass. The home comes
// from the system user database, so $HOME doesn't apply. Tests and scripts need it
// to keep the real store out of the way.
const homeEnv = "MYPASS_HOME"

// legacyBackupSuffix marks where `mypass upgrade` keeps the v1 store.
const legacyBackupSuffix = ".v1.bak"

func passdir() string {
	if dir := strings.TrimSpace(os.Getenv(homeEnv)); dir != "" {
		return dir
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir + "/.mypass"
}

func passfile() string {
	return passdir() + "/pass"
}

func configfile() string {
	return passdir() + "/config.yaml"
}
