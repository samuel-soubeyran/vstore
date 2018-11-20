package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestCreateEncodedSettingsFile(t *testing.T) {
	path, err := GetSettingsFilePath()
	if err != nil {
		t.Error("Couldn't get settings file path", err)
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		t.Error("Couldn't delete salt file", err)
	}

	key := MakeKey([]byte("testpassword"), []byte("testsalt"))
	settings := usersettings{Remote: "testremote", MasterKey: "testmasterkey"}
	CreateEncodedSettingsFile(&key, settings)

	file, err := os.Open(path)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}
	if len(b) == 0 {
		t.Error("Expecting content of settings file to not be empty")
	}
	plaintext, err := Decrypt(b, &key)
	if err != nil {
		t.Error("Couldn't decrypt file content", err)
	}
	if string(plaintext) != "{\"remote\":\"testremote\",\"master_key\":\"testmasterkey\"}" {
		t.Error("Expecting file content to decrypt as the json settings object, got", string(plaintext))
	}
}
