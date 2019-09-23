package manager

import (
	"io/ioutil"
	"os"

	"github.com/musenwill/mypass/util"
)

func read(crypto util.CryptoApi, path string) ([]byte, error) {
	encoded, err := util.ReadFromFile(path)
	if err != nil {
		return nil, err
	}
	return decryptFromBase64(crypto, string(encoded))
}

func write(crypto util.CryptoApi, conent []byte, path string) error {
	encoded, err := crypt2base64(crypto, conent)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte(encoded), os.ModePerm)
}
