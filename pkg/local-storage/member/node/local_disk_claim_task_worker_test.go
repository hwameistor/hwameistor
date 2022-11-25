package node

import (
	"context"
	"reflect"
	"testing"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/diskmonitor"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/storage"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Test_manager_getLocalDiskByName(t *testing.T) {
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
		localDiskName string
		nameSpace     string
	}
	client, _ := CreateFakeClient()
	// Create localDisk
	ld := GenFakeLocalDiskObject()
	ld.Name = fakeLocalDiskName
	err := client.Create(context.Background(), ld)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *apisv1alpha1.LocalDisk
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				localDiskName: fakeLocalDiskName,
				nameSpace:     fakeNamespace,
			},
			want:    ld,
			wantErr: false,
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
			got, err := m.getLocalDiskByName(tt.args.localDiskName, tt.args.nameSpace)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLocalDiskByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("getLocalDiskByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_getLocalDisksByDiskRefs(t *testing.T) {
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
		localDiskNames []string
		nameSpace      string
	}

	client, _ := CreateFakeClient()
	// Create localDisk
	ld := GenFakeLocalDiskObject()
	ld.Name = fakeLocalDiskName
	err := client.Create(context.Background(), ld)
	if err != nil {
		t.Errorf("Create LocalVolumeConvert fail %v", err)
	}
	var want []*apisv1alpha1.LocalDisk

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*apisv1alpha1.LocalDisk
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				localDiskNames: fakeLocalDiskNames,
				nameSpace:      fakeNamespace,
			},
			want:    want,
			wantErr: false,
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
			got, err := m.getLocalDisksByDiskRefs(tt.args.localDiskNames, tt.args.nameSpace)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLocalDisksByDiskRefs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("getLocalDisksByDiskRefs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_getLocalDisksByLocalDiskClaim(t *testing.T) {
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
		ldc *apisv1alpha1.LocalDiskClaim
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
		want    []*apisv1alpha1.LocalDevice
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				ldc: ldc,
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
			got, err := m.getLocalDisksByLocalDiskClaim(tt.args.ldc)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLocalDisksByLocalDiskClaim() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("getLocalDisksByLocalDiskClaim() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_getLocalDisksMapByLocalDiskClaim(t *testing.T) {
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
		ldc *apisv1alpha1.LocalDiskClaim
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
		want    map[string]*apisv1alpha1.LocalDevice
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				ldc: ldc,
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
			got, err := m.getLocalDisksMapByLocalDiskClaim(tt.args.ldc)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLocalDisksMapByLocalDiskClaim() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("getLocalDisksMapByLocalDiskClaim() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_listAllAvailableLocalDisksByLocalClaimDisk(t *testing.T) {
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
		ldc *apisv1alpha1.LocalDiskClaim
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
		want    []*apisv1alpha1.LocalDisk
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				ldc: ldc,
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
			got, err := m.listAllAvailableLocalDisksByLocalClaimDisk(tt.args.ldc)
			if (err != nil) != tt.wantErr {
				t.Errorf("listAllAvailableLocalDisksByLocalClaimDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("listAllAvailableLocalDisksByLocalClaimDisk() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_listAllInUseLocalDisksByLocalClaimDisk(t *testing.T) {
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
		ldc *apisv1alpha1.LocalDiskClaim
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
		want    []*apisv1alpha1.LocalDisk
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				ldc: ldc,
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
			got, err := m.listAllInUseLocalDisksByLocalClaimDisk(tt.args.ldc)
			if (err != nil) != tt.wantErr {
				t.Errorf("listAllInUseLocalDisksByLocalClaimDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("listAllInUseLocalDisksByLocalClaimDisk() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_listLocalDisksByLocalDiskClaim(t *testing.T) {
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
		ldc *apisv1alpha1.LocalDiskClaim
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
		want    []*apisv1alpha1.LocalDisk
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				ldc: ldc,
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
			got, err := m.listLocalDisksByLocalDiskClaim(tt.args.ldc)
			if (err != nil) != tt.wantErr {
				t.Errorf("listLocalDisksByLocalDiskClaim() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("listLocalDisksByLocalDiskClaim() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manager_processLocalDiskClaim(t *testing.T) {
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
			if err := m.processLocalDiskClaim(tt.args.localDiskNameSpacedName); (err != nil) != tt.wantErr {
				t.Errorf("processLocalDiskClaim() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manager_processLocalDiskClaimBound(t *testing.T) {
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
		claim *apisv1alpha1.LocalDiskClaim
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
		//{
		//	args: args{
		//		claim: ldc,
		//	},
		//},
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
			if err := m.processLocalDiskClaimBound(tt.args.claim); (err != nil) != tt.wantErr {
				t.Errorf("processLocalDiskClaimBound() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
