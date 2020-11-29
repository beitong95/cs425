package helper

import (
	"fmt"
	"io/ioutil"
	"os"
	. "structs"
	"bufio"
	"strings"
	"github.com/cespare/xxhash"
	"math/rand"
	"time"
)

func CreateFile() {
	os.Mkdir(Dir+"files_"+DatanodeHTTPServerPort, 0777)
	os.Mkdir(Dir+"maplejuicefiles", 0777)
	FileList = []string{}
}

// not use
func Get(fileName string) ([]byte, string) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		return data, "Not Found"
	}
	return data, "Found"
}

// not use
func Put(fileName string, buf []byte) {
	if _, err := os.Stat(Dir); os.IsNotExist(err) {
		// File does not exist
		CreateFile()
	}
	var path = Dir + "/" + fileName
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

// not used
func Delete(fileName string) {
	var path = Dir + "/" + fileName
	err := os.Remove(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	FileList = remove(fileName)
}

func List() []string {
	var c, err = ioutil.ReadDir(Dir + "files_" + DatanodeHTTPServerPort)
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
// range partition
func FastPartition(filename string, partitionCount uint64, id string) ([]string, error) {
	start := time.Now()
	if partitionCount <= 0 {
		panic("Paritioncount is 0")
	}
	file, err := os.Open(filename) 
	if err != nil {
		Logger.Fatal(err)
	}
	// count linenumber
	scanner := bufio.NewScanner(file)
	counter := 0
	for scanner.Scan() {
		counter++
	}
	delta := time.Now().Sub(start).String()
	//Write2Shell("count line time: " +  delta)
	//Write2Shell("line count: " +  fmt.Sprint(counter))
	if err := scanner.Err(); err != nil {
		Logger.Fatal(err)
	}
	file.Close()

	file, err = os.Open(filename) 
	if err != nil {
		Logger.Fatal(err)
	}
	defer file.Close()
	partitionIndex := 0
	partitionTarget := counter / int(partitionCount) 
	bufferIndex := 0
	scanner = bufio.NewScanner(file)
	linecount := 0

	res := []string{}
	filepointerList := []*os.File{}
	for i := 0; i < int(partitionCount); i++ {
		filename := "PartitionRes_" + id + "_" + fmt.Sprint(i)
		res = append(res, filename)
		resfile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)	
		if err != nil {
			Logger.Fatal(err)
		}
		defer resfile.Close()
		filepointerList = append(filepointerList, resfile)
	}

	for scanner.Scan() {
		linecount++
		if linecount % 10000 == 0 {
			//Write2Shell("line count: " + fmt.Sprint(linecount))
		}
		filepointerList[bufferIndex].WriteString(scanner.Text() + "\n")
		partitionIndex++
		if partitionIndex == partitionTarget && bufferIndex != int(partitionCount)-1{
			bufferIndex += 1
			partitionIndex = 0
		}
	}
	if err := scanner.Err(); err != nil {
		Logger.Fatal(err)
	}

	delta = time.Now().Sub(start).String()
	Write2Shell("partitoin time: " +  delta)
	return res, nil

}

/**
Name: HashPartition
Description: Hash partition a file into several small files for map step
Input: filename string, worker count int
Output: new filenames, error
**/
// question: what is the partition key?
// answer: use key + value as the key
// because all key value pairs may have the same key
func HashPartition(filename string, partitionCount uint64, id string) ([]string, error) {
	start := time.Now()
	file, err := os.Open(filename) 
	if err != nil {
		Logger.Fatal(err)
	}
	defer file.Close()

	buffer := make([]string, partitionCount)
	
	scanner := bufio.NewScanner(file)
	counter := 0
	for scanner.Scan() {
		counter++
	}
	delta := time.Now().Sub(start).String()
	Write2Shell("count line time: " +  delta)
	Write2Shell("line count: " +  fmt.Sprint(counter))
	
	for scanner.Scan() {
		counter++
		if counter % 10000 == 0 {
			Write2Shell("Current line: " + fmt.Sprint(counter))
		}
		text := scanner.Text() 
		//Write2Shell(text)
		hash := xxhash.Sum64([]byte(text + fmt.Sprint(rand.Intn(100))))
		//hash := xxhash.Sum64([]byte(text))
		//Write2Shell(fmt.Sprintf("%v", hash))
		partitionIndex := int(hash % partitionCount)
		//Write2Shell(fmt.Sprintf("%v", partitionIndex))
		buffer[partitionIndex] = buffer[partitionIndex] + text + "\n"
	}
	delta = time.Now().Sub(start).String()
	Write2Shell("After scanner time: " +  delta)
	
	if err := scanner.Err(); err != nil {
		Logger.Fatal(err)
	}

	res := []string{}
	for i := range buffer {
		res = append(res, "PartitionRes_" + id + "_" + fmt.Sprint(i))
	}

	for i, s := range buffer {
		err := ioutil.WriteFile(res[i], []byte(s), 0644)
		if err != nil {
			Logger.Fatal(err)
		}
	}
	delta = time.Now().Sub(start).String()
	Write2Shell("partitoin time: " +  delta)
	return res, nil
}

/**
Name: RangePartition
Description: Range partition a file into several small files for map step given limited workers
Input: filename string, worker count int
Output: new filenames, error
**/

func RangePartition(filename string, partitionCount uint64,id string) ([]string, error) {
	file, err := os.Open(filename) 
	if err != nil {
		Logger.Fatal(err)
	}
	defer file.Close()
	res := []string{}
	var keys = []string{}
	buffer := make(map[int][]string)
	temp := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		keys = append(keys,strings.Split(text," ")[0])
		temp[strings.Split(text," ")[0]] = strings.Split(text," ")[1]
	}
	var num = len(keys) / int(partitionCount)
	for i := 0; i < int(partitionCount); i++ {
		if i == int(partitionCount) - 1 {
			for _,key := range keys {
				buffer[i] = append(buffer[i],key + " " + temp[key])
			}
			break
		}
		for _,key := range keys[0:num] {
			buffer[i] = append(buffer[i],key + " " + temp[key])
		}
		keys = keys[num:]
	}
	for i := range buffer {
		res = append(res, id + "_" + fmt.Sprint(i))
	}
	for i, s := range buffer {
		content := strings.Join(s," ")
		err := ioutil.WriteFile(res[i], []byte(content), 0644)
		if err != nil {
			Logger.Fatal(err)
		}
	}
	return res, err
}