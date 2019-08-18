package main

import (
	"os"

	"github.com/musenwill/mypass/cmd"
)

func main() {
	cmd.New().Run(os.Args)
}
