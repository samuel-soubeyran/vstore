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
  "github.com/sethvargo/go-password/password"
  "strings"
  "bufio"
)

func PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("vstore reset : reset the local store")
	fmt.Println("vstore info : print vstore information")
  fmt.Println("vstore ls: list all files")
	fmt.Println("vstore get path/to/file : get content of file")
	fmt.Println("vstore get path/to/file /jsonpointer : get value at /jsonpointer, add value to clipboard")
  fmt.Println("vstore set path/to/file /jsonpointer [–g|-e] : set value at /jsonpointer using value in [clipboard|-g: generate random|-e enter")
  fmt.Println("vstore remove path/to/file")
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
    HandleErr(err, fmt.Sprintf("Couldn't reset at path %v", path))
		return err
	}
	fmt.Printf("Trying to delete: %s\n", path)
	err = os.RemoveAll(path)
	if err != nil {
		HandleErr(err,fmt.Sprintf("Couldn't delete %s\n", path))
		return err
	}
	fmt.Printf("Successfully deleted: %s\n", path)
	return nil
}

func ListFiles() error {
  path, err := GetStorePath()
  if err != nil {
    HandleErr(err, fmt.Sprintf("Couldn't get storepath: %v", path))
    return err
  }
  err = filepath.Walk(path,
    func(path string, info os.FileInfo, err error) error {
      if filepath.Base(path) == ".git" {
        return filepath.SkipDir
      }
      if err != nil {
        HandleErr(err, "Couldn't list files")
        return err
      }
      fmt.Println(path)
      return nil
  })
  if err != nil {
    HandleErr(err, fmt.Sprintf("Error while listing files at path %v", path))
  }
  return err
}

func GeneratePassword() (string, error) {
  value, err := password.Generate(10, 3, 2, false, false)
  if err != nil {
    HandleErr(err, "Couldn't generate value")
  }
  return value, err
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
    HandleErr(err, fmt.Sprintf("Couldn't parse input as integer %v", idx))
		fmt.Scanln(&idx)
		i, err = strconv.Atoi(idx)
	}
	if err != nil {
		return "", errors.New("Couldn't select file.")
	}
	if i >= len(paths) {
    return create_file_object(target)
	}
	return paths[i].Str, nil
}
func create_file_object(target string)(string, error) {
	storepath, err := GetStorePath()
	if err != nil {
    HandleErr(err, "Couldn't get store path")
		return "", err
	}
	path := filepath.Join(storepath, target)
	exists, _ := PathExists(path)
	if exists {
		return target, nil
	}
	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
    HandleErr(err, fmt.Sprintf("Couldn't create dirs %v", path))
		return "", err
	}
	return target, nil
}
func remove_file_object(target string) error {
  storepath, err := GetStorePath()
  if err != nil {
    HandleErr(err, "Couldn't get store path")
    return err
  }
  path := filepath.Join(storepath, target)
  err = os.Remove(path)
  if err != nil {
    HandleErr(err, fmt.Sprintf("Couldn't remove at path %v", path))
    return err
  }
  return StoreUpdateRemote(path)
}
func get_value_at_pointer(path string, jsonpointer string, key string) error {
	value, err := StoreGetValue(path, jsonpointer, key)
	if err != nil {
    HandleErr(err, "Couldn't get the value")
    return err
	}
	clipboard.WriteAll(value)
	fmt.Println(value)
  return nil
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
    HandleErr(err, "Couldn't get the settings")
    os.Exit(1)
	}

	// Update the store
	err = UpdateStore(settings.Remote)
	if err != nil {
    HandleErr(err, "Couldn't update the sore")
    os.Exit(1)
	}

	// Get content file path
	rel_filepath := args[1]
	path, err := FindObjectPath(rel_filepath, StdinSelector)
	if err != nil {
    HandleErr(err, fmt.Sprintf("Couldn't find path from %v", rel_filepath))
		os.Exit(1)
	}

	cmd := args[0]
  // case 0 : Create file
  if cmd == "create" && len(args) == 2 {
    rel_path, err := create_file_object(rel_filepath)
    if err != nil {
      HandleErr(err, "Couldn't create file")
      os.Exit(1)
    }
    storepath, err := GetStorePath()
    if err != nil {
      HandleErr(err, "Couldn't get store path")
      os.Exit(1)
    }
    path = filepath.Join(storepath, rel_path)
    err = StoreSetValue(path, "/touchobject", "create", settings.MasterKey)
    if err != nil {
      HandleErr(err, fmt.Sprintf("Couldn't write file at path %v", path))
      os.Exit(1)
    }
    os.Exit(0)
  }
  // Remove file
  if cmd == "remove" && len(args) == 2 {
    err := remove_file_object(rel_filepath)
    if err != nil {
      HandleErr(err, fmt.Sprintf("Couldn't remove file at path %v", rel_filepath))
      os.Exit(1)
    }
    os.Exit(0)
  }
	// case 1 : Get file content
	if cmd == "get" && len(args) == 2 {
		rawjson, err := GetRawJsonContent(path, settings.MasterKey)
		if err != nil {
			HandleErr(err, fmt.Sprintf("Couldn't get the content of file at path %v", path))
      os.Exit(1)
		}
		fmt.Println(string(rawjson))
		os.Exit(0)
	}
	jsonpointer := args[2]
	// case 2 : Get value of file at json pointer
	if cmd == "get" && len(args) == 3 {
    err := get_value_at_pointer(path, jsonpointer, settings.MasterKey)
		if err != nil {
      HandleErr(err, fmt.Sprintf("Couldn't read the value at path %v, json path: %v", path, jsonpointer))
      os.Exit(1)
    }
    os.Exit(0)
	}
	// case 3 : Set value of file at json pointer
	if cmd == "set" && len(args) <= 4 {
    value, err := "", errors.New("")
    if len(args) == 4 && args[3] == "-g" {
      value, err = GeneratePassword()
      if err != nil {
        HandleErr(err, "Couldn't generate value")
        os.Exit(1)
      }
    } else if len(args) == 4 && args[3] == "-e" {
      var str strings.Builder
      scanner := bufio.NewScanner(os.Stdin)
      fmt.Println("To register the value, break line and ctrl+d.\nEnter text -->")
	    for scanner.Scan() {
		    str.WriteString(scanner.Text()) // Println will add back the final '\n'
	    }
	    if err := scanner.Err(); err != nil {
		    HandleErr(err, "Couldn't read the value from stdin")
        os.Exit(1)
	    }
      value = str.String()
      fmt.Println("-->Storing")
    } else {
		  value, err = clipboard.ReadAll()
		  if err != nil {
			  HandleErr(err, "Couldn't read the value from clipboard")
        os.Exit(1)
		  }
    }
		err = StoreSetValue(path, jsonpointer, value, settings.MasterKey)
		if err != nil {
      HandleErr(err, fmt.Sprintf("Couldn't set the value at path: %v, jsonpointer: %v, value: %v ", path, jsonpointer, value))
		  os.Exit(1)
    }
		get_value_at_pointer(path, jsonpointer, settings.MasterKey)
		os.Exit(0)
	}
}
