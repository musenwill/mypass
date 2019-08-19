package util

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

func ReadFromFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
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
