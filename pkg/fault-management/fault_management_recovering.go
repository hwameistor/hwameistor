package faultmanagement

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	log "github.com/sirupsen/logrus"
)

const (
	XFSREPAIR = "xfs_repair"
	EXTREPAIR = "fsck"
	fsTypeKey = "csi.storage.k8s.io/fstype"
)

func (m *manager) processFaultTicketRecovering(faultTicket *v1alpha1.FaultTicket) error {
	logger := m.logger.WithFields(log.Fields{
		"faultTicket": faultTicket.Name,
		"faultType":   faultTicket.Spec.Type,
		"node":        faultTicket.Spec.NodeName,
		"source":      faultTicket.Spec.Source,
		"message":     faultTicket.Spec.Message,
	})
	logger.Debug("Starting faultTicket recovery")

	// TODO(ming): handler these fault according to the config that admin applied
	var err error
	switch faultTicket.Spec.Type {
	case v1alpha1.DiskFaultTicket:
		err = m.recoveringDiskFault(faultTicket)

	case v1alpha1.VolumeFaultTicket:
		err = m.recoveringVolumeFault(faultTicket)

	case v1alpha1.NodeFaultTicket:
		err = m.recoveringNodeFault(faultTicket)

	default:
		logger.Debug("Unknown Fault Type, ignore it")
	}

	return err
}

func (m *manager) recoveringNodeFault(faultTicket *v1alpha1.FaultTicket) error {
	return nil
}

func (m *manager) recoveringVolumeFault(faultTicket *v1alpha1.FaultTicket) error {
	logger := m.logger.WithFields(log.Fields{
		"nodeName":        faultTicket.Spec.NodeName,
		"volumeName":      faultTicket.Spec.Volume.Name,
		"volumePath":      faultTicket.Spec.Volume.Path,
		"volumeFaultType": faultTicket.Spec.Volume.FaultType,
	})
	logger.Debug("recover a volume fault")

	var err error
	switch faultTicket.Spec.Volume.FaultType {
	case v1alpha1.BadBlockFault:
		err = m.recoverVolumeFromBadblock(faultTicket)
	case v1alpha1.FileSystemFault:
		err = m.recoverVolumeFromFilesystem(faultTicket)
	default:
		err = fmt.Errorf("UNKNOWN volume fault type: %v", faultTicket.Spec.Volume.FaultType)
	}

	return err
}

func (m *manager) recoverVolumeFromBadblock(faultTicket *v1alpha1.FaultTicket) error {
	localVolume, err := m.localVolumeLister.Get(faultTicket.Spec.Volume.Name)
	if err != nil {
		return err
	}
	pvc, err := m.pvcLister.PersistentVolumeClaims(localVolume.Spec.PersistentVolumeClaimNamespace).Get(localVolume.Spec.PersistentVolumeClaimName)
	if err != nil {
		return err
	}
	sc, err := m.storageClassLister.Get(*pvc.Spec.StorageClassName)
	if err != nil {
		return err
	}

	fsType, ok := sc.Parameters[fsTypeKey]
	if !ok {
		return fmt.Errorf("no fstype found in storageclass %s parameters", sc.Name)
	}

	var repairBadblock exechelper.ExecParams
	switch fsType {
	case "xfs":
		repairBadblock = exechelper.ExecParams{CmdName: XFSREPAIR}
	case "ext2", "ext3", "ext4":
		repairBadblock = exechelper.ExecParams{CmdName: EXTREPAIR}
	}

	res := m.cmdExec.RunCommand(repairBadblock)
	if res.ExitCode != 0 || res.Error != nil {
		err = fmt.Errorf("failed to execute command => %v, result => %v", repairBadblock, res)
	}

	return err
}

func (m *manager) recoverVolumeFromFilesystem(faultTicket *v1alpha1.FaultTicket) error {
	return nil
}

func (m *manager) recoveringDiskFault(faultTicket *v1alpha1.FaultTicket) error {
	return nil
}