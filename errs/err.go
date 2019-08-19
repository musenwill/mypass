package errs

import "errors"

var (
	DecryptError     = errors.New("decrypt error")
	InvalidKey       = errors.New("invalid key")
	InvalidCsvRecord = errors.New("invalid record")
	DataNotFound     = errors.New("data not found")
	NoSuchFile       = errors.New("no such file")
	Uninited         = errors.New("uninited, command `mypass init` to init")
)
