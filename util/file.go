package util

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/musenwill/mypass/errs"
)

func ReadFromFile(path string) ([]byte, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		err = errs.NoSuchFile
	}
	return content, err
}

func ReadFromUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var b = &bytes.Buffer{}
	io.Copy(b, resp.Body)
	return b.Bytes(), nil
}
