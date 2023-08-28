package storage

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"

	apisv1alpha "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

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
			args:          args{poolName: apisv1alpha.PoolNameForHDD},
			wantPoolClass: apisv1alpha.DiskClassNameHDD,
			wantPoolType:  apisv1alpha.PoolTypeRegular,
		},
		{
			args:          args{poolName: apisv1alpha.PoolNameForSSD},
			wantPoolClass: apisv1alpha.DiskClassNameSSD,
			wantPoolType:  apisv1alpha.PoolTypeRegular,
		},
		{
			args:          args{poolName: apisv1alpha.PoolNameForNVMe},
			wantPoolClass: apisv1alpha.DiskClassNameNVMe,
			wantPoolType:  apisv1alpha.PoolTypeRegular,
		},
		{
			args:          args{poolName: "Unknown"},
			wantPoolClass: "",
			wantPoolType:  "",
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
		disk *apisv1alpha.LocalDevice
	}
	var disk = &apisv1alpha.LocalDevice{
		DevPath:       "/dev/sdb",
		Class:         apisv1alpha.DiskClassNameHDD,
		CapacityBytes: 10240,
		State:         apisv1alpha.DiskStateAvailable,
	}
	var disk2 = &apisv1alpha.LocalDevice{
		DevPath:       "/dev/sdb",
		Class:         apisv1alpha.DiskClassNameSSD,
		CapacityBytes: 10240,
		State:         apisv1alpha.DiskStateAvailable,
	}
	var disk3 = &apisv1alpha.LocalDevice{
		DevPath:       "/dev/sdb",
		Class:         apisv1alpha.DiskClassNameNVMe,
		CapacityBytes: 10240,
		State:         apisv1alpha.DiskStateAvailable,
	}
	var disk4 = &apisv1alpha.LocalDevice{
		DevPath:       "/dev/sdb",
		Class:         "Unknown",
		CapacityBytes: 10240,
		State:         apisv1alpha.DiskStateAvailable,
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			args:    args{disk: disk},
			want:    apisv1alpha.PoolNameForHDD,
			wantErr: false,
		},
		{
			args:    args{disk: disk2},
			want:    apisv1alpha.PoolNameForSSD,
			wantErr: false,
		},
		{
			args:    args{disk: disk3},
			want:    apisv1alpha.PoolNameForNVMe,
			wantErr: false,
		},
		{
			args:    args{disk: disk4},
			want:    "",
			wantErr: true,
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

func Test_localPoolManager_GetReplicas(t *testing.T) {
	var localVolumeReplicaM = map[string]*apisv1alpha.LocalVolumeReplica{}
	var localVolumeReplica = &apisv1alpha.LocalVolumeReplica{}
	localVolumeReplica.Spec.VolumeName = "volume1"
	localVolumeReplica.Spec.PoolName = "pool1"
	localVolumeReplica.Spec.NodeName = "node1"
	localVolumeReplica.Spec.RequiredCapacityBytes = 1240
	localVolumeReplica.Name = "test1"
	localVolumeReplicaM["test"] = localVolumeReplica

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalPoolManager(ctrl)
	m.
		EXPECT().
		GetReplicas().
		Return(localVolumeReplicaM, nil).
		Times(1)

	v, err := m.GetReplicas()
	fmt.Printf("Test_localPoolManager_GetReplicas err = %+v", err)
	if err != nil {
		t.Fatal()
	}
	fmt.Printf("Test_localPoolManager_GetReplicas v= %+v", v)
}

func Test_localPoolManager_ExtendPools(t *testing.T) {
	var localDisks = []*apisv1alpha.LocalDevice{}
	var disk = &apisv1alpha.LocalDevice{
		DevPath:       "/dev/sdb",
		Class:         apisv1alpha.DiskClassNameHDD,
		CapacityBytes: 10240,
		State:         apisv1alpha.DiskStateAvailable,
	}
	localDisks = append(localDisks, disk)

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockLocalPoolManager(ctrl)
	m.
		EXPECT().
		ExtendPools(localDisks).
		Return(nil).
		Times(1)

	_, err := m.ExtendPools(localDisks)
	if err != nil {
		t.Fatal()
	}
	fmt.Printf("Test_localPoolManager_ExtendPools v= %+v", err)
}
