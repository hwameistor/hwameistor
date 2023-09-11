package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/exechelper"
	"github.com/hwameistor/hwameistor/pkg/exechelper/nsexecutor"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
)

// consts
const (
	LVMask = 1
	VGMask = 1 << 1
	PVMask = 1 << 2
)

// LVMUnknownStatus this represents no error or this status is not set
const LVMUnknownStatus = "unknown"

// variables
var (
	ErrNotLVMByteNum   = errors.New("LVM byte format unrecognised")
	ErrReplicaNotFound = errors.New("volume replica not found on host")
	LVMTimeLayout      = "2006-01-02 15:04:05 -0700"
)

// for PV/disk
type pvsReport struct {
	Records []pvsReportRecord `json:"report,omitempty"`
}

type pvsReportRecord struct {
	Records []pvRecord `json:"pv,omitempty"`
}

type pvRecord struct {
	Name     string `json:"pv_name,omitempty"`
	PoolName string `json:"vg_name,omitempty"`
	PvAttr   string `json:"pv_attr,omitempty"`
	PvSize   string `json:"pv_size,omitempty"`
	PvFree   string `json:"pv_free,omitempty"`
}

// for VG/pools
type vgsReport struct {
	Records []vgsReportRecord `json:"report,omitempty"`
}

type vgsReportRecord struct {
	Records []vgRecord `json:"vg,omitempty"`
}

type vgRecord struct {
	Name           string `json:"vg_name"`
	PvCount        string `json:"pv_count"`
	LvCount        string `json:"lv_count"`
	SnapCount      string `json:"snap_count"`
	VgAttr         string `json:"vg_attr"`
	VgCapacityByte string `json:"vg_size"`
	VgFreeByte     string `json:"vg_free"`
}

// for LV, volume or thin pool
type lvsReport struct {
	Records []lvsReportRecord `json:"report,omitempty"`
}

type lvsReportRecord struct {
	Records []lvRecord `json:"lv,omitempty"`
}

type lvRecord struct {
	LvPath        string `json:"lv_path"`
	Name          string `json:"lv_name,omitempty"`
	PoolName      string `json:"vg_name,omitempty"`
	ThinPoolName  string `json:"pool_lv,omitempty"`
	LvCapacity    string `json:"lv_size"`
	Origin        string `json:"origin,omitempty"`
	DataPercent   string `json:"data_percent,omitempty"`
	LVSnapInvalid string `json:"lv_snapshot_invalid,omitempty"`
	LVMergeFailed string `json:"lv_merge_failed,omitempty"`
	SnapPercent   string `json:"snap_percent,omitempty"`
	LVMerging     string `json:"lv_merging,omitempty"`
	LVConverting  string `json:"lv_converting,omitempty"`
	LVTime        string `json:"lv_time,omitempty"`
}

// type vgStatus struct {
// 	// number of active PVs
// 	actPVCount int
// }

type lvStatus struct {
	// disks where a LV is spread cross
	disks []string
	// state of a LV, e.g. Ready, NotReady
	state apisv1alpha1.State
}

type localDevicesArray []*apisv1alpha1.LocalDevice

func (l localDevicesArray) string() (ds string) {
	for _, device := range l {
		ds = ds + device.DevPath + ","
	}
	return strings.TrimSuffix(ds, ",")
}

type localDevicesMap map[string]*apisv1alpha1.LocalDevice

func (l localDevicesMap) string() (ds string) {
	for _, device := range l {
		ds = ds + device.DevPath + ","
	}
	return strings.TrimSuffix(ds, ",")
}

type lvmExecutor struct {
	lm      *LocalManager
	cmdExec exechelper.Executor
	logger  *log.Entry
}

