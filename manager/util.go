package manager

import (
	"encoding/base64"
	"os"

	"github.com/musenwill/mypass/util"
)

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func createDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func createFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func crypt2base64(crypto util.CryptoApi, content []byte) (string, error) {
	encripted, err := crypto.Encrypt(content)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encripted), nil
}

func decryptFromBase64(crypto util.CryptoApi, encoded string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	decripted, err := crypto.Decrypt(decoded)
	if err != nil {
		return nil, err
	}

	return decripted, nil
}
