package util

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// maxFactorSize bounds what is read as a factor, a factor is a small file.
const maxFactorSize = 64 << 20

// ReadFromFile gives the content of a file as a factor.
func ReadFromFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read the factor file %v: %w", path, err)
	}
	return content, nil
}

// ReadFromURL gives the content behind a url as a factor. Only a complete
// answer counts: a failure status, a cut off or an oversized body is an error,
// the content would be a wrong factor otherwise.
func ReadFromURL(url string) ([]byte, error) {
	client := http.Client{Timeout: time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %v: %v", url, resp.Status)
	}

	content, err := io.ReadAll(io.LimitReader(resp.Body, maxFactorSize+1))
	if err != nil {
		return nil, fmt.Errorf("read %v: %w", url, err)
	}
	if len(content) > maxFactorSize {
		return nil, fmt.Errorf("%v is larger than %v bytes", url, int64(maxFactorSize))
	}
	return content, nil
}
