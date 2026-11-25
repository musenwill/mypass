package data

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/musenwill/mypass/errs"
)

const (
	// timeFormat is how Ct is stored, an instant in UTC. The nanoseconds keep
	// the order of versions written within one second.
	timeFormat = time.RFC3339Nano
	// DisplayTimeFormat is how Ct is shown, in local time.
	DisplayTimeFormat = "2006-01-02 15:04:05"
	// legacyTimeFormat is how v1 stored Ct. Its zone part is literal, so a v1
	// record keeps the local wall-clock time of the writing machine, labelled as UTC.
	legacyTimeFormat = "2006-01-02 15:04:05 +0000 UTC"
)

type Record struct {
	Sn       int
	Group    string
	Title    string
	Password string
	Describe string
	Ct       time.Time
}

func (p *Record) ToCsvRecord() []string {
	if strings.TrimSpace(p.Describe) == "" {
		p.Describe = "no describe"
	}
	return []string{strconv.Itoa(p.Sn), p.Group, p.Title, p.Password, p.Describe, p.Ct.UTC().Format(timeFormat)}
}

func FromCsvRecord(r []string) (*Record, error) {
	if len(r) != 6 {
		return nil, errs.InvalidCsvRecord
	}
	sn, err := strconv.Atoi(r[0])
	if err != nil {
		return nil, errs.InvalidCsvRecord
	}
	ct, err := time.Parse(timeFormat, r[5])
	if err != nil {
		return nil, err
	}
	return &Record{sn, r[1], r[2], r[3], r[4], ct}, nil
}

// FromLegacyCsvRecord reads a row of a v1 store. v1 has no sn, and its time is
// read as the local wall-clock time, see legacyTimeFormat.
func FromLegacyCsvRecord(r []string) (*Record, error) {
	if len(r) != 5 {
		return nil, errs.InvalidCsvRecord
	}
	ct, err := time.Parse(legacyTimeFormat, r[4])
	if err != nil {
		return nil, err
	}
	local := time.Date(ct.Year(), ct.Month(), ct.Day(), ct.Hour(), ct.Minute(), ct.Second(), 0, time.Local)
	return &Record{Group: r[0], Title: r[1], Password: r[2], Describe: r[3], Ct: local.UTC()}, nil
}

type Records struct {
	records []*Record
}

func (p *Records) GetRecords() []*Record {
	return p.records
}

func (p *Records) Len() int {
	return len(p.records)
}

// Less orders from the oldest to the newest.
func (p *Records) Less(i, j int) bool {
	return p.records[i].Ct.Before(p.records[j].Ct)
}

// Sort orders from the oldest to the newest. Versions sharing a time keep their
// stored order, so the one written last comes last.
func (p *Records) Sort() {
	sort.SliceStable(p.records, func(i, j int) bool {
		return p.Less(i, j)
	})
}

// Latest is the newest record. Versions sharing a time are told apart by their
// stored order, see Sort.
func (p *Records) Latest() *Record {
	if p.Len() <= 0 {
		return nil
	}

	p.Sort()
	return p.records[p.Len()-1]
}

// Accounts returns the latest record of every title, an account is a title.
func (p *Records) Accounts() *Records {
	var records []*Record
	for _, title := range p.Titles() {
		if latest := p.ByTitle(title).Latest(); latest != nil {
			records = append(records, latest)
		}
	}

	return &Records{records}
}

// LessByGroupTitle orders by group, then title, ignoring case.
func LessByGroupTitle(a, b *Record) bool {
	if ga, gb := strings.ToLower(a.Group), strings.ToLower(b.Group); ga != gb {
		return ga < gb
	}
	if ta, tb := strings.ToLower(a.Title), strings.ToLower(b.Title); ta != tb {
		return ta < tb
	}
	if a.Group != b.Group {
		return a.Group < b.Group
	}
	return a.Title < b.Title
}

func (p *Records) SortByGroupTitle() {
	sort.SliceStable(p.records, func(i, j int) bool {
		return LessByGroupTitle(p.records[i], p.records[j])
	})
}

// Renumber gives every account its position, starting from 1, in the group and
// title order. All versions of one account share the sn.
func (p *Records) Renumber() {
	accounts := p.Accounts()
	accounts.SortByGroupTitle()

	sns := make(map[string]int)
	for i, a := range accounts.records {
		sns[a.Title] = i + 1
	}
	for _, r := range p.records {
		r.Sn = sns[r.Title]
	}
}

func (p *Records) Groups() []string {
	return p.uniques(func(r *Record) string { return r.Group })
}

func (p *Records) Titles() []string {
	return p.uniques(func(r *Record) string { return r.Title })
}

func (p *Records) uniques(field func(*Record) string) []string {
	set := make(map[string]bool)
	for _, r := range p.records {
		set[field(r)] = true
	}

	var values []string
	for k := range set {
		values = append(values, k)
	}
	sort.Strings(values)

	return values
}

func (p *Records) ByTitle(title string) *Records {
	return p.filter(func(r *Record) bool { return r.Title == title })
}

func (p *Records) BySn(sn int) *Records {
	return p.filter(func(r *Record) bool { return sn > 0 && r.Sn == sn })
}

func (p *Records) GroupLike(groupLike string) *Records {
	return p.filter(func(r *Record) bool { return strings.Contains(r.Group, groupLike) })
}

func (p *Records) TitleLike(titleLike string) *Records {
	return p.filter(func(r *Record) bool { return strings.Contains(r.Title, titleLike) })
}

func (p *Records) filter(match func(*Record) bool) *Records {
	var records []*Record
	for _, r := range p.records {
		if match(r) {
			records = append(records, r)
		}
	}

	return &Records{records}
}

func (p *Records) Put(record *Record) {
	p.records = append(p.records, record)
}

// remove drops every record match is true for, and tells how many are gone.
func (p *Records) remove(match func(*Record) bool) int {
	kept := make([]*Record, 0, len(p.records))
	for _, r := range p.records {
		if !match(r) {
			kept = append(kept, r)
		}
	}

	removed := len(p.records) - len(kept)
	p.records = kept
	return removed
}

func (p *Records) DeleteTitle(title string) int {
	return p.remove(func(r *Record) bool { return r.Title == title })
}

func (p *Records) DeleteGroup(group string) int {
	return p.remove(func(r *Record) bool { return r.Group == group })
}
