package scheduler

import (
	"context"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	v1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	storagev1lister "k8s.io/client-go/listers/storage/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

const VolumeSnapshot = "VolumeSnapshot"

type LVMVolumeScheduler struct {
	fHandle   framework.Handle
	apiClient client.Client

	csiDriverName    string
	topoNodeLabelKey string

	replicaScheduler v1alpha1.VolumeScheduler
	hwameiStorCache  cache.Cache

	scLister storagev1lister.StorageClassLister
}

func NewLVMVolumeScheduler(f framework.Handle, scheduler v1alpha1.VolumeScheduler, hwameiStorCache cache.Cache, cli client.Client) VolumeScheduler {

	sche := &LVMVolumeScheduler{
		fHandle:          f,
		apiClient:        cli,
		topoNodeLabelKey: apis.TopologyNodeKey,
		csiDriverName:    v1alpha1.CSIDriverName,
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

// Score node according to volume nums and storage pool capacity.
// For now, we only consider storage capacity, calculate logic is as bellow:
// volume capacity / poolFreeCapacity less, score higher
func (s *LVMVolumeScheduler) Score(unboundPVCs []*corev1.PersistentVolumeClaim, node string) (int64, error) {
	var (
		err         error
		storageNode v1alpha1.LocalStorageNode
		scoreTotal  int64
	)

	if err = s.hwameiStorCache.Get(context.Background(), types.NamespacedName{Name: node}, &storageNode); err != nil {
		return 0, err
	}

	// score for each volume
	for _, volume := range unboundPVCs {
		score, err := s.scoreOneVolume(volume, &storageNode)
		if err != nil {
			return 0, err
		}
		scoreTotal += score
	}

	return int64(float64(scoreTotal) / float64(framework.MaxNodeScore*int64(len(unboundPVCs))) * float64(framework.MaxNodeScore)), err
}

func (s *LVMVolumeScheduler) scoreOneVolume(pvc *corev1.PersistentVolumeClaim, node *v1alpha1.LocalStorageNode) (int64, error) {
	if pvc.Spec.StorageClassName == nil {
		return 0, fmt.Errorf("storageclass is empty in pvc %s", pvc.Name)
	}
	relatedSC, err := s.scLister.Get(*pvc.Spec.StorageClassName)
	if err != nil {
		return 0, err
	}

	volumeClass := relatedSC.Parameters[v1alpha1.VolumeParameterPoolClassKey]
	volumeCapacity := pvc.Spec.Resources.Requests.Storage().Value()
	poolClass, err := buildStoragePoolName(volumeClass, v1alpha1.PoolTypeRegular)
	if err != nil {
		return 0, err
	}
	relatedPool := node.Status.Pools[poolClass]
	nodeFreeCapacity := relatedPool.FreeCapacityBytes

	log.WithFields(log.Fields{
		"volume":           pvc.GetName(),
		"volumeCapacity":   volumeCapacity,
		"node":             node.GetName(),
		"nodeFreeCapacity": nodeFreeCapacity,
	}).Debug("score node for one lvm-volume")

	return int64(float64(nodeFreeCapacity-volumeCapacity) / float64(nodeFreeCapacity) * float64(framework.MaxNodeScore)), nil
}

func (s *LVMVolumeScheduler) filterForExistingLocalVolumes(lvs []string, node *corev1.Node) (bool, error) {

	if len(lvs) == 0 {
		return true, nil
	}

	// Bound PVC already has volume created in the cluster. Just check if this node has the expected volume
	for _, lvName := range lvs {
		lv := &v1alpha1.LocalVolume{}
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

		// if volume is already Published, also check if this node is the published node, see #1155 for more details.
		if lv.Status.PublishedNodeName != "" && lv.Status.PublishedNodeName != node.Name {
			log.WithFields(log.Fields{"localvolume": lvName, "node": node.Name}).Debug("LocalVolume doesn't publish at this node")
			return false, fmt.Errorf("not published node")
		}
	}

	log.WithFields(log.Fields{"localvolumes": lvs, "node": node.Name}).Debug("Filtered in this node for all existing LVM volumes")
	return true, nil
}

func (s *LVMVolumeScheduler) filterForNewPVCs(pvcs []*corev1.PersistentVolumeClaim, node *corev1.Node) (bool, error) {
	if len(pvcs) == 0 {
		return true, nil
	}

	// the scheduled node must keep consistent with the existing snapshot node
	if ok, err := s.validateNodeForPVCsFromSnapshot(pvcs, node); !ok {
		return false, fmt.Errorf("node %v is not the expected snapshot node, error: %v", node.Name, err)
	}

	for _, pvc := range pvcs {
		log.WithField("pvc", pvc.Name).WithField("node", node.Name).Debug("New PVC")
	}
	lvs := []*v1alpha1.LocalVolume{}
	for i := range pvcs {
		lv, err := s.constructLocalVolumeForPVC(pvcs[i])
		if err != nil {
			return false, err
		}
		lvs = append(lvs, lv)
	}

	qualifiedNodes := s.replicaScheduler.GetNodeCandidates(lvs)
	if len(qualifiedNodes) < int(lvs[0].Spec.ReplicaNumber) {
		return false, fmt.Errorf("need %d node(s) to place volume, but only find %d node(s) meet the volume capacity requirements",
			int(lvs[0].Spec.ReplicaNumber), len(qualifiedNodes))
	}
	for _, qn := range qualifiedNodes {
		if qn.Name == node.Name {
			return true, nil
		}
	}

	return false, nil
}

// validateNodeForSnapshotVolume ensures that the node can be scheduled at the node where snapshot located
func (s *LVMVolumeScheduler) validateNodeForPVCsFromSnapshot(pvcs []*corev1.PersistentVolumeClaim, node *corev1.Node) (bool, error) {
	var vss []string
	for _, pvc := range pvcs {
		if !isVolumeFromSnapshot(pvc) {
			continue
		}
		vss = append(vss, fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Spec.DataSource.Name))
	}
	if len(vss) == 0 {
		return true, nil
	}

	// range each volumesnapshot and compare the node
	for _, vsNamespaceName := range vss {
		vs := v1.VolumeSnapshot{}
		if err := s.apiClient.Get(context.Background(), client.ObjectKey{
			Namespace: strings.Split(vsNamespaceName, "/")[0],
			Name:      strings.Split(vsNamespaceName, "/")[1]}, &vs); err != nil {
			return false, err
		}

		// check if snapshotcontent is already created and bounded
		if vs.Status.ReadyToUse == nil || vs.Status.BoundVolumeSnapshotContentName == nil {
			return false, fmt.Errorf("snapshot %s is not ready to use", vs.Name)
		}

		snapshot := v1alpha1.LocalVolumeSnapshot{}
		if err := s.apiClient.Get(context.Background(), client.ObjectKey{Name: *vs.Status.BoundVolumeSnapshotContentName}, &snapshot); err != nil {
			return false, err
		}

		if _, ok := utils.StrFind(snapshot.Spec.Accessibility.Nodes, node.Name); !ok {
			return false, fmt.Errorf("node %s is not matchable with snapshot accessibility node(s) %v", node.Name, snapshot.Spec.Accessibility.Nodes)
		}
	}

	return true, nil
}

func (s *LVMVolumeScheduler) constructLocalVolumeForPVC(pvc *corev1.PersistentVolumeClaim) (*v1alpha1.LocalVolume, error) {
	var scName string
	if pvc.Spec.DataSource != nil && pvc.Spec.DataSource.Kind == VolumeSnapshot {
		// for volume create from snapshot, use sc from the source volume
		if srcPVC, err := s.getSourcePVCFromSnapshot(pvc.Namespace, pvc.Spec.DataSource.Name); err != nil {
			return nil, err
		} else {
			scName = *srcPVC.Spec.StorageClassName
		}
	} else {
		scName = *pvc.Spec.StorageClassName
	}

	sc, err := s.scLister.Get(scName)
	if err != nil {
		return nil, err
	}
	localVolume := v1alpha1.LocalVolume{}
	poolName, err := buildStoragePoolName(
		sc.Parameters[v1alpha1.VolumeParameterPoolClassKey],
		sc.Parameters[v1alpha1.VolumeParameterPoolTypeKey])
	if err != nil {
		return nil, err
	}

	//localVolume.Name = pvc.Name
	localVolume.Spec.PersistentVolumeClaimName = pvc.Name
	localVolume.Spec.PersistentVolumeClaimNamespace = pvc.Namespace
	localVolume.Spec.PoolName = poolName
	storage := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	localVolume.Spec.RequiredCapacityBytes = storage.Value()
	replica, _ := strconv.Atoi(sc.Parameters[v1alpha1.VolumeParameterReplicaNumberKey])
	localVolume.Spec.ReplicaNumber = int64(replica)
	return &localVolume, nil
}

func (s *LVMVolumeScheduler) getSourcePVCFromSnapshot(vsNamespace, vsName string) (*corev1.PersistentVolumeClaim, error) {
	vs := v1.VolumeSnapshot{}
	if err := s.apiClient.Get(context.Background(), client.ObjectKey{Namespace: vsNamespace, Name: vsName}, &vs); err != nil {
		return nil, err
	}
	pvc := corev1.PersistentVolumeClaim{}
	if err := s.apiClient.Get(context.Background(), client.ObjectKey{Namespace: vsNamespace, Name: *vs.Spec.Source.PersistentVolumeClaimName}, &pvc); err != nil {
		return nil, err
	}
	return &pvc, nil
}

func buildStoragePoolName(poolClass string, poolType string) (string, error) {

	if poolClass == v1alpha1.DiskClassNameHDD && poolType == v1alpha1.PoolTypeRegular {
		return v1alpha1.PoolNameForHDD, nil
	}
	if poolClass == v1alpha1.DiskClassNameSSD && poolType == v1alpha1.PoolTypeRegular {
		return v1alpha1.PoolNameForSSD, nil
	}
	if poolClass == v1alpha1.DiskClassNameNVMe && poolType == v1alpha1.PoolTypeRegular {
		return v1alpha1.PoolNameForNVMe, nil
	}

	return "", fmt.Errorf("invalid pool info")
}

func (s *LVMVolumeScheduler) Reserve(pendingPVCs []*corev1.PersistentVolumeClaim, node string) error {
	return nil
}

func (s *LVMVolumeScheduler) Unreserve(pendingPVCs []*corev1.PersistentVolumeClaim, node string) error {
	return nil
}

func isVolumeFromSnapshot(pvc *corev1.PersistentVolumeClaim) bool {
	return pvc.Spec.DataSource != nil && pvc.Spec.DataSource.Kind == VolumeSnapshot
}
