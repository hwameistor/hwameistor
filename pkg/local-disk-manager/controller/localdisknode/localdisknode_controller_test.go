package localdisknode

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/tools/record"

	"github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	fakeLocalDiskClaimName       = "local-disk-claim-example"
	fakeLocalDiskClaimUID        = "local-disk-claim-example-uid"
	fakeLocalDiskName            = "local-disk-example"
	localDiskUID                 = "local-disk-example-uid"
	fakeNamespace                = "local-disk-manager-test"
	fakeNodename                 = "10-6-118-10"
	diskTypeHDD                  = "HDD"
	fakedevPath                  = "/dev/fake-sda"
	devType                      = "disk"
	vendorVMware                 = "VMware"
	proSCSI                      = "scsi"
	apiversion                   = "hwameistor.io/v1alpha1"
	localDiskKind                = "LocalDisk"
	localDiskNodeKind            = "LocalDiskNode"
	localDiskClaimKind           = "LocalDiskClaim"
	cap100G                int64 = 100 * 1024 * 1024 * 1024
	cap10G                 int64 = 10 * 1024 * 1024 * 1024
	fakeRecorder                 = record.NewFakeRecorder(100)
)

func TestReconcileLocalDiskNode_Reconcile(t *testing.T) {
	//cli, s := CreateFakeClientWithManager()
	cli, s := CreateFakeClient()

	// reconcile object
	r := ReconcileLocalDiskNode{
		client:   cli,
		scheme:   s,
		Recorder: fakeRecorder,
	}

	// create a group disk with same nodeName and same state
	// fakedevPath is random value
	generateGroupFreeDisk := func(nodeName string, state v1alpha1.LocalDiskState, count int) []*v1alpha1.LocalDisk {
		var fakedisks []*v1alpha1.LocalDisk
		for i := 0; i < count; i++ {
			time.Sleep(time.Nanosecond * 10)
			devPath := time.Now().Format(time.RFC3339Nano)

			disk := GenFakeLocalDiskObject()
			disk.Name = nodeName + "-" + devPath
			disk.Spec.NodeName = nodeName
			disk.Spec.DevicePath = devPath
			disk.Status.State = state

			fakedisks = append(fakedisks, disk)
		}
		return fakedisks
	}

	generateFreeDiskNode := func(nodeName string) *v1alpha1.LocalDiskNode {
		ldn := GenFakeLocalDiskNodeObject()
		ldn.Spec.AttachNode = nodeName
		ldn.SetName(nodeName)
		return ldn
	}

	createLocalDisks := func(cli client.Client, disks []*v1alpha1.LocalDisk) error {
		for _, disk := range disks {
			if err := cli.Create(context.Background(), disk); err != nil {
				return err
			}
		}
		return nil
	}

	createLocalDiskNode := func(cli client.Client, diskNode *v1alpha1.LocalDiskNode) error {
		return cli.Create(context.Background(), diskNode)
	}

	type createResource func(cli client.Client, resource interface{}) error
	type cleanResource func(cli client.Client, resource ...interface{})
	createLocalDiskResource := func(cli client.Client, resource interface{}) error {
		switch resource.(type) {
		case *v1alpha1.LocalDiskNode:
			return createLocalDiskNode(cli, resource.(*v1alpha1.LocalDiskNode))
		case []*v1alpha1.LocalDisk:
			return createLocalDisks(cli, resource.([]*v1alpha1.LocalDisk))
		default:
			return fmt.Errorf("unknown resource type")
		}
	}
	cleanLocalDiskResource := func(cli client.Client, resources ...interface{}) {
		for _, resource := range resources {
			switch resource.(type) {
			case []*v1alpha1.LocalDisk:
				for _, obj := range resource.([]*v1alpha1.LocalDisk) {
					if err := cli.Delete(context.Background(), obj); err != nil {
						fmt.Printf("Failed to delete disk %v\n", obj.GetName())
					} else {
						fmt.Printf("Succeed to delete disk %v\n", obj.GetName())
					}
				}
			default:
				_ = cli.Delete(context.Background(), resource.(*v1alpha1.LocalDiskNode))
			}
		}
	}

	testCases := []struct {
		description       string
		preReconcile      createResource
		postReconcile     cleanResource
		freeDisks         []*v1alpha1.LocalDisk
		freeNode          *v1alpha1.LocalDiskNode
		wantFreeDiskCount int64
	}{
		{
			description:   "Same node, free disks 1",
			preReconcile:  createLocalDiskResource,
			postReconcile: cleanLocalDiskResource,
			freeNode:      generateFreeDiskNode("node1"),
			freeDisks: append(
				generateGroupFreeDisk("node1", v1alpha1.LocalDiskAvailable, 1),
			),
			wantFreeDiskCount: 1,
		},
		{
			description:   "Different node, free disks 2",
			preReconcile:  createLocalDiskResource,
			postReconcile: cleanLocalDiskResource,
			freeNode:      generateFreeDiskNode("node2"),
			freeDisks: append(
				generateGroupFreeDisk("node1", v1alpha1.LocalDiskAvailable, 3),
				generateGroupFreeDisk("node2", v1alpha1.LocalDiskAvailable, 2)...,
			),
			wantFreeDiskCount: 2,
		},
		{
			description:   "Same node, free disks 2",
			preReconcile:  createLocalDiskResource,
			postReconcile: cleanLocalDiskResource,
			freeNode:      generateFreeDiskNode("node1"),
			freeDisks: append(
				generateGroupFreeDisk("node1", v1alpha1.LocalDiskBound, 3),
				generateGroupFreeDisk("node1", v1alpha1.LocalDiskAvailable, 2)...,
			),
			wantFreeDiskCount: 2,
		},
		{
			description:   "Different node, free disks 0",
			preReconcile:  createLocalDiskResource,
			postReconcile: cleanLocalDiskResource,
			freeNode:      generateFreeDiskNode("node1"),
			freeDisks: append(
				generateGroupFreeDisk("node1", v1alpha1.LocalDiskBound, 3),
				generateGroupFreeDisk("node2", v1alpha1.LocalDiskAvailable, 2)...,
			),
			wantFreeDiskCount: 0,
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.description, func(t *testing.T) {
			// create free disks
			err := testcase.preReconcile(r.client, testcase.freeDisks)
			if err != nil {
				t.Errorf("failed to create LocalDisks, err: %v", err)
			}

			// clean resources
			defer testcase.postReconcile(r.client, testcase.freeDisks, testcase.freeNode)

			// create disk node
			err = testcase.preReconcile(r.client, testcase.freeNode)
			if err != nil {
				t.Errorf("failed to create LocalDiskNode, err: %v", err)
			}

			// create reconcile request for LocalDiskNode
			request := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: testcase.freeNode.GetNamespace(), Name: testcase.freeNode.GetName()}}
			_, err = r.Reconcile(context.TODO(), request)
			if err != nil {
				t.Errorf("failed to reconcile, err: %v", err)
			}

			// refresh LocalDiskNode
			if err = r.client.Get(context.Background(), request.NamespacedName, testcase.freeNode); err != nil {
				t.Errorf("failed to refresh LocalDiskNode, err: %v", err)
			}
			refreshNodeDisks(testcase.freeNode, testcase.freeNode.Spec.AttachNode)

			if testcase.freeNode.Status.AllocatableDisk != testcase.wantFreeDiskCount {
				t.Errorf("Expected AllocatableDisk %d but actual get %d", testcase.wantFreeDiskCount, testcase.freeNode.Status.AllocatableDisk)
			}
		})
	}
}

