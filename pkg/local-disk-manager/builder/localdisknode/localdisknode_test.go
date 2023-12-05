package localdisknode

import (
	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	"testing"
	"time"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

func TestNewBuilder(t *testing.T) {
	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				DiskNode: &v1alpha1.LocalDiskNode{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBuilder(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithName(t *testing.T) {
	name := "testName"

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				DiskNode: &v1alpha1.LocalDiskNode{
					ObjectMeta: v1.ObjectMeta{
						Name: name,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.WithName(name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupAttachNode(t *testing.T) {
	attachNode := "testAttachNode"

	tests := []struct {
		name string
		want *Builder
	}{
		{
			want: &Builder{
				DiskNode: &v1alpha1.LocalDiskNode{
					Spec: v1alpha1.LocalDiskNodeSpec{
						NodeName: attachNode,
					},
				},
			},
		},
	}

	builder := NewBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := builder.SetupAttachNode(attachNode); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupAttachNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name    string
		builder *Builder
		want    *v1alpha1.LocalDiskNode
	}{
		{
			builder: &Builder{
				DiskNode: &v1alpha1.LocalDiskNode{},
			},
			want: &v1alpha1.LocalDiskNode{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.builder.Build(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Build() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	fakeLocalDiskNodeName        = "local-disk-node-example"
	fakeNodename                 = "10-6-118-10"
	fakeFreeDiskCount            = int64(1)
	fakeTotalDiskCount           = int64(1)
	fakePoolClass                = "HDD"
	fakePoolType                 = "REGULAR"
	fakeLocalDiskNodeUID         = "local-disk-node-uid"
	fakeTotalCapacityBytes int64 = 10 * 1024 * 1024 * 1024
	fakeFreeCapacityBytes  int64 = 10 * 1024 * 1024 * 1024

	LocalDiskNodeKind   = "LocalDiskNode"
	LocalDiskKind       = "LocalDisk"
	LeaseKind           = "Lease"
	LocalDiskClaimKind  = "LocalDiskClaim"
	LocalDiskVolumeKind = "LocalDiskVolume"

	apiversion    = "hwameistor.io/v1alpha1"
	fakeRecorder  = record.NewFakeRecorder(100)
	fakeTimestamp = time.Now()
)

func CreateFakeKubeClient() (*Kubeclient, error) {
	kubeclient := Kubeclient{}
	clientset := fake.NewSimpleClientset()
	kubeclient.SetClient(clientset)
	return &kubeclient, nil
}

func GenFakeLocalDiskNodeObject() *v1alpha1.LocalDiskNode {
	ldn := &v1alpha1.LocalDiskNode{}

	typeMeta := v1.TypeMeta{
		Kind:       LocalDiskNodeKind,
		APIVersion: apiversion,
	}

	objectMeta := v1.ObjectMeta{
		Name:              fakeLocalDiskNodeName,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalDiskNodeUID),
		CreationTimestamp: v1.Time{fakeTimestamp},
	}

	spec := v1alpha1.LocalDiskNodeSpec{
		NodeName: fakeLocalDiskNodeName,
	}

	status := v1alpha1.LocalDiskNodeStatus{
		TotalCapacity: fakeTotalCapacityBytes,
		FreeCapacity:  fakeFreeCapacityBytes,
		FreeDisk:      fakeFreeDiskCount,
		TotalDisk:     fakeTotalDiskCount,
	}

	ldn.Spec = spec
	ldn.TypeMeta = typeMeta
	ldn.ObjectMeta = objectMeta
	ldn.Status = status

	return ldn
}

func Test_KubeClient_Create(t *testing.T) {
	testcases := []struct {
		Description    string
		DiskNode       *v1alpha1.LocalDiskNode
		ExpectDiskNode *v1alpha1.LocalDiskNode
	}{
		// TODO: Add More test case
		{
			Description:    "It is a create localDiskNode test.",
			DiskNode:       GenFakeLocalDiskNodeObject(),
			ExpectDiskNode: GenFakeLocalDiskNodeObject(),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.Description, func(t *testing.T) {
			client, err := CreateFakeKubeClient()
			if err != nil {
				t.Fatal("create a fake client failed")
			}
			_, err = client.Create(testcase.DiskNode)
			if err != nil {
				t.Fatal("create a localDiskNode on fake client failed")
			}
			ldn, err := client.Get(testcase.DiskNode.Name)
			if err != nil {
				t.Fatal("get a specified localDiskNode on fake client failed")
			}
			if !reflect.DeepEqual(ldn, testcase.ExpectDiskNode) {
				t.Fatal("get a localDiskNode not same as expectDiskNode")
			}
		})
	}
}

func Test_KubeClient_Patch(t *testing.T) {
	type args struct {
		nodeName string
	}
	testcases := []struct {
		Description    string
		DiskNode       *v1alpha1.LocalDiskNode
		modifyArgs     args
		ExpectNodeName string
	}{
		// TODO: Add More test case
		{
			Description: "It is a patch localDiskNode test.",
			DiskNode:    GenFakeLocalDiskNodeObject(),
			modifyArgs: args{
				nodeName: fakeNodename,
			},
			ExpectNodeName: fakeNodename,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.Description, func(t *testing.T) {
			client, err := CreateFakeKubeClient()
			if err != nil {
				t.Fatal("create a fake client failed")
			}
			// ensure the source ldn exists in storage
			_, err = client.Create(testcase.DiskNode)
			if err != nil {
				t.Fatal("create a localDiskNode on fake client failed")
			}
			ldn := testcase.DiskNode.DeepCopy()
			ldn.Spec.NodeName = testcase.modifyArgs.nodeName
			err = client.Patch(testcase.DiskNode, ldn)
			if err != nil {
				// the origin object must exist in cluster
				t.Fatal("patch a localDiskNode on fake client failed")
			}
			newLdn, err := client.Get(testcase.DiskNode.Name)
			if err != nil {
				t.Fatal("get a specified localDiskNode on fake client failed")
			}
			if !reflect.DeepEqual(testcase.ExpectNodeName, newLdn.Spec.NodeName) {
				t.Fatal("get a localDiskNode not same as expectDiskNode")
			}
		})
	}
}
