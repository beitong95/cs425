// author: Beitong Tian
// time: 09/19/2020
package printtable

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	. "structs"
	"text/tabwriter"
)

func findLongestAttribute(m Membership) int {
	e := reflect.ValueOf(m)
	tableWidth := e.NumField()
	maxL := 0
	for i := 0; i < tableWidth; i++ {
		if len(e.Type().Field(i).Name) > maxL {
			maxL = len(e.Type().Field(i).Name)
		}
	}
	return maxL
}

func printMembershipListAsTable(membershipList []Membership) error {
	// handle edge cases
	if membershipList == nil {
		return errors.New("Passing nil to printMembershipListAsTable")
	}

	// initiallize print commands
	printCommand1 := "" // header format
	printCommand2 := "" // header separator format
	printCommand3 := "" // real values format
	maxL := 0

	// create header format and find the max length of attribute name
	e := reflect.ValueOf(Membership{})
	tableWidth := e.NumField()
	for i := 0; i < tableWidth; i++ {
		printCommand1 = printCommand1 + e.Type().Field(i).Name + "\t"
		// find max length of attribute name
		if l := len(e.Type().Field(i).Name); l > maxL {
			maxL = l
		}
	}

	// create table obj w
	w := new(tabwriter.Writer)
	cellWidth := maxL + 5
	w.Init(os.Stdout, cellWidth, 8, 0, ' ', 0)
	defer w.Flush()

	//print header, header separator and build the format for real values
	fmt.Fprintln(w, printCommand1)
	subPrintCommand2 := strings.Repeat("-", cellWidth-1) + "\t"
	printCommand2 = strings.Repeat(subPrintCommand2, tableWidth)
	fmt.Fprintln(w, printCommand2)
	printCommand3 = strings.Repeat("%v\t", tableWidth)
	printCommand3 = printCommand3 + "\n"

	//print real values
	for i := 0; i < len(membershipList); i++ {
		e := reflect.ValueOf(membershipList[i])
		s := make([]interface{}, tableWidth)
		for i := 0; i < tableWidth; i++ {
			s[i] = e.Field(i).Interface()
		}
		fmt.Fprintf(w, printCommand3, s...)
	}
	return nil
}