func (lvm *lvmExecutor) GetPools() (map[string]*apisv1alpha1.LocalPool, error) {
	// Get LVM status
	lvmStatus, err := lvm.getLVMStatus(LVMask | VGMask | PVMask)
	if err != nil {
		lvm.logger.WithError(err).Error("Failed to query LVM stats.")
		return nil, err
	}

	pools := make(map[string]*apisv1alpha1.LocalPool)

	for vgName, vg := range lvmStatus.vgs {
		if !strings.HasPrefix(vg.Name, apisv1alpha1.PoolNamePrefix) {
			continue
		}
		totalCapacityBytes, err := utils.ConvertLVMBytesToNumeric(vg.VgCapacityByte)
		if err != nil {
			lvm.logger.WithError(err).Errorf("Failed to convert LVM bytes into int64: %s\n.", vg.VgCapacityByte)
			return nil, err
		}
		freeCapacityBytes, err := utils.ConvertLVMBytesToNumeric(vg.VgFreeByte)
		if err != nil {
			lvm.logger.WithError(err).Errorf("Failed to convert LVM bytes into int64: %s\n", vg.VgFreeByte)
			return nil, err
		}

		poolClass, poolType := getPoolClassTypeByName(vgName)
		if len(poolClass) == 0 || len(poolType) == 0 {
			lvm.logger.Debugf("Failed to passe pool class and pool name: %s\n", vgName)
		}

		// Prepare PV status
		poolPVs := lvmStatus.getPVsByVGName(vgName)
		poolDisks := make([]apisv1alpha1.LocalDevice, 0, len(poolPVs))
		for _, pv := range poolPVs {
			pvCapacity, err := utils.ConvertLVMBytesToNumeric(pv.PvSize)
			if err != nil {
				lvm.logger.WithError(err).Errorf("Failed to convert LVM byte numbers int64: %s\n.", pv.PvSize)
				return nil, err
			}
			poolDisks = append(poolDisks, apisv1alpha1.LocalDevice{
				DevPath:       pv.Name,
				CapacityBytes: pvCapacity,
				Class:         poolClass,
				State:         apisv1alpha1.DiskStateInUse,
			})
		}

		// Prepare LV status
		poolLVs := lvmStatus.getLVsByVGName(vgName)
		poolVolumes := make([]string, 0, len(poolLVs))
		for _, lv := range poolLVs {
			poolVolumes = append(poolVolumes, lv.Name)
		}

		pools[vgName] = &apisv1alpha1.LocalPool{
			Name:                     vg.Name,
			Class:                    poolClass,
			Type:                     poolType,
			TotalCapacityBytes:       int64(totalCapacityBytes),
			UsedCapacityBytes:        int64(totalCapacityBytes) - int64(freeCapacityBytes),
			FreeCapacityBytes:        int64(freeCapacityBytes),
			VolumeCapacityBytesLimit: int64(totalCapacityBytes),
			TotalVolumeCount:         apisv1alpha1.LVMVolumeMaxCount,
			UsedVolumeCount:          int64(len(poolVolumes)),
			FreeVolumeCount:          apisv1alpha1.LVMVolumeMaxCount - int64(len(poolVolumes)),
			Disks:                    poolDisks,
			Volumes:                  poolVolumes,
		}
	}

	return pools, nil
}

func (lvm *lvmExecutor) GetReplicas() (map[string]*apisv1alpha1.LocalVolumeReplica, error) {
	// TODO
	lvmStatus, err := lvm.getLVMStatus(LVMask)
	if err != nil {
		lvm.logger.WithError(err).Error("Failed to query LV stats.")
		return nil, err
	}

	replicas := make(map[string]*apisv1alpha1.LocalVolumeReplica)

	for lvName, lv := range lvmStatus.lvs {
		if !strings.HasPrefix(lv.PoolName, apisv1alpha1.PoolNamePrefix) {
			continue
		}

		capacity, err := utils.ConvertLVMBytesToNumeric(lv.LvCapacity)
		if err != nil {
			lvm.logger.WithError(err).Errorf("Failed to get replica capacity, unrecognizied params %s.", lv.LvCapacity)
			return nil, err
		}

		replicaToTest := &apisv1alpha1.LocalVolumeReplica{
			Spec: apisv1alpha1.LocalVolumeReplicaSpec{
				VolumeName: lvName,
				PoolName:   lv.PoolName,
			},
			Status: apisv1alpha1.LocalVolumeReplicaStatus{
				StoragePath:            lv.LvPath,
				DevicePath:             lv.LvPath,
				AllocatedCapacityBytes: capacity,
				Synced:                 true,
			},
		}
		replica, _ := lvm.TestVolumeReplica(replicaToTest)
		replicas[lvName] = replica
		lvm.logger.WithField("volume", lvName).Debug("Detected a LVM volume")
	}

	lvm.logger.WithField("volumes", len(replicas)).Debug("Finshed LVM volume detection")
	return replicas, nil
}

var lvmExecutorInstance *lvmExecutor

func newLVMExecutor(lm *LocalManager) *lvmExecutor {
	if lvmExecutorInstance == nil {
		lvmExecutorInstance = &lvmExecutor{
			lm:      lm,
			cmdExec: nsexecutor.New(),
			logger:  log.WithField("Module", "NodeManager/lvmExecuter"),
		}
	}
	return lvmExecutorInstance
}

