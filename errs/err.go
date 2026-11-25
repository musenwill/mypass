package errs

import "errors"

var (
	DecryptError     = errors.New("decrypt error")
	InvalidPincode   = errors.New("invalid pincode, or the store file is corrupted")
	InvalidToken     = errors.New("invalid token, or the data is corrupted")
	InvalidCsvRecord = errors.New("invalid record")
	DataNotFound     = errors.New("data not found")
	SnNotFound       = errors.New("no such sn")
	NothingToDelete  = errors.New("no such group or title, nothing is deleted")
	Uninited         = errors.New("uninited, command `mypass init` to init")
	LegacyStore      = errors.New("store is in the old format, command `mypass upgrade` to upgrade")
	InvalidFormat    = errors.New("invalid store format")
	TokenUnset       = errors.New("token is not unlocked")
	EmptyStore       = errors.New("empty store")
)
