package healths

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"testing"
)

func Test_smartCtlr_IsDiskHealthy(t *testing.T) {
	var devPath string = "/dev/sdb"
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockDiskChecker(ctrl)
	m.
		EXPECT().
		IsDiskHealthy(devPath).
		Return(false, nil).
		Times(1)

	v, err := m.IsDiskHealthy(devPath)
	fmt.Printf("Test_smartCtlr_IsDiskHealthy v = %+v, err = %+v", v, err)
	if err != nil {
		t.Fatal()
	}
}

func Test_smartCtlr_GetLocalDisksAll(t *testing.T) {
	var deviceInfos []DeviceInfo
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockDiskChecker(ctrl)
	m.
		EXPECT().
		GetLocalDisksAll().
		Return(deviceInfos, nil).
		Times(1)

	v, err := m.GetLocalDisksAll()
	fmt.Printf("Test_smartCtlr_GetLocalDisksAll v = %+v, err = %+v", v, err)
	if err != nil {
		t.Fatal()
	}
}

func Test_smartCtlr_CheckHealthForLocalDisk(t *testing.T) {
	var device = &DeviceInfo{}
	device.Name = "sdb"
	device.Type = "HDD"
	device.InfoName = "TEST"
	device.Protocol = "ISCSI"
	checkResult := &DiskCheckResult{}
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()
	m := NewMockDiskChecker(ctrl)
	m.
		EXPECT().
		CheckHealthForLocalDisk(device).
		Return(checkResult, nil).
		Times(1)

	v, err := m.CheckHealthForLocalDisk(device)
	fmt.Printf("Test_smartCtlr_CheckHealthForLocalDisk v = %+v, err = %+v", v, err)
	if err != nil {
		t.Fatal()
	}
}