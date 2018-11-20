package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestMakeKey(t *testing.T) {
	salt := []byte("sdqfghjfdsfsdfgsdfgfsdfgsdgsdfgsdfgsdfgsdfgsdfdgsdfgsfgsdfgsdfgsfgsdfgsdfg")
	key := MakeKey([]byte("password"), salt)
	if len(key) != 32 {
		t.Error("Expecting key of length 32, got", len(key))
	}
}

func TestGetSalt(t *testing.T) {
	path, err := GetSaltFilePath()
	if err != nil {
		t.Error("Couldn't get salt file path", err)
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		t.Error("Couldn't delete salt file", err)
	}
	salt, err := GetSalt()
	if err != nil {
		t.Error(err)
	}
	if len(salt) < 8 {
		t.Error("Expecting salt of length at least 8, got", len(salt))
	}
	if err != nil {
		t.Error(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(b, salt) {
		t.Error("Expecting content of salt file to be same as salt got", string(b), string(salt))
	}
	salt2, err := GetSalt()
	if !bytes.Equal(salt, salt2) {
		t.Error("Second call to GetSalt should return same salt", string(salt), string(salt2))
	}
}
