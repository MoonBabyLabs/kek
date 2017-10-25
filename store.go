package kek

import (
	"strings"
	"github.com/mitchellh/go-homedir"
	"os"
	"io/ioutil"
	"encoding/json"
)

type Storer interface {
	Load(location string, unmarshallStruct interface{}) error
	Save(locale string, item interface{}) error
	Delete(locale string) error
	List(locale string) (map[string]bool, error)
}

type Store struct {

}

func (s Store) Save(locale string, content interface{}) error {
	marshallData, err := json.Marshal(content)

	if err != nil {
		return err
	}

	pathFilename := strings.Split(locale, "/")
	path := pathFilename[:len(pathFilename) - 1]
	pathString := strings.Join(path, "/")
	homeDir, _ := homedir.Dir()
	os.MkdirAll(homeDir + KEK_DIR + pathString, 0755)

	return ioutil.WriteFile(homeDir + KEK_DIR + locale, marshallData, 0755)
}

func (s Store) Load(locale string, unmarshallStruct interface{}) error {
	homeDir, _ := homedir.Dir()
	file, readErr := ioutil.ReadFile(homeDir + KEK_DIR + locale)

	if readErr != nil {
		return readErr
	}

	json.Unmarshal(file, &unmarshallStruct)

	return nil
}

func (s Store) Delete(locale string) error {
	homeDir, _ := homedir.Dir()
	err := os.Remove(homeDir + KEK_DIR + locale)

	return err
}


func (s Store) List(locale string) (map[string]bool, error) {
	homeDir, _ := homedir.Dir()
	listItems, err := ioutil.ReadDir(homeDir + KEK_DIR + locale)
	list := make(map[string]bool)
	total := len(listItems)

	if err != nil {
		return list, err
	}

	for i := 0; i < total; i++  {
		list[listItems[i].Name()] = true
	}

	return list, err
}