package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/musenwill/mypass/data"
	"github.com/urfave/cli"
)

func TestGetTarget(t *testing.T) {
	cases := []struct {
		args    []string
		target  *target
		invalid bool
	}{
		{args: []string{"-t", "qq"}, target: &target{title: "qq"}},
		{args: []string{"--title", "12306"}, target: &target{title: "12306"}},
		{args: []string{"-n", "5"}, target: &target{bySn: true, sn: 5}},
		{args: []string{"--sn", "12"}, target: &target{bySn: true, sn: 12}},
		{args: []string{}, invalid: true},
		{args: []string{"-t", "qq", "-n", "5"}, invalid: true},
	}

	for _, c := range cases {
		// the real flags, so that the short names are registered as they are in the app
		set := flag.NewFlagSet("get", flag.ContinueOnError)
		for _, f := range []cli.Flag{titleFlag, snFlag} {
			f.Apply(set)
		}
		if err := set.Parse(c.args); err != nil {
			t.Fatal(err)
		}

		got, err := getTarget(cli.NewContext(nil, set, nil))
		if c.invalid {
			if err == nil {
				t.Errorf("got no err expected an error on %v", c.args)
			}
			continue
		}
		if err != nil || *got != *c.target {
			t.Errorf("got %+v, %v expected %+v on %v", got, err, c.target, c.args)
		}
	}
}

// The list keeps the order given, which is group and title, and shows the sn of
// every account, with its time in local time.
func TestPrintRecords(t *testing.T) {
	now := time.Date(2019, 8, 18, 20, 31, 47, 0, time.Local)

	out := captureStdout(t, func() {
		printRecords(
			&data.Record{Sn: 1, Group: "ali", Title: "支付宝", Describe: "pay", Ct: now},
			&data.Record{Sn: 2, Group: "tencent", Title: "qq", Describe: "for qq", Ct: now},
		)
	})

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if act, exp := len(lines), 4; act != exp {
		t.Fatalf("got %v lines expected %v:\n%v", act, exp, out)
	}

	exp := []string{"  sn group ", "----", "   1 ali", "   2 tencent"}
	for i, prefix := range exp {
		if !strings.HasPrefix(lines[i], prefix) {
			t.Errorf("got %q expected it to start with %q", lines[i], prefix)
		}
	}
	if !strings.HasSuffix(lines[2], "2019-08-18 20:31:47 pay") {
		t.Errorf("got %q expected the local time and the describe", lines[2])
	}
	if !strings.HasSuffix(lines[3], "2019-08-18 20:31:47 for qq") {
		t.Errorf("got %q expected the local time and the describe", lines[3])
	}
}

func TestPrintRecordsV(t *testing.T) {
	now := time.Date(2019, 8, 18, 20, 31, 47, 0, time.Local)

	out := captureStdout(t, func() {
		printRecordsV(&data.Record{Sn: 4, Group: "tencent", Title: "qq", Password: "admin2", Describe: "for qq", Ct: now})
	})

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if act, exp := len(lines), 3; act != exp {
		t.Fatalf("got %v lines expected %v:\n%v", act, exp, out)
	}
	if !strings.HasPrefix(lines[0], "  sn group ") || !strings.Contains(lines[0], "password") {
		t.Errorf("got %q expected the sn and password columns", lines[0])
	}
	if !strings.HasPrefix(lines[2], "   4 tencent") || !strings.Contains(lines[2], "admin2") {
		t.Errorf("got %q expected the sn and the password", lines[2])
	}
}

func TestRandomPassword(t *testing.T) {
	for _, length := range []int{1, 12, maxKeyLen} {
		password, err := randomPassword(length)
		if err != nil {
			t.Fatal(err)
		}
		if act, exp := len(password), length; act != exp {
			t.Errorf("got %v chars expected %v", act, exp)
		}
		for _, c := range password {
			if !strings.ContainsRune(keyAlphabet, c) {
				t.Errorf("got %q expected characters of the alphabet only", string(c))
			}
		}
	}

	// a random source, and no character out of reach
	seen := make(map[byte]bool)
	for i := 0; i < len(keyAlphabet)*64; i++ {
		password, err := randomPassword(1)
		if err != nil {
			t.Fatal(err)
		}
		seen[password[0]] = true
	}
	if act, exp := len(seen), len(keyAlphabet); act != exp {
		t.Errorf("got %v different characters expected %v", act, exp)
	}
}

