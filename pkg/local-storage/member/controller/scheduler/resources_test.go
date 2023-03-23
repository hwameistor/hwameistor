package scheduler

import (
	"reflect"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

var (
	fakeLocalStorageNodeName       = "local-storage-node-example"
	fakeLocalStorageNodeUID        = "local-storage-node-uid"
	fakeLocalVolumeName            = "local-volume-example"
	fakeLocalVolumeUID             = "local-volume-uid"
	fakeNamespace                  = "local-volume-test"
	fakeNodenames                  = []string{"10-6-118-10"}
	fakeNodename                   = "10-6-118-10"
	fakeStorageIp                  = "10.6.118.11"
	fakeZone                       = "zone-test"
	fakeRegion                     = "region-test"
	fakeVgType                     = "LocalStorage_PoolHDD"
	fakeVgName                     = "vg-test"
	fakeVolName                    = "vol-test"
	fakePoolClass                  = "HDD"
	fakePoolType                   = "REGULAR"
	LocalStorageNodeKind           = "LocalStorageNode"
	fakeTotalCapacityBytes   int64 = 10 * 1024 * 1024 * 1024
	fakeFreeCapacityBytes    int64 = 8 * 1024 * 1024 * 1024
	fakeDiskCapacityBytes    int64 = 2 * 1024 * 1024 * 1024

	apiversion      = "hwameistor.io/v1alpha1"
	LocalVolumeKind = "LocalVolume"
	// fakeRecorder    = record.NewFakeRecorder(100)

	// defaultDRBDStartPort      = 43001
	// defaultHAVolumeTotalCount = 1000
)

func Test_newResources(t *testing.T) {
	type args struct {
		maxHAVolumeCount int
	}
	var resource = &resources{
		logger:               log.WithField("Module", "Scheduler/Resources"),
		allocatedResourceIDs: make(map[string]int),
		freeResourceIDList:   make([]int, 0, 10),
		maxHAVolumeCount:     10,
		allocatedStorages:    newStorageCollection(),
		totalStorages:        newStorageCollection(),
		storageNodes:         map[string]*v1alpha1.LocalStorageNode{},
	}
	tests := []struct {
		name string
		args args
		want *resources
	}{
		// TODO: Add test cases.
		{
			args: args{maxHAVolumeCount: 10},
			want: resource,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newResources(tt.args.maxHAVolumeCount, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resources_Score(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		vol      *v1alpha1.LocalVolume
		nodeName string
	}

	var vol = &v1alpha1.LocalVolume{}
	vol.Name = fakeVolName

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantScore int64
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			args:    args{vol: vol, nodeName: fakeNodename},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			gotScore, err := r.Score(tt.args.vol, tt.args.nodeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Score() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotScore != tt.wantScore {
				t.Errorf("Score() gotScore = %v, want %v", gotScore, tt.wantScore)
			}
		})
	}
}

func Test_resources_addAllocatedStorage(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		vol *v1alpha1.LocalVolume
	}
	var vol = &v1alpha1.LocalVolume{}
	vol.Name = fakeVolName

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args: args{vol: vol},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			r.addAllocatedStorage(tt.args.vol)
		})
	}
}

func Test_resources_addTotalStorage(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		node *v1alpha1.LocalStorageNode
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			r.addTotalStorage(tt.args.node)
		})
	}
}

func Test_resources_allocateResourceID(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		volName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{volName: fakeVolName},
			wantErr: true,
			want:    -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			got, err := r.allocateResourceID(tt.args.volName)
			if (err != nil) != tt.wantErr {
				t.Errorf("allocateResourceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("allocateResourceID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resources_delTotalStorage(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		node *v1alpha1.LocalStorageNode
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			r.delTotalStorage(tt.args.node)
		})
	}
}

func Test_resources_getNodeCandidates(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		vol *v1alpha1.LocalVolume
	}

	var vol = &v1alpha1.LocalVolume{}
	vol.Name = fakeVolName

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*v1alpha1.LocalStorageNode
		wantErr bool
	}{
		// TODO: Add test cases.
		//{
		//	args:    args{vol: vol},
		//	wantErr: true,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			got, err := r.getNodeCandidates(tt.args.vol)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNodeCandidates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNodeCandidates() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resources_getResourceIDForVolume(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		vol *v1alpha1.LocalVolume
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			got, err := r.getResourceIDForVolume(tt.args.vol)
			if (err != nil) != tt.wantErr {
				t.Errorf("getResourceIDForVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getResourceIDForVolume() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resources_handleNodeAdd(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			r.handleNodeAdd(tt.args.obj)
		})
	}
}

func Test_resources_handleNodeDelete(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			r.handleNodeDelete(tt.args.obj)
		})
	}
}

