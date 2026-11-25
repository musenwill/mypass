package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/musenwill/mypass/data"
)

// column is a column of a record table: its header, the cell it shows for a
// record, and the width of the cell in display cells.
type column struct {
	header string
	cell   func(*data.Record) string
	width  int
}

var (
	snColumn       = column{"sn", func(r *data.Record) string { return strconv.Itoa(r.Sn) }, 4}
	groupColumn    = column{"group", func(r *data.Record) string { return r.Group }, 16}
	titleColumn    = column{"title", func(r *data.Record) string { return r.Title }, 32}
	passwordColumn = column{"password", func(r *data.Record) string { return r.Password }, 32}
	createAtColumn = column{"create at", func(r *data.Record) string { return createAt(r) }, 20}
	// describe ends the line, its width isn't padded to
	describeColumn = column{"describe", func(r *data.Record) string { return r.Describe }, 0}
)

// printRecords lists accounts in the order given. It is the group and title
// order, so the accounts of one group stay together.
func printRecords(records ...*data.Record) {
	printTable([]column{snColumn, groupColumn, titleColumn, createAtColumn, describeColumn}, records)
}

// printRecordsV lists versions of accounts, with their passwords.
func printRecordsV(records ...*data.Record) {
	printTable([]column{snColumn, groupColumn, titleColumn, passwordColumn, createAtColumn, describeColumn}, records)
}

// createAt is shown in local time.
func createAt(r *data.Record) string {
	return r.Ct.Local().Format(data.DisplayTimeFormat)
}

// printTable prints the header, a line as wide as the table, then a line per record.
func printTable(columns []column, records []*data.Record) {
	header := tableRow(columns, func(c column) string { return c.header })

	lines := make([]string, 0, len(records))
	for _, r := range records {
		lines = append(lines, tableRow(columns, func(c column) string { return c.cell(r) }))
	}

	maxWidth := runewidth.StringWidth(header)
	for _, line := range lines {
		if width := runewidth.StringWidth(line); width > maxWidth {
			maxWidth = width
		}
	}

	fmt.Println(header)
	fmt.Println(strings.Repeat("-", maxWidth))
	for _, line := range lines {
		fmt.Println(line)
	}
}

// tableRow is a line of a record table, a cell for every column. The first
// column, the sn, is right aligned, the others are left aligned. A cell keeps
// a space after it however wide it is, so two cells never run together. The
// last cell ends the line.
func tableRow(columns []column, cell func(column) string) string {
	var line strings.Builder
	for i, c := range columns {
		text := cell(c)
		if i == len(columns)-1 {
			line.WriteString(text)
			break
		}

		pad := c.width - runewidth.StringWidth(text)
		if i == 0 {
			if pad < 0 {
				pad = 0
			}
			line.WriteString(strings.Repeat(" ", pad))
			line.WriteString(text)
			pad = 1
		} else {
			line.WriteString(text)
			if pad < 1 {
				pad = 1
			}
		}
		line.WriteString(strings.Repeat(" ", pad))
	}
	return line.String()
}
