package csi

import (
	"testing"

	"github.com/golang/mock/gomock"
)

func Test_plugin_Run(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	var stopCh <-chan struct{}

	m := NewMockDriver(ctrl)
	m.
		EXPECT().
		Run(stopCh).
		Return().
		Times(1)

	m.Run(stopCh)
}
