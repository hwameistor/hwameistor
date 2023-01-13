package csi

import (
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
)

func Test_server_GracefulStop(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockServer(ctrl)
	m.
		EXPECT().
		GracefulStop().
		Return().
		Times(1)

	m.GracefulStop()
}

func Test_server_Init(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var endpoint = "127.0.0.1"

	m := NewMockServer(ctrl)
	m.
		EXPECT().
		Init(endpoint).
		Return().
		Times(1)

	m.Init(endpoint)
}

func Test_server_Run(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var ids csi.IdentityServer
	var cs csi.ControllerServer
	var ns csi.NodeServer

	m := NewMockServer(ctrl)
	m.
		EXPECT().
		Run(ids, cs, ns).
		Return().
		Times(1)

	m.Run(ids, cs, ns)
}

func Test_server_Stop(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockServer(ctrl)
	m.
		EXPECT().
		Stop().
		Return().
		Times(1)

	m.Stop()
}
