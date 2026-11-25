package data

import (
	"testing"
	"time"

	"github.com/musenwill/mypass/errs"
)

type csvCase struct {
	record    *Record
	csvRecord []string
}

var cst = time.FixedZone("CST", 8*3600)

func TestCsvRecord(t *testing.T) {
	ct, err := time.Parse(timeFormat, "2019-08-18T12:31:47Z")
	if err != nil {
		t.Fatal(err)
	}

	cases := []csvCase{
		{
			&Record{7, "tencent", "qq", "admin", "", ct},
			[]string{"7", "tencent", "qq", "admin", "no describe", "2019-08-18T12:31:47Z"},
		},
		{
			&Record{7, "tencent", "qq", "admin", "for qq", ct},
			[]string{"7", "tencent", "qq", "admin", "for qq", "2019-08-18T12:31:47Z"},
		},
	}

	for _, c := range cases {
		act := c.record.ToCsvRecord()
		if !compareStrs(act, c.csvRecord) {
			t.Errorf("got %v expected %v", act, c.csvRecord)
		}
	}
}

func TestFromCsvRecord(t *testing.T) {
	ct, err := time.Parse(timeFormat, "2019-08-18T12:31:47Z")
	if err != nil {
		t.Fatal(err)
	}

	cases := []csvCase{
		{
			&Record{7, "tencent", "qq", "admin", "no describe", ct},
			[]string{"7", "tencent", "qq", "admin", "no describe", "2019-08-18T12:31:47Z"},
		},
		{
			&Record{7, "tencent", "qq", "admin", "for qq", ct},
			[]string{"7", "tencent", "qq", "admin", "for qq", "2019-08-18T12:31:47Z"},
		},
	}

	for _, c := range cases {
		act, err := FromCsvRecord(c.csvRecord)
		if err != nil || !compareStrs(act.ToCsvRecord(), c.record.ToCsvRecord()) {
			t.Errorf("got %v, %v expected %v", act, err, c.record)
		}
	}
}

// Ct is stored as an instant, whatever zone it was created in.
func TestCsvRecordKeepsInstant(t *testing.T) {
	ct := time.Date(2019, 8, 18, 20, 31, 47, 0, cst)
	record := &Record{1, "tencent", "qq", "admin", "for qq", ct}

	csvRecord := record.ToCsvRecord()
	if act, exp := csvRecord[5], "2019-08-18T12:31:47Z"; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}

	parsed, err := FromCsvRecord(csvRecord)
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.Ct.Equal(ct) {
		t.Errorf("got %v expected %v", parsed.Ct, ct)
	}

	// nanoseconds are kept, two versions written within one second stay in order
	withNanos := time.Date(2019, 8, 18, 20, 31, 47, 123456789, cst)
	parsed, err = FromCsvRecord((&Record{1, "tencent", "qq", "admin", "for qq", withNanos}).ToCsvRecord())
	if err != nil {
		t.Fatal(err)
	}
	if act, exp := parsed.Ct.Format(time.RFC3339Nano), "2019-08-18T12:31:47.123456789Z"; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}
}

// Versions sharing a time are told apart by their stored order, the one written
// last is the latest.
func TestLatestOfSameTime(t *testing.T) {
	ct := time.Date(2019, 8, 18, 12, 31, 47, 0, time.UTC)
	records := &Records{[]*Record{
		{Title: "qq", Password: "v1", Ct: ct},
		{Title: "qq", Password: "v2", Ct: ct},
		{Title: "qq", Password: "v3", Ct: ct},
	}}

	if act, exp := records.Latest().Password, "v3"; act != exp {
		t.Errorf("got %v expected %v to be the latest", act, exp)
	}

	records.Sort()
	var got []string
	for _, r := range records.records {
		got = append(got, r.Password)
	}
	if exp := []string{"v1", "v2", "v3"}; !compareStrs(got, exp) {
		t.Errorf("got %v expected %v, oldest first", got, exp)
	}
}

func TestFromCsvRecordRejectsBadRows(t *testing.T) {
	cases := map[string][]string{
		"too few columns":  {"1", "tencent", "qq", "admin", "for qq"},
		"too many columns": {"1", "tencent", "qq", "admin", "for qq", "2019-08-18T12:31:47Z", "extra"},
		"bad sn":           {"x", "tencent", "qq", "admin", "for qq", "2019-08-18T12:31:47Z"},
		"bad time":         {"1", "tencent", "qq", "admin", "for qq", "2019-08-18 12:31:47"},
	}

	for name, row := range cases {
		_, err := FromCsvRecord(row)
		if err == nil {
			t.Errorf("got no err expected error on %v", name)
			continue
		}
		if name != "bad time" && err != errs.InvalidCsvRecord {
			t.Errorf("got %v expected invalid record on %v", err, name)
		}
	}
}

