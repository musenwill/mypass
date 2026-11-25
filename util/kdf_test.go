package util

import (
	"bytes"
	"testing"
)

func fastKdf() KdfParams {
	return KdfParams{Name: kdfArgon2id, Memory: 8, Time: 1, Threads: 1}
}

func TestKdfDerive(t *testing.T) {
	p := fastKdf()
	key, err := p.Derive([]byte("pin"), []byte("salt"))
	if err != nil {
		t.Fatal(err)
	}

	if act, exp := len(key), KeyLen; act != exp {
		t.Errorf("got %v key len expected %v", act, exp)
	}

	again, _ := p.Derive([]byte("pin"), []byte("salt"))
	if !bytes.Equal(key, again) {
		t.Error("got different keys expected the derivation to be deterministic")
	}

	cases := map[string][]byte{
		"another salt":   mustDerive(t, p, []byte("pin"), []byte("another salt")),
		"another factor": mustDerive(t, p, []byte("another pin"), []byte("salt")),
	}
	for name, other := range cases {
		if bytes.Equal(key, other) {
			t.Errorf("got same key expected keys differ on %v", name)
		}
	}

	longer, _ := KdfParams{Name: kdfArgon2id, Memory: 8, Time: 2, Threads: 1}.Derive([]byte("pin"), []byte("salt"))
	if bytes.Equal(key, longer) {
		t.Error("got same key expected keys differ on params")
	}
}

func TestKdfCostIndependentOfFactorSize(t *testing.T) {
	p := fastKdf()
	small := []byte("pin")
	large := bytes.Repeat(small, 1024*1024)

	_, err := p.Derive(large, []byte("salt"))
	if err != nil {
		t.Fatal(err)
	}
	if key, _ := p.Derive(small, []byte("salt")); len(key) != KeyLen {
		t.Errorf("got %v key len expected %v", len(key), KeyLen)
	}
}

func TestKdfValidate(t *testing.T) {
	if err := DefaultKdfParams.Validate(); err != nil {
		t.Errorf("got %v expected default params valid", err)
	}

	cases := map[string]KdfParams{
		"unknown kdf":    {Name: "scrypt", Memory: 65536, Time: 3, Threads: 1},
		"zero time":      {Name: kdfArgon2id, Memory: 65536, Time: 0, Threads: 1},
		"huge time":      {Name: kdfArgon2id, Memory: 65536, Time: KdfMaxTime + 1, Threads: 1},
		"zero threads":   {Name: kdfArgon2id, Memory: 65536, Time: 3, Threads: 0},
		"too many th":    {Name: kdfArgon2id, Memory: 65536, Time: 3, Threads: 17},
		"memory too low": {Name: kdfArgon2id, Memory: 7, Time: 3, Threads: 1},
		"huge memory":    {Name: kdfArgon2id, Memory: KdfMaxMemory + 1, Time: 3, Threads: 1},
		"mem < 8*th":     {Name: kdfArgon2id, Memory: 16, Time: 3, Threads: 4},
	}
	for name, p := range cases {
		if _, err := p.Derive([]byte("pin"), []byte("salt")); err == nil {
			t.Errorf("got no err expected error on %v", name)
		}
	}
}

func mustDerive(t *testing.T, p KdfParams, factor, salt []byte) []byte {
	t.Helper()
	key, err := p.Derive(factor, salt)
	if err != nil {
		t.Fatal(err)
	}
	return key
}
