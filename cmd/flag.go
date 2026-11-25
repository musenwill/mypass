package cmd

import (
	"strings"

	"github.com/urfave/cli"
)

const (
	flagGit      = "git"
	flagGroup    = "group, g"
	flagTitle    = "title, t"
	flagSn       = "sn, n"
	flagDescribe = "describe, d"
	flagPrint    = "print, p"
	flagLen      = "len, l"
)

// urfave/cli v1 registers every name of a flag, "title, t" here, as a flag of
// its own, and copies the value to the other names after parsing (v1.21.0). The
// actions don't depend on that: they read a flag through all of its names with
// these helpers, so they work on a bare FlagSet too, as in the tests.

func flagNames(names string) []string {
	var trimmed []string
	for _, name := range strings.Split(names, ",") {
		trimmed = append(trimmed, strings.TrimSpace(name))
	}
	return trimmed
}

// flagIsSet tells whether the flag was given, by any of its names.
func flagIsSet(c *cli.Context, names string) bool {
	for _, name := range flagNames(names) {
		if c.IsSet(name) {
			return true
		}
	}
	return false
}

// flagSetWith gives the name the flag was given with, or its first name.
func flagSetWith(c *cli.Context, names string) string {
	for _, name := range flagNames(names) {
		if c.IsSet(name) {
			return name
		}
	}
	return flagNames(names)[0]
}

func flagString(c *cli.Context, names string) string {
	return c.String(flagSetWith(c, names))
}

func flagInt(c *cli.Context, names string) int {
	return c.Int(flagSetWith(c, names))
}

func flagBool(c *cli.Context, names string) bool {
	return c.Bool(flagSetWith(c, names))
}
