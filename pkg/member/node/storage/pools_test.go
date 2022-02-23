package storage

import (
	"fmt"
	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	"github.com/golang/mock/gomock"
	"testing"
)

func Test_localPoolManager_ExtendPoolsInfo(t *testing.T) {
	var localDiskM = map[string]*localstoragev1alpha1.LocalDisk{}
	localDiskM["/dev/sdb"] = &localstoragev1alpha1.LocalDisk{
		DevPath:       "/dev/sdb",
		Class:         localstoragev1alpha1.DiskClassNameHDD,
		CapacityBytes: 10240,
		State:         localstoragev1alpha1.DiskStateAvailable,
	}
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalPoolManager(ctrl)
	m.
		EXPECT().
		ExtendPoolsInfo(localDiskM).
		Return(nil, nil).
		Times(1)

	v, err := m.ExtendPoolsInfo(localDiskM)
	fmt.Printf("Test_localPoolManager_ExtendPoolsInfo err = %+v", err)
	if err != nil {
		t.Fatal()
	}
	fmt.Printf("Test_localPoolManager_ExtendPoolsInfo v= %+v", v)
}

func Test_getPoolClassTypeByName(t *testing.T) {
	type args struct {
		poolName string
	}
	tests := []struct {
		name          string
		args          args
		wantPoolClass string
		wantPoolType  string
	}{
		{
			args:          args{poolName: localstoragev1alpha1.PoolNameForHDD},
			wantPoolClass: localstoragev1alpha1.DiskClassNameHDD,
			wantPoolType:  localstoragev1alpha1.PoolTypeRegular,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPoolClass, gotPoolType := getPoolClassTypeByName(tt.args.poolName)
			if gotPoolClass != tt.wantPoolClass {
				t.Errorf("getPoolClassTypeByName() gotPoolClass = %v, want %v", gotPoolClass, tt.wantPoolClass)
			}
			if gotPoolType != tt.wantPoolType {
				t.Errorf("getPoolClassTypeByName() gotPoolType = %v, want %v", gotPoolType, tt.wantPoolType)
			}
		})
	}
}

func Test_getPoolNameAccordingDisk(t *testing.T) {
	type args struct {
		disk *localstoragev1alpha1.LocalDisk
	}
	var disk = &localstoragev1alpha1.LocalDisk{
		DevPath:       "/dev/sdb",
		Class:         localstoragev1alpha1.DiskClassNameHDD,
		CapacityBytes: 10240,
		State:         localstoragev1alpha1.DiskStateAvailable,
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			args:    args{disk: disk},
			want:    localstoragev1alpha1.PoolNameForHDD,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPoolNameAccordingDisk(tt.args.disk)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPoolNameAccordingDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getPoolNameAccordingDisk() got = %v, want %v", got, tt.want)
			}
		})
	}
}
