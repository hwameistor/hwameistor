package localdiskvolume

import (
	"context"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/node/registry"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/node/volume"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/udev"
	"path"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	lscsi "github.com/hwameistor/hwameistor/pkg/local-storage/member/csi"
)

const (
	// LocalDiskFinalizer for the LocalDiskVolume CR
	LocalDiskFinalizer string = "localdisk.hwameistor.io/finalizer"
)

// DiskVolumeHandler
type DiskVolumeHandler struct {
	client.Client
	record.EventRecorder
	Ldv          *v1alpha1.LocalDiskVolume
	hostVM       volume.Manager
	hostRegistry registry.Manager
	mounter      lscsi.Mounter
}

// NewLocalDiskHandler
func NewLocalDiskVolumeHandler(cli client.Client, recorder record.EventRecorder) *DiskVolumeHandler {
	logger := log.WithField("Module", "CSIPlugin")
	return &DiskVolumeHandler{
		Client:        cli,
		EventRecorder: recorder,
		mounter:       lscsi.NewLinuxMounter(logger),
		hostVM:        volume.New(),
		hostRegistry:  registry.New(),
	}
}

func (v *DiskVolumeHandler) ReconcileCreated() (reconcile.Result, error) {
	volumeName := v.Ldv.Name
	volumeType := v.Ldv.Spec.DiskType
	selectedDisk := ""
	// devPath is unreliable， use localdisk to find the by-path or by-id path
	for _, links := range v.Ldv.Status.DevLinks {
		for _, linkName := range links {
			linkName = strings.Split(linkName, "/")[len(strings.Split(linkName, "/"))-1]
			if v.hostRegistry.DiskSymbolLinkExist(linkName) {
				selectedDisk = linkName
			}
		}
	}
	return reconcile.Result{}, v.hostVM.CreateVolume(volumeName, types.GetLocalDiskPoolName(volumeType), selectedDisk)
}

func (v *DiskVolumeHandler) ReconcileMount() (reconcile.Result, error) {
	var err error
	var result reconcile.Result
	var volPath = v.GetVolumePath()
	var mountPoints = v.GetMountPoints()

	if volPath == "" || len(mountPoints) == 0 {
		log.Infof("DevPath or MountPoints is empty, no operation here")
		return result, nil
	}

	// Mount RawBlock or FileSystem Volumes
	for _, mountPoint := range mountPoints {
		if mountPoint.Phase != v1alpha1.MountPointToBeMounted {
			continue
		}

		// check if the volume can mount safely, more details see issue #1116
		if _, err = v.CanSafelyMount(); err != nil {
			log.WithError(err).Errorf("Volume %s can't be safely mount to %s", volPath, mountPoint.TargetPath)
			result.Requeue = true
			continue
		}

		switch mountPoint.VolumeCap.AccessType {
		case v1alpha1.VolumeCapability_AccessType_Mount:
			err = v.MountFileSystem(volPath, mountPoint.TargetPath, mountPoint.FsTye, mountPoint.MountOptions...)
		case v1alpha1.VolumeCapability_AccessType_Block:
			err = v.MountRawBlock(volPath, mountPoint.TargetPath)
		default:
			// record and skip this mountpoint
			v.RecordEvent(corev1.EventTypeWarning, "ValidAccessType", "AccessType of MountPoint %s "+
				"is invalid, ignore it", mountPoint)
		}

		if err != nil {
			log.WithError(err).Errorf("Failed to mount %s to %s", volPath, mountPoint.TargetPath)
			result.Requeue = true
			continue
		}
		v.UpdateMountPointPhase(mountPoint.TargetPath, v1alpha1.MountPointMounted)
		// once a volume is attached success, the disk will be wiped when volume is deleted
		v.SetCanWipe(true)
	}

	if !result.Requeue {
		v.SetupVolumeStatus(v1alpha1.VolumeStateReady)
	}

	return result, v.UpdateLocalDiskVolume()
}

func (v *DiskVolumeHandler) ReconcileUnmount() (reconcile.Result, error) {
	var err error
	var result reconcile.Result
	var mountPoints = v.GetMountPoints()

	for _, mountPoint := range mountPoints {
		if mountPoint.Phase == v1alpha1.MountPointToBeUnMount {
			if err = v.UnMount(mountPoint.TargetPath); err != nil {
				log.WithError(err).Errorf("Failed to unmount %s", mountPoint.TargetPath)
				result.Requeue = true
				continue
			}

			v.RemoveMountPoint(mountPoint.TargetPath)
		}
	}
	if !result.Requeue {
		// Considering that a disk will be mounted to multiple pods,
		// if a mount point is processed, the entire LocalDiskVolume will be setup to the Ready state
		v.SetupVolumeStatus(v1alpha1.VolumeStateReady)
	}

	return result, v.UpdateLocalDiskVolume()
}

