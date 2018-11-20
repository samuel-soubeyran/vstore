package main

import (
	"errors"
	"fmt"
	"github.com/sahilm/fuzzy"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func PrintUsage() {
	fmt.Println("Usage: goscrypt reset |  password path/to/obj property [value]")
	fmt.Println("If a value is specified, the value will be upserted to the object property at path")
	fmt.Println("If no value is specified, return the value of the property at path. Support fuzzy path matching. If ambiguous, return list of potential path")
}

func Reset() error {
	path, err := GetRootPath()
	if err != nil {
		return err
	}
	fmt.Printf("Trying to delete: %s\n", path)
	err = os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Couldn't delete %s\n", path)
		return err
	}
	fmt.Printf("Successfully deleted: %s\n", path)
	return nil
}
func stdinselector(paths fuzzy.Matches) (string, error) {
	for i := 0; i < len(paths); i++ {
		fmt.Printf(" %d => %s\n", i, paths[i].Str)
	}
	fmt.Printf(" %d => %s\n", len(paths), " ... enter new file path")
	var idx string
	fmt.Scanln(&idx)
	i, err := strconv.Atoi(idx)
	try := 0
	for err != nil && try < 3 {
		try++
		log.Fatal(err)
		fmt.Scanln(&idx)
		i, err = strconv.Atoi(idx)
	}
	if err != nil {
		return "", errors.New("Couldn't parse input as integer.")
	}
	if i >= len(paths) {
		fmt.Print("new file path: ")
		var relpath string
		fmt.Scanln(&relpath)
		storepath, err := GetStorePath()
		if err != nil {
			return "", err
		}
		path := filepath.Join(storepath, relpath)
		exists, _ := PathExists(path)
		if exists {
			return relpath, nil
		}
		err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return "", err
		}
		return relpath, nil
	}
	return paths[i].Str, nil
}

func main() {
  log.SetFlags(log.LstdFlags | log.Lshortfile)
	args := os.Args[1:]
	if len(args) == 0 || len(args) > 4 {
		PrintUsage()
		os.Exit(1)
	}
	if len(args) == 1 {
		if args[0] != "reset" {
			PrintUsage()
			os.Exit(1)
		}
		Reset()
		os.Exit(0)
	}
	if len(args) < 3 || len(args) > 4 {
		PrintUsage()
		os.Exit(1)
	}
  log.Println("Get Salt")
	salt, err := GetSalt()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
  log.Println("Create AES Key from password and salt")
	key := MakeKey([]byte(args[0]), salt)
  log.Println("Get settings")
	settings, err := GetSettings(&key)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
  log.Println("Update Store")
	err = UpdateStore(settings.Remote)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	objectName := args[1]
  log.Println("Find object path")
	path, err := FindObjectPath(objectName, stdinselector)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	property := args[2]
	if len(args) > 3 {
		value := args[3]
    log.Println("Set value in store")
		err := StoreSetValue(path, property, value, settings.MasterKey)
		if err != nil {
      log.Fatal("Couldn't set the value: ", err)
			os.Exit(1)
		}
	} else {
    log.Println("Get store value")
		value, err := StoreGetValue(path, property, settings.MasterKey)
		if err != nil {
			log.Fatal("Couldn't get the value", err)
			os.Exit(1)
		}
		fmt.Println(value)
		os.Exit(0)
	}
}
