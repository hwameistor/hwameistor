package scheduler

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1lister "k8s.io/client-go/listers/core/v1"
	storagev1lister "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/tools/cache"
	volumehelpers "k8s.io/component-helpers/storage/volume"
	controllercache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type testVolumeScheduler struct {
	VolumeScheduler
	driver string
	filter func([]string, []*corev1.PersistentVolumeClaim, *corev1.Node) (bool, error)
	score  func([]*corev1.PersistentVolumeClaim, string) (int64, error)
}

func (s *testVolumeScheduler) CSIDriverName() string { return s.driver }
func (s *testVolumeScheduler) Filter(lvs []string, pvcs []*corev1.PersistentVolumeClaim, node *corev1.Node) (bool, error) {
	if s.filter != nil {
		return s.filter(lvs, pvcs, node)
	}
	return true, nil
}
func (s *testVolumeScheduler) Score(pvcs []*corev1.PersistentVolumeClaim, node string) (int64, error) {
	return s.score(pvcs, node)
}

func testPVC(name, driver, volumeName, selectedNode string, phase corev1.PersistentVolumeClaimPhase) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: "test",
			Annotations: map[string]string{volumehelpers.AnnSelectedNode: selectedNode},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &driver,
			VolumeName:       volumeName,
			Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("1Gi"),
			}},
		},
		Status: corev1.PersistentVolumeClaimStatus{Phase: phase},
	}
}

func testScheduler(t *testing.T, pvcs ...*corev1.PersistentVolumeClaim) (*Scheduler, *corev1.Pod, cache.Indexer) {
	t.Helper()
	pvcIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	pvIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	scIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	add := func(indexer cache.Indexer, obj interface{}) {
		t.Helper()
		if err := indexer.Add(obj); err != nil {
			t.Fatal(err)
		}
	}
	for _, driver := range []string{"lvm.hwameistor.io", "disk.hwameistor.io", "other.csi.io"} {
		add(scIndexer, &storagev1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{Name: driver}, Provisioner: driver,
			Parameters: map[string]string{
				v1alpha1.VolumeParameterPoolClassKey:     v1alpha1.DiskClassNameHDD,
				v1alpha1.VolumeParameterReplicaNumberKey: "1",
			},
		})
	}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "test"}}
	for _, pvc := range pvcs {
		add(pvcIndexer, pvc)
		pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
			Name:         pvc.Name,
			VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: pvc.Name}},
		})
		if pvc.Spec.VolumeName != "" {
			add(pvIndexer, &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{Name: pvc.Spec.VolumeName},
				Spec: corev1.PersistentVolumeSpec{PersistentVolumeSource: corev1.PersistentVolumeSource{
					CSI: &corev1.CSIPersistentVolumeSource{VolumeHandle: "handle-" + pvc.Spec.VolumeName},
				}},
			})
		}
	}
	return &Scheduler{
		lvmScheduler:  &testVolumeScheduler{driver: "lvm.hwameistor.io"},
		diskScheduler: &testVolumeScheduler{driver: "disk.hwameistor.io"},
		pvcLister:     corev1lister.NewPersistentVolumeClaimLister(pvcIndexer),
		pvLister:      corev1lister.NewPersistentVolumeLister(pvIndexer),
		scLister:      storagev1lister.NewStorageClassLister(scIndexer),
	}, pod, pvcIndexer
}

