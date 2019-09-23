package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/mattn/go-runewidth"
	"github.com/musenwill/mypass/data"
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
	fmt.Print("Enter password: ")
	input, err := gopass.GetPasswd()
	if err != nil {
		return "", err
	}
	password := string(input)
	password = strings.TrimSpace(password)

	return password, nil
}

func inputPincode() (string, string, error) {
	return inputFactor("pincode")
}

func inputToken() (string, string, error) {
	return inputFactor("token")
}

func inputFactor(name string) (string, string, error) {
	fmt.Printf("Enter %v%v: ", name, factorType.list())
	input, err := gopass.GetPasswd()
	if err != nil {
		return "", "", err
	}
	text := string(input)

	slices := strings.Split(text, " ")
	slen := len(slices)
	if slen == 1 {
		return "str", slices[0], nil
	} else if slen == 2 {
		t, content := slices[0], slices[1]
		t = strings.TrimSpace(t)
		content = strings.TrimSpace(content)
		if !factorType.contains(t) {
			return "", "", fmt.Errorf(`supported types are: %v`, factorType.list())
		}
		return t, content, nil
	} else {
		return "", "", errors.New(`invalid input, expected "type content"`)
	}
}

func printRecords(records ...*data.Record) {
	header := "%-16s%-32s%-32s %s\n"
	fmt.Printf(header, "group", "title", "create at", "describe")
	fmt.Println("--------------------------------------------------------------------------------")
	for _, r := range records {
		fmt.Print(fixLen(r.Group, -1, 16), fixLen(r.Title, -1, 32),
			fixLen(r.Ct.String(), -1, 32), fixLen(r.Describe, -1, 32))
		fmt.Println()
	}
}

func printRecordsV(records ...*data.Record) {
	header := "%-16s%-32s%-32s%-32s %s\n"
	fmt.Printf(header, "group", "title", "password", "create at", "describe")
	fmt.Println("--------------------------------------------------------------------------------")
	for _, r := range records {
		fmt.Print(fixLen(r.Group, -1, 16), fixLen(r.Title, -1, 32),
			fixLen(r.Password, -1, 32), fixLen(r.Ct.String(), -1, 32), fixLen(r.Describe, -1, 32))
		fmt.Println()
	}
}

/* align: -1 left, 0 centre, 1 right */
func fixLen(text string, align int, fixLen int) string {
	textWidth := runewidth.StringWidth(text)
	if textWidth >= fixLen {
		return text
	}

	var printText string
	if align < 0 {
		printText = text + strings.Repeat(" ", fixLen-textWidth)
	} else if align == 0 {
		leftWidth := (fixLen - textWidth) / 2
		rightWidth := fixLen - leftWidth
		printText = strings.Repeat(" ", leftWidth) + text + strings.Repeat(" ", rightWidth)
	} else {
		printText = strings.Repeat(" ", fixLen-textWidth) + text
	}

	return printText
}
