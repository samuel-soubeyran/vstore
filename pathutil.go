package main

import (
	"github.com/sahilm/fuzzy"
	"os"
	"path/filepath"
  "fmt"
)

const (
	ROOT_FOLDER_NAME = "vstore"
)

func FindObjectPath(objectName string, selector func(string, fuzzy.Matches) (string, error)) (string, error) {
	matches, err := GetFuzzyPath(objectName)
	if err != nil {
		return "", err
	}
  storepath, err := GetStorePath()
  if err != nil {
    return "", err
  }
	if len(matches) == 1 {
		return filepath.Join(storepath, matches[0].Str), nil
	} else {
		path, err := selector(objectName, matches)
		if err != nil {
			return "", err
		}
		return filepath.Join(storepath, path), nil
	}
}

func GetFuzzyPath(objectName string) (fuzzy.Matches, error) {
	storepath, err := GetStorePath()
	if err != nil {
		return nil, err
	}
  exists, _ := PathExists(storepath)
  if !exists {
    return fuzzy.Matches{}, nil
  }
	files, err := FilePathWalkDir(storepath)
	if err != nil {
		return nil, err
	}
	matches := fuzzy.Find(objectName, files)
	return matches, nil
}

func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
      storepath, err := GetStorePath()
      if err != nil {
        return err
      }
      relpath, err := filepath.Rel(storepath, path)
      if err != nil {
        HandleErr(err, fmt.Sprintf("Couldn't get a relative path from %v with base %v", path, storepath))
        return err
      }
			files = append(files, relpath)
		}
		return nil
	})
	return files, err
}

func GetRootPath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ROOT_FOLDER_NAME), nil
}
func CreateRootDir() (string, error) {
	dirPath, err := GetRootPath()
	if err != nil {
		return "", err
	}
	err = os.Mkdir(dirPath, os.ModePerm)
	if !os.IsNotExist(err) {
		return "", err
	}
	return dirPath, nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