func (lvm *lvmExecutor) CreateVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error) {

	// // if strip is enabled, the number of stripes should be equal to the one of active PVs in the VG.
	// // in another word, the striped volume replica should be spread across all the active PVs in the VG
	// vgStatus, err := lvm.vgdisplay(replica.Spec.PoolName)
	// if err != nil {
	// 	return nil, err
	// }
	// // if strip is disabled, the strip number should be 1
	// stripNum := 1
	// if replica.Spec.Striped {
	// 	stripNum = vgStatus.actPVCount
	// }

	options := []string{
		"--size", utils.ConvertNumericToLVMBytes(replica.Spec.RequiredCapacityBytes),
		"--stripes", fmt.Sprintf("%d", 1),
	}
	if err := lvm.lvcreate(replica.Spec.VolumeName, replica.Spec.PoolName, options); err != nil {
		if !strings.Contains(err.Error(), ErrorLocalVolumeExistsInVolumeGroup.Error()) {
			return nil, err
		}
	}

	// query current status of the replica
	record, err := lvm.lvRecord(replica.Spec.VolumeName, replica.Spec.PoolName)
	if err != nil {
		return nil, err
	}
	allocatedCapacityBytes, err := utils.ConvertLVMBytesToNumeric(record.LvCapacity)
	if err != nil {
		return nil, err
	}
	status, err := lvm.lvdisplay(record.LvPath)
	if err != nil {
		return nil, err
	}
	newReplica := replica.DeepCopy()
	newReplica.Status.AllocatedCapacityBytes = allocatedCapacityBytes
	newReplica.Status.StoragePath = record.LvPath
	newReplica.Status.DevicePath = record.LvPath
	newReplica.Status.Disks = status.disks

	return newReplica, nil
}

func (lvm *lvmExecutor) ExpandVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica, newCapacityBytes int64) (*apisv1alpha1.LocalVolumeReplica, error) {

	if replica.Status.AllocatedCapacityBytes == newCapacityBytes {
		return replica, nil
	}

	newLVMCapacityBytes := utils.NumericToLVMBytes(newCapacityBytes)

	// for compatibility
	storagePath := replica.Status.StoragePath
	if len(storagePath) == 0 {
		storagePath = replica.Status.DevicePath
	}
	if err := lvm.lvextend(storagePath, newLVMCapacityBytes, []string{}); err != nil {
		return nil, err
	}

	// query current status of the replica
	record, err := lvm.lvRecord(replica.Spec.VolumeName, replica.Spec.PoolName)
	if err != nil {
		return nil, err
	}
	allocatedCapacityBytes, err := utils.ConvertLVMBytesToNumeric(record.LvCapacity)
	if err != nil {
		return nil, err
	}
	newReplica := replica.DeepCopy()
	newReplica.Status.AllocatedCapacityBytes = allocatedCapacityBytes

	return newReplica, nil
}

func (lvm *lvmExecutor) DeleteVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error {
	// for compatibility
	storagePath := replica.Status.StoragePath
	if len(storagePath) == 0 {
		storagePath = replica.Status.DevicePath
	}

	return lvm.lvremove(storagePath, []string{})
}

func (lvm *lvmExecutor) TestVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error) {
	// for compatibility
	storagePath := replica.Status.StoragePath
	if len(storagePath) == 0 {
		storagePath = replica.Status.DevicePath
	}
	status, err := lvm.lvdisplay(storagePath)
	if err != nil {
		return nil, err
	}
	newReplica := replica.DeepCopy()
	newReplica.Status.Synced = true
	newReplica.Status.Disks = status.disks
	newReplica.Status.State = status.state

	return newReplica, nil
}

func (lvm *lvmExecutor) ExtendPools(localDevices []*apisv1alpha1.LocalDevice) (bool, error) {
	lvm.logger.Debugf("Start extending pool disk(s): %s, count: %d", localDevicesArray(localDevices).string(), len(localDevices))

	extend := false
	existingPVMap, err := lvm.getExistingPVs()
	if err != nil {
		lvm.logger.WithError(err).Error("Failed to getExistingPVs.")
		return false, err
	}

	// find out new disks which is not exist in vg(i.g LocalStorage_XXX)
	disksToBeExtends := make(map[string][]*apisv1alpha1.LocalDevice)
	for _, disk := range localDevices {
		poolName, err := getPoolNameAccordingDisk(disk)
		if err != nil {
			lvm.logger.WithError(err)
			continue
		}
		if disksToBeExtends[poolName] == nil {
			disksToBeExtends[poolName] = make([]*apisv1alpha1.LocalDevice, 0)
		}
		if _, ok := existingPVMap[disk.DevPath]; ok {
			continue
		}
		disksToBeExtends[poolName] = append(disksToBeExtends[poolName], disk)
	}

	for poolName, classifiedDisks := range disksToBeExtends {
		if len(classifiedDisks) == 0 {
			continue
		}

		lvm.logger.Debugf("Adding disk(s): %+v to pool %s", localDevicesArray(classifiedDisks).string(), poolName)
		if err := lvm.extendPool(poolName, classifiedDisks); err != nil {
			lvm.logger.WithError(err).Error("Add available disk failed.")
			return extend, err
		}
		extend = true
	}

	return extend, nil
}

