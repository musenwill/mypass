package data

import (
	"errors"
	"strings"
)

type Record struct {
	Group, Title, Password, Describe string
}

func (p *Record) CsvRecord() []string {
	if strings.Trim(p.Describe, "\t\n\r0x20") == "" {
		p.Describe = "no describe"
	}
	return []string{p.Group, p.Title, p.Password, p.Describe}
}

func FromCsvRecord(r []string) (*Record, error) {
	if len(r) != 4 {
		return nil, errors.New("invalid record")
	}
	return &Record{r[0], r[1], r[2], r[3]}, nil
}

func EqualRecord(a, b *Record) bool {
	return a.Group == b.Group && a.Title == b.Title && a.Password == b.Password && a.Describe == b.Describe
}
