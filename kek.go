package main

import (
	"fmt"
	"os/exec"
	"log"
	"strings"
	"os"
	"io/ioutil"
	"encoding/json"
	"github.com/satori/go.uuid"
	"github.com/revel/modules/csrf/app"
	"time"
	"bytes"
	"encoding/gob"
)

type ContentType struct {
	Fields map[string]map[string]string
}

func main() {
	out, err := exec.Command("date").Output()
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

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The date is %s\n", out)
}

func Get(args []string) {
	tp := args[1]

	switch tp {
	case "collection":
		fmt.Printf("Content is : %s\n", GetCollection(args[2]))
	}
}

func Init(args []string) {
	os.Mkdir(".kek", 0755)
	conf := make(map[string]string)
	kekname, _ := csrf.RandomString(9)
	conf["kekname"] = kekname
	conf["kuid"] = uuid.NewV4().String()

	for k, v := range args {
		if v == "-name" {
			conf["kekname"] = args[k+1]
		} else if v == "-owner.email" {
			conf["owner.email"] = args[k+1]
		} else if v == "-owner.phone" {
			conf["owner.phone"] = args[k+1]
		}else if v == "-owner.first_name" {
			conf["owner.first_name"] = args[k+1]
		} else if v == "-owner.last_name" {
			conf["owner.last_name"] = args[k+1]
		} else if v == "-company.name" {
			conf["company.name"] = args[k+1]
		} else if v == "-company.email" {
			conf["company.email"] = args[k+1]
		} else if v == "-company.phone" {
			conf["company.phone"] = args[k+1]
		}
	}

	cdata, e := json.Marshal(conf)
	log.Print(e)
	ioutil.WriteFile(".kek/config", cdata, 0755)
}

func Add(args []string) {
	contents := make(map[string]interface{})

	for k, v :=range args{
		if strings.Contains(v, "-file") {
			f := args[k + 1]
			cnts, _ :=ioutil.ReadFile(f)
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
	ct := ContentType{}
	log.Print(name, details)
	fileCont := make([]byte, 0)
	cdata := GetConfContent()
	kcMap := make(map[interface{}]interface{})
	kcData, _ := ioutil.ReadFile(".kek/classmap")


	for k, arg := range details {
		if arg == "-file" {
			fileArg := k+1
			fileCont, _ := ioutil.ReadFile(details[fileArg])
			json.Unmarshal(fileCont, ct.Fields)
		}
	}
	tp := uuid.NewV5(uuid.NamespaceDNS, cdata["kuid"] + "/" + name)
	kcMap[name] = kcMap
	kcb, _ := GetBytes(kcMap)
	kcData = append(kcData, kcb...)
	os.Mkdir(".kek/" + tp.String(), 0755)
	ioutil.WriteFile(".kek/classmap", kcData, 0755)
	ioutil.WriteFile(".kek/types/" + tp.String(), fileCont, 0755)
}

func AddItem(name string, cType string, data map[string]interface{}) {
	cdata := GetConfContent()
	log.Print(name, cType, data)
	tpUid := uuid.NewV5(uuid.NamespaceDNS, cdata["kuid"] + "/" + cType)
	cuid := uuid.NewV5(tpUid, name)
	data["_kekspace"] = cdata["kuid"]
	data["_kekclass"] = tpUid.String()
	data["_created_date"] = time.Now()
	data["_updated_date"] = time.Now()
	data["_kuid"] = cuid
	js, _ := json.Marshal(data)
	log.Print(data)
	log.Print(string(js[:]))
	ioutil.WriteFile(".kek/" + tpUid.String() + "/" + cuid.String(), js, 0755)
}

func GetBytes(key map[interface]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GetCollection(kClass string) []string  {
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

func Run(command string, args ...string) {
	var cmd *exec.Cmd
	if len(args) == 0 {
		fmt.Printf("run: %s\n", command)
		args := strings.Split(command, " ")
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		fmt.Printf("run: %s %s\n", command, strings.Join(args, " "))
		cmd = exec.Command(command, args...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("command failed: %s\n", command)
		panic(err)
	}
}

