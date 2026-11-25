package data

import (
	"fmt"
	"testing"

	"github.com/musenwill/mypass/errs"
)

func TestCRUD(t *testing.T) {
	storage, err := New("")
	if err != nil {
		t.Error(err)
	}

	// empty store
	groups, err := storage.ListGroups()
	if err != nil {
		t.Error(err)
	} else {
		if act, exp := len(groups), 0; act != exp {
			t.Errorf("got %v expected %v", act, exp)
		}
	}

	all, err := storage.ListAll()
	if err != nil {
		t.Error(err)
	} else {
		if act, exp := len(all), 0; act != exp {
			t.Errorf("got %v expected %v", act, exp)
		}
	}

	_, err = storage.Get("qq")
	if err == nil {
		t.Error("got no err expected no data error")
	}

	// put records into store
	qqRecord := []string{"0", "tencent", "qq", "admin", "for qq"}
	record, err := storage.Put(qqRecord[1], qqRecord[2], qqRecord[3], qqRecord[4])
	if err != nil {
		t.Error(err)
		return
	}
	qqRecord = append(qqRecord, record.Ct.UTC().Format(timeFormat))

	wxRecord := []string{"0", "tencent", "wx", "admin", "for wx"}
	record, err = storage.Put(wxRecord[1], wxRecord[2], wxRecord[3], wxRecord[4])
	if err != nil {
		t.Error(err)
		return
	}
	wxRecord = append(wxRecord, record.Ct.UTC().Format(timeFormat))

	// get from no empty store
	groups, err = storage.ListGroups()
	if err != nil {
		t.Error(err)
	} else {
		if act, exp := len(groups), 1; act != exp {
			t.Errorf("got %v expected %v", act, exp)
		}
	}

	all, err = storage.ListAll()
	if err != nil {
		t.Error(err)
	} else {
		if act, exp := len(all), 2; act != exp {
			t.Errorf("got %v expected %v", act, exp)
		}
	}

	r, err := storage.Get("qq")
	if err != nil {
		t.Error(err)
	}
	if !compareStrs(r.ToCsvRecord(), qqRecord) {
		t.Errorf("got %v expected %v", r, qqRecord)
	}

	r, err = storage.Get("wx")
	if err != nil {
		t.Error(err)
	}
	if !compareStrs(r.ToCsvRecord(), wxRecord) {
		t.Errorf("got %v expected %v", r, wxRecord)
	}

	// sns are given on save, in the group + title order
	if _, err := storage.Save(); err != nil {
		t.Error(err)
	}
	if r, err := storage.GetBySn(1); err != nil || r.Title != "qq" {
		t.Errorf("got %v, %v expected qq", r, err)
	}

	// delete from store
	storage.DeleteTitle("qq")

	_, err = storage.Get("qq")
	if err == nil {
		t.Error("got no err expected no data error")
	}
	if _, err := storage.Save(); err != nil {
		t.Error(err)
	}
	if r, err := storage.GetBySn(1); err != nil || r.Title != "wx" {
		t.Errorf("got %v, %v expected wx to take sn 1 after qq is deleted", r, err)
	}
}

// A store as it was written by v1, 5 columns and the v1 time.
const legacyContent = `tencent,qq,admin2,for qq,2019-08-18 12:31:47 +0000 UTC
tencent,qq,admin1,for qq,2019-07-18 12:31:47 +0000 UTC
tencent,qq,admin0,for qq,2019-06-18 12:31:47 +0000 UTC
tencent,wx,admin,for wx,2019-08-18 12:31:47 +0000 UTC
tencent,mail.qq.com,admin,for qq mail,2019-08-18 12:31:47 +0000 UTC
ali,支付宝,password,no describe,2019-08-18 12:31:47 +0000 UTC
ali,淘宝,password,no describe,2019-08-18 12:31:47 +0000 UTC
`

// The same store in the v2 layout, sns are stale on purpose.
const content = `0,tencent,qq,admin2,for qq,2019-08-18T12:31:47Z
0,tencent,qq,admin1,for qq,2019-07-18T12:31:47Z
0,tencent,qq,admin0,for qq,2019-06-18T12:31:47Z
0,tencent,wx,admin,for wx,2019-08-18T12:31:47Z
0,tencent,mail.qq.com,admin,for qq mail,2019-08-18T12:31:47Z
0,ali,支付宝,password,no describe,2019-08-18T12:31:47Z
0,ali,淘宝,password,no describe,2019-08-18T12:31:47Z
`

