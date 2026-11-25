package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/musenwill/mypass/util"
)

// A factor is typed as a string, or as "file <path>" or "url <url>" to take a
// file content as the factor.
const (
	factorStr  = "str"
	factorFile = "file"
	factorUrl  = "url"

	// factorTypes are the types, as they are shown in a prompt and in an error.
	factorTypes = "[" + factorStr + " " + factorFile + " " + factorUrl + "]"
)

func inputPassword() (string, error) {
	for {
		fmt.Print("Enter password: ")
		input, err := gopass.GetPasswd()
		if err != nil {
			return "", err
		}

		password := strings.TrimSpace(string(input))
		if password != "" {
			return password, nil
		}
		// an empty password opens nothing
		fmt.Println("it can't be empty, try again")
	}
}

func inputPincode() ([]byte, error) {
	return inputFactor("pincode")
}

func inputToken() ([]byte, error) {
	return inputFactor("token")
}

// A factor being set is asked for twice, a mistyped one would lock the store
// for good. An empty one protects nothing, it is asked for again.
func inputPincodeTwice() ([]byte, error) {
	return inputFactorTwice("pincode")
}

func inputTokenTwice() ([]byte, error) {
	return inputFactorTwice("token")
}

func inputFactorTwice(name string) ([]byte, error) {
	for {
		first, err := inputFactor(name)
		if err != nil {
			return nil, err
		}
		second, err := inputFactor(name + " again")
		if err != nil {
			return nil, err
		}
		switch {
		case len(first) == 0 || len(second) == 0:
			fmt.Println("it can't be empty, try again")
		case !bytes.Equal(first, second):
			fmt.Println("the two inputs differ, try again")
		default:
			return first, nil
		}
	}
}

// inputFactor asks for a factor, a pincode or a token, without echo. It returns
// the content the factor stands for.
func inputFactor(name string) ([]byte, error) {
	fmt.Printf("Enter %v%v: ", name, factorTypes)
	input, err := gopass.GetPasswd()
	if err != nil {
		return nil, err
	}

	return parseFactor(string(input))
}

// parseFactor reads a line typed for a factor. A line without a type is a
// string, "file <path>" and "url <url>" take the content they point at.
func parseFactor(line string) ([]byte, error) {
	slices := strings.Split(line, " ")
	if len(slices) == 1 {
		return []byte(slices[0]), nil
	}
	if len(slices) != 2 {
		return nil, errors.New(`invalid input, expected "type content"`)
	}

	t, content := strings.TrimSpace(slices[0]), strings.TrimSpace(slices[1])
	switch t {
	case factorStr:
		return []byte(content), nil
	case factorFile:
		return util.ReadFromFile(content)
	case factorUrl:
		return util.ReadFromURL(content)
	}
	return nil, fmt.Errorf(`supported types are: %v`, factorTypes)
}
