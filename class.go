package main

import (
	"io/ioutil"
	"log"
	"github.com/MoonBabyLabs/revchain"
	"github.com/satori/go.uuid"
	"encoding/json"
	"strings"
	"time"
	"strconv"
)

const CLASS_DIR = ".kek/classes/"
const KEK_DIR = ".kek/"

type Kekclass struct {
	Id uuid.UUID
	Slug	  string
	Fields    []KekField
	KekDocs   []KekDoc
	KekDocLocations []string
}

func (kc Kekclass) GenerateSlug(content string) string {
	strings.Replace(content, " ", "-", -1)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	fs, _ := ioutil.ReadDir(CLASS_DIR)
	fCounts := 0

	for _, f := range fs {
		if strings.Contains(f.Name(), content) {
			fCounts++
		}
	}

	content = content + "." + ts

	if fCounts > 0 {
		content = content + strconv.Itoa(fCounts)
	}

	return content
}

func NewClass(name string, fields []KekField) (Kekclass, error) {
	kc := Kekclass{}
	kc.Fields = fields
	kc.Slug = kc.GenerateSlug()
	kc.Id = uuid.NewV4()
	kcd, marshallErr := json.Marshal(kc)

	if marshallErr != nil {
		return kc, marshallErr
	}

	return kc, nil
}

func GetClass(id uuid.UUID, withDocs bool, withDocRevs bool) (Kekclass, error) {
	kclass := Kekclass{}
	file, err := ioutil.ReadFile(CLASS_DIR + id.String())

	if err != nil {
		return kclass, err
	}

	json.Unmarshal(file, kclass)

	if withDocs {
		kclass.KekDocs = make([]KekDoc, len(kclass.KekDocLocations))

		for k, docLoc := range kclass.KekDocLocations {
			kdoc := KekDoc{}
			kFile, err := ioutil.ReadFile(KEK_DIR + "/docs/" + docLoc)

			if err != nil {
				log.Fatal(err)
			}

			json.Unmarshal(kFile, kdoc)

			if withDocRevs {
				rData := make([]revchain.Block, 0)
				rFile, reverr := ioutil.ReadFile(KEK_DIR + "/docs/" + docLoc + ".rev")

				if reverr != nil {
					log.Fatal(reverr)
				}

				json.Unmarshal(rFile, rData)
				kdoc.Blocks = rData
			}

			kclass.KekDocs[k] = kdoc
		}
	}

	return kclass, nil
}