// Saving keeps the record order, and renumbers the accounts.
const renumbered = `4,tencent,qq,admin2,for qq,2019-08-18T12:31:47Z
4,tencent,qq,admin1,for qq,2019-07-18T12:31:47Z
4,tencent,qq,admin0,for qq,2019-06-18T12:31:47Z
5,tencent,wx,admin,for wx,2019-08-18T12:31:47Z
3,tencent,mail.qq.com,admin,for qq mail,2019-08-18T12:31:47Z
1,ali,支付宝,password,no describe,2019-08-18T12:31:47Z
2,ali,淘宝,password,no describe,2019-08-18T12:31:47Z
`

func loadStore(t *testing.T) Store {
	store, err := New(content)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestLoadAndSave(t *testing.T) {
	store := loadStore(t)
	result, err := store.Save()
	if err != nil {
		t.Fatal(err)
	}
	if act, exp := result, renumbered; act != exp {
		t.Errorf("got \n%vexpected \n%v", act, exp)
	}

	// saving again changes nothing
	again, err := New(result)
	if err != nil {
		t.Fatal(err)
	}
	if act, err := again.Save(); err != nil || act != result {
		t.Errorf("got %v, %v expected %v", act, err, result)
	}
}

func TestGroups(t *testing.T) {
	store := loadStore(t)
	groups, err := store.ListGroups()
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(groups), 2; act != exp {
		t.Errorf("got %v groups expected %v", act, exp)
	}
}

func TestTitles(t *testing.T) {
	store := loadStore(t)
	titles, err := store.ListTitles()
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(titles), 5; act != exp {
		t.Errorf("got %v titles expected %v", act, exp)
	}
}

// accounts come sorted by group and title, each with its sn
func TestListAll(t *testing.T) {
	store := loadStore(t)
	all, err := store.ListAll()
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(all), 5; act != exp {
		t.Errorf("got %v accounts expected %v", act, exp)
	}

	var got []string
	for _, r := range all {
		got = append(got, r.Group+"/"+r.Title)
	}
	exp := []string{"ali/支付宝", "ali/淘宝", "tencent/mail.qq.com", "tencent/qq", "tencent/wx"}
	if !compareStrs(got, exp) {
		t.Errorf("got %v expected %v", got, exp)
	}
	for i, r := range all {
		if act, exp := r.Sn, i+1; act != exp {
			t.Errorf("got sn %v expected %v on %v", act, exp, r.Title)
		}
	}
}

func TestGet(t *testing.T) {
	store := loadStore(t)
	record, err := store.Get("qq")
	if err != nil {
		t.Error(err)
	}
	exp := []string{"4", "tencent", "qq", "admin2", "for qq", "2019-08-18T12:31:47Z"}
	if act := record.ToCsvRecord(); !compareStrs(act, exp) {
		t.Errorf("got %v expected %v", act, exp)
	}
}

func TestGetBySn(t *testing.T) {
	store := loadStore(t)

	record, err := store.GetBySn(1)
	if err != nil || record.Title != "支付宝" {
		t.Errorf("got %v, %v expected 支付宝", record, err)
	}
	if _, err := store.GetBySn(6); err != errs.SnNotFound {
		t.Errorf("got %v expected no such sn", err)
	}
}

func TestGetHistory(t *testing.T) {
	store := loadStore(t)
	records, err := store.GetHistory("qq")
	if err != nil {
		t.Error(err)
	}
	if act, exp := len(records), 3; act != exp {
		t.Errorf("got %v records expected %v", act, exp)
	}

	// newest first
	var got []string
	for _, r := range records {
		got = append(got, r.Password)
	}
	if exp := []string{"admin2", "admin1", "admin0"}; !compareStrs(got, exp) {
		t.Errorf("got %v expected %v", got, exp)
	}
}

func TestDeleteGroup(t *testing.T) {
	store := loadStore(t)
	if act, exp := store.DeleteGroup("tencent"), 5; act != exp {
		t.Errorf("got %v records removed expected %v", act, exp)
	}

	groups, err := store.ListGroups()
	if err != nil {
		t.Error(err)
	}
	if act, exp := len(groups), 1; act != exp {
		t.Errorf("got %v groups expected %v", act, exp)
	}

	// nothing is left in the group, so nothing goes
	if act := store.DeleteGroup("tencent"); act != 0 {
		t.Errorf("got %v records removed expected none", act)
	}
}

func TestDeleteTitle(t *testing.T) {
	store := loadStore(t)
	if act, exp := store.DeleteTitle("qq"), 3; act != exp {
		t.Errorf("got %v records removed expected %v", act, exp)
	}

	titles, err := store.ListTitles()
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(titles), 4; act != exp {
		t.Errorf("got %v titles expected %v", act, exp)
	}

	// the versions of an account go together
	if act := store.DeleteTitle("wx"); act != 1 {
		t.Errorf("got %v records removed expected the one version", act)
	}
}

func TestDeleteRenumbersOnSave(t *testing.T) {
	store := loadStore(t)
	store.DeleteTitle("qq")
	if _, err := store.Save(); err != nil {
		t.Fatal(err)
	}

	var got []string
	for _, r := range store.GetRecords() {
		got = append(got, fmt.Sprintf("%v:%v", r.Sn, r.Password))
	}
	// 支付宝 and 淘宝 keep 1 and 2, mail.qq.com and wx move up
	exp := []string{"4:admin", "3:admin", "1:password", "2:password"}
	if !compareStrs(got, exp) {
		t.Errorf("got %v expected %v", got, exp)
	}
}

func TestFilterByGroup(t *testing.T) {
	store := loadStore(t)
	records, err := store.Filter("cent", "")
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(records), 3; act != exp {
		t.Errorf("got %v records expected %v", act, exp)
	}
	var got []string
	for _, r := range records {
		got = append(got, r.Title)
	}
	if exp := []string{"mail.qq.com", "qq", "wx"}; !compareStrs(got, exp) {
		t.Errorf("got %v expected %v", got, exp)
	}
}

func TestFilterByTitle(t *testing.T) {
	store := loadStore(t)
	records, err := store.Filter("", "宝")
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(records), 2; act != exp {
		t.Errorf("got %v records expected %v", act, exp)
	}
}

func TestFilterByGroupAndTitle(t *testing.T) {
	store := loadStore(t)
	records, err := store.Filter("cent", "qq")
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(records), 2; act != exp {
		t.Errorf("got %v records expected %v", act, exp)
	}
	var got []string
	for _, r := range records {
		got = append(got, r.Title)
	}
	if exp := []string{"mail.qq.com", "qq"}; !compareStrs(got, exp) {
		t.Errorf("got %v expected %v", got, exp)
	}
}

// An account is its latest version, living in the group of that version. A
// group it was moved out of neither lists it, nor is it a group any more.
func TestAccountMovedToAnotherGroup(t *testing.T) {
	store, err := New("1,tencent,qq,v2,for qq,2019-08-18T12:31:48Z\n" +
		"1,ali,qq,v1,for qq,2019-08-18T12:31:47Z\n")
	if err != nil {
		t.Fatal(err)
	}

	if records, err := store.Filter("ali", ""); err != nil || len(records) != 0 {
		t.Errorf("got %v, %v expected no account in ali, qq moved to tencent", records, err)
	}
	if groups, err := store.ListGroups(); err != nil || !compareStrs(groups, []string{"tencent"}) {
		t.Errorf("got %v, %v expected tencent alone", groups, err)
	}

	// the old version stays, as the history of qq
	history, err := store.GetHistory("qq")
	if err != nil || len(history) != 2 {
		t.Errorf("got %v, %v expected both versions of qq", history, err)
	}
}

func TestNewLegacy(t *testing.T) {
	store, err := NewLegacy(legacyContent)
	if err != nil {
		t.Fatal(err)
	}

	records := store.GetRecords()
	if act, exp := len(records), 7; act != exp {
		t.Errorf("got %v records expected %v", act, exp)
	}

	// v1 has no sn, the loaded store numbers its accounts right away
	cases := make(map[*Record]int)
	for _, r := range records {
		cases[r] = map[string]int{"支付宝": 1, "淘宝": 2, "mail.qq.com": 3, "qq": 4, "wx": 5}[r.Title]
	}
	assertSns(t, cases)
}
