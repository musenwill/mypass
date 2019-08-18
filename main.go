package main

import (
	"os"

	"github.com/musenwill/mypass/cmd"
)

func main() {
	cmd.New().Run(os.Args)
}

// dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
// if err != nil {
// 	return nil, err
// }
// storePath := dir + "/" + fileName
