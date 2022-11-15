package node

import (
	"context"
	"testing"

	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/diskmonitor"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/storage"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Test_manager_processLocalDisk(t *testing.T) {
	type fields struct {
		name                    string
		namespace               string
		apiClient               client.Client
		informersCache          cache.Cache
		replicaRecords          map[string]string
		storageMgr              *storage.LocalManager
		diskEventQueue          *diskmonitor.EventQueue
		volumeTaskQueue         *common.TaskQueue
		volumeReplicaTaskQueue  *common.TaskQueue
		localDiskClaimTaskQueue *common.TaskQueue
		localDiskTaskQueue      *common.TaskQueue
		configManager           *configManager
		logger                  *log.Entry
	}
	type args struct {
		localDiskNameSpacedName string
	}

	client, _ := CreateFakeClient()
	// Create localDisk
	ld := GenFakeLocalDiskObject()
	ld.Name = fakeLocalDiskName
	err := client.Create(context.Background(), ld)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	// Create LocalDiskClaim
	ldc := GenFakeLocalDiskClaimObject()
	ldc.Name = fakeLocalDiskClaimName
	ldc.Spec.NodeName = fakeNodename
	err = client.Create(context.Background(), ldc)
	if err != nil {
		t.Errorf("Create LocalDiskClaim fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				localDiskNameSpacedName: fakelocalDiskNameSpacedName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				name:                    fakeManagerName,
				namespace:               fakeNamespace,
				apiClient:               client,
				replicaRecords:          map[string]string{},
				volumeTaskQueue:         common.NewTaskQueue("VolumeTask", maxRetries),
				volumeReplicaTaskQueue:  common.NewTaskQueue("VolumeReplicaTask", maxRetries),
				localDiskClaimTaskQueue: common.NewTaskQueue("LocalDiskClaim", maxRetries),
				localDiskTaskQueue:      common.NewTaskQueue("localDisk", maxRetries),
				// healthCheckQueue:        common.NewTaskQueue("HealthCheckTask", maxRetries),
				diskEventQueue: diskmonitor.NewEventQueue("DiskEvents"),
				logger:         log.WithField("Module", "NodeManager"),
			}
			if err := m.processLocalDisk(tt.args.localDiskNameSpacedName); (err != nil) != tt.wantErr {
				t.Errorf("processLocalDisk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