func (v *DiskVolumeHandler) ReconcileToBeDeleted() (reconcile.Result, error) {
	var result reconcile.Result
	var mountPoints = v.GetMountPoints()

	if len(mountPoints) > 0 {
		log.Infof("Volume %s remains %d mountpoint, no operation here",
			v.Ldv.Name, len(mountPoints))
		result.Requeue = true
		return result, nil
	}

	return result, v.DeleteLocalDiskVolume()
}

func (v *DiskVolumeHandler) ReconcileDeleted() (reconcile.Result, error) {
	// 1. delete volume
	if err := v.hostVM.DeleteVolume(v.Ldv.Name, types.GetLocalDiskPoolName(v.Ldv.Spec.DiskType)); err != nil {
		return reconcile.Result{}, err
	}

	// 2. wipe disk
	if err := v.WipeDisk(); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, v.Delete(context.Background(), v.Ldv)
}

// WipeDisk use wipefs to wipe disk
func (v *DiskVolumeHandler) WipeDisk() error {
	logCtx := log.Fields{"volume": v.Ldv.GetName(), "localdisk": v.GetBoundDisk(), "devName": v.GetDevPath()}
	if !v.GetCanWipe() {
		log.WithFields(logCtx).Debug("disk will not be wiped")
		return nil
	}

	wipefs := fmt.Sprintf("wipefs -af %s", v.GetDevPath())
	log.WithFields(logCtx).Debugf("disk will be wiped by cmd %s", wipefs)

	if _, err := utils.Bash(wipefs); err != nil {
		log.WithFields(logCtx).WithError(err).Error("failed to wipe disk")
		return err
	}
	log.WithFields(logCtx).Debug("disk wipe success")
	return nil
}

func (v *DiskVolumeHandler) GetLocalDiskVolume(key client.ObjectKey) (volume *v1alpha1.LocalDiskVolume, err error) {
	volume = &v1alpha1.LocalDiskVolume{}
	err = v.Get(context.Background(), key, volume)
	return
}

func (v *DiskVolumeHandler) RecordEvent(eventtype, reason, messageFmt string, args ...interface{}) {
	v.EventRecorder.Eventf(v.Ldv, eventtype, reason, messageFmt, args...)
}

func (v *DiskVolumeHandler) UpdateLocalDiskVolume() error {
	return v.Update(context.Background(), v.Ldv)
}

func (v *DiskVolumeHandler) DeleteLocalDiskVolume() error {
	v.SetupVolumeStatus(v1alpha1.VolumeStateDeleted)
	return v.UpdateLocalDiskVolume()
}

func (v *DiskVolumeHandler) RemoveFinalizers() error {
	// range all finalizers and notify all owners
	v.Ldv.Finalizers = nil
	return nil
}

func (v *DiskVolumeHandler) AddFinalizers(finalizer []string) {
	v.Ldv.Finalizers = finalizer
	//return
}

func (v *DiskVolumeHandler) GetMountPoints() []v1alpha1.MountPoint {
	return v.Ldv.Status.MountPoints
}

func (v *DiskVolumeHandler) GetDevPath() string {
	return v.Ldv.Status.DevPath
}

func (v *DiskVolumeHandler) GetCanWipe() bool {
	return v.Ldv.Spec.CanWipe
}

func (v *DiskVolumeHandler) SetCanWipe(canWipe bool) {
	v.Ldv.Spec.CanWipe = canWipe
}

func (v *DiskVolumeHandler) GetVolumePath() string {
	return v.Ldv.Status.VolumePath
}

func (v *DiskVolumeHandler) MountRawBlock(devPath, mountPoint string) error {
	// fixme: should check mount points again when do mount
	// mount twice will cause error
	if err := v.mounter.MountRawBlock(devPath, mountPoint); err != nil {
		v.RecordEvent(corev1.EventTypeWarning, "MountFailed", "Failed to mount block %s to %s due to err: %v",
			devPath, mountPoint, err)
		return err
	}

	v.RecordEvent(corev1.EventTypeNormal, "MountSuccess", "MountRawBlock %s to %s successfully",
		devPath, mountPoint)
	return nil
}

