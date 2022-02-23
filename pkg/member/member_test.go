package member

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"

	localapis "github.com/hwameistor/local-storage/pkg/apis"
	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestMember(t *testing.T) {
	tests := []struct {
		name string
		want localapis.LocalStorageMember
	}{
		{
			want:  &localStorageMember{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Member(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Member() = %v, want %v", got, tt.want)
			}
			fmt.Printf("TestMember Member()= %+v", Member())
		})
	}
}

func Test_newMember(t *testing.T) {
	tests := []struct {
		name string
		want localapis.LocalStorageMember
	}{
		{
			want:  &localStorageMember{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newMember(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMember() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_localStorageMember_ConfigureBase(t *testing.T) {
	var name string = "test2"
	var namespace string = "test"
	var systemConfig localstoragev1alpha1.SystemConfig
	var cli client.Client
	var informersCache cache.Cache


	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		ConfigureBase(name, namespace, systemConfig, cli, informersCache).
		Return(m).
		Times(1)

	v := m.ConfigureBase(name, namespace, systemConfig, cli, informersCache)

	fmt.Printf("Test_localStorageMember_ConfigureBase v= %+v", v)
}

func Test_localStorageMember_ConfigureNode(t *testing.T) {
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		ConfigureNode().
		Return(m).
		Times(1)

	v := m.ConfigureNode()

	fmt.Printf("Test_localStorageMember_ConfigureNode v= %+v", v)
}

func Test_localStorageMember_ConfigureController(t *testing.T) {
	var scheme = &runtime.Scheme{}
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		ConfigureController(scheme).
		Return(m).
		Times(1)

	v := m.ConfigureController(scheme)

	fmt.Printf("Test_localStorageMember_ConfigureController v= %+v", v)
}

func Test_localStorageMember_ConfigureCSIDriver(t *testing.T) {
	var driverName string = "test_driver"
	var sockAddr string = "1.1.1.1"
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		ConfigureCSIDriver(driverName, sockAddr).
		Return(m).
		Times(1)

	v := m.ConfigureCSIDriver(driverName, sockAddr)

	fmt.Printf("Test_localStorageMember_ConfigureCSIDriver v= %+v", v)
}

func Test_localStorageMember_ConfigureRESTServer(t *testing.T) {
	var httpPort int = 8080

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		ConfigureRESTServer(httpPort).
		Return(m).
		Times(1)

	v := m.ConfigureRESTServer(httpPort)

	fmt.Printf("Test_localStorageMember_ConfigureRESTServer v= %+v", v)
}

func Test_localStorageMember_Run(t *testing.T) {
	var stopCh <-chan struct{}
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		Run(stopCh).
		Return().
		Times(1)

	m.Run(stopCh)

	fmt.Printf("Test_localStorageMember_Run ends")
}
