package scheduler

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/controller/disk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/controller/volume"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	storagev1lister "k8s.io/client-go/listers/storage/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sort"
	"strings"
	"sync"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/driver/identity"
)

// diskVolumeSchedulerPlugin implement the Scheduler interface
// defined in github.com/hwameistor/scheduler/pkg/scheduler/scheduler.go: Scheduler
type diskVolumeSchedulerPlugin struct {
	diskNodeHandler   disk.Manager
	diskVolumeHandler volume.Manager
	scLister          storagev1lister.StorageClassLister
}

func NewDiskVolumeSchedulerPlugin(scLister storagev1lister.StorageClassLister) *diskVolumeSchedulerPlugin {
	return &diskVolumeSchedulerPlugin{
		diskNodeHandler:   disk.New(),
		diskVolumeHandler: volume.New(),
		scLister:          scLister,
	}
}

// Filter whether the node meets the storage requirements of pod runtime.
// The following two types of situations need to be met at the same time:
// 1. If the pod uses a created volume, we need to ensure that the volume is located at the scheduled node.
// 2. If the pod uses a pending volume, we need to ensure that the scheduled node can meet the requirements of the volume.
func (s *diskVolumeSchedulerPlugin) Filter(boundVolumes []string, pendingVolumes []*v1.PersistentVolumeClaim, node *v1.Node) (bool, error) {
	// return directly if no disk volume
	if len(boundVolumes) == 0 && len(pendingVolumes) == 0 {
		return true, nil
	}
	logCtx := log.Fields{
		"node":           node.GetName(),
		"boundVolumes":   strings.Join(boundVolumes, ","),
		"pendingVolumes": stringVolumes(pendingVolumes),
	}
	log.WithFields(logCtx).Debug("Start filter node")

	if ready, err := s.diskNodeHandler.NodeIsReady(node.GetName()); err != nil {
		log.WithError(err).WithFields(logCtx).Error("failed to get LocalDiskNode info")
		return false, err
	} else if !ready {
		log.WithFields(logCtx).Info("node is not ready")
		return false, nil
	}

	// step1: filter bounded volumes
	ok, err := s.filterExistVolumes(boundVolumes, node.GetName())
	if err != nil {
		log.WithFields(logCtx).WithError(err).Errorf("failed to filter node %s for bounded volumes %v due to error: %s", node.GetName(), boundVolumes, err.Error())
		return false, err
	}
	if !ok {
		log.WithFields(logCtx).Infof("node %s is not suitable because of bounded volumes %v is already located on the other node", node.GetName(), boundVolumes)
		return false, nil
	}

	// step2: filter pending volumes
	ok, err = s.filterPendingVolumes(pendingVolumes, node.GetName())
	if err != nil {
		log.WithFields(logCtx).WithError(err).Infof("failed to filter node %s for pending volumes due to error: %s", node.GetName(), err.Error())
		return false, err
	}
	if !ok {
		log.WithFields(logCtx).Infof("node %s is not suitable", node.GetName())
		return false, nil
	}

	log.WithFields(logCtx).Debug("Filter node success")
	return true, nil
}

func stringVolumes(pvs []*v1.PersistentVolumeClaim) (s string) {
	var ss []string
	for _, pv := range pvs {
		ss = append(ss, pv.GetName())
	}
	return strings.Join(ss, ",")
}

// Reserve disk needed by the volumes
func (s *diskVolumeSchedulerPlugin) Reserve(pendingVolumes []*v1.PersistentVolumeClaim, node string) error {
	return nil
}

// Unreserve disk reserved by the volumes on the node
func (s *diskVolumeSchedulerPlugin) Unreserve(pendingVolumes []*v1.PersistentVolumeClaim, node string) error {
	return nil
}

func (s *diskVolumeSchedulerPlugin) Score(unboundPVCs []*v1.PersistentVolumeClaim, node string) (int64, error) {
	return framework.MinNodeScore, nil
}

func (s *diskVolumeSchedulerPlugin) removeDuplicatePVC(pendingVolumes []*v1.PersistentVolumeClaim) (pvs []*v1.PersistentVolumeClaim) {
	pvcMap := map[string]*v1.PersistentVolumeClaim{}
	for i, pvc := range pendingVolumes {
		if _, ok := pvcMap[pvc.GetName()]; ok {
			continue
		} else {
			pvcMap[pvc.GetName()] = pendingVolumes[i]
			pvs = append(pvs, pendingVolumes[i])
		}
	}
	return
}