func (v *DiskVolumeHandler) MountFileSystem(devPath, mountPoint, fsType string, options ...string) error {
	// fixme: should check mount points again when do mount
	// mount twice will cause error
	if err := v.mounter.FormatAndMount(devPath, mountPoint, fsType, options); err != nil {
		v.RecordEvent(corev1.EventTypeWarning, "MountFailed", "Failed to mount filesystem %s to %s due to err: %v",
			devPath, mountPoint, err)
		return err
	}

	v.RecordEvent(corev1.EventTypeNormal, "MountSuccess", "MountFileSystem %s to %s successfully",
		devPath, mountPoint)
	return nil
}

func (v *DiskVolumeHandler) UnMount(mountPoint string) error {
	if err := v.mounter.Unmount(mountPoint); err != nil {
		// fixme: need consider raw block
		if !v.IsDevMountPoint(mountPoint) {
			v.RecordEvent(corev1.EventTypeWarning, "UnMountSuccess", "Unmount skipped due to mountpoint %s is empty or not mounted by disk %v",
				mountPoint, v.Ldv.Status.DevLinks)
			return nil
		}

		v.RecordEvent(corev1.EventTypeWarning, "UnMountFailed", "Failed to umount %s due to err: %v",
			mountPoint, err)
		return err
	}

	v.RecordEvent(corev1.EventTypeNormal, "UnMountSuccess", "Unmount %s successfully",
		mountPoint)
	return nil
}

// IsDevMountPoint judge if this mountpoint is mounted by the dev
func (v *DiskVolumeHandler) IsDevMountPoint(mountPoint string) bool {
	for _, p := range v.mounter.GetDeviceMountPoints(v.Ldv.Status.DevPath) {
		if p == mountPoint {
			return true
		}
	}
	return false
}

func (v *DiskVolumeHandler) VolumeState() v1alpha1.State {
	return v.Ldv.Status.State
}

func (v *DiskVolumeHandler) ExistMountPoint(targetPath string) bool {
	for _, mountPoint := range v.GetMountPoints() {
		if mountPoint.TargetPath == targetPath {
			return true
		}
	}

	return false
}

func (v *DiskVolumeHandler) AppendMountPoint(targetPath string, volCap *csi.VolumeCapability) {
	mountPoint := v1alpha1.MountPoint{TargetPath: targetPath, Phase: v1alpha1.MountPointToBeMounted}
	switch volCap.AccessType.(type) {
	case *csi.VolumeCapability_Block:
		mountPoint.VolumeCap = v1alpha1.VolumeCapability{AccessType: v1alpha1.VolumeCapability_AccessType_Block}
	case *csi.VolumeCapability_Mount:
		mountPoint.FsTye = volCap.GetMount().FsType
		mountPoint.MountOptions = volCap.GetMount().MountFlags
		mountPoint.VolumeCap = v1alpha1.VolumeCapability{AccessType: v1alpha1.VolumeCapability_AccessType_Mount}
	}
	v.Ldv.Status.MountPoints = append(v.Ldv.Status.MountPoints, mountPoint)
}

func (v *DiskVolumeHandler) RemoveMountPoint(targetPath string) {
	for i, mountPoint := range v.GetMountPoints() {
		if mountPoint.TargetPath == targetPath {
			v.Ldv.Status.MountPoints = append(v.Ldv.Status.MountPoints[:i], v.Ldv.Status.MountPoints[i+1:]...)
			return
		}
	}
}

func (v *DiskVolumeHandler) UpdateMountPointPhase(targetPath string, phase v1alpha1.State) {
	for i, mountPoint := range v.GetMountPoints() {
		if mountPoint.TargetPath == targetPath {
			v.Ldv.Status.MountPoints[i].Phase = phase
			return
		}
	}
}

// UpdateDevPathAccordingVolume devPath can change after machine restart, so update it here according to the volume
func (v *DiskVolumeHandler) UpdateDevPathAccordingVolume() {
	if vol := v.hostRegistry.GetVolumeByName(v.Ldv.Name); vol != nil {
		v.Ldv.Status.DevPath = vol.AttachPath
		return
	}
}

// WaitVolumeReady wait LocalDiskVolume Ready
func (v *DiskVolumeHandler) WaitVolumeReady(ctx context.Context) error {
	if _, ok := ctx.Deadline(); !ok {
		return fmt.Errorf("no deadline is set")
	}

	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context error occured when wait volume ready: %v", ctx.Err())
		case <-timer.C:
		}

		if err := v.RefreshVolume(); err != nil {
			return err
		}
		if v.VolumeState() == v1alpha1.VolumeStateReady {
			return nil
		}
		timer.Reset(1 * time.Second)
	}
}

