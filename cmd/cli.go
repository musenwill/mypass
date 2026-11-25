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

var (
	gitFlag = cli.StringFlag{
		Name:     flagGit,
		Usage:    "your git repository address `URL`",
		Required: true,
	}
	groupFlag = cli.StringFlag{
		Name:  flagGroup,
		Usage: "group name",
	}
	groupFlagR = cli.StringFlag{
		Name:     flagGroup,
		Usage:    "group name",
		Required: true,
	}
	titleFlag = cli.StringFlag{
		Name:  flagTitle,
		Usage: "title name",
	}
	titleFlagR = cli.StringFlag{
		Name:     flagTitle,
		Usage:    "title name",
		Required: true,
	}
	snFlag = cli.IntFlag{
		Name:  flagSn,
		Usage: "sn of the account, listed by the list commands",
	}
	describeFlag = cli.StringFlag{
		Name:  flagDescribe,
		Usage: "a simple description of account",
	}
	printFlag = cli.BoolFlag{
		Name:  flagPrint,
		Usage: "whether to print password to console, default is copied to clipboard",
	}
	lenFlagR = cli.IntFlag{
		Name:     flagLen,
		Usage:    "length of key",
		Required: true,
	}
)

func New() *cli.App {
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
			Usage:  "init your local password storage, can be found in ~/.mypass directory",
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
			Usage:  "list all entries, sorted by group and title",
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
			Flags:  []cli.Flag{groupFlag, titleFlag},
			Action: filter,
		},
		{
			Name:   "delete",
			Usage:  "delete accounts, -t to delete by account name, -g to delete all accounts in the group",
			Flags:  []cli.Flag{groupFlag, titleFlag},
			Action: delete,
		},
		{
			Name:   "put",
			Usage:  "add new password, it becomes a new version of the account when the title exists",
			Flags:  []cli.Flag{groupFlagR, titleFlagR, describeFlag},
			Action: put,
		},
		{
			Name:   "get",
			Usage:  "get password by account name or by sn, password copied to clipboard on default, printed to console if -p specified",
			Flags:  []cli.Flag{titleFlag, snFlag, printFlag},
			Action: get,
		},
		{
			Name:   "history",
			Usage:  "print all versions passwords of the account",
			Flags:  []cli.Flag{titleFlagR},
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
		{
			Name:   "upgrade",
			Usage:  "upgrade the store from v1 to v2, re-encrypted with the new scheme, the v1 store is kept as a backup",
			Action: upgrade,
		},
	}

	return app
}
