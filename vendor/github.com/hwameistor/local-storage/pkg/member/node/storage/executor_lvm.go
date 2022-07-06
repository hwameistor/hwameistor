package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/exechelper"
	"github.com/hwameistor/local-storage/pkg/exechelper/nsexecutor"
	"github.com/hwameistor/local-storage/pkg/utils"

	log "github.com/sirupsen/logrus"
)

// consts
const (
	LVMask = 1
	VGMask = 1 << 1
	PVMask = 1 << 2
)

// variables
var (
	ErrNotLVMByteNum = errors.New("LVM byte format unrecognised")
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
	LvPath       string `json:"lv_path"`
	Name         string `json:"lv_name,omitempty"`
	PoolName     string `json:"vg_name,omitempty"`
	ThinPoolName string `json:"pool_lv,omitempty"`
	LvCapacity   string `json:"lv_size"`
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

type lvmExecutor struct {
	lm      *LocalManager
	cmdExec exechelper.Executor
	logger  *log.Entry
}

func (lvm *lvmExecutor) ExtendPoolsInfo(disks map[string]*apisv1alpha1.LocalDisk) (map[string]*apisv1alpha1.LocalPool, error) {
	oldRegistryDisks := lvm.lm.registry.Disks()
	localDisks := mergeRegistryDiskMap(oldRegistryDisks, disks)

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
		pvRecords := lvmStatus.getPVsByVGName(vgName)
		disks := make([]apisv1alpha1.LocalDisk, 0, len(pvRecords))

		for _, pv := range pvRecords {
			if _, exists := localDisks[pv.Name]; !exists {
				continue
			}
			pvcap, err := utils.ConvertLVMBytesToNumeric(pv.PvSize)
			if err != nil {
				lvm.logger.WithError(err).Errorf("Failed to convert LVM byte numbers int64: %s\n.", pv.PvSize)
				return nil, err
			}
			typedPv := &apisv1alpha1.LocalDisk{
				DevPath:       pv.Name,
				Class:         localDisks[pv.Name].Class,
				CapacityBytes: pvcap,
				State:         localDisks[pv.Name].State,
			}
			disks = append(disks, *typedPv)
		}

		// Prepare LV status
		lvRecords := lvmStatus.getLVsByVGName(vgName)
		volumes := make([]string, 0, len(lvRecords))
		for _, lv := range lvRecords {
			volumes = append(volumes, lv.Name)
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
			UsedVolumeCount:          int64(len(volumes)),
			FreeVolumeCount:          apisv1alpha1.LVMVolumeMaxCount - int64(len(volumes)),
			Disks:                    disks,
			Volumes:                  volumes,
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

func (lvm *lvmExecutor) ExtendPools(availableLocalDisks []*apisv1alpha1.LocalDisk) error {
	lvm.logger.Debugf("Adding available disk %+v, count: %d\n.", availableLocalDisks, len(availableLocalDisks))

	existingPVMap, err := lvm.getExistingPVs()
	if err != nil {
		lvm.logger.WithError(err).Error("Failed to getExistingPVs.")
		return err
	}

	disksToBeExtends := make(map[string][]*apisv1alpha1.LocalDisk)
	for _, disk := range availableLocalDisks {
		poolName, err := getPoolNameAccordingDisk(disk)
		if err != nil {
			lvm.logger.WithError(err)
			continue
		}
		if disksToBeExtends[poolName] == nil {
			disksToBeExtends[poolName] = make([]*apisv1alpha1.LocalDisk, 0, len(availableLocalDisks))
		}
		if _, ok := existingPVMap[disk.DevPath]; ok {
			continue
		}
		disksToBeExtends[poolName] = append(disksToBeExtends[poolName], disk)
	}

	for poolName, classifiedDisks := range disksToBeExtends {
		if err := lvm.extendPool(poolName, classifiedDisks); err != nil {
			lvm.logger.WithError(err).Error("Add available disk failed.")
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

func (lvm *lvmExecutor) extendPool(poolName string, disks []*apisv1alpha1.LocalDisk) error {
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
		CmdArgs: []string{"-o", "lv_path,lv_name,vg_name,lv_attr,lv_size,pool_lv,origin,data_percent,metadata_percent,move_pv,mirror_log,copy_percent,convert_lv", "--reportformat", "json", "--units", "B"},
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
