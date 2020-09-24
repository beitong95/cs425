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

/**
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
**/
func PrintMembershipListAsTable(membershipList map[string]Membership) error {
	// handle edge cases
	if membershipList == nil {
		return errors.New("Passing nil to printMembershipListAsTable")
	}

	e := reflect.ValueOf(Membership{})
	membershipAttributeCount := e.NumField()
	//	fmt.Println(membershipAttributeCount)
	// 1 is ID
	tableWidth := membershipAttributeCount + 1
	//	fmt.Println(tableWidth)
	tableHeight := len(membershipList)
	//	fmt.Println(tableHeight)
	maxL := make([]int, tableWidth)

	//get table header info
	s1 := make([]interface{}, tableWidth)
	keyName := "ID"
	s1[0] = keyName
	maxL[0] = len(keyName)
	for i := 1; i < tableWidth; i++ {
		attrName := e.Type().Field(i - 1).Name
		s1[i] = attrName
		maxL[i] = len(attrName)
	}

	// get table body info
	s3 := make([][]interface{}, tableHeight)
	i := 0
	for k, _ := range membershipList {
		e := reflect.ValueOf(membershipList[k])
		s3[i] = make([]interface{}, tableWidth)
		s3[i][0] = k
		if l := len(k); l > maxL[0] {
			maxL[0] = l
		}
		for j := 1; j < tableWidth; j++ {
			s3[i][j] = e.Field(j - 1).Interface()
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
