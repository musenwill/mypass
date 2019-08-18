package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type ft struct {
	str, file, url string
}

func (p *ft) list() []string {
	return []string{p.str, p.file, p.url}
}

func (p *ft) contains(t string) bool {
	for _, s := range p.list() {
		if s == t {
			return true
		}
	}
	return false
}

var factorType = &ft{
	"str", "file", "url",
}

func inputPassword() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter password: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return text, nil
}

func inputPincode() (string, string, error) {
	return inputFactor("pincode")
}

func inputToken() (string, string, error) {
	return inputFactor("token")
}

func inputFactor(name string) (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter %v%v: ", name, factorType.list())
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	slices := strings.Split(text, " ")
	slen := len(slices)
	if slen == 1 {
		return "str", slices[0], nil
	} else if slen == 2 {
		t, content := slices[0], slices[1]
		if !factorType.contains(t) {
			return "", "", fmt.Errorf(`supported types are: %v`, factorType.list())
		}
		return t, content, nil
	} else {
		return "", "", errors.New(`invalid input, expected "type content"`)
	}
}
