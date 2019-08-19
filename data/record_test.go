package data

import (
	"testing"
	"time"
)

type csvCase struct {
	record    *Record
	csvRecord []string
}

func TestCsvRecord(t *testing.T) {
	ct, err := time.Parse(timeFormat, timeFormat)
	if err != nil {
		t.Fatal(err)
	}

	cases := []csvCase{
		{
			&Record{"tencent", "qq", "admin", "", ct},
			[]string{"tencent", "qq", "admin", "no describe", timeFormat},
		},
		{
			&Record{"tencent", "qq", "admin", "for qq", ct},
			[]string{"tencent", "qq", "admin", "for qq", timeFormat},
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
	ct, err := time.Parse(timeFormat, timeFormat)
	if err != nil {
		t.Fatal(err)
	}

	cases := []csvCase{
		{
			&Record{"tencent", "qq", "admin", "no describe", ct},
			[]string{"tencent", "qq", "admin", "no describe", timeFormat},
		},
		{
			&Record{"tencent", "qq", "admin", "for qq", ct},
			[]string{"tencent", "qq", "admin", "for qq", timeFormat},
		},
	}

	for _, c := range cases {
		act, err := FromCsvRecord(c.csvRecord)
		if err != nil || !c.record.Equal(act) {
			t.Errorf("got %v expected %v, %v", act, c.record, err)
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
