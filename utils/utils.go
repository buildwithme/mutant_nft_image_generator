package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func EnsureDir(dirName string) error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	return err
}

func ReadAll(fileNamePath string) []byte {
	inputFile, errFile := os.Open(fileNamePath)
	if errFile != nil {
		Fatal(errFile)
	}

	body, errReadFile := ioutil.ReadAll(inputFile)
	if errReadFile != nil {
		Fatal(errReadFile)
	}

	return body
}

func PrintJson(input interface{}) {
	body, err := json.Marshal(input)
	if err != nil {
		Fatal(err)
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "\t")
	if err != nil {
		Fatal(err)
	}
	fmt.Printf(string(prettyJSON.Bytes()))
}

func GetExtension(name string) string {

	index := strings.Index(name, ".") + 1

	return name[index:]
}


func FleExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func ExistIn(name string, names[]string) bool {
	for _, nameValue := range names {
		if name == nameValue {
			return true
		}
	}
	return false
}
