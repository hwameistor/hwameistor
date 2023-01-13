package csi

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
)

func Test_linuxMounter_GetDeviceMountPoints(t *testing.T) {

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var devPath string = "/dev/sdb"
	var res []string

	m := NewMockMounter(ctrl)
	m.
		EXPECT().
		GetDeviceMountPoints(devPath).
		Return(res).
		Times(1)

	v := m.GetDeviceMountPoints(devPath)

	fmt.Printf("Test_linuxMounter_GetDeviceMountPoints v= %+v", v)

}

func Test_linuxMounter_Unmount(t *testing.T) {

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var mountPoint string = "/dev/tmp"

	m := NewMockMounter(ctrl)
	m.
		EXPECT().
		Unmount(mountPoint).
		Return(nil).
		Times(1)

	v := m.Unmount(mountPoint)

	fmt.Printf("Test_linuxMounter_Unmount v= %+v", v)

}

func Test_linuxMounter_BindMount(t *testing.T) {

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var mountPoint string = "/dev/tmp"
	var devPath string = "/dev/sdb"

	m := NewMockMounter(ctrl)
	var err error
	m.
		EXPECT().
		BindMount(devPath, mountPoint).
		Return(err).
		Times(1)

	v := m.BindMount(devPath, mountPoint)

	fmt.Printf("Test_linuxMounter_BindMount v= %+v", v)

}

func Test_linuxMounter_MountRawBlock(t *testing.T) {

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var mountPoint string = "/dev/tmp"
	var devPath string = "/dev/sdb"

	var err error

	m := NewMockMounter(ctrl)
	m.
		EXPECT().
		MountRawBlock(devPath, mountPoint).
		Return(err).
		Times(1)

	v := m.MountRawBlock(devPath, mountPoint)

	fmt.Printf("Test_linuxMounter_MountRawBlock v= %+v", v)

}

func Test_linuxMounter_FormatAndMount(t *testing.T) {

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var mountPoint string = "/dev/tmp"
	var devPath string = "/dev/sdb"
	var fsType string = "xfs"
	var options []string

	m := NewMockMounter(ctrl)
	m.
		EXPECT().
		FormatAndMount(devPath, mountPoint, fsType, options).
		Return(nil).
		Times(1)

	v := m.FormatAndMount(devPath, mountPoint, fsType, options)

	fmt.Printf("Test_linuxMounter_FormatAndMount v= %+v", v)

}
