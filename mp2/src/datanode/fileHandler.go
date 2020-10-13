package datanode

import (
	"fmt"
	"io/ioutil"
	"os"
)

func CreateFile() {
	err := os.Mkdir("../File", 0755)
	fmt.Println(err)
}

func Get(fileName string) ([]byte, string) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		return data,"Not Found"
	}
	return data,"Found"
}

func Put(fileName string, buf []byte) {
	if _, err := os.Stat("../File"); os.IsNotExist(err) {
		// File does not exist
		CreateFile()
	}
	var path = "../File/" + fileName
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// filename exists
		Delete(fileName)
	}
	err := ioutil.WriteFile(path, buf, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func Delete(fileName string) {
	var path = "../File/" + fileName
	err := os.Remove(path) 
    if err != nil { 
        fmt.Println(err)
    }
}

func List() []string {
	var c, err = ioutil.ReadDir("../File")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var output []string
	for _, entry := range c {
        output = append(output,entry.Name())
	}
	return output
}