func (lvm *lvmExecutor) ResizePhysicalVolumes(localDevices map[string]*apisv1alpha1.LocalDevice) error {
	lvm.logger.Debugf("Resizing pvsize, device(s): %+v, count: %d", localDevicesMap(localDevices).string(), len(localDevices))

	for _, disk := range localDevices {
		if err := lvm.pvresize(disk.DevPath); err != nil {
			lvm.logger.WithError(err).Errorf("Failed to resize pv: %+v", disk.DevPath)
			return err
		}
	}

	return nil
}

func (lvm *lvmExecutor) ConsistencyCheck(crdReplicas map[string]*apisv1alpha1.LocalVolumeReplica) {

	lvm.logger.Debug("Consistency Checking for LVM volume ...")

	replicas, err := lvm.GetReplicas()
	if err != nil {
		lvm.logger.Error("Failed to collect volume replicas info from OS")
		return
	}

	for volName, crd := range crdReplicas {
		lvm.logger.WithField("volume", volName).Debug("Checking VolumeReplica CRD")
		replica, exists := replicas[volName]
		if !exists {
			lvm.logger.WithField("volume", volName).WithError(fmt.Errorf("not found on Host")).Warning("Volume replica consistency check failed")
			continue
		}
		if crd.Status.StoragePath != replica.Status.StoragePath {
			lvm.logger.WithFields(log.Fields{
				"volume":   volName,
				"crd.path": crd.Status.StoragePath,
				"rep.path": replica.Status.StoragePath,
			}).WithError(fmt.Errorf("mismatched storage path")).Warning("Volume replica consistency check failed")
		}
		if crd.Status.AllocatedCapacityBytes != replica.Status.AllocatedCapacityBytes {
			lvm.logger.WithFields(log.Fields{
				"volume":       volName,
				"crd.capacity": crd.Status.AllocatedCapacityBytes,
				"rep.capacity": replica.Status.AllocatedCapacityBytes,
			}).WithError(fmt.Errorf("mismatched allocated capacity")).Warning("Volume replica consistency check failed")
		}
	}

	for volName, replica := range replicas {
		lvm.logger.WithField("volume", volName).Debug("Checking volume replica on Host")
		crd, exists := crdReplicas[volName]
		if !exists {
			lvm.logger.WithField("volume", volName).WithError(fmt.Errorf("not found the CRD")).Warning("Volume replica consistency check failed")
			continue
		}
		if crd.Status.StoragePath != replica.Status.StoragePath {
			lvm.logger.WithFields(log.Fields{
				"volume":   volName,
				"crd.path": crd.Status.StoragePath,
				"rep.path": replica.Status.StoragePath,
			}).WithError(fmt.Errorf("mismatched storage path")).Warning("Volume replica consistency check failed")
		}
		if crd.Status.AllocatedCapacityBytes != replica.Status.AllocatedCapacityBytes {
			lvm.logger.WithFields(log.Fields{
				"volume":       volName,
				"crd.capacity": crd.Status.AllocatedCapacityBytes,
				"rep.capacity": replica.Status.AllocatedCapacityBytes,
			}).WithError(fmt.Errorf("mismatched allocated capacity")).Warning("Volume replica consistency check failed")
		}
	}

	lvm.logger.Debug("Consistency check completed")
}

// CreateVolumeReplicaSnapshot creates a new COW volume replica snapshot
func (lvm *lvmExecutor) CreateVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) error {
	logCtx := lvm.logger.WithFields(log.Fields{
		"volumeSnapshot":        replicaSnapshot.Spec.VolumeSnapshotName,
		"volumeReplicaSnapshot": replicaSnapshot.Name,
		"sourceVolume":          replicaSnapshot.Spec.SourceVolume,
		"sourceVolumeReplica":   replicaSnapshot.Spec.SourceVolumeReplica,
	})
	logCtx.Debug("Start creating volume replica snapshot")

	existReplicas, err := lvm.GetReplicas()
	if err != nil {
		return err
	}

	if _, ok := existReplicas[replicaSnapshot.Spec.SourceVolume]; !ok {
		logCtx.WithError(err).Error("Failed to get source volume replica on host")
		return ErrReplicaNotFound
	}

	// use volume snapshot name as snapshot volume key - avoid duplicate volume replica snapshot with the same snapshot
	if err = lvm.lvSnapCreate(replicaSnapshot.Spec.VolumeSnapshotName, path.Join(replicaSnapshot.Spec.PoolName, replicaSnapshot.Spec.SourceVolume),
		replicaSnapshot.Spec.RequiredCapacityBytes); err != nil {
		lvm.logger.WithError(err).Error("Failed to create volume replica snapshot")
		return err
	}

	lvm.logger.Debugf("Volume replica snapshot created: %s", replicaSnapshot.Name)
	return nil
}

