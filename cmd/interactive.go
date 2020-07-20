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
	dict := make(map[string][]*data.Record)
	for _, r := range records {
		lst := dict[r.Group]
		if lst == nil {
			lst = make([]*data.Record, 0)
		}
		lst = append(lst, r)
		dict[r.Group] = lst
	}

	header := "%-16s%-32s%-32s %s\n"
	header = fmt.Sprintf(header, "group", "title", "create at", "describe")
	maxWidth := runewidth.StringWidth(header)

	var sb strings.Builder
	for _, lst := range dict {
		for _, r := range lst {
			line := strings.Join([]string{fixWidth(r.Group, -1, 16), fixWidth(r.Title, -1, 32),
				fixWidth(r.Ct.String(), -1, 32), fixWidth(r.Describe, -1, 32)}, "")
			if width := runewidth.StringWidth(line); width > maxWidth {
				maxWidth = width
			}
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	gridLine := strings.Repeat("-", maxWidth)
	fmt.Print(header)
	fmt.Println(gridLine)
	fmt.Print(sb.String())
}

func printRecordsV(records ...*data.Record) {
	dict := make(map[string][]*data.Record)
	for _, r := range records {
		lst := dict[r.Group]
		if lst == nil {
			lst = make([]*data.Record, 0)
		}
		lst = append(lst, r)
		dict[r.Group] = lst
	}

	header := "%-16s%-32s%-32s%-32s %s\n"
	header = fmt.Sprintf(header, "group", "title", "password", "create at", "describe")
	maxWidth := runewidth.StringWidth(header)

	var sb strings.Builder
	for _, lst := range dict {
		for _, r := range lst {
			line := strings.Join([]string{fixWidth(r.Group, -1, 16), fixWidth(r.Title, -1, 32),
				fixWidth(r.Password, -1, 32), fixWidth(r.Ct.String(), -1, 32), fixWidth(r.Describe, -1, 32)}, "")
			if width := runewidth.StringWidth(line); width > maxWidth {
				maxWidth = width
			}
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	gridLine := strings.Repeat("-", maxWidth)
	fmt.Print(header)
	fmt.Println(gridLine)
	fmt.Print(sb.String())
}

/* align: -1 left, 0 centre, 1 right */
func fixWidth(text string, align int, fixWidth int) string {
	textWidth := runewidth.StringWidth(text)
	if textWidth >= fixWidth {
		return text
	}

	var printText string
	if align < 0 {
		printText = text + strings.Repeat(" ", fixWidth-textWidth)
	} else if align == 0 {
		leftWidth := (fixWidth - textWidth) / 2
		rightWidth := fixWidth - leftWidth
		printText = strings.Repeat(" ", leftWidth) + text + strings.Repeat(" ", rightWidth)
	} else {
		printText = strings.Repeat(" ", fixWidth-textWidth) + text
	}

	return printText
}
