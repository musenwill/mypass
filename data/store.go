package data

import (
	"bytes"
	"encoding/csv"
	"strings"
	"time"

	"github.com/musenwill/mypass/errs"
)

type Store interface {
	GetRecords() []*Record
	// Save renumbers the accounts, then serializes them.
	Save() (string, error)
	// ListGroups lists the groups the accounts are in, a group every account
	// has left isn't one.
	ListGroups() ([]string, error)
	ListTitles() ([]string, error)
	// ListAll, Filter and Get return the latest version of an account.
	ListAll() ([]*Record, error)
	Filter(groupLike, titleLike string) ([]*Record, error)
	Put(group, title, password, describe string) (*Record, error)
	// Add puts a record as it is, with its own time.
	Add(record *Record)
	Get(title string) (*Record, error)
	GetBySn(sn int) (*Record, error)
	GetHistory(title string) ([]*Record, error)
	// DeleteTitle and DeleteGroup remove every version of an account, or of
	// the accounts of a group, and tell how many records are gone.
	DeleteTitle(title string) int
	DeleteGroup(group string) int
}

type storage struct {
	records *Records
}

func New(content string) (Store, error) {
	return parse(content, FromCsvRecord)
}

// NewLegacy reads the csv of a v1 store.
func NewLegacy(content string) (Store, error) {
	return parse(content, FromLegacyCsvRecord)
}

func parse(content string, from func([]string) (*Record, error)) (Store, error) {
	reader := csv.NewReader(strings.NewReader(content))

	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var records []*Record
	for _, l := range lines {
		r, err := from(l)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	all := &Records{records}
	// the file keeps sns of its last save, line them up again in case they are stale
	all.Renumber()

	return &storage{all}, nil
}

func (p *storage) GetRecords() []*Record {
	return p.records.GetRecords()
}

func (p *storage) Save() (string, error) {
	p.records.Renumber()

	var csvRecords [][]string
	for _, r := range p.records.records {
		csvRecords = append(csvRecords, r.ToCsvRecord())
	}

	buf := new(bytes.Buffer)
	// WriteAll flushes what it wrote, a failure shows in its error
	if err := csv.NewWriter(buf).WriteAll(csvRecords); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (p *storage) ListGroups() ([]string, error) {
	return p.records.Accounts().Groups(), nil
}

func (p *storage) ListTitles() ([]string, error) {
	return p.records.Titles(), nil
}

// ListAll lists every account, sorted by group and title.
func (p *storage) ListAll() ([]*Record, error) {
	return sortedAccounts(p.records.Accounts()), nil
}

// Filter lists the accounts whose group contains groupLike and whose title
// contains titleLike. An account is matched by its latest version, the one
// listed, not by a version it was moved out of.
func (p *storage) Filter(groupLike, titleLike string) ([]*Record, error) {
	accounts := p.records.Accounts()
	if strings.TrimSpace(groupLike) != "" {
		accounts = accounts.GroupLike(groupLike)
	}
	if strings.TrimSpace(titleLike) != "" {
		accounts = accounts.TitleLike(titleLike)
	}

	return sortedAccounts(accounts), nil
}

// sortedAccounts sorts accounts by group and title.
func sortedAccounts(accounts *Records) []*Record {
	accounts.SortByGroupTitle()
	return accounts.records
}

func (p *storage) Put(group, title, password, describe string) (*Record, error) {
	record := &Record{Group: group, Title: title, Password: password, Describe: describe, Ct: time.Now()}
	p.records.Put(record)
	return record, nil
}

func (p *storage) Add(record *Record) {
	p.records.Put(record)
}

func (p *storage) Get(title string) (*Record, error) {
	latest := p.records.ByTitle(title).Latest()
	if latest == nil {
		return nil, errs.DataNotFound
	}

	return latest, nil
}

func (p *storage) GetBySn(sn int) (*Record, error) {
	latest := p.records.BySn(sn).Latest()
	if latest == nil {
		return nil, errs.SnNotFound
	}

	return latest, nil
}

// GetHistory gives all versions of an account, the newest first.
func (p *storage) GetHistory(title string) ([]*Record, error) {
	records := p.records.ByTitle(title)
	records.Sort()

	history := make([]*Record, records.Len())
	for i, r := range records.records {
		history[records.Len()-1-i] = r
	}
	return history, nil
}

func (p *storage) DeleteTitle(title string) int {
	return p.records.DeleteTitle(title)
}

func (p *storage) DeleteGroup(group string) int {
	return p.records.DeleteGroup(group)
}
