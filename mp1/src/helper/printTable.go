// author: Beitong Tian
// time: 09/19/2020
package helper

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	. "structs"
)

func PrintMembershipListAsTable(membershipList []Membership) error {
	// handle edge cases
	if membershipList == nil {
		return errors.New("Passing nil to printMembershipListAsTable")
	}

	e := reflect.ValueOf(Membership{})
	tableWidth := e.NumField()
	tableHeight := len(membershipList)
	maxL := make([]int, tableWidth)

	//get table header info
	s1 := make([]interface{}, tableWidth)
	for i := 0; i < tableWidth; i++ {
		attrName := e.Type().Field(i).Name
		s1[i] = attrName
		maxL[i] = len(attrName)
	}

	// get table body info
	s3 := make([][]interface{}, tableHeight)
	for i := 0; i < tableHeight; i++ {
		e := reflect.ValueOf(membershipList[i])
		s3[i] = make([]interface{}, tableWidth)
		for j := 0; j < tableWidth; j++ {
			s3[i][j] = e.Field(j).Interface()
			str := fmt.Sprintln(s3[i][j])
			if l := len(str); l > maxL[j] {
				maxL[j] = l
			}
		}
	}
	//fmt.Println("max length ", maxL)

	// get table border info
	tableWidthByCharacter := 0
	pad := 3
	for i := 0; i < tableWidth; i++ {
		tableWidthByCharacter += maxL[i]
	}
	tableWidthByCharacter += pad*tableWidth + (1*tableWidth + 1)

	// create print format command
	printCommand := ""
	for i := 0; i < tableWidth; i++ {
		printCommand = printCommand + "|%-" + fmt.Sprintf("%v", maxL[i]+3) + "v"
	}
	printCommand = printCommand + "|\n"
	//fmt.Printf("%#v\n", printCommand)

	// print border
	border := strings.Repeat("-", tableWidthByCharacter)
	fmt.Printf("%v\n", border)

	// print header
	fmt.Printf(printCommand, s1...)
	s2 := make([]interface{}, tableWidth)
	for i := 0; i < tableWidth; i++ {
		s2[i] = strings.Repeat("-", maxL[i]+3)
	}
	fmt.Printf(printCommand, s2...)

	// print body
	for i := 0; i < tableHeight; i++ {
		fmt.Printf(printCommand, s3[i]...)
	}

	// print border
	fmt.Printf("%v\n", border)

	return nil
}
