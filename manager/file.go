package manager

import (
	"encoding/base64"
	"io/ioutil"
	"os"

	"github.com/musenwill/mypass/util"
)

func read(crypto util.CryptoApi, path string) ([]byte, error) {
	encoded, err := util.ReadFromFile(path)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return nil, err
	}

	decripted, err := crypto.Decrypt(decoded)
	if err != nil {
		return nil, err
	}

	return decripted, nil
}

func write(crypto util.CryptoApi, conent []byte, path string) error {
	encripted, err := crypto.Encrypt(conent)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(encripted)
	return ioutil.WriteFile(path, []byte(encoded), os.ModePerm)
}
