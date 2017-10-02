package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"github.com/MoonBabyLabs/revchain"
	"github.com/revel/modules/csrf/app"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)


type KekField struct {
	Required bool
	Name     string
	Default  interface{}
	Type     string
}

type KekDoc struct {
	Id         uuid.UUID
	Attributes map[string]interface{}
	CreatedAt  time.Time
	UpdatedAt  time.Time
	*revchain.Chain
	Kekspace   Kekspace
	Kekclasses []Kekclass
	Related    []KekDoc
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

func GetKekDoc(id uuid.UUID, withRevs bool) (KekDoc, error) {
	kek := KekDoc{}

	return kek, nil
}


func main() {
	argsWithoutProg := os.Args[1:]
	fa := argsWithoutProg[0]
	log.Print(fa)

	switch fa {
	case "init":
		Init(argsWithoutProg)
		break
	case "add":
		Add(argsWithoutProg)
		break
	case "get":
		log.Print("Get some content")
		Get(argsWithoutProg)
		break
	}
}

func Get(args []string) (interface{}, error) {
	tp := args[1]
	uid, uidErr := uuid.FromString(args[2])

	if uidErr != nil {
		return uid, uidErr
	}

	switch tp {
	case "class":
		kClass, kErr := GetKekClass(uid, true, true)

		return kClass, kErr

		break
	case "doc":
		kDoc, kErr := GetKekDoc(uid, true)

		return kDoc, kErr
		break
	}

	return nil, errors.New("need to follow format: kek get [class or doc]")
}

func Init(args []string) {
	os.Mkdir(".kek", 0755)
	kspace := Kekspace{}
	kekname, _ := csrf.RandomString(9)
	kspace.Name = kekname
	kspace.Id = uuid.NewV4()
	kspace.Owners = make([]Contact, 1)

	for k, v := range args {
		if v == "-name" {
			kspace.Name = args[k+1]
		} else if v == "-owner.email" {
			kspace.Owners[0].Email = args[k+1]
		} else if v == "-owner.phone" {
			kspace.Owners[0].Phone = args[k+1]
		} else if v == "-owner.first_name" {
			kspace.Owners[0].FirstName = args[k+1]
		} else if v == "-owner.last_name" {
			kspace.Owners[0].LastName = args[k+1]
		}
	}

	cdata, _ := json.Marshal(kspace)
	ioutil.WriteFile(".kek/space", cdata, 0755)
}

func Add(args []string) {
	contents := make(map[string]interface{})

	for k, v := range args {
		if strings.Contains(v, "-file") {
			f := args[k+1]
			cnts, _ := ioutil.ReadFile(f)
			json.Unmarshal(cnts, contents)
			log.Print(contents)
		}
	}
	switch args[1] {
	case "class":
		log.Print("Generate new content type")
		AddClass(args[2], args[3:])
	case "item":
		AddItem(args[2], args[3], contents)
		log.Print("Add new content item")
	}
}

func AddClass(name string, details []string) {
	ct := Kekclass{}
	fileCont := make([]byte, 0)
	cdata := GetConfContent()
	kcMap := make(map[interface{}]interface{})
	kcData, _ := ioutil.ReadFile(".kek/classmap")

	for k, arg := range details {
		if arg == "-file" {
			fileArg := k + 1
			fileCont, _ := ioutil.ReadFile(details[fileArg])
			json.Unmarshal(fileCont, ct.Fields)
		}
	}
	tp := uuid.NewV5(uuid.NamespaceDNS, cdata["kuid"]+"/"+name)
	kcMap[name] = kcMap
	kcb, _ := GetBytes(kcMap)
	kcData = append(kcData, kcb...)
	os.Mkdir(".kek/"+tp.String(), 0755)
	ioutil.WriteFile(".kek/classmap", kcData, 0755)
	ioutil.WriteFile(".kek/types/"+tp.String(), fileCont, 0755)
}

func AddItem(name string, cType string, data map[string]interface{}) {
	cdata := GetConfContent()
	log.Print(name, cType, data)
	tpUid := uuid.NewV5(uuid.NamespaceDNS, cdata["kuid"]+"/"+cType)
	cuid := uuid.NewV5(tpUid, name)
	data["_kekspace"] = cdata["kuid"]
	data["_kekclass"] = tpUid.String()
	data["_created_date"] = time.Now()
	data["_updated_date"] = time.Now()
	data["_kuid"] = cuid
	js, _ := json.Marshal(data)
	log.Print(data)
	log.Print(string(js[:]))
	ioutil.WriteFile(".kek/"+tpUid.String()+"/"+cuid.String(), js, 0755)
}

func GetBytes(key map[interface{}]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GetCollection(kClass string) []string {
	keks, _ := ioutil.ReadDir(".kek/" + kClass)
	list := make([]string, len(keks))

	for k, kek := range keks {
		fi, _ := ioutil.ReadFile(".kek/" + kClass + "/" + kek.Name())
		list[k] = string(fi[:])
	}

	return list
}

func GetItems(items []map[string]string) {
	list := make([][]byte, len(items))
	i := 0
	for _, it := range items {
		for k, v := range it {
			fi, _ := ioutil.ReadFile(".kek/" + k + "/" + v)
			list[i] = fi
			i++
		}
	}
}

func GetConfContent() map[string]string {
	cdata := make(map[string]string)
	conf, _ := ioutil.ReadFile(".kek/conf")
	json.Unmarshal(conf, cdata)

	return cdata
}