func Test_resources_handleNodeUpdate(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		oldObj interface{}
		newObj interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			r.handleNodeUpdate(tt.args.oldObj, tt.args.newObj)
		})
	}
}

func Test_resources_handleVolumeUpdate(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		oldObj interface{}
		newObj interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			r.handleVolumeUpdate(tt.args.oldObj, tt.args.newObj)
		})
	}
}

func Test_resources_initilizeResources(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	client, _ := CreateFakeClient()
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		{
			fields: fields{apiClient: client},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				logger:               log.WithField("Module", "Scheduler/Resources"),
				allocatedResourceIDs: make(map[string]int),
				freeResourceIDList:   make([]int, 0, 10),
				maxHAVolumeCount:     10,
				allocatedStorages:    newStorageCollection(),
				totalStorages:        newStorageCollection(),
				storageNodes:         map[string]*v1alpha1.LocalStorageNode{},
				apiClient:            client,
			}
			r.initilizeResources()
		})
	}
}

// GenFakeLocalVolumeObject Create lv request
func GenFakeLocalVolumeObject() *v1alpha1.LocalVolume {
	lv := &v1alpha1.LocalVolume{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalVolumeKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalVolumeName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalVolumeUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := v1alpha1.LocalVolumeSpec{
		RequiredCapacityBytes: fakeDiskCapacityBytes,
		ReplicaNumber:         1,
		PoolName:              fakeVgType,
		Delete:                false,
		Convertible:           true,
		Accessibility: v1alpha1.AccessibilityTopology{
			Nodes:   fakeNodenames,
			Regions: []string{fakeRegion},
			Zones:   []string{fakeZone},
		},
		Config: &v1alpha1.VolumeConfig{
			Convertible:           true,
			Initialized:           true,
			ReadyToInitialize:     true,
			RequiredCapacityBytes: fakeDiskCapacityBytes,
			ResourceID:            5,
			Version:               11,
			VolumeName:            fakeLocalVolumeName,
			Replicas: []v1alpha1.VolumeReplica{
				{
					Hostname: fakeNodename,
					ID:       1,
					IP:       fakeStorageIp,
					Primary:  true,
				},
			},
		},
	}

	lv.ObjectMeta = ObjectMata
	lv.TypeMeta = TypeMeta
	lv.Spec = Spec
	lv.Status.State = v1alpha1.VolumeStateCreating
	lv.Status.AllocatedCapacityBytes = fakeTotalCapacityBytes - fakeFreeCapacityBytes
	lv.Status.PublishedNodeName = fakeNodename
	lv.Status.Replicas = []string{fakeLocalVolumeName}

	return lv
}

// CreateFakeClient Create LocalVolume resource
func CreateFakeClient() (client.Client, *runtime.Scheme) {
	lv := GenFakeLocalVolumeObject()
	lvList := &v1alpha1.LocalVolumeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalVolumeKind,
			APIVersion: apiversion,
		},
	}

	lsn := GenFakeLocalStorageNodeObject()
	lsnList := &v1alpha1.LocalStorageNodeList{
		TypeMeta: metav1.TypeMeta{
			Kind:       LocalStorageNodeKind,
			APIVersion: apiversion,
		},
	}

	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lv)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lvList)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lsn)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, lsnList)
	return fake.NewClientBuilder().WithScheme(s).Build(), s
}

