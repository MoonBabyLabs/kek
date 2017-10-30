package kek

import (
	"os"
	"io/ioutil"
	"encoding/json"
	"strings"
	"github.com/mitchellh/go-homedir"
)

const DOC_DIR = "d/"
const FIELD_DIR = "f/"
const KEK_DIR = "/.kek/"


// Deprecated as you should be using the kek.store{}.Save method
func Save(kekLocale string, content interface{}) error {
	marshallData, err := json.Marshal(content)

	if err != nil {
		return err
	}

	pathFilename := strings.Split(kekLocale, "/")
	path := pathFilename[:len(pathFilename) - 1]
	pathString := strings.Join(path, "/")
	homeDir, _ := homedir.Dir()
	os.MkdirAll(homeDir + KEK_DIR + pathString, 0755)

	return ioutil.WriteFile(homeDir + KEK_DIR + kekLocale, marshallData, 0755)
}

// Deprecated as you should be using the kek.store{}.Delete method
func Delete(locale string) error {
	homeDir, _ := homedir.Dir()
	err := os.Remove(homeDir + KEK_DIR + locale)

	return err
}

// Deprecated as you should be using the kek.store{}.Load method
func Load(locale string, unmarshallStruct interface{}) (interface{}, error) {
	homeDir, _ := homedir.Dir()
	file, readErr := ioutil.ReadFile(homeDir + KEK_DIR + locale)

	if readErr != nil {
		return unmarshallStruct, readErr
	}

	json.Unmarshal(file, &unmarshallStruct)

	return unmarshallStruct, nil
}

// Deprecated as you should use kek.store{}.List()
func List(locale string, limit int) (map[string]bool, error) {
	homeDir, _ := homedir.Dir()
	listItems, err := ioutil.ReadDir(homeDir + KEK_DIR + locale)
	list := make(map[string]bool)

	if err != nil {
		return list, err
	}

	if limit < 1 {
		limit = len(listItems)
	}

	for ind := 0; ind < limit; ind++  {
		if ind == len(listItems) {
			break
		}
		list[listItems[ind].Name()] = true
	}

	return list, err
}