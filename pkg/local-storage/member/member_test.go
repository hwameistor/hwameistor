package member

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	memmock "github.com/hwameistor/hwameistor/pkg/local-storage/member/controller"
)

func TestMember(t *testing.T) {
	tests := []struct {
		name string
		want apis.LocalStorageMember
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
		want apis.LocalStorageMember
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
	var name = "test2"
	var hostName = "node1"
	var namespace = "test"
	var systemConfig apisv1alpha1.SystemConfig
	var cli client.Client
	var informersCache cache.Cache
	var recorder record.EventRecorder

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := memmock.NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		ConfigureBase(name, hostName, namespace, systemConfig, cli, informersCache, recorder).
		Return(m).
		Times(1)

	v := m.ConfigureBase(name, hostName, namespace, systemConfig, cli, informersCache, recorder)

	fmt.Printf("Test_localStorageMember_ConfigureBase v= %+v", v)
}

func Test_localStorageMember_ConfigureNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := memmock.NewMockLocalStorageMember(ctrl)
	m.
		EXPECT().
		ConfigureNode(nil).
		Return(m).
		Times(1)

	v := m.ConfigureNode(nil)

	fmt.Printf("Test_localStorageMember_ConfigureNode v= %+v", v)
}

func Test_localStorageMember_ConfigureController(t *testing.T) {
	var scheme = &runtime.Scheme{}
	ctrl := gomock.NewController(t)
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
	var driverName = "test_driver"
	var sockAddr = "1.1.1.1"
	ctrl := gomock.NewController(t)
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
	var httpPort = 8080

	ctrl := gomock.NewController(t)
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
	ctrl := gomock.NewController(t)
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
	var cm apis.ControllerManager
	ctrl := gomock.NewController(t)
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
	var nm apis.NodeManager
	ctrl := gomock.NewController(t)
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
	ctrl := gomock.NewController(t)
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
	ctrl := gomock.NewController(t)
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
	ctrl := gomock.NewController(t)
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
