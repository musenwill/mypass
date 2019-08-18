package manager

import (
	"log"
	"os/user"
)

var (
	userDir = ""
)

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	userDir = usr.HomeDir
}

func passdir() string {
	return userDir + "/.mypass"
}

func passfile() string {
	return passdir() + "/pass"
}

func configfile() string {
	return passdir() + "/config.yaml"
}
