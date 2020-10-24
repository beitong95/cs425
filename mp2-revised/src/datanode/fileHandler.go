package datanode

import (
	"constant"
	"fmt"
	"io/ioutil"
	"os"
)

func CreateFile() {
	err := os.Mkdir(constant.Dir, 0777)
	fmt.Println(err)
	FileList = []string{}
}

func Get(fileName string) ([]byte, string) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		return data, "Not Found"
	}
	return data, "Found"
}

func Put(fileName string, buf []byte) {
	if _, err := os.Stat(constant.Dir); os.IsNotExist(err) {
		// File does not exist
		CreateFile()
	}
	var path = constant.Dir + "/" + fileName
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// filename exists
		Delete(fileName)
	}
	err := ioutil.WriteFile(path, buf, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	FileList = append(FileList, fileName)
}

func remove(filename string) []string {
	for i, file := range FileList {
		if file == filename {
			if i == len(FileList)-1 {
				return FileList[:i]
			}
			return append(FileList[:i], FileList[i+1:]...)
		}
	}
	return FileList
}

func Delete(fileName string) {
	var path = constant.Dir + "/" + fileName
	err := os.Remove(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	FileList = remove(fileName)
}

func List() []string {
	var c, err = ioutil.ReadDir(constant.Dir)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var output []string
	for _, entry := range c {
		output = append(output, entry.Name())
	}
	return output
}
