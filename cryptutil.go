package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"golang.org/x/crypto/pbkdf2"
	"io"
  "fmt"
)

const (
	PW_SALT_BYTES = 32
  PW_KEY_BYTES = 32
)

func MakeKey(password []byte, salt [PW_SALT_BYTES]byte) [PW_KEY_BYTES]byte {
  dk := pbkdf2.Key(password, salt[:], 4096, PW_KEY_BYTES, sha512.New)
	var arr [32]byte
	copy(arr[:], dk)
	return arr
}

// https://github.com/gtank/cryptopasta/blob/master/encrypt.go
// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
func Decrypt(ciphertext []byte, key *[PW_KEY_BYTES]byte) (plaintext []byte, err error) {
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
func Encrypt(plaintext []byte, key *[PW_KEY_BYTES]byte) (ciphertext []byte, err error) {
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
func GenerateSalt() ([PW_SALT_BYTES]byte, error) {
	salt := make([]byte, PW_SALT_BYTES)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
    HandleErr(err, "Couldn't get enough entropy")
		return [PW_SALT_BYTES]byte{}, err
	}
  var saltByteArr [PW_SALT_BYTES]byte
  copy(saltByteArr[:], salt)
	return saltByteArr, nil
}