func (lvm *lvmExecutor) DeleteVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) error {
	logCtx := lvm.logger.WithFields(log.Fields{
		"volumeSnapshot":        replicaSnapshot.Spec.VolumeSnapshotName,
		"volumeReplicaSnapshot": replicaSnapshot.Name,
		"sourceVolume":          replicaSnapshot.Spec.SourceVolume,
		"sourceVolumeReplica":   replicaSnapshot.Spec.SourceVolumeReplica,
	})
	logCtx.Debug("Start deleting volume replica snapshot")

	if err := lvm.lvremove(path.Join(replicaSnapshot.Spec.PoolName, replicaSnapshot.Spec.VolumeSnapshotName), []string{}); err != nil {
		logCtx.WithError(err).Error("Failed to delete volume replica snapshot from host")
		return err
	}

	lvm.logger.Debugf("Volume replica snapshot deleted: %s", replicaSnapshot.Name)
	return nil
}

func (lvm *lvmExecutor) UpdateVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) (*apisv1alpha1.LocalVolumeReplicaSnapshotStatus, error) {
	//TODO implement me
	panic("implement me")
}

// GetVolumeReplicaSnapshot returns a volume replica snapshot attribute including state and creation time
func (lvm *lvmExecutor) GetVolumeReplicaSnapshot(replicaSnapshot *apisv1alpha1.LocalVolumeReplicaSnapshot) (*apisv1alpha1.LocalVolumeReplicaSnapshotStatus, error) {
	logCtx := lvm.logger.WithFields(log.Fields{
		"volumeSnapshot":        replicaSnapshot.Spec.VolumeSnapshotName,
		"volumeReplicaSnapshot": replicaSnapshot.Name,
		"sourceVolume":          replicaSnapshot.Spec.SourceVolume,
		"sourceVolumeReplica":   replicaSnapshot.Spec.SourceVolumeReplica,
	})
	logCtx.Debug("Getting a volume replica snapshot")

	lvmState, err := lvm.getLVMStatus(LVMask)
	if err != nil {
		logCtx.WithError(err).Error("Failed to query LV stats")
		return nil, err
	}

	actualSnapshotStatus := apisv1alpha1.LocalVolumeReplicaSnapshotStatus{}
	snapshotVolume, ok := lvmState.lvs[replicaSnapshot.Spec.VolumeSnapshotName]
	if !ok {
		return nil, ErrorSnapshotNotFound
	}

	capacity, err := utils.ConvertLVMBytesToNumeric(snapshotVolume.LvCapacity)
	if err != nil {
		logCtx.WithError(err).Errorf("Failed to get replica snapshot capacity, unrecognizied params %s", snapshotVolume.LvCapacity)
		return nil, err
	}

	// parse lv create time to UTC time
	cTime, err := time.Parse(LVMTimeLayout, snapshotVolume.LVTime)
	if err != nil {
		logCtx.WithError(err).Errorf("Failed to convert replica snapshot creation time to UTC Time, unrecognizied params %s", snapshotVolume.LVTime)
		return nil, err
	}

	newTime := metav1.NewTime(cTime.Local())
	actualSnapshotStatus.CreationTime = newTime.DeepCopy()
	actualSnapshotStatus.AllocatedCapacityBytes = capacity
	actualSnapshotStatus.Attribute.Invalid = len(snapshotVolume.LVSnapInvalid) > 0 && snapshotVolume.LVSnapInvalid != LVMUnknownStatus
	actualSnapshotStatus.Attribute.Merging = len(snapshotVolume.LVMerging) > 0
	if actualSnapshotStatus.Attribute.Invalid {
		actualSnapshotStatus.Message = fmt.Sprintf("snapshot is invalid")
		actualSnapshotStatus.State = apisv1alpha1.VolumeStateNotReady
	} else {
		actualSnapshotStatus.State = apisv1alpha1.VolumeStateReady
	}

	return &actualSnapshotStatus, nil
}

// RollbackVolumeReplicaSnapshot rollback snapshot to the source volume
func (lvm *lvmExecutor) RollbackVolumeReplicaSnapshot(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	logCtx := lvm.logger.WithFields(log.Fields{
		"sourceVolume":                snapshotRestore.Spec.SourceVolumeSnapshot,
		"volumeSnapshot":              snapshotRestore.Spec.VolumeSnapshotRestore,
		"targetVolume":                snapshotRestore.Spec.TargetVolume,
		"sourceVolumeReplicaSnapshot": snapshotRestore.Spec.SourceVolumeReplicaSnapshot,
	})
	logCtx.Debug("Rolling back a volume replica snapshot")

	replicaSnapshot := &apisv1alpha1.LocalVolumeReplicaSnapshot{}
	if err := lvm.lm.apiClient.Get(context.Background(), client.ObjectKey{Name: snapshotRestore.Spec.SourceVolumeReplicaSnapshot}, replicaSnapshot); err != nil {
		logCtx.WithError(err).Error("Failed to get VolumeReplicaSnapshot")
		return err
	}

	var options = []string{"--merge", fmt.Sprintf("%s/%s", replicaSnapshot.Spec.PoolName, replicaSnapshot.Spec.VolumeSnapshotName)}

	// lvconvert --merge LocalStorage_PoolHDD/snapshot-name
	if err := lvm.lvconvert(options...); err != nil {
		logCtx.WithError(err).Error("Failed to convert snapshot to source volume")
		return err
	}

	logCtx.Error("Successfully to convert snapshot to source volume")
	return nil
}