// filterExistVolumes compare the tobe scheduled node is equal to the node where volume already located at
func (s *diskVolumeSchedulerPlugin) filterExistVolumes(boundVolumes []string, tobeScheduleNode string) (bool, error) {
	for _, name := range boundVolumes {
		vol, err := s.diskVolumeHandler.GetVolumeInfo(name)
		if err != nil {
			log.WithError(err).Errorf("failed to get vol %s info", name)
			return false, err
		}
		log.Debugf("exist vol node: %s, tobeSchedulerNode: %s", vol.AttachNode, tobeScheduleNode)
		if vol.AttachNode != tobeScheduleNode {
			log.Infof("bounded vol is located at node %s,so node %s is not suitable", vol.AttachNode, tobeScheduleNode)
			return false, nil
		}
	}

	return true, nil
}

func (s *diskVolumeSchedulerPlugin) convertPVCToDiskRequest(pvc *v1.PersistentVolumeClaim, node string) (types.Disk, error) {
	sc, err := s.getParamsFromStorageClass(pvc)
	if err != nil {
		log.WithError(err).Errorf("failed to parse params from StorageClass")
		return types.Disk{}, err
	}

	storage := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	return types.Disk{
		AttachNode: node,
		Capacity:   storage.Value(),
		DiskType:   sc.DiskType,
	}, nil
}

func (s *diskVolumeSchedulerPlugin) getParamsFromStorageClass(volume *v1.PersistentVolumeClaim) (*StorageClassParams, error) {
	// sc here can't be empty,
	// more info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1
	if volume.Spec.StorageClassName == nil {
		return nil, fmt.Errorf("storageclass in pvc %s can't be empty", volume.GetName())
	}

	sc, err := s.scLister.Get(*volume.Spec.StorageClassName)
	if err != nil {
		return nil, err
	}

	return parseParams(sc.Parameters), nil
}

// filterPendingVolumes select free disks for pending pvc
func (s *diskVolumeSchedulerPlugin) filterPendingVolumes(pendingVolumes []*v1.PersistentVolumeClaim, tobeScheduleNode string) (bool, error) {
	avaSortDisks, err := s.diskNodeHandler.GetNodeAvailableDisks(tobeScheduleNode)
	if err != nil {
		return false, err
	}

	pendingSortVolumes := s.removeDuplicatePVC(pendingVolumes)
	if len(pendingVolumes) > len(avaSortDisks) {
		log.WithFields(log.Fields{"avaSortDisks": len(avaSortDisks), "pendingVolumes": len(pendingVolumes)}).Info("No enough free disks")
		return false, nil
	}

	// descending order
	sort.Sort(sort.Reverse(utils.ByDiskSize(avaSortDisks)))
	sort.Sort(sort.Reverse(utils.ByVolumeCapacity(pendingSortVolumes)))

	// we should order available disks and persistent volume claim by d type(e.g. HDD,SSD etc.)
	var classSortDisks, classSortVolumes sync.Map
	for _, d := range avaSortDisks {
		v, ok := classSortDisks.Load(d.DiskType)
		var classDisks []types.Disk
		if ok {
			classDisks = v.([]types.Disk)
		}

		classDisks = append(classDisks, d)
		classSortDisks.Store(d.DiskType, classDisks)
	}

	for _, vol := range pendingSortVolumes {
		params, err := s.getParamsFromStorageClass(vol)
		if err != nil {
			return false, err
		}
		var classVolumes []*v1.PersistentVolumeClaim
		v, ok := classSortVolumes.Load(params.DiskType)
		if ok {
			classVolumes = v.([]*v1.PersistentVolumeClaim)
		}

		classVolumes = append(classVolumes, vol)
		classSortVolumes.Store(params.DiskType, classVolumes)
	}

	var meetup = true
	// compare request storage capacity and available disk capacity
	classSortVolumes.Range(func(key, value any) bool {
		volumeType := key.(string)
		classPendingVolumes := value.([]*v1.PersistentVolumeClaim)
		if len(classPendingVolumes) == 0 {
			return true
		}
		v, ok := classSortDisks.Load(volumeType)
		if !ok {
			log.WithFields(log.Fields{"volumeType": volumeType, "volumes": len(classPendingVolumes)}).Info("There is no matchable type disk available")
			meetup = false
			return meetup
		}
		classAvailableDisks := v.([]types.Disk)
		for i, pendingVolume := range classPendingVolumes {
			if pendingVolume.Spec.Resources.Requests.Storage().Value() <= classAvailableDisks[i].Capacity {
				continue
			}
			log.WithFields(log.Fields{"index": i, "pendingVolume": pendingVolume.GetName(), "requestCapacity": pendingVolume.Spec.Resources.Requests.Storage().Value()}).
				Info("Can't meetup volume request storage capacity")
			meetup = false
			return meetup
		}
		return true
	})
	return meetup, nil
}

func (s *diskVolumeSchedulerPlugin) CSIDriverName() string {
	return identity.DriverName
}
