package errs

import "errors"

var (
	DecryptError     = errors.New("decrypt error")
	InvalidKey       = errors.New("invalid key")
	InvalidCsvRecord = errors.New("invalid record")
	DataNotFound     = errors.New("data not found")
)
