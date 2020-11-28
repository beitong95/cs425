package main

import (
	"os"
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
		key := fields[0]
		votes := strings.Split(key,",")
		var A = ""
		var B = ""
		var C = ""
		var D = ""
		for _,vote := range votes {
			var candidate = strings.Split(vote,".")
			if candidate[0] == "A" {
				A = candidate[1]
			} else if candidate[0] == "B" {
				B = candidate[1]
			} else if candidate[0] == "C" {
				C = candidate[1]
			} else {
				D = candidate[1]
			}
		}
		var output = ""
		if A < B {
			output += "A,B\t1\n"
		} else {
			output += "B,A\t1\n"
		}
		if A < C {
			output += "A,C\t1\n"
		} else {
			output += "C,A\t1\n"
		}
		if A < D {
			output += "A,D\t1\n"
		} else {
			output += "D,A\t1\n"
		}
		if B < C {
			output += "B,C\t1\n"
		} else {
			output += "C,B\t1\n"
		}
		if B < D {
			output += "B,D\t1\n"
		} else {
			output += "D,B\t1\n"
		}
		if C < D {
			output += "C,D\t1\n"
		} else {
			output += "D,C\t1\n"
		}
		destFile.Write([]byte(output))
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	
}
