// author: Beitong Tian

// time: 09/19/2020

package printtable

import (
	. "structs"
	"testing"
)

var nilMembership []Membership
var emptyMembership = make([]Membership, 0)
var simpleMembership []Membership = []Membership{{"1", 1, 1}}
var complexMembership []Membership = []Membership{{"1", 1, 1},
	{"cs425", 1.123, 1.123},
	{"mp1", 1.123, 1.123},
}

func Test_printMembershipListAsTable(t *testing.T) {
	// args' structure
	type args struct {
		membershipList []Membership
	}
	// test wrapper's struct
	tests := []struct {
		name string
		args args
	}{
		{"test nil", args{nilMembership}},
		{"test empty", args{emptyMembership}},
		{"test simple", args{simpleMembership}},
		{"test complex", args{complexMembership}},
	}
	for _, tt := range tests {
		// call test functions with test id (name) and our to be tested function
		t.Run(tt.name, func(t *testing.T) {
			err := printMembershipListAsTable(tt.args.membershipList)
			// check nil input error
			if tt.name == "test nil" {
				expected := "Passing nil to printMembershipListAsTable"
				if err.Error() != expected {
					t.Errorf("Error actual = %v, but expected = %v\n", err, expected)
				}
			}

		})
	}
}
