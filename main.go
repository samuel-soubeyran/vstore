package main

import (
	"errors"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/sahilm/fuzzy"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("vstore reset : reset the local store")
	fmt.Println("vstore info : print vstore information")
  fmt.Println("vstore ls: list all files")
	fmt.Println("vstore get path/to/file : get content of file")
	fmt.Println("vstore get path/to/file /jsonpointer : get value at /jsonpointer, add value to clipboard")
	fmt.Println("vstore set path/to/file /jsonpointer : set value at /jsonpointer using value in clipboard")
  fmt.Println("vstore create path/to/file: force create file")
}

func PrintInfo() (error){
	path, err := GetRootPath()
	if err != nil {
		return err
	}
	fmt.Println("Info")
	fmt.Printf("store path: %s\n", path)
  return nil
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

func ListFiles() error {
  path, err := GetStorePath()
  if err != nil {
    return err
  }
  err = filepath.Walk(path,
    func(path string, info os.FileInfo, err error) error {
      if filepath.Base(path) == ".git" {
        return filepath.SkipDir
      }
      if err != nil {
        return err
      }
      fmt.Println(path)
      return nil
  })
  return err
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
    return create_file_object(target)
	}
	return paths[i].Str, nil
}
func create_file_object(target string)(string, error) {
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
func get_value_at_pointer(path string, jsonpointer string, key string) {
	value, err := StoreGetValue(path, jsonpointer, key)
	if err != nil {
		log.Fatal("Couldn't get the value ", err)
	}
	clipboard.WriteAll(value)
	fmt.Println(value)
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
		if args[0] == "info" {
			PrintInfo()
			os.Exit(0)
		}
		if args[0] == "reset" {
			Reset()
			os.Exit(0)
		}
    if args[0] == "ls" {
      ListFiles()
      os.Exit(0)
    }
		PrintUsage()
		os.Exit(1)
	}

	// Get the settings
	password := os.Getenv("VSTORE_PASSWORD")
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
	rel_filepath := args[1]
	path, err := FindObjectPath(rel_filepath, StdinSelector)
	if err != nil {
		log.Fatal(err)
	}

	cmd := args[0]
  // case 0 : Create file
  if cmd == "create" && len(args) == 2 {
    rel_path, err := create_file_object(rel_filepath)
    if err != nil {
      log.Fatal("Couldn't create file", err)
    }
    storepath, err := GetStorePath()
    if err != nil {
      log.Fatal("Couldn't get store path")
    }
    path = filepath.Join(storepath, rel_path)
    err = StoreSetValue(path, "/touchobject", "create", settings.MasterKey)
    if err != nil {
      log.Fatal("Couldn't write file")
    }
    os.Exit(0)
  }
	// case 1 : Get file content
	if cmd == "get" && len(args) == 2 {
		rawjson, err := GetRawJsonContent(path, settings.MasterKey)
		if err != nil {
			log.Fatal("Couldn't get the content of file ", err)
		}
		fmt.Println(string(rawjson))
		os.Exit(0)
	}
	jsonpointer := args[2]
	// case 2 : Get value of file at json pointer
	if cmd == "get" && len(args) == 3 {
		get_value_at_pointer(path, jsonpointer, settings.MasterKey)
		os.Exit(0)
	}
	// case 3 : Set value of file at json pointer
	if cmd == "set" && len(args) == 3 {
		value, err := clipboard.ReadAll()
		if err != nil {
			log.Fatal("Couldn't read the value from clipboard")
		}
		err = StoreSetValue(path, jsonpointer, value, settings.MasterKey)
		if err != nil {
			log.Fatal("Couldn't set the value: ", err)
		}
		get_value_at_pointer(path, jsonpointer, settings.MasterKey)
		os.Exit(0)
	}
}
