// author: Beitong Tian

// time: 09/19/2020

package helper

import (
	. "structs"
	"testing"
)

var testMap1 map[string]Membership = make(map[string]Membership)

func TestPrintMembershipListAsTable(t *testing.T) {
	testMap1["11111"] = Membership{1111, 1111}
	testMap1["11112"] = Membership{1111, 1111}
	type args struct {
		membershipList map[string]Membership
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"simpleTest1", args{testMap1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PrintMembershipListAsTable(tt.args.membershipList); (err != nil) != tt.wantErr {
				t.Errorf("PrintMembershipListAsTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
