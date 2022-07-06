package utils

import (
	"testing"
)

func TestConvertLVMBytesToNumeric(t *testing.T) {
	type args struct {
		lvmbyte string
	}
	var lvmb = "10240B"
	var lvmb2 = "10240"
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			args:    args{lvmbyte: lvmb},
			want:    10240,
			wantErr: false,
		},
		{
			args:    args{lvmbyte: lvmb2},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertLVMBytesToNumeric(tt.args.lvmbyte)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertLVMBytesToNumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertLVMBytesToNumeric() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertNumericToLVMBytes(t *testing.T) {
	type args struct {
		num int64
	}
	var num = int64(4194305)
	var num2 = int64(0)

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{num: num},
			want: "8388608B",
		},
		{
			args: args{num: num2},
			want: "4194304B",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertNumericToLVMBytes(tt.args.num); got != tt.want {
				t.Errorf("ConvertNumericToLVMBytes() = %v, want %v, tt.num = %v", got, tt.want, tt.args.num)
			}
		})
	}
}

func TestNumericToLVMBytes(t *testing.T) {
	type args struct {
		bytes int64
	}
	var num = int64(4194305)
	var num2 = int64(4)

	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			args: args{bytes: num},
			want: 8388608,
		},
		{
			args: args{bytes: num2},
			want: 4194304,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NumericToLVMBytes(tt.args.bytes); got != tt.want {
				t.Errorf("NumericToLVMBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
