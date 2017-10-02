package main

import (
	"github.com/satori/go.uuid"
	"io/ioutil"
	"errors"
	"encoding/json"
	"github.com/revel/modules/csrf/app"
	"time"
	"strconv"
)

// The current space that occupies a document. Spaces can contain many different kek content items and class indexes.
// Many users can access and contribute to a single kekspace. A version control repository is a fair comparison.
type Kekspace struct {
	Owners  []Contact
	Name    string
	Id      uuid.UUID
	Classes []Kekclass
}

type KekspaceConfig struct {
	Owners []Contact
	Name string
}

const KEK_LOC = ".kek/space"

func (ks Kekspace) load() (Kekspace, error) {
	config, err := ioutil.ReadFile(KEK_LOC)

	if err != nil {
		spaceEr := errors.New("No kekspace initialized. Need to init a space first")

		return ks, spaceEr
	}

	umK := json.Unmarshal(config, ks)

	if umK != nil {
		return ks, umK
	}

	return ks, nil
}

func (ks Kekspace) New(config KekspaceConfig) (Kekspace, error) {
	ks.Owners = make([]Contact, len(config.Owners))

	if len(config.Name) == 0 {
		name, csrfError := csrf.RandomString(8)

		if csrfError != nil {
			return ks, csrfError
		}

		t := time.Now().Unix()
		ti := strconv.FormatInt(t, 10)
		ks.Name = name + ti
	}

	ksd, me := json.Marshal(ks)

	if me != nil {
		return ks, nil
	}

	ioutil.WriteFile(KEK_LOC, ksd, 0755)

	return ks, nil
}