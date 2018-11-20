package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	ENCRYPTED_SETTINGS_FILE = "settings.json.enc"
)

type usersettings struct {
	Remote    string `json:"remote"`
	MasterKey string `json:"master_key"`
}

func GetSettingsFilePath() (string, error) {
	path, err := GetRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, ENCRYPTED_SETTINGS_FILE), nil
}


func GetSettings(password string) (usersettings, error) {
	salt, err := GetSalt()
	if err != nil {
		return usersettings{}, err
	}
  key := MakeKey([]byte(password), salt)

	settingsPath, err := GetSettingsFilePath()
	if err != nil {
		return usersettings{}, err
	}
	file, err := os.Open(settingsPath)
	if err != nil {
		if !os.IsNotExist(err) {
      HandleErr(err, "Couldn't open settings file")
			return usersettings{}, err
		}
		return CreateSettings(&key)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
    HandleErr(err, "Couldn't read settings file")
		return usersettings{}, err
	}
	rawSettings, err := Decrypt(b, &key)
	if err != nil {
    HandleErr(err, "Couldn't decrypt settings file content")
		return usersettings{}, err
	}
	var userSettings usersettings
	err = json.Unmarshal(rawSettings, &userSettings)
  if err != nil {
    HandleErr(err, "Couldn't read settings as JSON object")
  }
	return userSettings, err
}

func CreateSettings(key *[32]byte) (usersettings, error) {
	fmt.Println("Could not find settings file. Creating a new one.")
	var masterKey string
	var remote string
	fmt.Print("master key: ")
	fmt.Scanln(&masterKey)
	fmt.Print("remote: ")
	fmt.Scanln(&remote)
	settings := usersettings{Remote: remote, MasterKey: masterKey}
	err := CreateEncodedSettingsFile(key, settings)
	if err != nil {
		return usersettings{}, err
	}
	return settings, nil
}

func CreateEncodedSettingsFile(key *[32]byte, settings usersettings) error {
	b, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	encrypted, err := Encrypt(b, key)
	if err != nil {
    HandleErr(err, "Couldn't encrypt user settings")
		return err
	}
	path, err := GetSettingsFilePath()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, encrypted, 0644)
	if err != nil {
    HandleErr(err, "Couldn't write settings file")
		return err
	}
	return nil
}
