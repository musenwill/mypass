package manager

import (
	"os"
	"path/filepath"
)

// writeFileAtomic writes to a temp file beside path first, and renames it over
// path only after the temp file is on the disk. The old file is left untouched
// when anything fails. The file ends up with mode 0600, only its owner can read it.
func writeFileAtomic(path string, content []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// nothing left to remove once the rename is done
	defer os.Remove(tmpName)

	if _, err = tmp.Write(content); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmpName, path); err != nil {
		return err
	}

	// make the rename durable too, it fails silently where directories can't be synced
	if dir, err := os.Open(filepath.Dir(path)); err == nil {
		dir.Sync()
		dir.Close()
	}
	return nil
}
