package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
  "fmt"
)

const (
	PW_SALT_BYTES = 32
	SALT_FILE     = "salt"
)

func GetSaltFilePath() (string, error) {
	path, err := GetRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, SALT_FILE), nil
}
func GetSalt() ([]byte, error) {
	_, err := CreateRootDir()
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	path, err := GetSaltFilePath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
      HandleErr(err, "Expected error while opening saltfile.")
			return nil, err
		}
		return CreateSalt()
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
    HandleErr(err, "Error while reading saltfile")
		return nil, err
	}
	return b, nil
}
func MakeKey(password []byte, salt []byte) [32]byte {
	dk := pbkdf2.Key(password, salt, 4096, 32, sha512.New)
	var arr [32]byte
	copy(arr[:], dk)
	return arr
}

// https://github.com/gtank/cryptopasta/blob/master/encrypt.go
// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
func Decrypt(ciphertext []byte, key *[32]byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
    HandleErr(err, fmt.Sprintf("Couldn't create cipher with key %v", key))
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
    HandleErr(err, "Couldn't create new GCM from block")
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}

// https://github.com/gtank/cryptopasta/blob/master/encrypt.go
// Encrypt encrypts data using 256-bit AES-GCM.  This both hides the content of
// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Output takes the
// form nonce|ciphertext|tag where '|' indicates concatenation.
func Encrypt(plaintext []byte, key *[32]byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
    HandleErr(err, fmt.Sprintf("Couldn't create cipher from key %v", key))
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
    HandleErr(err, "Couldn't create new GCM from block")
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
    HandleErr(err, "Couldn't get enough entropy")
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, PW_SALT_BYTES)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
    HandleErr(err, "Couldn't get enough entropy")
		return nil, err
	}
	return salt, nil
}
func CreateSalt() ([]byte, error) {
	path, err := GetSaltFilePath()
	if err != nil {
		return nil, err
	}
	salt, err := GenerateSalt()
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(path, salt, 0644)
	if err != nil {
    HandleErr(err, "Couldn't write saltfile")
		return nil, err
	}
	return salt, nil
}
