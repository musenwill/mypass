package data

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strings"
	"time"

	"github.com/musenwill/mypass/errs"
)

const fileName = ".pass"

type Store interface {
	Save() (string, error)
	All() ([][]string, error)
	ListGroups() ([]string, error)
	ListTitles() ([]string, error)
	ListAll() ([][]string, error)
	Filter(groupLike, titleLike string) ([][]string, error)
	Put(group, title, password, describe string) (*Record, error)
	Get(title string) ([]string, error)
	GetHistory(title string) ([][]string, error)
	DeleteTitle(title string) error
	DeleteGroup(group string) error
}

type storage struct {
	records *Records
}

func New(content string) (Store, error) {
	reader := csv.NewReader(strings.NewReader(content))
	if reader == nil {
		return nil, errors.New("failed create csv reader")
	}

	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var records []*Record
	for _, l := range lines {
		r, err := FromCsvRecord(l)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return &storage{&Records{records}}, nil
}

func (p *storage) Save() (string, error) {
	var csvRecords [][]string
	for _, r := range p.records.records {
		csvRecords = append(csvRecords, r.ToCsvRecord())
	}

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)
	if writer == nil {
		return "", errors.New("failed create csv writer")
	}

	err := writer.WriteAll(csvRecords)
	if err != nil {
		return "", err
	}
	writer.Flush()

	return buf.String(), nil
}

func (p *storage) All() ([][]string, error) {
	titles := p.records.Titles()

	var rs []*Record
	for _, title := range titles {
		latest := p.records.ByTitle(title).Latest()
		if latest != nil {
			rs = append(rs, latest)
		}
	}

	var results [][]string
	for _, record := range rs {
		results = append(results, record.ToCsvRecord())
	}

	return results, nil
}

func (p *storage) ListGroups() ([]string, error) {
	return p.records.Groups(), nil
}

func (p *storage) ListTitles() ([]string, error) {
	return p.records.Titles(), nil
}

func (p *storage) ListAll() ([][]string, error) {
	titles, err := p.ListTitles()
	if err != nil {
		return nil, err
	}

	var records []*Record
	for _, title := range titles {
		latest := p.records.ByTitle(title).Latest()
		if latest != nil {
			records = append(records, latest)
		}
	}

	var results [][]string
	for _, record := range records {
		results = append(results, record.ToCsvRecord())
	}

	return results, nil
}

func (p *storage) Filter(groupLike, titleLike string) ([][]string, error) {
	records := p.records
	if strings.TrimSpace(groupLike) != "" {
		records = records.GroupLike(groupLike)
	}
	if strings.TrimSpace(titleLike) != "" {
		records = records.TitleLike(titleLike)
	}

	titles := records.Titles()

	var rs []*Record
	for _, title := range titles {
		latest := records.ByTitle(title).Latest()
		if latest != nil {
			rs = append(rs, latest)
		}
	}

	var results [][]string
	for _, record := range rs {
		results = append(results, record.ToCsvRecord())
	}

	return results, nil
}

func (p *storage) Put(group, title, password, describe string) (*Record, error) {
	record := &Record{group, title, password, describe, time.Now()}
	p.records.Put(record)
	return record, nil
}

func (p *storage) Get(title string) ([]string, error) {
	latest := p.records.ByTitle(title).Latest()
	if latest == nil {
		return nil, errs.DataNotFound
	}

	return latest.ToCsvRecord(), nil
}

func (p *storage) GetHistory(title string) ([][]string, error) {
	records := p.records.ByTitle(title)
	records.Sort()

	var results [][]string
	for _, record := range records.records {
		results = append(results, record.ToCsvRecord())
	}

	return results, nil
}

func (p *storage) DeleteTitle(title string) error {
	p.records.DeleteTitle(title)
	return nil
}

func (p *storage) DeleteGroup(group string) error {
	p.records.DeleteGroup(group)
	return nil
}
