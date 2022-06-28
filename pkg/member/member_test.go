package member

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/runtime"

	localapis "github.com/hwameistor/local-storage/pkg/apis"
	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	memmock "github.com/hwameistor/local-storage/pkg/member/controller"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestMember(t *testing.T) {
	tests := []struct {
		name string
		want localapis.LocalStorageMember
	}{
		{
			want: &localStorageMember{},
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
			want: &localStorageMember{},
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
	var systemConfig apisv1alpha1.SystemConfig
	var cli client.Client
	var informersCache cache.Cache

	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := memmock.NewMockLocalStorageMember(ctrl)
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

	m := memmock.NewMockLocalStorageMember(ctrl)
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

	m := memmock.NewMockLocalStorageMember(ctrl)
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

	m := memmock.NewMockLocalStorageMember(ctrl)
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

	m := memmock.NewMockLocalStorageMember(ctrl)
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

	m := memmock.NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		Run(stopCh).
		Return().
		Times(1)

	m.Run(stopCh)

	fmt.Printf("Test_localStorageMember_Run ends")
}

func Test_localStorageMember_Controller(t *testing.T) {
	var cm localapis.ControllerManager
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := memmock.NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		Controller().
		Return(cm).
		Times(1)

	cm = m.Controller()

	fmt.Printf("Test_localStorageMember_Controller cm = %v", cm)
}

func Test_localStorageMember_Node(t *testing.T) {
	var nm localapis.NodeManager
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := memmock.NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		Node().
		Return(nm).
		Times(1)

	nm = m.Node()

	fmt.Printf("Test_localStorageMember_Node cm = %v", nm)
}

func Test_localStorageMember_Name(t *testing.T) {
	var str string
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := memmock.NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		Name().
		Return(str).
		Times(1)

	str = m.Name()

	fmt.Printf("Test_localStorageMember_Name str = %v", str)
}

func Test_localStorageMember_Version(t *testing.T) {
	var version string
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := memmock.NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		Version().
		Return(version).
		Times(1)

	version = m.Version()

	fmt.Printf("Test_localStorageMember_Version version = %v", version)
}

func Test_localStorageMember_DriverName(t *testing.T) {
	var driverName string
	// 创建gomock控制器，用来记录后续的操作信息
	ctrl := gomock.NewController(t)
	// 断言期望的方法都被执行
	// Go1.14+的单测中不再需要手动调用该方法
	defer ctrl.Finish()

	m := memmock.NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		DriverName().
		Return(driverName).
		Times(1)

	driverName = m.DriverName()

	fmt.Printf("Test_localStorageMember_DriverName driverName = %v", driverName)
}
