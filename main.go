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
	fmt.Println("Usage:")
	fmt.Println("vstore reset : reset the local store")
	fmt.Println("vstore password path/to/file : get content of file")
	fmt.Println("vstore password path/to/file /jsonpointer : get value at /jsonpointer")
	fmt.Println("vstore password path/to/file /jsonpointer value : set value at /jsonpointer")
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

func StdinSelector(target string, paths fuzzy.Matches) (string, error) {
	for i := 0; i < len(paths); i++ {
		fmt.Printf(" %d => %s\n", i, paths[i].Str)
	}
	fmt.Printf(" %d => %s\n", len(paths), " ... new file path")
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
		storepath, err := GetStorePath()
		if err != nil {
			return "", err
		}
		path := filepath.Join(storepath, target)
		exists, _ := PathExists(path)
		if exists {
			return target, nil
		}
		err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			return "", err
		}
		return target, nil
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
	// Reset the store
	if len(args) == 1 {
		if args[0] != "reset" {
			PrintUsage()
			os.Exit(1)
		}
		Reset()
		os.Exit(0)
	}

	if len(args) < 2 || len(args) > 4 {
		PrintUsage()
		os.Exit(1)
	}
	
	// Get the settings
  password := args[0]
	settings, err := GetSettings(password)
	if err != nil {
		log.Fatal(err)
	}

	// Update the store
	err = UpdateStore(settings.Remote)
	if err != nil {
		log.Fatal(err)
	}

	// Get content file path
	filepath := args[1]
	path, err := FindObjectPath(filepath, StdinSelector)
	if err != nil {
		log.Fatal(err)
	}

	// case 1 : Get file content
	if len(args) == 2 {
		rawjson, err := GetRawJsonContent(path, settings.MasterKey)
		if err != nil {
			log.Fatal("Couldn't get the content of file ", err)
		}
		fmt.Println(string(rawjson))
		os.Exit(0)
	}
	jsonpointer := args[2]
	// case 2 : Get value of file at json pointer
	if len(args) == 3 {
		value, err := StoreGetValue(path, jsonpointer, settings.MasterKey)
		if err != nil {
			log.Fatal("Couldn't get the value ", err)
		}
		fmt.Println(value)
		os.Exit(0)
	}
	// case 3 : Set value of file at json pointer
	if len(args) > 3 {
		value := args[3]
		err := StoreSetValue(path, jsonpointer, value, settings.MasterKey)
		if err != nil {
      log.Fatal("Couldn't set the value: ", err)
		}
	}
}