func TestFromLegacyCsvRecord(t *testing.T) {
	original := time.Local
	defer func() { time.Local = original }()

	row := []string{"tencent", "qq", "admin", "for qq", "2019-08-18 20:31:47 +0000 UTC"}

	// v1 kept the local wall-clock time of the writing machine, labelled as UTC
	time.Local = cst
	record, err := FromLegacyCsvRecord(row)
	if err != nil {
		t.Fatal(err)
	}

	exp := time.Date(2019, 8, 18, 20, 31, 47, 0, cst).UTC()
	if !record.Ct.Equal(exp) {
		t.Errorf("got %v expected %v", record.Ct, exp)
	}
	if act, exp := record.Sn, 0; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}
	if act, exp := record.ToCsvRecord()[5], "2019-08-18T12:31:47Z"; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}

	// written on a machine in utc, the same row means another instant
	time.Local = time.UTC
	record, err = FromLegacyCsvRecord(row)
	if err != nil {
		t.Fatal(err)
	}
	if act, exp := record.ToCsvRecord()[5], "2019-08-18T20:31:47Z"; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}
}

func TestFromLegacyCsvRecordRejectsBadRows(t *testing.T) {
	cases := map[string][]string{
		"too few columns":  {"tencent", "qq", "admin", "for qq"},
		"too many columns": {"tencent", "qq", "admin", "for qq", "2019-08-18 20:31:47 +0000 UTC", "extra"},
		"bad time":         {"tencent", "qq", "admin", "for qq", "2019-08-18 20:31:47"},
	}

	for name, row := range cases {
		if _, err := FromLegacyCsvRecord(row); err == nil {
			t.Errorf("got no err expected error on %v", name)
		}
	}
}

func TestAccounts(t *testing.T) {
	now := time.Now()
	records := &Records{[]*Record{
		{Group: "ali", Title: "alipay", Ct: now},
		{Group: "tencent", Title: "qq", Password: "v1", Ct: now.AddDate(0, 0, -2)},
		{Group: "ali", Title: "taobao", Ct: now},
		{Group: "tencent", Title: "qq", Password: "v2", Ct: now},
	}}

	// an account is a title, all its versions count once
	if act, exp := records.Accounts().Len(), 3; act != exp {
		t.Errorf("got %v accounts expected %v", act, exp)
	}

	latest := records.ByTitle("qq").Latest()
	if act, exp := latest.Password, "v2"; act != exp {
		t.Errorf("got %v expected the latest version %v", act, exp)
	}
}

func TestSortByGroupTitle(t *testing.T) {
	records := &Records{[]*Record{
		{Group: "tencent", Title: "wx"},
		{Group: "ali", Title: "淘宝"},
		{Group: "Tencent", Title: "QQ"},
		{Group: "ali", Title: "alipay"},
		{Group: "tencent", Title: "qq"},
	}}
	records.SortByGroupTitle()

	var got []string
	for _, r := range records.records {
		got = append(got, r.Group+"/"+r.Title)
	}
	exp := []string{"ali/alipay", "ali/淘宝", "Tencent/QQ", "tencent/qq", "tencent/wx"}
	if !compareStrs(got, exp) {
		t.Errorf("got %v expected %v", got, exp)
	}
}

func TestRenumber(t *testing.T) {
	now := time.Now()
	wx := &Record{Group: "tencent", Title: "wx", Ct: now}
	qqOld := &Record{Group: "tencent", Title: "qq", Ct: now.AddDate(0, 0, -1)}
	qq := &Record{Group: "tencent", Title: "qq", Ct: now}
	alipay := &Record{Group: "ali", Title: "alipay", Ct: now}
	records := &Records{[]*Record{wx, qqOld, alipay, qq}}

	records.Renumber()

	// numbers follow the group + title order, all versions of an account share one
	cases := map[*Record]int{alipay: 1, qq: 2, qqOld: 2, wx: 3}
	assertSns(t, cases)

	// so they move when accounts come and go
	records.DeleteTitle("alipay")
	records.Renumber()

	assertSns(t, map[*Record]int{qq: 1, qqOld: 1, wx: 2})
}

func TestGroupsAndTitlesSorted(t *testing.T) {
	records := &Records{[]*Record{
		{Group: "tencent", Title: "wx"},
		{Group: "ali", Title: "淘宝"},
		{Group: "tencent", Title: "qq"},
		{Group: "ali", Title: "支付宝"},
	}}

	if act, exp := records.Groups(), []string{"ali", "tencent"}; !compareStrs(act, exp) {
		t.Errorf("got %v expected %v", act, exp)
	}
	if act, exp := records.Titles(), []string{"qq", "wx", "支付宝", "淘宝"}; !compareStrs(act, exp) {
		t.Errorf("got %v expected %v", act, exp)
	}
}

func TestBySn(t *testing.T) {
	records := &Records{[]*Record{
		{Sn: 2, Title: "wx"},
		{Sn: 1, Title: "alipay"},
		{Sn: 2, Title: "qq"},
	}}

	if act, exp := records.BySn(2).Len(), 2; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}
	if act, exp := records.BySn(9).Len(), 0; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}
	// 0 is the sn of an account before its first save, never a valid lookup
	if act, exp := records.BySn(0).Len(), 0; act != exp {
		t.Errorf("got %v expected %v", act, exp)
	}
}

func assertSns(t *testing.T, cases map[*Record]int) {
	t.Helper()

	for r, exp := range cases {
		if act := r.Sn; act != exp {
			t.Errorf("got sn %v expected %v on %v", act, exp, r.Title)
		}
	}
}

func compareStrs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	length := len(a)
	for i := 0; i < length; i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