func TestSchedulerPVCStates(t *testing.T) {
	for _, driver := range []string{"lvm.hwameistor.io", "disk.hwameistor.io"} {
		for _, tt := range []struct {
			name, volumeName, selectedNode, node string
			phase                                corev1.PersistentVolumeClaimPhase
			wantErr                              string
			wantExisting, wantNew                bool
		}{
			{name: "selected node continues provisioning", phase: corev1.ClaimPending, selectedNode: "node-a", node: "node-a"},
			{name: "selected node rejects another node", phase: corev1.ClaimPending, selectedNode: "node-a", node: "node-b", wantErr: "selected node node-a"},
			{name: "new claim uses capacity filter", phase: corev1.ClaimPending, node: "node-b", wantNew: true},
			{name: "pending with PV still uses selected node", phase: corev1.ClaimPending, volumeName: "pv", selectedNode: "node-a", node: "node-a"},
			{name: "pending with PV rejects another node", phase: corev1.ClaimPending, volumeName: "pv", selectedNode: "node-a", node: "node-b", wantErr: "selected node node-a"},
			{name: "volumeName alone does not change pending classification", phase: corev1.ClaimPending, volumeName: "pv", node: "node-a", wantNew: true},
			{name: "bound ignores historical selection", phase: corev1.ClaimBound, volumeName: "pv", selectedNode: "node-b", node: "node-a", wantExisting: true},
			{name: "lost remains unhealthy", phase: corev1.ClaimLost, volumeName: "pv", selectedNode: "node-a", node: "node-a", wantErr: "unhealthy"},
			{name: "unknown remains unhealthy", selectedNode: "node-a", node: "node-a", wantErr: "unhealthy"},
		} {
			t.Run(driver+"/"+tt.name, func(t *testing.T) {
				pvc := testPVC("claim", driver, tt.volumeName, tt.selectedNode, tt.phase)
				s, pod, _ := testScheduler(t, pvc)
				// Selected claims must not trigger the new-PVC affinity patch either.
				if tt.selectedNode != "" {
					pod.Spec.Affinity = &corev1.Affinity{}
				}
				backend := s.lvmScheduler.(*testVolumeScheduler)
				if driver == "disk.hwameistor.io" {
					backend = s.diskScheduler.(*testVolumeScheduler)
				}
				filterCalls, scoreCalls := 0, 0
				backend.filter = func(lvs []string, pvcs []*corev1.PersistentVolumeClaim, node *corev1.Node) (bool, error) {
					filterCalls++
					if tt.wantExisting {
						if len(lvs) != 1 || lvs[0] != "handle-pv" {
							t.Fatalf("expected PV CSI handle, got %v", lvs)
						}
					} else if len(lvs) != 0 {
						t.Fatalf("unexpected existing volumes: %v", lvs)
					}
					if tt.wantNew {
						if len(pvcs) != 1 || pvcs[0].Name != pvc.Name {
							t.Fatalf("expected claim in capacity filter, got %v", pvcs)
						}
					} else if len(pvcs) != 0 {
						t.Fatalf("unexpected new claims: %v", pvcs)
					}
					return true, nil
				}
				backend.score = func(pvcs []*corev1.PersistentVolumeClaim, node string) (int64, error) {
					scoreCalls++
					if len(pvcs) != 1 || pvcs[0].Name != pvc.Name {
						t.Fatalf("unexpected claims in Score: %v", pvcs)
					}
					return 50, nil
				}
				allowed, err := s.Filter(pod, &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: tt.node}})
				if tt.wantErr != "" {
					if allowed || err == nil || !strings.Contains(err.Error(), tt.wantErr) || filterCalls != 0 {
						t.Fatalf("Filter = (%v, %v), calls = %d; want rejection containing %q before volume filtering", allowed, err, filterCalls, tt.wantErr)
					}
					return
				}
				if !allowed || err != nil || filterCalls != 1 {
					t.Fatalf("Filter = (%v, %v), calls = %d", allowed, err, filterCalls)
				}
				score, err := s.Score(pod, tt.node)
				if err != nil || (tt.wantNew && (scoreCalls != 1 || score != 50)) || (!tt.wantNew && (scoreCalls != 0 || score != 0)) {
					t.Fatalf("Score = (%d, %v), calls = %d", score, err, scoreCalls)
				}
			})
		}
	}
}

func TestSchedulerSelectedNodeRemoved(t *testing.T) {
	pvc := testPVC("claim", "lvm.hwameistor.io", "", "node-a", corev1.ClaimPending)
	s, pod, indexer := testScheduler(t, pvc)
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-b"}}
	if allowed, err := s.Filter(pod, node); allowed || err == nil {
		t.Fatalf("expected selected-node rejection, got (%v, %v)", allowed, err)
	}
	updated := pvc.DeepCopy()
	delete(updated.Annotations, volumehelpers.AnnSelectedNode)
	if err := indexer.Update(updated); err != nil {
		t.Fatal(err)
	}
	calls := 0
	s.lvmScheduler.(*testVolumeScheduler).filter = func(_ []string, pvcs []*corev1.PersistentVolumeClaim, _ *corev1.Node) (bool, error) {
		calls++
		if len(pvcs) != 1 || pvcs[0].Name != pvc.Name {
			t.Fatalf("expected a fresh capacity check, got %v", pvcs)
		}
		return true, nil
	}
	if allowed, err := s.Filter(pod, node); !allowed || err != nil || calls != 1 {
		t.Fatalf("expected fresh selection after annotation removal, got (%v, %v), calls = %d", allowed, err, calls)
	}
}

func TestSchedulerConflictingSelectedNodes(t *testing.T) {
	s, pod, _ := testScheduler(t,
		testPVC("lvm", "lvm.hwameistor.io", "", "node-a", corev1.ClaimPending),
		testPVC("disk", "disk.hwameistor.io", "", "node-b", corev1.ClaimPending),
	)
	for _, name := range []string{"node-a", "node-b"} {
		if allowed, err := s.Filter(pod, &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: name}}); allowed || err == nil {
			t.Fatalf("conflicting selections must reject %s, got (%v, %v)", name, allowed, err)
		}
	}
}

