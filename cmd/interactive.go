package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var supportedTypes = []string{"str", "file", "url"}

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
	fmt.Printf("Enter %v(str,file,url): ", name)
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
		if !inStrings(slices, t) {
			return "", "", fmt.Errorf(`supported types are: %v`, supportedTypes)
		}
		return t, content, nil
	} else {
		return "", "", errors.New(`invalid input, expected "type content"`)
	}
}

func inStrings(slices []string, val string) bool {
	for _, s := range slices {
		if s == val {
			return true
		}
	}
	return false
}