// CreateFakeClient Create localDisk and LocalDiskNode resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDisk{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDiskList{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDiskNode{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.LocalDiskNodeList{})
	return fake.NewFakeClientWithScheme(s, &v1alpha1.LocalDisk{}, &v1alpha1.LocalDiskNode{}), s
}

// GenFakeLocalDiskObject Create disk
func GenFakeLocalDiskObject() *v1alpha1.LocalDisk {
	ld := &v1alpha1.LocalDisk{}

	TypeMeta := metav1.TypeMeta{
		Kind:       localDiskKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeNodename + fakedevPath,
		Namespace:         fakeNamespace,
		UID:               types.UID(localDiskUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := v1alpha1.LocalDiskSpec{
		NodeName:     fakeNodename,
		DevicePath:   fakedevPath,
		Capacity:     cap100G,
		HasPartition: false,
		HasRAID:      false,
		RAIDInfo:     v1alpha1.RAIDInfo{},
		HasSmartInfo: false,
		SmartInfo:    v1alpha1.SmartInfo{},
		DiskAttributes: v1alpha1.DiskAttributes{
			Type:     diskTypeHDD,
			DevType:  devType,
			Vendor:   vendorVMware,
			Protocol: proSCSI,
		},
		State: v1alpha1.LocalDiskActive,
	}

	Status := v1alpha1.LocalDiskStatus{State: v1alpha1.LocalDiskAvailable}

	ld.TypeMeta = TypeMeta
	ld.ObjectMeta = ObjectMata
	ld.Spec = Spec
	ld.Status = Status
	return ld
}

func GenFakeLocalDiskNodeObject() *v1alpha1.LocalDiskNode {
	ldn := &v1alpha1.LocalDiskNode{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1alpha1.LocalDiskNodeSpec{
			AttachNode: fakeNodename,
		},
	}
	return ldn
}

func refreshNodeDisks(node *v1alpha1.LocalDiskNode, wantNode string) {
	node.Status.TotalDisk = 0
	node.Status.AllocatableDisk = 0
	for name, disk := range node.Status.Disks {
		if strings.HasPrefix(name, wantNode) {
			node.Status.TotalDisk++
			if disk.Status == string(v1alpha1.LocalDiskAvailable) {
				node.Status.AllocatableDisk++
			}
		}
	}
}
