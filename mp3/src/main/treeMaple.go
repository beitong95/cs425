package main

import (
	"fmt"
	"os"
	"strconv"
	"bufio"
	"strings"
)

func main() {
	args := os.Args[1:]
	file := args[0]
	dest := args[1]
	// open destination file
	destFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 0644)	
	if err != nil {
		panic(err)
	}
	defer destFile.Close()

	// process file
	f, err := os.Open(file) 
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Fields(text)
		if len(fields) == 1 {
			continue
		}
		fmt.Println(fields)
		key := fields[0]
		value := fields[1]
		intValue,_ := strconv.Atoi(value)
		newValue := fmt.Sprint(intValue) 
		output := key + "\t" + newValue + "\n"
		destFile.Write([]byte(output))
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	
}
