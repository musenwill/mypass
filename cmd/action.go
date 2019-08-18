package cmd

import (
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
	return nil
}

func filter(c *cli.Context) error {
	group := c.String("group")
	title := c.String("title")

	fmt.Println(group)
	fmt.Println(title)

	return nil
}

func delete(c *cli.Context) error {
	group := c.String("group")
	title := c.String("title")

	fmt.Println(group)
	fmt.Println(title)

	return nil
}

func put(c *cli.Context) error {
	group := c.String("group")
	title := c.String("title")
	describe := c.String("describe")

	fmt.Println(group)
	fmt.Println(title)
	fmt.Println(describe)

	return nil
}

func get(c *cli.Context) error {
	title := c.String("title")
	print := c.Bool("print")

	fmt.Println(title)
	fmt.Println(print)

	return nil
}

func history(c *cli.Context) error {
	title := c.String("title")

	fmt.Println(title)
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
