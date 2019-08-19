package data

import (
	"errors"
	"sort"
	"strings"
	"time"
)

const timeFormat = "2006-01-02 15:04:05 +0000 UTC"

type Record struct {
	Group, Title, Password, Describe string
	Ct                               time.Time
}

func (p *Record) ToCsvRecord() []string {
	if strings.TrimSpace(p.Describe) == "" {
		p.Describe = "no describe"
	}
	return []string{p.Group, p.Title, p.Password, p.Describe, p.Ct.Format(timeFormat)}
}

func (p *Record) Equal(r *Record) bool {
	return p.Group == r.Group && p.Title == r.Title && p.Password == r.Password && p.Describe == r.Describe && p.Ct.Equal(r.Ct)
}

func FromCsvRecord(r []string) (*Record, error) {
	if len(r) != 5 {
		return nil, errors.New("invalid record")
	}
	ct, err := time.Parse(timeFormat, r[4])
	if err != nil {
		return nil, err
	}
	return &Record{r[0], r[1], r[2], r[3], ct}, nil
}

type Records struct {
	records []*Record
}

func ensureRecordsImplSort() {
	var _ sort.Interface = &Records{}
}

func (p *Records) Len() int {
	return len(p.records)
}

func (p *Records) Less(i, j int) bool {
	ri := p.records[i]
	rj := p.records[j]
	return rj.Ct.Before(ri.Ct)
}

func (p *Records) Swap(i, j int) {
	p.records[i], p.records[j] = p.records[j], p.records[i]
}

func (p *Records) Sort() {
	sort.Sort(p)
}

func (p *Records) Latest() *Record {
	if p.Len() <= 0 {
		return nil
	}

	p.Sort()
	return p.records[0]
}

func (p *Records) Groups() []string {
	set := make(map[string]bool)

	for _, r := range p.records {
		set[r.Group] = true
	}

	var groups []string
	for k := range set {
		groups = append(groups, k)
	}

	return groups
}

func (p *Records) Titles() []string {
	set := make(map[string]bool)

	for _, r := range p.records {
		set[r.Title] = true
	}

	var titles []string
	for k := range set {
		titles = append(titles, k)
	}

	return titles
}

func (p *Records) ByGroup(group string) *Records {
	var records []*Record

	for _, r := range p.records {
		if r.Group == group {
			records = append(records, r)
		}
	}

	return &Records{records}
}

func (p *Records) ByTitle(title string) *Records {
	var records []*Record

	for _, r := range p.records {
		if r.Title == title {
			records = append(records, r)
		}
	}

	return &Records{records}
}

func (p *Records) GroupLike(groupLike string) *Records {
	var records []*Record

	for _, r := range p.records {
		if strings.Contains(r.Group, groupLike) {
			records = append(records, r)
		}
	}

	return &Records{records}
}

func (p *Records) TitleLike(titleLike string) *Records {
	var records []*Record

	for _, r := range p.records {
		if strings.Contains(r.Title, titleLike) {
			records = append(records, r)
		}
	}

	return &Records{records}
}

func (p *Records) Put(record *Record) {
	p.records = append(p.records, record)
}

func (p *Records) DeleteTitle(title string) {
	var newRecords []*Record

	for _, r := range p.records {
		if r.Title != title {
			newRecords = append(newRecords, r)
		}
	}

	p.records = newRecords
}

func (p *Records) DeleteGroup(group string) {
	var newRecords []*Record

	for _, r := range p.records {
		if r.Group != group {
			newRecords = append(newRecords, r)
		}
	}

	p.records = newRecords
}
