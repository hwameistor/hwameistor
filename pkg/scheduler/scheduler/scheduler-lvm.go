package scheduler

import (
	"context"
	"fmt"
	"strconv"

	localstorageapis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage"
	localstoragev1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage/v1alpha1"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	storagev1lister "k8s.io/client-go/listers/storage/v1"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LVMVolumeScheduler struct {
	fHandle   framework.FrameworkHandle
	apiClient client.Client

	csiDriverName    string
	topoNodeLabelKey string

	replicaScheduler localstoragev1alpha1.VolumeScheduler
	hwameiStorCache  cache.Cache

	scLister storagev1lister.StorageClassLister
}

func NewLVMVolumeScheduler(f framework.FrameworkHandle, scheduler localstoragev1alpha1.VolumeScheduler, hwameiStorCache cache.Cache, cli client.Client) VolumeScheduler {

	sche := &LVMVolumeScheduler{
		fHandle:          f,
		apiClient:        cli,
		topoNodeLabelKey: localstorageapis.TopologyNodeKey,
		csiDriverName:    localstoragev1alpha1.CSIDriverName,
		replicaScheduler: scheduler,
		hwameiStorCache:  hwameiStorCache,
		scLister:         f.SharedInformerFactory().Storage().V1().StorageClasses().Lister(),
	}

	return sche
}

func (s *LVMVolumeScheduler) CSIDriverName() string {
	return s.csiDriverName
}

func (s *LVMVolumeScheduler) Filter(lvs []string, pendingPVCs []*corev1.PersistentVolumeClaim, node *corev1.Node) (bool, error) {
	canSchedule, err := s.filterForExistingLocalVolumes(lvs, node)
	if err != nil {
		return false, err
	}
	if !canSchedule {
		return false, fmt.Errorf("filtered out the node %s", node.Name)
	}

	return s.filterForNewPVCs(pendingPVCs, node)
}

func (s *LVMVolumeScheduler) filterForExistingLocalVolumes(lvs []string, node *corev1.Node) (bool, error) {

	if len(lvs) == 0 {
		return true, nil
	}

	// Bound PVC already has volume created in the cluster. Just check if this node has the expected volume
	for _, lvName := range lvs {
		lv := &localstoragev1alpha1.LocalVolume{}
		if err := s.hwameiStorCache.Get(context.Background(), types.NamespacedName{Name: lvName}, lv); err != nil {
			log.WithFields(log.Fields{"localvolume": lvName}).WithError(err).Error("Failed to fetch LocalVolume")
			return false, err
		}
		if lv.Spec.Config == nil {
			log.WithFields(log.Fields{"localvolume": lvName}).Error("Not found replicas info in the LocalVolume")
			return false, fmt.Errorf("pending localvolume")
		}
		isLocalNode := false
		for _, rep := range lv.Spec.Config.Replicas {
			if rep.Hostname == node.Name {
				isLocalNode = true
				break
			}
		}
		if !isLocalNode {
			log.WithFields(log.Fields{"localvolume": lvName, "node": node.Name}).Debug("LocalVolume doesn't locate at this node")
			return false, fmt.Errorf("not right node")
		}
	}

	log.WithFields(log.Fields{"localvolumes": lvs, "node": node.Name}).Debug("Filtered in this node for all existing LVM volumes")
	return true, nil
}

func (s *LVMVolumeScheduler) filterForNewPVCs(pvcs []*corev1.PersistentVolumeClaim, node *corev1.Node) (bool, error) {

	if len(pvcs) == 0 {
		return true, nil
	}
	for _, pvc := range pvcs {
		log.WithField("pvc", pvc.Name).WithField("node", node.Name).Debug("New PVC")
	}
	lvs := []*localstoragev1alpha1.LocalVolume{}
	for i := range pvcs {
		lv, err := s.constructLocalVolumeForPVC(pvcs[i])
		if err != nil {
			return false, err
		}
		lvs = append(lvs, lv)
	}

	qualifiedNodes := s.replicaScheduler.GetNodeCandidates(lvs)
	if len(qualifiedNodes) < int(lvs[0].Spec.ReplicaNumber) {
		return false, fmt.Errorf("no enough nodes")
	}
	for _, qn := range qualifiedNodes {
		if qn.Name == node.Name {
			return true, nil
		}
	}

	return false, nil
}

func (s *LVMVolumeScheduler) constructLocalVolumeForPVC(pvc *corev1.PersistentVolumeClaim) (*localstoragev1alpha1.LocalVolume, error) {

	sc, err := s.scLister.Get(*pvc.Spec.StorageClassName)
	if err != nil {
		return nil, err
	}
	localVolume := localstoragev1alpha1.LocalVolume{}
	poolName, err := buildStoragePoolName(
		sc.Parameters[localstoragev1alpha1.VolumeParameterPoolClassKey],
		sc.Parameters[localstoragev1alpha1.VolumeParameterPoolTypeKey])
	if err != nil {
		return nil, err
	}

	localVolume.Spec.PoolName = poolName
	storage := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	localVolume.Spec.RequiredCapacityBytes = storage.Value()
	replica, _ := strconv.Atoi(sc.Parameters[localstoragev1alpha1.VolumeParameterReplicaNumberKey])
	localVolume.Spec.ReplicaNumber = int64(replica)
	return &localVolume, nil
}

func buildStoragePoolName(poolClass string, poolType string) (string, error) {

	if poolClass == localstoragev1alpha1.DiskClassNameHDD && poolType == localstoragev1alpha1.PoolTypeRegular {
		return localstoragev1alpha1.PoolNameForHDD, nil
	}
	if poolClass == localstoragev1alpha1.DiskClassNameSSD && poolType == localstoragev1alpha1.PoolTypeRegular {
		return localstoragev1alpha1.PoolNameForSSD, nil
	}
	if poolClass == localstoragev1alpha1.DiskClassNameNVMe && poolType == localstoragev1alpha1.PoolTypeRegular {
		return localstoragev1alpha1.PoolNameForNVMe, nil
	}

	return "", fmt.Errorf("invalid pool info")
}

func (s *LVMVolumeScheduler) Reserve(pendingPVCs []*corev1.PersistentVolumeClaim, node string) error {
	return nil
}

func (s *LVMVolumeScheduler) Unreserve(pendingPVCs []*corev1.PersistentVolumeClaim, node string) error {
	return nil
}
