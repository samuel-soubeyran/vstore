package main

import (
	"encoding/json"
	"fmt"
	"github.com/samuel-soubeyran/gojsonpointer"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	REPO_FOLDER_NAME  = "repo"
	STORE_FOLDER_NAME = "store"
	AUTHOR_NAME       = "vstore"
	AUTHOR_EMAIL      = ""
)

func GetStorePath() (string, error) {
	path, err := GetRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, REPO_FOLDER_NAME, STORE_FOLDER_NAME), nil
}
func GetRepoPath() (string, error) {
	path, err := GetRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(path, REPO_FOLDER_NAME), nil
}

func UpdateStore(remote string) error {
	path, err := GetRepoPath()
	if err != nil {
		return err
	}
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// store does not exist
			return CreateStore(remote)
		}
		return err
	}
	repo, err := git.PlainOpen(path)
	if err != nil {
		HandleErr(err, fmt.Sprintf("Couldn't open the repository at path %v", path))
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		HandleErr(err, "Couldn't get the repository worktree")
		return err
	}
	err = worktree.Pull(&git.PullOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		HandleErr(err, "Couldn't pull the worktree")
		return err
	}
	return nil
}

func CreateStore(remote string) error {
	dirPath, err := GetRepoPath()
	if err != nil {
		return err
	}
	err = os.Mkdir(dirPath, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		HandleErr(err, "Couldn't create the repo directory")
		return err
	}
	_, err = git.PlainClone(dirPath, false, &git.CloneOptions{
		URL: remote,
	})
	if err != nil {
		HandleErr(err, "Couldn't clone the repository from remote")
		return err
	}
	return nil
}
func GetRawJsonContent(path string, masterPassword string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	} else {
		b, err := ioutil.ReadAll(file)
		file.Close()
		if err != nil {
			HandleErr(err, fmt.Sprintf("Couldn't read content file at path %v", path))
			return nil, err
		}
		var salt [PW_SALT_BYTES]byte
		copy(salt[:], b[:32])
		b = b[32:]
		key := MakeKey([]byte(masterPassword), salt)
		// decode to json object
		return Decrypt(b, &key)
	}
}
func GetJsonContent(path string, masterPassword string) (map[string]interface{}, error) {
	// read file content
	jsonDocument := map[string]interface{}{}
	rawjson, err := GetRawJsonContent(path, masterPassword)
	if err != nil {
		if os.IsNotExist(err) {
			return jsonDocument, nil
		}
		return nil, err
	}
	err = json.Unmarshal(rawjson, &jsonDocument)
	if err != nil {
		HandleErr(err, "Couldn't read content file as JSON object")
		return nil, err
	}
	return jsonDocument, err
}

func StoreSetValue(path string, property string, value string, masterPassword string) error {
	jsonDocument, err := GetJsonContent(path, masterPassword)
	if err != nil {
		return err
	}
	// update value
	pointer, err := gojsonpointer.NewJsonPointer(property)
	if err != nil {
		HandleErr(err, fmt.Sprintf("%v is not a valid JSON pointer", property))
		return err
	}
	_, err = pointer.Set(jsonDocument, value)
	if err != nil {
		HandleErr(err, fmt.Sprintf("Couldn't update content file at path %v with property %v and value %v", path, property, value))
		return err
	}
	nb, err := json.Marshal(jsonDocument)
	if err != nil {
		HandleErr(err, "Couldn't marshal JSON content")
		return err
	}
	// encrypt
	salt, err := GenerateSalt()
	if err != nil {
		return err
	}
	key := MakeKey([]byte(masterPassword), salt)
	encrypted, err := Encrypt(nb, &key)
	if err != nil {
		return err
	}
	// overwrite file with new encrypted content
	err = ioutil.WriteFile(path, append(salt[:], encrypted...), 0644)
	if err != nil {
		HandleErr(err, fmt.Sprintf("Couldn't write content file at path %v", path))
		return err
	}
	// git commit and push
	repoPath, err := GetRepoPath()
	if err != nil {
		return err
	}
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		HandleErr(err, fmt.Sprintf("Couldn't open repo at path %v", repoPath))
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		HandleErr(err, "Couldn't get worktree")
		return err
	}
	relpath, err := filepath.Rel(repoPath, path)
	if err != nil {
		HandleErr(err, fmt.Sprintf("Couldn't find rel path from %v for path %v", repoPath, path))
		return err
	}
	_, err = worktree.Add(relpath)
	if err != nil {
		HandleErr(err, fmt.Sprintf("Couldn't add file %v to index", relpath))
		return err
	}
	now := time.Now()
	signature := object.Signature{
		Name:  AUTHOR_NAME,
		Email: AUTHOR_EMAIL,
		When:  now,
	}
	_, err = worktree.Commit(fmt.Sprintf("Update content at %s", relpath), &git.CommitOptions{
		Author: &signature,
	})
	if err != nil {
		HandleErr(err, "Couldn't commit content change")
		return err
	}
	err = repo.Push(&git.PushOptions{})
	if err != nil {
		HandleErr(err, "Couldn't push commit to remote")
	}
	return err
}

func StoreGetValue(path string, property string, masterPassword string) (string, error) {
	jsonDocument, err := GetJsonContent(path, masterPassword)
	if err != nil {
		return "", err
	}
	// get value
	pointer, err := gojsonpointer.NewJsonPointer(property)
	if err != nil {
		HandleErr(err, fmt.Sprintf("%v is not a valid JSON pointer", property))
		return "", err
	}
	value, _, err := pointer.Get(jsonDocument)
	if err != nil {
		HandleErr(err, fmt.Sprintf("Couldn't get value at %v for content file at path %v", property, path))
	}
	return value.(string), nil
}
