package service

import (
	"os"
	"io/ioutil"
	"encoding/json"
	"go/build"
	"strings"
)

const KEK_SPACE_CONFIG = "space"
const CLASS_DIR = "classes/"
const DOC_DIR = "d/"
const FIELD_DIR = "f/"
const KEK_DIR = ".kek/"


type KekField struct {
	Required bool
	Name     string
	Default  interface{}
	Type     string
}

type Contact struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Id         string `json:"id"`
	City       string `json:"city"`
	Region     string `json:"region"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

func Save(kekLocale string, content interface{}) error {
	marshallData, err := json.Marshal(content)

	if err != nil {
		return err
	}

	pathFilename := strings.Split(kekLocale, "/")
	path := pathFilename[:len(pathFilename) - 1]
	pathString := strings.Join(path, "/")
	os.MkdirAll(build.Default.GOPATH + "/" + KEK_DIR + pathString, 0755)

	return ioutil.WriteFile(build.Default.GOPATH + "/" + KEK_DIR + kekLocale, marshallData, 0755)
}

func Delete(locale string) error {
	return os.Remove(build.Default.GOPATH + "/" + KEK_DIR + locale)
}

func Load(locale string, unmarshallStruct interface{}) (interface{}, error) {
	file, readErr := ioutil.ReadFile(build.Default.GOPATH + "/" + KEK_DIR + locale)

	if readErr != nil {
		return unmarshallStruct, readErr
	}

	json.Unmarshal(file, unmarshallStruct)

	return unmarshallStruct, nil
}

func List(locale string, limit int) (map[string]bool, error) {
	listItems, err := ioutil.ReadDir(build.Default.GOPATH + "/" + KEK_DIR + locale)
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