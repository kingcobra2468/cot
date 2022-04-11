package crypto

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/kingcobra2468/cot/internal/config"
	"github.com/patrickmn/go-cache"
)

var (
	keyCache      = cache.New(cache.NoExpiration, 0)
	cotPrivateKey = ""
)

var (
	errInvalidService = errors.New("client number doesn't exist")
	cryptoConfig      *config.Encryption
)

func Decrypt(clientNumber, message string) (string, error) {
	key, found := keyCache.Get(clientNumber)
	if !found {
		return "", errInvalidService
	}

	if cryptoConfig.Base64Encoding {
		cipherText, err := b64.StdEncoding.DecodeString(message)
		if err != nil {
			return "", err
		}
		message = string(cipherText)
	}

	var err error
	if cryptoConfig.SignatureVerification {
		message, err = helper.DecryptVerifyMessageArmored(string(key.(string)), cotPrivateKey, []byte(cryptoConfig.Passphrase), string(message))
	} else {
		message, err = helper.DecryptMessageArmored(cotPrivateKey, []byte(cryptoConfig.Passphrase), string(message))
	}

	return message, err
}

func Encrypt(clientNumber, message string) (string, error) {
	key, found := keyCache.Get(clientNumber)
	if !found {
		return "", errInvalidService
	}

	var err error
	if cryptoConfig.SignatureVerification {
		message, err = helper.EncryptSignMessageArmored(string(key.(string)), cotPrivateKey, []byte(cryptoConfig.Passphrase), message)
	} else {
		message, err = helper.EncryptMessageArmored(string(key.(string)), message)
	}
	if err != nil {
		return "", err
	}

	message = b64.StdEncoding.EncodeToString([]byte(message))

	return message, nil
}

func LoadClientNumberKeys(path string) error {
	keyPaths, err := filepath.Glob(filepath.Join(path, "*.asc"))
	if err != nil {
		return err
	}
	for _, path := range keyPaths {
		publicKey, err := os.ReadFile(path)
		if err != nil {
			fmt.Println(err)
			// TODO: log error
			continue
		}

		fileName := filepath.Base(path)
		clientNumber := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		keyCache.Set(clientNumber, string(publicKey), cache.NoExpiration)
	}

	return nil
}

func SetConfig(c *config.Encryption) error {
	cryptoConfig = c

	privateKey, err := os.ReadFile(c.PrivateKeyFile)
	if err != nil {
		return err
	}
	cotPrivateKey = string(privateKey)

	return nil
}
