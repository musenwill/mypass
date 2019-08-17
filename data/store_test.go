package data

import "testing"

func TestCRUD(t *testing.T) {
	storage, err := NewStoreage("")
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
	err = storage.Put(qqRecord[0], qqRecord[1], qqRecord[2], qqRecord[3])
	if err != nil {
		t.Error(err)
		return
	}

	wxRecord := []string{"tencent", "wx", "admin", "for wx"}
	err = storage.Put(wxRecord[0], wxRecord[1], wxRecord[2], wxRecord[3])
	if err != nil {
		t.Error(err)
		return
	}

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
	if !compareStrs(r, qqRecord) {
		t.Errorf("got %v expected %v", r, qqRecord)
	}

	r, err = storage.Get("wx")
	if err != nil {
		t.Error(err)
	}
	if !compareStrs(r, wxRecord) {
		t.Errorf("got %v expected %v", r, qqRecord)
	}

	// delete from store
	err = storage.Delete("qq")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = storage.Get("qq")
	if err == nil {
		t.Error("got no err expected no data error")
	}
}