func (v *DiskVolumeHandler) WaitVolume(ctx context.Context, state v1alpha1.State) error {
	if _, ok := ctx.Deadline(); !ok {
		return fmt.Errorf("no deadline is set")
	}

	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context error occured when wait volume ready: %v", ctx.Err())
		case <-timer.C:
		}

		if err := v.RefreshVolume(); err != nil {
			return err
		}
		if v.VolumeState() == state {
			return nil
		}
		timer.Reset(1 * time.Second)
	}
}

// WaitVolumeUnmounted wait a special mountpoint is unmounted
func (v *DiskVolumeHandler) WaitVolumeUnmounted(ctx context.Context, mountPoint string) error {
	if _, ok := ctx.Deadline(); !ok {
		return fmt.Errorf("no deadline is set")
	}

	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context error occured when wait volume ready: %v", ctx.Err())
		case <-timer.C:
		}

		if err := v.RefreshVolume(); err != nil {
			return err
		}
		if v.VolumeState() == v1alpha1.VolumeStateReady && !v.ExistMountPoint(mountPoint) {
			return nil
		}
		timer.Reset(1 * time.Second)
	}
}

func (v *DiskVolumeHandler) RefreshVolume() error {
	newVolume, err := v.GetLocalDiskVolume(client.ObjectKey{
		Namespace: v.Ldv.GetNamespace(),
		Name:      v.Ldv.GetName()})
	if err != nil {
		return err
	}

	v.For(newVolume)
	return nil
}

func (v *DiskVolumeHandler) SetupVolumeStatus(status v1alpha1.State) {
	v.Ldv.Status.State = status
}

func (v *DiskVolumeHandler) CheckFinalizers() error {
	finalizers := v.Ldv.GetFinalizers()
	_, ok := utils.StrFind(finalizers, LocalDiskFinalizer)
	if ok {
		return nil
	}

	// volume is ready for delete, don't append finalizer
	if v.VolumeState() == v1alpha1.VolumeStateDeleted {
		return nil
	}

	// add localdisk finalizers to prevent resource deleted by mistake
	finalizers = append(finalizers, LocalDiskFinalizer)
	v.AddFinalizers(finalizers)
	if err := v.UpdateLocalDiskVolume(); err != nil {
		return err
	}

	return v.RefreshVolume()
}

func (v *DiskVolumeHandler) GetBoundDisk() string {
	return v.Ldv.Status.LocalDiskName
}

func (v *DiskVolumeHandler) For(volume *v1alpha1.LocalDiskVolume) {
	v.Ldv = volume
}

// CanSafelyMount returns true if the volume is safe to be mounted, false otherwise.
// case 1: if there are multiple disks having the same identifier, return false to prevent mount the wrong disk
func (v *DiskVolumeHandler) CanSafelyMount() (bool, error) {
	// this can not be happened, stop the mount process
	if len(v.Ldv.Status.DevLinks) == 0 {
		return false, fmt.Errorf("no devlinks found for volume %s", v.Ldv.Name)
	}

	// skip if no id_path found
	if len(v.Ldv.Status.DevLinks[v1alpha1.LinkByID]) == 0 {
		return true, nil
	}

	// NOTE: don't use id_path here because it might be different from the dev-link used by the volume
	// 1. get the symbol link that actually used by the volume
	deviceLink, err := getDeviceLinkByVolume(v.Ldv.Status.VolumePath)
	if err != nil {
		return false, err
	}

	// 2. list all devices and check whether this link exists at 2 or more different devices
	allDevices, err := udev.ListAllBlockDevices()
	if err != nil {
		return false, err
	}

	var matchedDevices []manager.Attribute
	for i, device := range allDevices {
		if _, exist := utils.StrFind(device.DevLinks, deviceLink); exist {
			matchedDevices = append(matchedDevices, allDevices[i])
		}
	}
	if len(matchedDevices) > 1 {
		return false, fmt.Errorf("device %s and %s has the same device link %s", matchedDevices[0].DevName, matchedDevices[1].DevName, deviceLink)
	}

	return true, nil
}

// this is only used for disk volume
func getDeviceLinkByVolume(volumePath string) (string, error) {
	// device path example: ../disk/pci-0000:03:00.0-scsi-0:0:30:0
	devicePath, err := utils.Bash(fmt.Sprintf("readlink -n %v", volumePath))
	if err != nil {
		return "", err
	}
	// after convert: /etc/hwameistor/LocalDisk_PoolHDD/disk/pci-0000:03:00.0-scsi-0:0:30:0
	devicePath = path.Join(types.GetLocalDiskPoolPathFromVolume(volumePath), strings.TrimPrefix(devicePath, "../"))

	// final output: /dev/disk/by-path/pci-0000:03:00.0-scsi-0:0:30:0
	return utils.Bash(fmt.Sprintf("readlink -n %v", devicePath))
}
