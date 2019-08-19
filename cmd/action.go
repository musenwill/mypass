package cmd

import (
	"errors"
	"fmt"

	"github.com/musenwill/mypass/util"

	"github.com/musenwill/mypass/manager"
	"github.com/urfave/cli"
)

func initStore(c *cli.Context) error {
	gitUrl := c.String("git")
	srv := manager.New()
	return srv.Init(gitUrl)
}

func groups(c *cli.Context) error {
	if err := empty(); err != nil {
		return err
	}

	srv, err := load()
	if err != nil {
		return err
	}

	results, err := srv.Groups()
	if err != nil {
		return err
	}

	fmt.Println(results)

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

	results, err := srv.Titles()
	if err != nil {
		return err
	}

	fmt.Println(results)

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

	fmt.Println(result)

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
	print := c.Bool("print")

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

	fmt.Println(result, print)

	return nil
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

	fmt.Println(result)

	return nil
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
