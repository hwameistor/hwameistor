package node

import (
	"fmt"
	"testing"

	ldmv1alpha1 "github.com/cherry-io/local-disk-manager/pkg/apis/cherry/v1alpha1"
	"k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClientset "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_manager_getLocalDiskByName(t *testing.T) {
	//type fields struct {
	//	name                    string
	//	namespace               string
	//	apiClient               client.Client
	//	informersCache          cache.Cache
	//	replicaRecords          map[string]string
	//	storageMgr              *storage.LocalManager
	//	healthCheckQueue        *common.TaskQueue
	//	diskEventQueue          *diskmonitor.EventQueue
	//	volumeTaskQueue         *common.TaskQueue
	//	volumeReplicaTaskQueue  *common.TaskQueue
	//	localDiskClaimTaskQueue *common.TaskQueue
	//	localDiskTaskQueue      *common.TaskQueue
	//	configManager           *configManager
	//	logger                  *log.Entry
	//}
	//type args struct {
	//	localDiskName string
	//	nameSpace     string
	//}
	//tests := []struct {
	//	name    string
	//	fields  fields
	//	args    args
	//	want    *ldmv1alpha1.LocalDisk
	//	wantErr bool
	//}{
	//	// TODO: Add test cases.
	//	{
	//		name: "",
	//		//fields: manager,
	//		args: args{localDiskName: "k8s-node1-sdb", nameSpace: "local-disk-manager-system"},
	//		want: CreateFakeWant(t),
	//	},
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		m := &manager{
	//			name:                    tt.fields.name,
	//			namespace:               tt.fields.namespace,
	//			apiClient:               CreateFakeClient(t),
	//			informersCache:          tt.fields.informersCache,
	//			replicaRecords:          tt.fields.replicaRecords,
	//			storageMgr:              tt.fields.storageMgr,
	//			healthCheckQueue:        tt.fields.healthCheckQueue,
	//			diskEventQueue:          tt.fields.diskEventQueue,
	//			volumeTaskQueue:         tt.fields.volumeTaskQueue,
	//			volumeReplicaTaskQueue:  tt.fields.volumeReplicaTaskQueue,
	//			localDiskClaimTaskQueue: tt.fields.localDiskClaimTaskQueue,
	//			localDiskTaskQueue:      tt.fields.localDiskTaskQueue,
	//			configManager:           tt.fields.configManager,
	//			logger:                  log.WithField("Module", "NodeManager"),
	//		}
	//		got, err := m.getLocalDiskByName(tt.args.localDiskName, tt.args.nameSpace)
	//		if (err != nil) != tt.wantErr {
	//			t.Errorf("getLocalDiskByName() error = %v, wantErr %v", err, tt.wantErr)
	//			return
	//		}
	//
	//		if got == nil {
	//			t.Errorf("getLocalDiskByName() got nil, got = %v", got)
	//			return
	//		}
	//
	//		if !reflect.DeepEqual(got.Name, tt.want.Name) {
	//			t.Errorf("getLocalDiskByName() got = %v, want %v", got, tt.want)
	//		}
	//	})
	//}
}

func CreateFakeClient(t *testing.T) client.Client {

	fakeClient := fakeClientset.NewFakeClient()
	if fakeClient == nil {
		fmt.Println("FakeClient is not created")
	}
	return fakeClient
}

func CreateFakeLocalDiskClaim(t *testing.T) *ldmv1alpha1.LocalDiskClaim {

	var claim = &ldmv1alpha1.LocalDiskClaim{}
	claim.Name = "node-localdiskclaim-1"
	claim.Namespace = "local-disk-manager-system"
	claim.Spec.NodeName = "k8s-node1"

	claim.Spec.Description.DiskType = "HDD"

	diskRefs := []*v1.ObjectReference{}
	diskRefs1 := &v1.ObjectReference{}
	diskRefs1.Name = "k8s-node1-sdb"
	diskRefs = append(diskRefs, diskRefs1)
	claim.Spec.DiskRefs = diskRefs

	claim.Status.Status = ldmv1alpha1.LocalDiskClaimStatusBound

	return claim
}

func CreateFakeWant(t *testing.T) *ldmv1alpha1.LocalDisk {
	var ld = &ldmv1alpha1.LocalDisk{}
	ld.Name = "k8s-node1-sdb"
	return ld
}
