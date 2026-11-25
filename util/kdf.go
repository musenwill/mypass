package util

import (
	"crypto/sha256"
	"errors"

	"golang.org/x/crypto/argon2"
)

const (
	kdfArgon2id = "argon2id"

	// KdfMaxMemory and KdfMaxTime bound the params read from a store file,
	// so a tampered header can't make the derivation run forever.
	KdfMaxMemory uint32 = 1 << 20 // KiB
	KdfMaxTime   uint32 = 64
)

var ErrUnsupportedKdf = errors.New("unsupported kdf params")

// KdfParams tells how a key is derived from a factor. They live in the store
// file, so they can be tuned without another file format.
type KdfParams struct {
	Name    string `json:"kdf"`
	Memory  uint32 `json:"m"` // KiB
	Time    uint32 `json:"t"`
	Threads uint8  `json:"p"`
}

// DefaultKdfParams is used when a new salt is generated.
var DefaultKdfParams = KdfParams{Name: kdfArgon2id, Memory: 65536, Time: 3, Threads: 1}

// Derive turns a factor, string or file content, into a KeyLen key.
// The factor is hashed first, so the cost doesn't grow with the factor size.
// Brute forcing is not helped by that, every candidate still pays one kdf.
func (p KdfParams) Derive(factor, salt []byte) ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	digest := sha256.Sum256(factor)
	switch p.Name {
	case kdfArgon2id:
		return argon2.IDKey(digest[:], salt, p.Time, p.Memory, p.Threads, KeyLen), nil
	default:
		return nil, ErrUnsupportedKdf
	}
}

func (p KdfParams) Validate() error {
	if p.Name != kdfArgon2id {
		return ErrUnsupportedKdf
	}
	// argon2 needs at least 8 blocks of 1KiB per thread
	if p.Time < 1 || p.Time > KdfMaxTime || p.Threads < 1 || p.Threads > 16 {
		return ErrUnsupportedKdf
	}
	if uint64(p.Memory) < 8*uint64(p.Threads) || p.Memory > KdfMaxMemory {
		return ErrUnsupportedKdf
	}
	return nil
}
