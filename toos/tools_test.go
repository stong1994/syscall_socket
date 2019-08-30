package tools

import (
	"net"
	"testing"
)

func TestString2IPV4(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name    string
		args    args
		want    net.IP
		wantErr error
	}{
		{
			"test_nil",
			args{"1,2,3,4"},
			nil,
			ErrIPNotValid,
		},
		{
			"normal",
			args{"255.255.255.255"},
			net.ParseIP("255.255.255.255"),
			nil,
		},
	}

	checkIP := func(want, got net.IP) bool {
		if len(want) != len(got) {
			return false
		}
		for i, w := range want {
			if w != got[i] {
				return false
			}
		}
		return true
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := String2IPV4(tt.args.ip)
			if err  != tt.wantErr {
				t.Errorf("String2IPV4() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// todo
			if !checkIP(got, tt.want) {
				t.Errorf("String2IPV4() = %v, want %v", got, tt.want)
			}
		})
	}
}
