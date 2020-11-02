// author: Beitong Tian

// time: 09/19/2020

package helper

import (
	. "structs"
	"testing"
)

var nilMembership []Membership
var emptyMembership = make([]Membership, 0)
var simpleMembership []Membership = []Membership{{"1", 1, 1}}
var complexMembership []Membership = []Membership{{"1", 1, 1},
	{"cs425", 2, 3},
	{"mp1", 3, 1111111111111},
}

func Test_PrintMembershipListAsTable(t *testing.T) {
	// args' structure
	type args struct {
		membershipList []Membership
	}
	// test wrapper's struct
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test nil", args{nilMembership}, true},
		{"test empty", args{emptyMembership}, false},
		{"test simple", args{simpleMembership}, false},
		{"test complex", args{complexMembership}, false},
	}
	for _, tt := range tests {
		// call test functions with test id (name) and our to be tested function
		t.Run(tt.name, func(t *testing.T) {
			err := PrintMembershipListAsTable(tt.args.membershipList)
			// check nil input error
			if (err != nil) != tt.wantErr {
				t.Errorf("Error actual = %v, but expected = %v\n", err, tt.wantErr)
			}
		})
	}
}