// RestoreVolumeReplicaSnapshot restore snapshot to a new volume
func (lvm *lvmExecutor) RestoreVolumeReplicaSnapshot(snapshotRestore *apisv1alpha1.LocalVolumeReplicaSnapshotRestore) error {
	panic("have not implemented restore")
	return nil
}

func (lvm *lvmExecutor) getExistingPVs() (map[string]struct{}, error) {
	existingPVsMap := make(map[string]struct{})

	vgsReport, err := lvm.vgs()
	if err != nil {
		lvm.logger.WithError(err).Error("Failed to query VG.")
		return existingPVsMap, err
	}

	poolNamesMap := make(map[string]struct{})
	// Find out if pool exists
	for _, vgsReportRecords := range vgsReport.Records {
		for _, vgRecord := range vgsReportRecords.Records {
			poolNamesMap[vgRecord.Name] = struct{}{}
		}
	}

	// Fetch pvs info
	pvsReport, err := lvm.pvs()
	if err != nil {
		lvm.logger.WithError(err).Error("Failed to query PV.")
		return existingPVsMap, err
	}

	for _, pvsReportRecords := range pvsReport.Records {
		for _, pvRecord := range pvsReportRecords.Records {
			if _, ok := poolNamesMap[pvRecord.PoolName]; ok {
				existingPVsMap[pvRecord.Name] = struct{}{}
			}
		}
	}

	return existingPVsMap, nil
}

func (lvm *lvmExecutor) extendPool(poolName string, disks []*apisv1alpha1.LocalDevice) error {
	if len(disks) == 0 {
		lvm.logger.Info("Empty disk list given.")
		return nil
	}

	vgsReport, err := lvm.vgs()
	if err != nil {
		lvm.logger.WithError(err).Error("Failed to query VG.")
		return err
	}

	diskDevPaths := make([]string, 0, len(disks))
	for _, disk := range disks {
		diskDevPaths = append(diskDevPaths, disk.DevPath)
	}

	// Find out if pool exists
	for _, vgsReportRecords := range vgsReport.Records {
		for _, vgRecord := range vgsReportRecords.Records {
			if vgRecord.Name == poolName {
				lvm.logger.Debugf("Pool %s already exists, adding disks to pool..", poolName)
				err := lvm.vgextend(poolName, diskDevPaths)
				if err != nil {
					lvm.logger.WithError(err).Error("Failed to extend pool when adding available disk to default pool.")
					return err
				}
				return nil
			}
		}
	}

	lvm.logger.Debugf("Pool %s not exists, creating...", poolName)
	if err := lvm.vgcreate(poolName, diskDevPaths); err != nil {
		lvm.logger.WithError(err).Errorf("Failed to create pool %s.", poolName)
		return err
	}

	return nil
}

// ======== Helpers ============

func (lvm *lvmExecutor) vgcreate(vgName string, pvs []string) error {
	// when add device into a VG by vgcreate, PV of device will be created automatically
	params := exechelper.ExecParams{
		CmdName: "vgcreate",
		CmdArgs: append([]string{vgName, "-y"}, pvs...),
	}
	res := lvm.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error
}

func (lvm *lvmExecutor) vgextend(vgName string, pvs []string) error {
	// when add device into a VG by vgextend, PV of device will be created automatically
	params := exechelper.ExecParams{
		CmdName: "vgextend",
		CmdArgs: append([]string{vgName, "-y"}, pvs...),
	}
	res := lvm.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error

}

// func (lvm *lvmExecutor) vgdisplay(vgName string) (*vgStatus, error) {
// 	params := exechelper.ExecParams{
// 		CmdName: "vgdisplay",
// 		CmdArgs: []string{vgName},
// 	}
// 	res := lvm.cmdExec.RunCommand(params)
// 	if res.ExitCode != 0 {
// 		return nil, res.Error
// 	}
// 	status := vgStatus{actPVCount: -1}
// 	for _, line := range strings.Split(res.OutBuf.String(), "\n") {
// 		str := strings.Trim(line, " ")
// 		if len(str) == 0 {
// 			continue
// 		}
// 		if strings.Contains(str, "Act PV") {
// 			cnt, err := strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(str, "Act PV", "")))
// 			if err != nil {
// 				return nil, err
// 			}
// 			status.actPVCount = cnt
// 		}
// 	}
// 	return &status, nil
// }

