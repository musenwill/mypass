package data

import (
	"testing"
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
	qqRecord := []string{"tencent", "qq", "admin", "for qq"}
	record, err := storage.Put(qqRecord[0], qqRecord[1], qqRecord[2], qqRecord[3])
	if err != nil {
		t.Error(err)
		return
	}
	qqRecord = append(qqRecord, record.Ct.Format(timeFormat))

	wxRecord := []string{"tencent", "wx", "admin", "for wx"}
	record, err = storage.Put(wxRecord[0], wxRecord[1], wxRecord[2], wxRecord[3])
	if err != nil {
		t.Error(err)
		return
	}
	wxRecord = append(wxRecord, record.Ct.Format(timeFormat))

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
		t.Errorf("got %v expected %v", r, qqRecord)
	}

	// delete from store
	err = storage.DeleteTitle("qq")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = storage.Get("qq")
	if err == nil {
		t.Error("got no err expected no data error")
	}
}

const content = `tencent,qq,admin2,for qq,2019-08-18 12:31:47 +0000 UTC
tencent,qq,admin1,for qq,2019-07-18 12:31:47 +0000 UTC
tencent,qq,admin0,for qq,2019-06-18 12:31:47 +0000 UTC
tencent,wx,admin,for wx,2019-08-18 12:31:47 +0000 UTC
tencent,mail.qq.com,admin,for qq mail,2019-08-18 12:31:47 +0000 UTC
ali,支付宝,password,no describe,2019-08-18 12:31:47 +0000 UTC
ali,淘宝,password,no describe,2019-08-18 12:31:47 +0000 UTC
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
	if act, exp, actLen, expLen := result, content, len(result), len(content); act != exp || actLen != expLen {
		t.Errorf("got \n%vexpected %v \ngot actLen %v expLen %v", act, exp, actLen, expLen)
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
		t.Errorf("got %v groups expected %v", act, exp)
	}
}

func TestListAll(t *testing.T) {
	store := loadStore(t)
	all, err := store.ListAll()
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(all), 5; act != exp {
		t.Errorf("got %v groups expected %v", act, exp)
	}
}

func TestGet(t *testing.T) {
	store := loadStore(t)
	record, err := store.Get("qq")
	if err != nil {
		t.Error(err)
	}
	exp := []string{"tencent", "qq", "admin2", "for qq", "2019-08-18 12:31:47 +0000 UTC"}
	if act, exp := record.ToCsvRecord(), exp; !compareStrs(act, exp) {
		t.Errorf("got %v expected %v", act, exp)
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
}

func TestDeleteGroup(t *testing.T) {
	store := loadStore(t)
	err := store.DeleteGroup("tencent")
	if err != nil {
		t.Error(err)
	}

	groups, err := store.ListGroups()
	if err != nil {
		t.Error(err)
	}
	if act, exp := len(groups), 1; act != exp {
		t.Errorf("got %v groups expected %v", act, exp)
	}
}

func TestDeleteTitle(t *testing.T) {
	store := loadStore(t)
	err := store.DeleteTitle("qq")
	if err != nil {
		t.Error(err)
	}

	titles, err := store.ListTitles()
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(titles), 4; act != exp {
		t.Errorf("got %v groups expected %v", act, exp)
	}
}

func TestFilterByGroup(t *testing.T) {
	store := loadStore(t)
	records, err := store.Filter("cent", "")
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(records), 3; act != exp {
		t.Errorf("got %v groups expected %v", act, exp)
	}
}

func TestFilterByTitle(t *testing.T) {
	store := loadStore(t)
	records, err := store.Filter("", "宝")
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(records), 2; act != exp {
		t.Errorf("got %v groups expected %v", act, exp)
	}
}

func TestFilterByGroupAndTitle(t *testing.T) {
	store := loadStore(t)
	records, err := store.Filter("cent", "qq")
	if err != nil {
		t.Error(err)
	}

	if act, exp := len(records), 2; act != exp {
		t.Errorf("got %v groups expected %v", act, exp)
	}
}
