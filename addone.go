package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]
	key := args[0]
	value := args[1]
	i, _:= strconv.Atoi(value)
	i = i + 1
	fmt.Println(key + "\t" + fmt.Sprint(i))

	
}