func TestInputFactor(t *testing.T) {
	content := "the content of a factor file"
	path := filepath.Join(t.TempDir(), "factor")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		input   string
		content string
		invalid bool
	}{
		"a string":     {input: "p@ss", content: "p@ss"},
		"a file":       {input: "file " + path, content: content},
		"unknown type": {input: "dir /tmp", invalid: true},
		"no content":   {input: "file", content: "file"},
		"too many":     {input: "file a b", invalid: true},
		"no such file": {input: "file /no/such/file", invalid: true},
	}

	for name, c := range cases {
		var got []byte
		var err error
		feedStdin(t, c.input+"\n", func() {
			got, err = inputFactor("pincode")
		})

		if c.invalid {
			if err == nil {
				t.Errorf("got no err expected an error on %v", name)
			}
			continue
		}
		if err != nil || string(got) != c.content {
			t.Errorf("got %q, %v expected %q on %v", string(got), err, c.content, name)
		}
	}
}

// A factor being set is asked for again on a mismatch.
func TestInputFactorTwice(t *testing.T) {
	var got []byte
	var err error
	out := feedStdin(t, "abc\ndef\nabc\nabc\n", func() {
		got, err = inputPincodeTwice()
	})

	if err != nil || string(got) != "abc" {
		t.Errorf("got %q, %v expected abc", string(got), err)
	}
	if !strings.Contains(out, "the two inputs differ, try again") {
		t.Errorf("got %q expected the mismatch to be reported", out)
	}
}

// Every flag reads the same by its short name as by its long one, see flag.go.
// A bare FlagSet is used here, so the copying of values between names, which
// the library does on parsing, doesn't help.
func TestShortNames(t *testing.T) {
	flags := []cli.Flag{gitFlag, groupFlag, titleFlag, snFlag, describeFlag, printFlag, lenFlagR}
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	for _, f := range flags {
		f.Apply(set)
	}
	if err := set.Parse([]string{"-g", "ali", "-t", "qq", "-n", "5", "-d", "desc", "-p", "-l", "16", "--git", "url"}); err != nil {
		t.Fatal(err)
	}
	c := cli.NewContext(nil, set, nil)

	cases := map[string]struct {
		got, exp string
	}{
		"group":    {flagString(c, flagGroup), "ali"},
		"title":    {flagString(c, flagTitle), "qq"},
		"sn":       {fmt.Sprint(flagInt(c, flagSn)), "5"},
		"describe": {flagString(c, flagDescribe), "desc"},
		"print":    {fmt.Sprint(flagBool(c, flagPrint)), "true"},
		"len":      {fmt.Sprint(flagInt(c, flagLen)), "16"},
		"git":      {flagString(c, flagGit), "url"},
	}
	for name, c := range cases {
		if c.got != c.exp {
			t.Errorf("got %v expected %v on %v", c.got, c.exp, name)
		}
	}
	for _, names := range []string{flagGit, flagGroup, flagTitle, flagSn, flagDescribe, flagPrint, flagLen} {
		if !flagIsSet(c, names) {
			t.Errorf("got not set expected %v to be set", names)
		}
	}
}

// A flag not given reads as not set, whatever its names.
func TestFlagNotSet(t *testing.T) {
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	for _, f := range []cli.Flag{titleFlag, snFlag} {
		f.Apply(set)
	}
	if err := set.Parse([]string{"-n", "5"}); err != nil {
		t.Fatal(err)
	}
	c := cli.NewContext(nil, set, nil)

	if flagIsSet(c, flagTitle) {
		t.Error("got title set expected it not set")
	}
	if act, exp := flagString(c, flagTitle), ""; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}
}

// feedStdin runs do with text on stdin, and gives what was printed on stdout.
func feedStdin(t *testing.T, text string, do func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString(text); err != nil {
		t.Fatal(err)
	}
	w.Close()

	original := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = original }()

	return captureStdout(t, do)
}

// captureStdout runs do, and gives what it printed on stdout.
func captureStdout(t *testing.T, do func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	do()
	w.Close()
	os.Stdout = original
	return <-done
}
