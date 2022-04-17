// crypto handles the PGP encryption/decryption as well as the
// Base64 encoding/decoding that can take place.
package crypto

import (
	b64 "encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/golang/glog"
	"github.com/kingcobra2468/cot/internal/config"
	"github.com/patrickmn/go-cache"
)

var (
	// stores the PGP public key for each client number
	keyCache      = cache.New(cache.NoExpiration, 0)
	cotPrivateKey = ""
	cryptoConfig  *config.Encryption
)

var (
	errInvalidClientNumber = errors.New("client number doesn't exist")
)

// Decrypt attempts to decrypt a PGP message sent by a client number.
func Decrypt(clientNumber, message string) (string, error) {
	publicKey, found := keyCache.Get(clientNumber)
	if !found {
		return "", errInvalidClientNumber
	}

	// perform base64 decoding if enabled
	if cryptoConfig.Base64Encoding {
		cipherText, err := b64.StdEncoding.DecodeString(message)
		if err != nil {
			return "", err
		}
		message = string(cipherText)
	}
	var err error
	if cryptoConfig.SignatureVerification {
		message, err = helper.DecryptVerifyMessageArmored(string(publicKey.(string)), cotPrivateKey, []byte(cryptoConfig.Passphrase), message)
	} else {
		message, err = helper.DecryptMessageArmored(cotPrivateKey, []byte(cryptoConfig.Passphrase), message)
	}

	return message, err
}

// Encrypt attempts to encrypt a plaintext message into a PGP ASCII-armored message for a
// given client number.
func Encrypt(clientNumber, message string) (string, error) {
	publicKey, found := keyCache.Get(clientNumber)
	if !found {
		return "", errInvalidClientNumber
	}

	var err error
	if cryptoConfig.SignatureVerification {
		message, err = helper.EncryptSignMessageArmored(string(publicKey.(string)), cotPrivateKey, []byte(cryptoConfig.Passphrase), message)
	} else {
		message, err = helper.EncryptMessageArmored(string(publicKey.(string)), message)
	}
	if err != nil {
		return "", err
	}

	message = b64.StdEncoding.EncodeToString([]byte(message))

	return message, nil
}

// LoadClientNumberKeys attempts to register every public key (ending with .asc extension) to a client number.
// The name of the file should match 1:1 with what is specified in the configuration file.
func LoadClientNumberKeys(path string) error {
	keyPaths, err := filepath.Glob(filepath.Join(path, "*.asc"))
	if err != nil {
		return err
	}
	for _, path := range keyPaths {
		publicKey, err := os.ReadFile(path)
		if err != nil {
			glog.Errorln(err)
			continue
		}

		// extract the client number from the filename
		fileName := filepath.Base(path)
		clientNumber := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		keyCache.Set(clientNumber, string(publicKey), cache.NoExpiration)
	}

	return nil
}

// SetConfig sets up the crypto module by providing the encryption configuration.
func SetConfig(c *config.Encryption) error {
	cryptoConfig = c

	privateKey, err := os.ReadFile(c.PrivateKeyFile)
	if err != nil {
		return err
	}
	cotPrivateKey = string(privateKey)

	return nil
}