func (lvm *lvmExecutor) lvcreate(lvName string, vgName string, options []string) error {
	params := exechelper.ExecParams{
		CmdName: "lvcreate",
		CmdArgs: append(options, vgName, "-n", lvName, "-y"),
	}
	res := lvm.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error
}

func (lvm *lvmExecutor) lvSnapCreate(snapName string, sourceVolumePath string, snapSize int64, options ...string) error {
	snapSizeStr := utils.ConvertNumericToLVMBytes(snapSize)
	params := exechelper.ExecParams{
		CmdName: "lvcreate",
		CmdArgs: append(options,
			"--snapshot",
			"--name", snapName,
			// only supported read-only snapshots. By default, lvm snapshot is readable and writeable
			"--permission", "r",
			"--size", snapSizeStr,
			sourceVolumePath,
		),
	}
	res := lvm.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error
}

func (lvm *lvmExecutor) lvRecord(lvName string, vgName string) (*lvRecord, error) {
	lvsReport, err := lvm.lvs()
	if err != nil {
		return nil, err
	}
	for _, lvsReportRecords := range lvsReport.Records {
		for _, lvRecord := range lvsReportRecords.Records {
			if lvRecord.Name == lvName && lvRecord.PoolName == vgName {
				return &lvRecord, nil
			}
		}
	}
	return nil, fmt.Errorf("not found")
}

func (lvm *lvmExecutor) lvdisplay(lvPath string) (*lvStatus, error) {
	params := exechelper.ExecParams{
		CmdName: "lvdisplay",
		CmdArgs: []string{"-m", lvPath},
	}
	res := lvm.cmdExec.RunCommand(params)
	if res.ExitCode != 0 {
		return nil, res.Error
	}

	status := lvStatus{
		disks: []string{},
		state: apisv1alpha1.VolumeStateNotReady,
	}

	for _, line := range strings.Split(res.OutBuf.String(), "\n") {
		str := strings.TrimSpace(line)
		if len(str) == 0 {
			continue
		}
		if strings.Contains(str, "Physical volume") {
			disk := strings.TrimSpace(strings.ReplaceAll(str, "Physical volume", ""))
			status.disks = append(status.disks, disk)
			continue
		}
		if strings.Contains(str, "LV Status") {
			stateStr := strings.TrimSpace(strings.ReplaceAll(str, "LV Status", ""))
			lvm.logger.WithFields(log.Fields{"lvPath": lvPath, "status": stateStr}).Debug("lv status")
			if stateStr == "NOT available" {
				continue
			}
			if stateStr == "available" {
				status.state = apisv1alpha1.VolumeStateReady
			}
		}
	}
	return &status, nil
}

func (lvm *lvmExecutor) pvs() (*pvsReport, error) {
	params := exechelper.ExecParams{
		CmdName: "pvs",
		CmdArgs: []string{"--reportformat", "json", "--units", "B"},
	}
	res := lvm.cmdExec.RunCommand(params)
	if res.ExitCode != 0 {
		lvm.logger.WithError(res.Error).Error("Failed to discover PVs")
		return nil, res.Error
	}
	report := &pvsReport{}

	if err := json.Unmarshal(res.OutBuf.Bytes(), report); err != nil {
		lvm.logger.WithError(err).Error("Failed to parse PVs output")
		return nil, err
	}
	return report, nil
}

func (lvm *lvmExecutor) vgs() (*vgsReport, error) {
	params := exechelper.ExecParams{
		CmdName: "vgs",
		CmdArgs: []string{"--reportformat", "json", "--units", "B"},
	}
	res := lvm.cmdExec.RunCommand(params)
	if res.ExitCode != 0 {
		lvm.logger.WithError(res.Error).Error("Failed to discover VGs")
		return nil, res.Error
	}
	report := &vgsReport{}

	if err := json.Unmarshal(res.OutBuf.Bytes(), report); err != nil {
		lvm.logger.WithError(err).Error("Failed to parse VGs output")
		return nil, err
	}
	return report, nil
}

func (lvm *lvmExecutor) lvs() (*lvsReport, error) {
	params := exechelper.ExecParams{
		CmdName: "lvs",
		CmdArgs: []string{"-a", "-o", "lv_path,lv_name,vg_name,lv_attr,lv_size,pool_lv,origin,data_percent,metadata_percent,move_pv,mirror_log,copy_percent,convert_lv," +
			"lv_snapshot_invalid,lv_merge_failed,snap_percent,lv_device_open,lv_merging,lv_converting,lv_time", "--reportformat", "json", "--units", "B"},
	}
	res := lvm.cmdExec.RunCommand(params)
	if res.ExitCode != 0 {
		lvm.logger.WithError(res.Error).Error("Failed to discover LVs")
		return nil, res.Error
	}
	report := &lvsReport{}

	if err := json.Unmarshal(res.OutBuf.Bytes(), report); err != nil {
		lvm.logger.WithError(err).Error("Failed to parse LVs output")
		return nil, err
	}
	return report, nil
}

