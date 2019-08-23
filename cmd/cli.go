package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"
)

var (
	// Version should be a git tag
	Version = ""
	Name    = ""
)

func New() *cli.App {
	gitFlag := cli.StringFlag{
		Name:     "git",
		Usage:    "your git repository address `URL`",
		Required: true,
	}
	groupFlag := cli.StringFlag{
		Name:  "group, g",
		Usage: "group name",
	}
	groupFlagR := cli.StringFlag{
		Name:     "group, g",
		Usage:    "group name",
		Required: true,
	}
	titleGlag := cli.StringFlag{
		Name:  "title, t",
		Usage: "title name",
	}
	titleGlagR := cli.StringFlag{
		Name:     "title, t",
		Usage:    "title name",
		Required: true,
	}
	describeFlag := cli.StringFlag{
		Name:  "describe, d",
		Usage: "a simple description of account",
	}
	printFlag := cli.BoolFlag{
		Name:  "print, p",
		Usage: "whether to print password to console, default is copied to clipboard",
	}
	lenFlagR := cli.IntFlag{
		Name:     "len, l",
		Usage:    "length of key",
		Required: true,
	}

	app := cli.NewApp()
	app.ErrWriter = os.Stdout
	app.EnableBashCompletion = true
	app.Name = Name
	app.Usage = "A terminal password manager"
	app.Version = Version
	app.Author = "musenwill"
	app.Email = "musenwill@qq.com"
	app.Copyright = fmt.Sprintf("Copyright © 2019 - %v musenwill. All Rights Reserved.", time.Now().Year())
	app.Description = description

	app.Commands = []cli.Command{
		{
			Name:   "init",
			Usage:  "init you local password storeage, can be found in ~/.mypass directory",
			Flags:  []cli.Flag{gitFlag},
			Action: initStore,
		},
		{
			Name:   "old",
			Usage:  "list all entries which haven't been updated since 6 months ago",
			Action: oldPasswords,
		},
		{
			Name:   "all",
			Usage:  "list all entries",
			Action: all,
		},
		{
			Name:   "groups",
			Usage:  "list all existing groups",
			Action: groups,
		},
		{
			Name:   "titles",
			Usage:  "list all existing titles (or accounts)",
			Action: titles,
		},
		{
			Name:   "filter",
			Usage:  "fuzzy query your accounts by group name or title",
			Flags:  []cli.Flag{groupFlag, titleGlag},
			Action: filter,
		},
		{
			Name:   "delete",
			Usage:  "delete accounts, -t to delete by account name, -g to delete all accounts in the group",
			Flags:  []cli.Flag{groupFlag, titleGlag},
			Action: delete,
		},
		{
			Name:   "put",
			Usage:  "add new password",
			Flags:  []cli.Flag{groupFlagR, titleGlagR, describeFlag},
			Action: put,
		},
		{
			Name:   "get",
			Usage:  "get password by account name, password copied to clipboard on default, printed to console if -p specified",
			Flags:  []cli.Flag{titleGlagR, printFlag},
			Action: get,
		},
		{
			Name:   "history",
			Usage:  "print all versions passwords of the account",
			Flags:  []cli.Flag{titleGlagR},
			Action: history,
		},
		{
			Name:   "key",
			Usage:  "reset main key",
			Action: resetKey,
		},
		{
			Name:   "keygen",
			Usage:  "generate random key with given length",
			Flags:  []cli.Flag{lenFlagR, printFlag},
			Action: genkey,
		},
	}

	return app
}
