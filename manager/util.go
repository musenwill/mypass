package manager

import (
	"encoding/base64"
	"os"

	"github.com/musenwill/mypass/util"
)

// pathExists tells whether there is something at path.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func crypt2base64(crypto util.CryptoApi, content []byte) (string, error) {
	encrypted, err := crypto.Encrypt(content)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func decryptFromBase64(crypto util.CryptoApi, encoded string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	decrypted, err := crypto.Decrypt(decoded)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}
