package cmd

import (
	"fmt"

	"github.com/musenwill/mypass/data"
	"github.com/musenwill/mypass/manager"
	"github.com/urfave/cli"
)

func initStore(c *cli.Context) error {
	gitUrl := c.String("git")
	return manager.InitStore(gitUrl)
}

func groups(c *cli.Context) error {

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

func load() (data.Store, error) {
	t, pincode, err := inputPincode()
	if err != nil {
		return nil, err
	}
	fmt.Println(t, pincode)

	t, token, err := inputToken()
	if err != nil {
		return nil, err
	}
	fmt.Println(t, token)

	return nil, nil
}