func TestSchedulerIgnoresOtherDriversSelectedNode(t *testing.T) {
	s, pod, _ := testScheduler(t, testPVC("other", "other.csi.io", "", "node-a", corev1.ClaimPending))
	if allowed, err := s.Filter(pod, &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-b"}}); !allowed || err != nil {
		t.Fatalf("another CSI driver's annotation must not constrain HwameiStor, got (%v, %v)", allowed, err)
	}
}

type testStorageCache struct {
	controllercache.Informers
	client.Reader
}

type testReplicaScheduler struct {
	v1alpha1.VolumeScheduler
	candidates func([]*v1alpha1.LocalVolume) []*v1alpha1.LocalStorageNode
}

func (s *testReplicaScheduler) GetNodeCandidates(lvs []*v1alpha1.LocalVolume) []*v1alpha1.LocalStorageNode {
	return s.candidates(lvs)
}

func TestSchedulerMixedPVCsDuringProvisioning(t *testing.T) {
	for _, tt := range []struct {
		name, volumeName, selectedNode, nodeBError string
		phase                                      corev1.PersistentVolumeClaimPhase
		noCapacity                                 bool
	}{
		{name: "selected volume not created yet", phase: corev1.ClaimPending, selectedNode: "node-a", nodeBError: "selected node node-a"},
		{name: "created volume still pending as in issue 1863", phase: corev1.ClaimPending, volumeName: "pv", selectedNode: "node-a", nodeBError: "selected node node-a"},
		{name: "bound volume uses actual replica despite old annotation", phase: corev1.ClaimBound, volumeName: "pv", selectedNode: "node-b", nodeBError: "not the published node"},
		{name: "selected node still checks the other new claim capacity", phase: corev1.ClaimPending, selectedNode: "node-a", nodeBError: "selected node node-a", noCapacity: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oldPVC := testPVC("existing", "lvm.hwameistor.io", tt.volumeName, tt.selectedNode, tt.phase)
			newPVC := testPVC("new", "lvm.hwameistor.io", "", "", corev1.ClaimPending)
			s, pod, _ := testScheduler(t, oldPVC, newPVC)
			scheme := runtime.NewScheme()
			if err := v1alpha1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			objects := []client.Object{
				&v1alpha1.LocalStorageNode{ObjectMeta: metav1.ObjectMeta{Name: "node-a"}},
				&v1alpha1.LocalStorageNode{ObjectMeta: metav1.ObjectMeta{Name: "node-b"}},
			}
			// Leave both PV and LocalVolume absent until creation has completed.
			if tt.volumeName != "" {
				objects = append(objects, &v1alpha1.LocalVolume{
					ObjectMeta: metav1.ObjectMeta{Name: "handle-pv"},
					Spec: v1alpha1.LocalVolumeSpec{Config: &v1alpha1.VolumeConfig{
						Replicas: []v1alpha1.VolumeReplica{{Hostname: "node-a"}},
					}},
				})
			}
			cli := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
			candidateCalls := 0
			s.lvmScheduler = &LVMVolumeScheduler{
				csiDriverName: "lvm.hwameistor.io", apiClient: cli,
				hwameiStorCache: &testStorageCache{Reader: cli}, scLister: s.scLister,
				replicaScheduler: &testReplicaScheduler{candidates: func(lvs []*v1alpha1.LocalVolume) []*v1alpha1.LocalStorageNode {
					candidateCalls++
					if len(lvs) != 1 || lvs[0].Spec.PersistentVolumeClaimName != "new" {
						t.Fatalf("capacity selection must receive only the genuinely new claim: %v", lvs)
					}
					if tt.noCapacity {
						return nil
					}
					return []*v1alpha1.LocalStorageNode{
						{ObjectMeta: metav1.ObjectMeta{Name: "node-a"}},
						{ObjectMeta: metav1.ObjectMeta{Name: "node-b"}},
					}
				}},
			}
			if allowed, err := s.Filter(pod, &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-b"}}); allowed || err == nil || !strings.Contains(err.Error(), tt.nodeBError) || candidateCalls != 0 {
				t.Fatalf("expected node-b rejection before capacity selection, got (%v, %v), calls = %d", allowed, err, candidateCalls)
			}
			allowed, err := s.Filter(pod, &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-a"}})
			if candidateCalls != 1 {
				t.Fatalf("expected new claim capacity check, calls = %d", candidateCalls)
			}
			if tt.noCapacity {
				if allowed || err == nil || !strings.Contains(err.Error(), "capacity and replica requirements") {
					t.Fatalf("expected insufficient capacity rejection, got (%v, %v)", allowed, err)
				}
			} else if !allowed || err != nil {
				t.Fatalf("expected node-a to pass, got (%v, %v)", allowed, err)
			}
		})
	}
}
