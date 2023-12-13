package utils

import (
	"reflect"
	"testing"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

func TestBuildStoragePoolName(t *testing.T) {
	type args struct {
		poolClass string
		poolType  string
	}
	var poolClass1 = apisv1alpha1.DiskClassNameHDD
	var poolType = apisv1alpha1.PoolTypeRegular

	var poolClass2 = apisv1alpha1.DiskClassNameSSD
	var poolClass3 = apisv1alpha1.DiskClassNameNVMe
	var poolClass4 = ""

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			args:    args{poolClass: poolClass1, poolType: poolType},
			want:    apisv1alpha1.PoolNameForHDD,
			wantErr: false,
		},
		{
			args:    args{poolClass: poolClass2, poolType: poolType},
			want:    apisv1alpha1.PoolNameForSSD,
			wantErr: false,
		},
		{
			args:    args{poolClass: poolClass3, poolType: poolType},
			want:    apisv1alpha1.PoolNameForNVMe,
			wantErr: false,
		},
		{
			args:    args{poolClass: poolClass4, poolType: poolType},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildStoragePoolName(tt.args.poolClass)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildStoragePoolName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BuildStoragePoolName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertBytesToStr(t *testing.T) {
	type args struct {
		size int64
	}
	var size = int64(1024)
	var size1 = int64(0)
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{size: size},
			want: "1024B",
		},
		{
			args: args{size: size1},
			want: "0B",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertBytesToStr(tt.args.size); got != tt.want {
				t.Errorf("ConvertBytesToStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBytes(t *testing.T) {
	type args struct {
		sizeStr string
	}
	var sizeStr = "1024"
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			args: args{sizeStr: sizeStr},
			want: 1024,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBytes(tt.args.sizeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveStringItem(t *testing.T) {
	type args struct {
		items        []string
		itemToDelete string
	}
	var items = []string{"apple", "banana", "orange"}
	var itemToDelete = "orange"
	var want = []string{"apple", "banana"}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			args: args{items: items, itemToDelete: itemToDelete},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveStringItem(tt.args.items, tt.args.itemToDelete); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveStringItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeName(t *testing.T) {
	type args struct {
		name string
	}
	var name = "abcde"
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{name: name},
			want: "abcde",
		},
		{
			args: args{name: "a.b.c.d"},
			want: "a-b-c-d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeName(tt.args.name); got != tt.want {
				t.Errorf("SanitizeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPCIDiskInfo_IsNVMe(t *testing.T) {
	type fields struct {
		isNVMe   bool
		pciName  string
		diskName string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
		{
			fields: fields{
				isNVMe: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &PCIDiskInfo{
				isNVMe:   tt.fields.isNVMe,
				pciName:  tt.fields.pciName,
				diskName: tt.fields.diskName,
			}
			if got := d.IsNVMe(); got != tt.want {
				t.Errorf("IsNVMe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddUniqueStringItem1(t *testing.T) {
	type args struct {
		items     []string
		itemToAdd string
	}
	var items = []string{"apple", "banana", "orange"}
	var itemToAdd = "orange"
	var want = []string{"apple", "banana", "orange"}
	var items1 = []string{"apple", "banana", "orange"}
	var itemToAdd1 = "apple"
	var want1 = []string{"apple", "banana", "orange"}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
		{
			args: args{items: items, itemToAdd: itemToAdd},
			want: want,
		},
		{
			args: args{items: items1, itemToAdd: itemToAdd1},
			want: want1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddUniqueStringItem(tt.args.items, tt.args.itemToAdd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddUniqueStringItem() = %v, want %v", got, tt.want)
			}
		})
	}
}
