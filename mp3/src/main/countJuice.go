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
	key := args[0]
	file := args[1]
	dest := args[2]
	// open destination file
	destFile, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)	
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
	sum := 0
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Fields(text)
		value := fields[1]
		intValue,_ := strconv.Atoi(value)
		sum = sum + intValue
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	output := key + "\t" + fmt.Sprint(sum) + "\n"
	destFile.Write([]byte(output))
	
	


	
}