func (lvm *lvmExecutor) lvremove(lvPath string, options []string) error {
	params := exechelper.ExecParams{
		CmdName: "lvremove",
		CmdArgs: []string{lvPath, "-y"},
	}
	params.CmdArgs = append(params.CmdArgs, options...)
	res := lvm.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error
}

func (lvm *lvmExecutor) lvextend(lvPath string, newCapacityBytes int64, options []string) error {
	params := exechelper.ExecParams{
		CmdName: "lvextend",
		CmdArgs: []string{"--size", fmt.Sprintf("%db", newCapacityBytes), lvPath, "-y"},
	}
	params.CmdArgs = append(params.CmdArgs, options...)
	res := lvm.cmdExec.RunCommand(params)
	errcontent := res.ErrBuf.String()
	if res.ExitCode == 0 || strings.Contains(errcontent, "matches existing size") {
		return nil
	}
	return res.Error
}

func (lvm *lvmExecutor) pvresize(pv string, options ...string) error {
	params := exechelper.ExecParams{
		CmdName: "pvresize",
		CmdArgs: []string{pv, "-y"},
	}
	params.CmdArgs = append(params.CmdArgs, options...)
	res := lvm.cmdExec.RunCommand(params)
	if res.ExitCode == 0 {
		return nil
	}
	return res.Error
}

type lvmStatus struct {
	lvs    map[string]lvRecord
	pvs    map[string]pvRecord
	vgs    map[string]vgRecord
	logger *log.Entry
}

func (ls *lvmStatus) getPVsByVGName(vgname string) []pvRecord {
	queryRst := make([]pvRecord, 0)
	for _, pv := range ls.pvs {
		if pv.PoolName == vgname {
			queryRst = append(queryRst, pv)
		}
	}
	return queryRst
}

func (ls *lvmStatus) getLVsByVGName(vgname string) []lvRecord {
	queryRst := make([]lvRecord, 0)
	for _, lv := range ls.lvs {
		if lv.PoolName == vgname {
			queryRst = append(queryRst, lv)
		}
	}
	return queryRst
}

func (lvm *lvmExecutor) getLVMStatus(masks int) (*lvmStatus, error) {
	status := &lvmStatus{
		lvs:    make(map[string]lvRecord),
		pvs:    make(map[string]pvRecord),
		vgs:    make(map[string]vgRecord),
		logger: log.WithField("Module", "NodeManager/lvmExecuter"),
	}

	if masks&LVMask == LVMask {
		// Fetch lvs info
		lvsReport, err := lvm.lvs()
		if err != nil {
			lvm.logger.WithError(err).Error("Failed to get LVM (lv) status.")
			return nil, err
		}
		for _, lvsReportRecords := range lvsReport.Records {
			for _, lvRecord := range lvsReportRecords.Records {
				// for merging snapshot: "lv_name":"[snapcontent-49246190-f939-4e17-912a-394ddc088299]"
				if strings.HasPrefix(lvRecord.Name, "[") && strings.HasSuffix(lvRecord.Name, "]") {
					lvRecord.Name = strings.TrimPrefix(lvRecord.Name, "[")
					lvRecord.Name = strings.TrimSuffix(lvRecord.Name, "]")
				}
				status.lvs[lvRecord.Name] = lvRecord
			}
		}
	}

	if masks&VGMask == VGMask {
		// Fetch vgs info
		vgsReport, err := lvm.vgs()
		if err != nil {
			lvm.logger.WithError(err).Error("Failed to get LVM (vg) status.")
			return nil, err
		}
		for _, vgsReportRecords := range vgsReport.Records {
			for _, vgRecord := range vgsReportRecords.Records {
				status.vgs[vgRecord.Name] = vgRecord
			}
		}
	}

	if masks&PVMask == PVMask {
		// Fetch pvs info
		pvsReport, err := lvm.pvs()
		if err != nil {
			lvm.logger.WithError(err).Error("Failed to get LVM (pv) status.")
			return nil, err
		}
		for _, pvsReportRecords := range pvsReport.Records {
			for _, pvRecord := range pvsReportRecords.Records {
				status.pvs[pvRecord.Name] = pvRecord
			}
		}
	}

	return status, nil
}

func (lvm *lvmExecutor) lvconvert(options ...string) error {
	params := exechelper.ExecParams{
		CmdName: "lvconvert",
		CmdArgs: options,
	}
	res := lvm.cmdExec.RunCommand(params)
	if res.ExitCode != 0 {
		lvm.logger.WithError(res.Error).Error("Failed to do lvconvert")
		return res.Error
	}

	return nil
}
