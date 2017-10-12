package main

import (
	"os"
	"io/ioutil"
	"encoding/json"
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
	FirstName  string
	LastName   string
	Email      string
	Phone      string
	Id         string
	City       string
	Region     string
	PostalCode string
	Country    string
}


func main() {
	//argsWithoutProg := os.Args[1:]
	//fa := argsWithoutProg[0]
}

func Save(kekLocale string, content interface{}) error {
	marshallData, err := json.Marshal(content)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(KEK_DIR + kekLocale, marshallData, 0755)
}

func Delete(locale string) error {
	return os.Remove(KEK_DIR + locale)
}

func Load(locale string, unmarshallStruct interface{}) error {
	file, readErr := ioutil.ReadFile(KEK_DIR + locale)

	if readErr != nil {
		return readErr
	}

	json.Unmarshal(file, unmarshallStruct)

	return nil
}

func List(locale string, limit int) (map[string]bool, error) {
	listItems, err := ioutil.ReadDir(KEK_DIR + locale)
	list := make(map[string]bool)

	if err != nil {
		return list, err
	}

	if limit < 1 {
		limit = len(listItems)
	}

	for ind := 0; ind < limit; ind++  {
		list[listItems[ind].Name()] = true
	}

	return list, err
}