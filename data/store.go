package data

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strings"
)

const fileName = ".pass"

type Api interface {
	Save() (string, error)
	ListGroups() ([]string, error)
	ListAll() ([][]string, error)
	Filter(groupLike, titleLike string) ([][]string, error)
	Put(group, title, password, describe string) error
	Get(title string) ([]string, error)
	Delete(title string) error
}

type storage struct {
	records []*Record
}

func NewStoreage(content string) (Api, error) {
	s := &storage{}

	reader := csv.NewReader(strings.NewReader(content))
	if reader == nil {
		return nil, errors.New("failed create csv reader")
	}

	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, l := range lines {
		r, err := FromCsvRecord(l)
		if err != nil {
			return nil, err
		}
		s.records = append(s.records, r)
	}

	return s, nil

	// dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	// if err != nil {
	// 	return nil, err
	// }
	// storePath := dir + "/" + fileName
}

func (p *storage) Save() (string, error) {
	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)
	if writer == nil {
		return "", errors.New("failed create csv writer")
	}

	var csvRecords [][]string
	for _, r := range p.records {
		csvRecords = append(csvRecords, r.CsvRecord())
	}

	err := writer.WriteAll(csvRecords)
	if err != nil {
		return "", err
	}
	writer.Flush()

	return buf.String(), nil
}

func (p *storage) ListGroups() ([]string, error) {
	set := make(map[string]bool)

	for _, r := range p.records {
		set[r.Group] = true
	}

	var groups []string
	for k := range set {
		groups = append(groups, k)
	}

	return groups, nil
}

func (p *storage) ListAll() ([][]string, error) {
	var records [][]string

	for _, r := range p.records {
		records = append(records, r.CsvRecord())
	}

	return records, nil
}

func (p *storage) Filter(groupLike, titleLike string) ([][]string, error) {
	return nil, nil
}

func (p *storage) Put(group, title, password, describe string) error {
	p.records = append(p.records, &Record{group, title, password, describe})
	return nil
}

func (p *storage) Get(title string) ([]string, error) {
	for _, r := range p.records {
		if r.Title == title {
			return r.CsvRecord(), nil
		}
	}

	return nil, errors.New("not found")
}

func (p *storage) Delete(title string) error {
	var newRecords []*Record

	for _, r := range p.records {
		if r.Title != title {
			newRecords = append(newRecords, r)
		}
	}

	p.records = newRecords

	return nil
}
