package helper

import (
	"fmt"
	"testing"
)

func Test_GetLocalIP(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{"simpleTest", "172.22.156.12", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLocalIP()
			if (err != nil) != tt.wantErr {
				t.Errorf("getLocalIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("IP: ", got)
			if got != tt.want {
				t.Errorf("getLocalIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
