package service

import (
	"reflect"
	"testing"
)

func Test_parseCmds(t *testing.T) {
	type args struct {
		cmds []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"simpleTest1", args{[]int{1}}, []int{1}},
		{"simpleTest2", args{[]int{2}}, []int{2}},
		{"simpleTest3", args{[]int{3}}, []int{3}},
		{"simpleTest4", args{[]int{4}}, []int{}},
		{"simpleTest5", args{[]int{1, 2}}, []int{2}},
		{"simpleTest6", args{[]int{2, 1}}, []int{1}},
		{"simpleTest7", args{[]int{3, 4}}, []int{3}},
		{"simpleTest8", args{[]int{4, 3}}, []int{3}},
		{"simpleTest9", args{[]int{1, 2, 1, 2, 3, 4, 4, 3}}, []int{2, 3}},
	}
	for _, tt := range tests {
		if tt.name == "simpleTest7" {
		} else {
			isJoin = false
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := parseCmds(tt.args.cmds); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseCmds() = %v, want %v", got, tt.want)
			}
		})
	}
}
