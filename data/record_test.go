package data

import "testing"

type csvCase struct {
	record    *Record
	csvRecord []string
}

func TestCsvRecord(t *testing.T) {
	cases := []csvCase{
		{
			&Record{"tencent", "qq", "admin", ""},
			[]string{"tencent", "qq", "admin", "no describe"},
		},
		{
			&Record{"tencent", "qq", "admin", "for qq"},
			[]string{"tencent", "qq", "admin", "for qq"},
		},
	}

	for _, c := range cases {
		act := c.record.CsvRecord()
		if !compareStrs(act, c.csvRecord) {
			t.Errorf("got %v expected %v", act, c.csvRecord)
		}
	}
}

func TestFromCsvRecord(t *testing.T) {
	cases := []csvCase{
		{
			&Record{"tencent", "qq", "admin", "no describe"},
			[]string{"tencent", "qq", "admin", "no describe"},
		},
		{
			&Record{"tencent", "qq", "admin", "for qq"},
			[]string{"tencent", "qq", "admin", "for qq"},
		},
	}

	for _, c := range cases {
		act, err := FromCsvRecord(c.csvRecord)
		if err != nil || !EqualRecord(act, c.record) {
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
