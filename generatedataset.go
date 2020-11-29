package main

import(
	"os"
	"math/rand"
)
func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}
func main() {
	var isExist = fileExists("votes.txt")
	if isExist {
		e := os.Remove("votes.txt")
		if e != nil {
			panic(e)
		}
	}
	var destFile, err = os.OpenFile("votes.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil{
		panic(err)
	}
	defer destFile.Close()
	//rand.Seed(42)
	res := []string{"A.1,B.2,C.3,D.4\t1\n","A.1,B.2,C.4,D.3\t1\n","A.1,B.3,C.2,D.4\t1\n","A.1,B.3,C.4,D.2\t1\n","A.1,B.4,C.2,D.3\t1\n","A.1,B.4,C.3,D.2\t1\n","A.2,B.1,C.3,D.4\t1\n","A.2,B.1,C.4,D.3\t1\n","A.2,B.3,C.1,D.4\t1\n","A.2,B.3,C.4,D.1\t1\n","A.2,B.4,C.3,D.1\t1\n","A.2,B.4,C.1,D.3\t1\n","A.3,B.2,C.1,D.4\t1\n","A.3,B.2,C.4,D.1\t1\n","A.3,B.1,C.2,D.4\t1\n","A.3,B.1,C.4,D.2\t1\n","A.3,B.4,C.1,D.2\t1\n","A.3,B.4,C.2,D.1\t1\n","A.4,B.1,C.2,D.3\t1\n","A.4,B.1,C.3,D.2\t1\n","A.4,B.2,C.1,D.3\t1\n","A.4,B.2,C.3,D.1\t1\n","A.4,B.3,C.1,D.2\t1\n","A.4,B.3,C.2,D.1\t1\n"}
	for i := 0; i < 6000000; i++ {
		seed := rand.Intn(len(res))
		output := res[seed]
		destFile.Write([]byte(output))
	}
}
