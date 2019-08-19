package cmd

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/atotto/clipboard"
	"github.com/musenwill/mypass/util"

	"github.com/musenwill/mypass/manager"
	"github.com/urfave/cli"
)

func after(c *cli.Context) error {
	rand.Seed(time.Now().UnixNano())
	if rand.Float64() > 0.2 {
		return nil
	}

	srv, err := load()
	if err != nil {
		return err
	}

	now := time.Now()
	halfYearAgo := now
	// halfYearAgo := now.AddDate(0, -6, 0)

	result, err := srv.Olds(halfYearAgo)
	if err != nil {
		return err
	}

	if len(result) > 0 {
		fmt.Println("password of these accounts were updated 6 months ago, they may be in risk, suggest update them now")
		printRecords(result...)
	}

	return nil
}

func initStore(c *cli.Context) error {
	gitUrl := c.String("git")
	srv := manager.New()
	return srv.Init(gitUrl)
}

func all(c *cli.Context) error {
	if err := empty(); err != nil {
		return err
	}

	srv, err := load()
	if err != nil {
		return err
	}

	result, err := srv.All()
	if err != nil {
		return err
	}

	printRecords(result...)

	return nil
}

func groups(c *cli.Context) error {
	if err := empty(); err != nil {
		return err
	}

	srv, err := load()
	if err != nil {
		return err
	}

	result, err := srv.Groups()
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", result)

	return nil
}

func titles(c *cli.Context) error {
	if err := empty(); err != nil {
		return err
	}

	srv, err := load()
	if err != nil {
		return err
	}

	result, err := srv.Titles()
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", result)

	return nil
}

func filter(c *cli.Context) error {
	group := c.String("group")
	title := c.String("title")

	if err := empty(); err != nil {
		return err
	}

	srv, err := load()
	if err != nil {
		return err
	}

	result, err := srv.Filter(group, title)
	if err != nil {
		return err
	}

	printRecords(result...)

	return nil
}

func delete(c *cli.Context) error {
	group := c.String("group")
	title := c.String("title")

	if err := empty(); err != nil {
		return err
	}

	srv, err := load()
	if err != nil {
		return err
	}

	err = srv.Delete(group, title)
	if err != nil {
		return err
	}

	return srv.Save()
}

func put(c *cli.Context) error {
	group := c.String("group")
	title := c.String("title")
	describe := c.String("describe")

	srv, err := load()
	if err != nil {
		return err
	}

	password, err := inputPassword()
	if err != nil {
		return err
	}

	err = srv.Put(group, title, password, describe)
	if err != nil {
		return err
	}

	srv.Save()

	return nil
}

func get(c *cli.Context) error {
	title := c.String("title")
	var _ = c.Bool("print")

	if err := empty(); err != nil {
		return err
	}

	srv, err := load()
	if err != nil {
		return err
	}

	result, err := srv.Get(title)
	if err != nil {
		return err
	}

	fmt.Println("your password has copied to clipboard")
	return clipboard.WriteAll(result.Password)
}

func history(c *cli.Context) error {
	title := c.String("title")

	if err := empty(); err != nil {
		return err
	}

	srv, err := load()
	if err != nil {
		return err
	}

	result, err := srv.History(title)
	if err != nil {
		return err
	}

	printRecords(result...)

	return nil
}

func resetKey(c *cli.Context) error {
	srv, err := load()
	if err != nil {
		return err
	}

	fmt.Println("please set your new key below")

	t, pincodeSource, err := inputPincode()
	if err != nil {
		return err
	}
	pincode, err := factor(t, pincodeSource)
	if err != nil {
		return err
	}

	t, tokenSource, err := inputToken()
	if err != nil {
		return err
	}
	token, err := factor(t, tokenSource)
	if err != nil {
		return err
	}

	crypto := util.NewCrypto(pincode, token)
	srv.SetCrypto(crypto)

	return srv.Save()
}

func empty() error {
	srv := manager.New()
	empty, err := srv.Empty()
	if err != nil {
		return err
	}
	if empty {
		return errors.New("empty store")
	}
	return nil
}

func load() (manager.SrvApi, error) {
	t, pincodeSource, err := inputPincode()
	if err != nil {
		return nil, err
	}
	pincode, err := factor(t, pincodeSource)
	if err != nil {
		return nil, err
	}

	t, tokenSource, err := inputToken()
	if err != nil {
		return nil, err
	}
	token, err := factor(t, tokenSource)
	if err != nil {
		return nil, err
	}

	srv := manager.New()
	err = srv.Load(pincode, token)
	return srv, err
}

func factor(t, source string) ([]byte, error) {
	if t == factorType.str {
		return []byte(source), nil
	} else if t == factorType.file {
		return util.ReadFromFile(source)
	} else if t == factorType.url {
		return util.ReadFromUrl(source)
	} else {
		return nil, fmt.Errorf("unsupported factor type %v, expected are %v", t, factorType.list())
	}
}