func Test_resources_predicate(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	var vol = &v1alpha1.LocalVolume{}
	vol.Name = fakeVolName
	nodeName := "test_node_name1"
	nodeName2 := "test"

	type args struct {
		vol      *v1alpha1.LocalVolume
		nodeName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args:    args{vol: vol, nodeName: nodeName},
			wantErr: true,
		},
		{
			args:    args{vol: vol, nodeName: nodeName2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				logger:               log.WithField("Module", "Scheduler/Resources"),
				allocatedResourceIDs: make(map[string]int),
				freeResourceIDList:   make([]int, 0, 10),
				maxHAVolumeCount:     10,
				allocatedStorages:    newStorageCollection(),
				totalStorages:        newStorageCollection(),
				storageNodes:         map[string]*v1alpha1.LocalStorageNode{},
			}
			r.storageNodes["test"] = &v1alpha1.LocalStorageNode{}
			if err := r.predicate(tt.args.vol, tt.args.nodeName); (err != nil) != tt.wantErr {
				t.Errorf("predicate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func GenFakeLocalStorageNodeObject() *v1alpha1.LocalStorageNode {
	lsn := &v1alpha1.LocalStorageNode{}

	TypeMeta := metav1.TypeMeta{
		Kind:       LocalStorageNodeKind,
		APIVersion: apiversion,
	}

	ObjectMata := metav1.ObjectMeta{
		Name:              fakeLocalStorageNodeName,
		Namespace:         fakeNamespace,
		ResourceVersion:   "",
		UID:               types.UID(fakeLocalStorageNodeUID),
		CreationTimestamp: metav1.Time{Time: time.Now()},
	}

	Spec := v1alpha1.LocalStorageNodeSpec{
		HostName:  fakeNodename,
		StorageIP: fakeStorageIp,
		Topo: v1alpha1.Topology{
			Zone:   fakeZone,
			Region: fakeRegion,
		},
	}

	disks := make([]v1alpha1.LocalDevice, 0, 10)
	var localdisk1 v1alpha1.LocalDevice
	localdisk1.DevPath = "/dev/sdf"
	localdisk1.State = v1alpha1.DiskStateAvailable
	localdisk1.Class = fakePoolClass
	localdisk1.CapacityBytes = fakeDiskCapacityBytes
	disks = append(disks, localdisk1)

	volumes := make([]string, 0, 5)
	volumes = append(volumes, "volume-test1")

	pools := make(map[string]v1alpha1.LocalPool)
	pools[fakeVgType] = v1alpha1.LocalPool{
		Name:                     fakeVgName,
		Class:                    fakePoolClass,
		Type:                     fakePoolType,
		TotalCapacityBytes:       int64(fakeTotalCapacityBytes),
		UsedCapacityBytes:        int64(fakeTotalCapacityBytes) - int64(fakeFreeCapacityBytes),
		FreeCapacityBytes:        int64(fakeFreeCapacityBytes),
		VolumeCapacityBytesLimit: int64(fakeTotalCapacityBytes),
		TotalVolumeCount:         v1alpha1.LVMVolumeMaxCount,
		UsedVolumeCount:          int64(len(volumes)),
		FreeVolumeCount:          v1alpha1.LVMVolumeMaxCount - int64(len(volumes)),
		Disks:                    disks,
		Volumes:                  volumes,
	}

	lsn.ObjectMeta = ObjectMata
	lsn.TypeMeta = TypeMeta
	lsn.Spec = Spec
	lsn.Status.State = v1alpha1.NodeStateReady
	lsn.Status.Pools = pools
	return lsn
}

func Test_resources_recycleAllocatedStorage(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		vol *v1alpha1.LocalVolume
	}
	var vol = &v1alpha1.LocalVolume{}
	vol.Name = "test"
	var allocatedResourceIDs = make(map[string]int)
	allocatedResourceIDs["test"] = 10

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args:   args{vol: vol},
			fields: fields{allocatedResourceIDs: allocatedResourceIDs},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			r.recycleAllocatedStorage(tt.args.vol)
		})
	}
}

func Test_resources_recycleResourceID(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		vol *v1alpha1.LocalVolume
	}
	var vol = &v1alpha1.LocalVolume{}
	vol.Name = "test"
	var allocatedResourceIDs = make(map[string]int)
	allocatedResourceIDs["test"] = 10

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			args:   args{vol: vol},
			fields: fields{allocatedResourceIDs: allocatedResourceIDs},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			r.recycleResourceID(tt.args.vol)
		})
	}
}

func Test_resources_score(t *testing.T) {
	type fields struct {
		apiClient            client.Client
		allocatedResourceIDs map[string]int
		freeResourceIDList   []int
		maxHAVolumeCount     int
		allocatedStorages    *storageCollection
		totalStorages        *storageCollection
		storageNodes         map[string]*v1alpha1.LocalStorageNode
		lock                 sync.Mutex
		logger               *log.Entry
	}
	type args struct {
		vol      *v1alpha1.LocalVolume
		nodeName string
	}
	var vol = &v1alpha1.LocalVolume{}
	vol.Name = fakeVolName
	nodeName := "test_node_name1"
	nodeName2 := "test"

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantScore int64
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			args:    args{vol: vol, nodeName: nodeName},
			wantErr: true,
		},
		{
			args:    args{vol: vol, nodeName: nodeName2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resources{
				apiClient:            tt.fields.apiClient,
				allocatedResourceIDs: tt.fields.allocatedResourceIDs,
				freeResourceIDList:   tt.fields.freeResourceIDList,
				maxHAVolumeCount:     tt.fields.maxHAVolumeCount,
				allocatedStorages:    tt.fields.allocatedStorages,
				totalStorages:        tt.fields.totalStorages,
				storageNodes:         tt.fields.storageNodes,
				lock:                 tt.fields.lock,
				logger:               tt.fields.logger,
			}
			gotScore, err := r.score(tt.args.vol, tt.args.nodeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("score() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotScore != tt.wantScore {
				t.Errorf("score() gotScore = %v, want %v", gotScore, tt.wantScore)
			}
		})
	}
}
