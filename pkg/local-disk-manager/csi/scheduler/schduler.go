package scheduler

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	storagev1lister "k8s.io/client-go/listers/storage/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/diskmanager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/driver/identity"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/csi/volumemanager"
)

// diskVolumeSchedulerPlugin implement the Scheduler interface
// defined in github.com/hwameistor/scheduler/pkg/scheduler/scheduler.go: Scheduler
type diskVolumeSchedulerPlugin struct {
	diskNodeHandler   diskmanager.DiskManager
	diskVolumeHandler volumemanager.VolumeManager
	scLister          storagev1lister.StorageClassLister
}

func NewDiskVolumeSchedulerPlugin(scLister storagev1lister.StorageClassLister) *diskVolumeSchedulerPlugin {
	return &diskVolumeSchedulerPlugin{
		diskNodeHandler:   diskmanager.NewLocalDiskManager(),
		diskVolumeHandler: volumemanager.NewLocalDiskVolumeManager(),
		scLister:          scLister,
	}
}

// Filter whether the node meets the storage requirements of pod runtime.
// The following two types of situations need to be met at the same time:
// 1. If the pod uses a created volume, we need to ensure that the volume is located at the scheduled node.
// 2. If the pod uses a pending volume, we need to ensure that the scheduled node can meet the requirements of the volume.
func (s *diskVolumeSchedulerPlugin) Filter(boundVolumes []string, pendingVolumes []*v1.PersistentVolumeClaim, node *v1.Node) (bool, error) {
	logCtx := log.Fields{
		"boundVolumes":   strings.Join(boundVolumes, ","),
		"node":           node.GetName(),
		"pendingVolumes": listPendingVolumes(pendingVolumes),
	}
	log.WithFields(logCtx).Debug("start disk volume filter")

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

	log.WithFields(logCtx).Debug("succeed filter disk volume")
	return true, nil
}

func listPendingVolumes(pvs []*v1.PersistentVolumeClaim) (s string) {
	for _, pv := range pvs {
		s = pv.GetName() + ","
	}
	return strings.TrimSuffix(s, ",")
}

// Reserve disk needed by the volumes
func (s *diskVolumeSchedulerPlugin) Reserve(pendingVolumes []*v1.PersistentVolumeClaim, node string) error {
	log.WithFields(log.Fields{"node": node, "volumes": listPendingVolumes(pendingVolumes)}).Debug("reserving disk")
	for _, pvc := range pendingVolumes {
		diskReq, err := s.convertPVCToDiskRequest(pvc, node)
		if err != nil {
			return err
		}
		if err = s.diskNodeHandler.ReserveDiskForVolume(diskReq, pvc.GetNamespace()+"-"+pvc.GetName()); err != nil {
			return err
		}
	}
	return nil
}

// Unreserve disk reserved by the volumes on the node
func (s *diskVolumeSchedulerPlugin) Unreserve(pendingVolumes []*v1.PersistentVolumeClaim, node string) error {
	log.WithFields(log.Fields{"node": node, "volumes": pendingVolumes}).Debug("unreserving disk")
	for _, pvc := range pendingVolumes {
		if err := s.diskNodeHandler.UnReserveDiskForPVC(pvc.GetNamespace() + "-" + pvc.GetName()); err != nil {
			return err
		}
	}
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
		volume, err := s.diskVolumeHandler.GetVolumeInfo(name)
		if err != nil {
			log.WithError(err).Errorf("failed to get volume %s info", name)
			return false, err
		}
		log.Debugf("exist volume node: %s, tobeSchedulerNode: %s", volume.AttachNode, tobeScheduleNode)
		if volume.AttachNode != tobeScheduleNode {
			log.Infof("bounded volume is located at node %s,so node %s is not suitable", volume.AttachNode, tobeScheduleNode)
			return false, nil
		}
	}

	return true, nil
}

func (s *diskVolumeSchedulerPlugin) convertPVCToDiskRequest(pvc *v1.PersistentVolumeClaim, node string) (diskmanager.Disk, error) {
	sc, err := s.getParamsFromStorageClass(pvc)
	if err != nil {
		log.WithError(err).Errorf("failed to parse params from StorageClass")
		return diskmanager.Disk{}, err
	}

	storage := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	return diskmanager.Disk{
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
	pendingVolumes = s.removeDuplicatePVC(pendingVolumes)
	var reqDisks []diskmanager.Disk
	for _, pvc := range pendingVolumes {
		disk, err := s.convertPVCToDiskRequest(pvc, tobeScheduleNode)
		if err != nil {
			return false, err
		}
		reqDisks = append(reqDisks, disk)
	}

	return s.diskNodeHandler.FilterFreeDisks(reqDisks)
}

func (s *diskVolumeSchedulerPlugin) CSIDriverName() string {
	return identity.DriverName
}
