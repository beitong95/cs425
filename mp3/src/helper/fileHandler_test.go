package helper

import (
	"reflect"
	"testing"
)

func TestHashPartition(t *testing.T) {
	type args struct {
		filename       string
		partitionCount uint64
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{"test1", args{"111.txt", 3, "666"}, []string{"666_0.txt", "666_1.txt", "666_2.txt"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashPartition(tt.args.filename, tt.args.partitionCount, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPartition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HashPartition() = %v, want %v", got, tt.want)
			}
		})
	}
}
