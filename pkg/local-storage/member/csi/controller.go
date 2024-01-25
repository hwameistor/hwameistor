package csi

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/fields"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

var (
	_ csi.ControllerServer = (*plugin)(nil)

	createLock sync.Mutex
)

const (
	RetryInterval = 1 * time.Second
	// SnapshotSize 1 GB snapshot size by default
	SnapshotSize = "1073741824"
)

// ControllerGetCapabilities implementation
func (p *plugin) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	p.logger.Debug("ControllerGetCapabilities")

	return &csi.ControllerGetCapabilitiesResponse{Capabilities: p.csCaps}, nil
}

// CreateVolume implementation, idempotent
func (p *plugin) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	createLock.Lock()
	defer createLock.Unlock()

	if req.VolumeContentSource != nil {
		if req.VolumeContentSource.GetSnapshot() != nil {
			return p.restoreVolumeFromSnapshot(ctx, req)
		}
		if req.VolumeContentSource.GetVolume() != nil {
			return p.cloneVolume(ctx, req)
		}
		p.logger.WithFields(log.Fields{"type": req.VolumeContentSource.Type}).Error("Invalid VolumeContentSource")
	}

	logCtx := p.logger.WithFields(log.Fields{
		"volume":           req.Name,
		"requiredCapacity": req.CapacityRange.RequiredBytes,
		"limitedCapacity":  req.CapacityRange.LimitBytes,
		"parameters":       req.Parameters,
		"topology":         req.AccessibilityRequirements,
		"capabilities":     req.VolumeCapabilities,
	})
	logCtx.Debug("CreateVolume")
	var (
		volume         = apisv1alpha1.LocalVolume{}
		resp           = &csi.CreateVolumeResponse{Volume: &csi.Volume{}}
		lastSavedError error
	)

	// 1. create volume if not exist
	if err := p.apiClient.Get(ctx, client.ObjectKey{Name: req.Name}, &volume); err != nil {
		if !errors.IsNotFound(err) {
			return resp, status.Errorf(codes.Internal, "failed to get volume %v", err)
		}

		if err = p.createEmptyVolumeFromRequest(ctx, req); err != nil {
			return resp, status.Errorf(codes.InvalidArgument, "failed to create volume %v", err)
		}
	}

	// 2. check if volume is ready to use per 2 seconds
	if err := wait.PollUntil(RetryInterval, func() (done bool, err error) {
		logCtx.Debugf("Checking if LocalVolume ready to use")
		if err = p.apiClient.Get(ctx, client.ObjectKey{Name: req.Name}, &volume); err != nil {
			lastSavedError = status.Errorf(codes.Unavailable, "failed to get volume %s %v", volume.Name, err)
			return false, nil
		}

		if volume.Status.State != apisv1alpha1.VolumeStateReady {
			lastSavedError = status.Errorf(codes.Unavailable, "LocalVolume %s is NotReady", volume.Name)
			return false, nil
		}

		// volume finally ready update response info
		resp.Volume = &csi.Volume{
			VolumeId:      volume.Name,
			CapacityBytes: volume.Status.AllocatedCapacityBytes,
			VolumeContext: req.Parameters,
		}
		lastSavedError = nil
		return true, nil
	}, ctx.Done()); err != nil {
		// must be timeout error
		lastSavedError = status.Errorf(codes.Unavailable, "LocalVolume %s is NotReady(deadline exceeded)", volume.Name)
	}

	// complete the request when context deadline exceeded or volume is ready to use
	if lastSavedError != nil {
		logCtx.Debugf("CreateVolume failed due to at context deadline exceeded, lastSavedError => %v", lastSavedError)
	} else {
		logCtx.Debugf("CreateVolume successfully, volume info => %v", resp.Volume)
	}

	return resp, lastSavedError
}

// createEmptyVolumeFromRequest creates a volume with the given request - without snapshot and clone
func (p *plugin) createEmptyVolumeFromRequest(ctx context.Context, req *csi.CreateVolumeRequest) error {
	params, err := parseParameters(req)
	if err != nil {
		p.logger.WithError(err).Error("Failed to parse parameters")
		return err
	}

	lvg, err := p.getLocalVolumeGroupOrCreate(req, params)
	if err != nil {
		p.logger.WithError(err).Error("Failed to get or create LocalVolumeGroup")
		return err
	}

	// return directly if exist already
	vol := &apisv1alpha1.LocalVolume{}
	if err = p.apiClient.Get(ctx, types.NamespacedName{Name: req.Name}, vol); err == nil {
		return nil
	}

	if !errors.IsNotFound(err) {
		p.logger.WithFields(log.Fields{"volName": req.Name, "error": err.Error()}).Error("Failed to query volume")
		return err
	}

	// create volume if not exist
	vol.Name = req.Name
	vol.Spec.RequiredCapacityBytes = req.CapacityRange.RequiredBytes
	vol.Spec.Convertible = params.convertible
	vol.Spec.PersistentVolumeClaimName = params.pvcName
	vol.Spec.PersistentVolumeClaimNamespace = params.pvcNamespace
	vol.Spec.VolumeGroup = lvg.Name
	vol.Spec.Accessibility.Nodes = lvg.Spec.Accessibility.Nodes
	vol.Spec.VolumeQoS = apisv1alpha1.VolumeQoS{
		Throughput: params.throughput,
		IOPS:       params.iops,
	}

	// override blow parameters when creating volume from snapshot
	if len(params.snapshot) > 0 {
		sourceVolume, err := getSourceVolumeFromSnapshot(params.snapshot, p.apiClient)
		if err != nil {
			p.logger.WithFields(log.Fields{"volName": vol.Name, "snapshot": params.snapshot}).WithError(err).Error("Failed to get source volume from snapshot")
			return err
		}
		vol.Spec.ReplicaNumber = sourceVolume.Spec.ReplicaNumber
		vol.Spec.PoolName = sourceVolume.Spec.PoolName
	} else {
		vol.Spec.ReplicaNumber = params.replicaNumber
		vol.Spec.PoolName = params.poolName
	}

	p.logger.WithFields(log.Fields{"volume": vol}).Debug("Creating a volume")
	return p.apiClient.Create(ctx, vol)
}

func (p *plugin) getLocalVolumeGroupOrCreate(req *csi.CreateVolumeRequest, params *volumeParameters) (*apisv1alpha1.LocalVolumeGroup, error) {

	// case 1. if the pvc is in a LVG, return it
	// case 2. if the pvc is not in any LVG, create a new one
	// case 3. if the pvc is not in any LVG but associated pvc is in a LVG, add it into the LVG

	if req.AccessibilityRequirements == nil || len(req.AccessibilityRequirements.Requisite) != 1 {
		p.logger.WithFields(log.Fields{"volume": req.Name, "accessibility": req.AccessibilityRequirements}).Error("Not found accessibility requirements")
		return nil, fmt.Errorf("not found accessibility requirements")
	}
	requiredNodeName := req.AccessibilityRequirements.Requisite[0].Segments[apis.TopologyNodeKey]

	lvg, lvs, err := p.getAssociatedVolumeGroupAndVolumesForPVC(params.pvcNamespace, params.pvcName)
	// // fetch the local volume group by PVC
	// lvg, err := p.getLocalVolumeGroupByPVC(params.pvcNamespace, params.pvcName)
	p.logger.WithFields(log.Fields{
		"lvg":       lvg,
		"lvs":       lvs,
		"err":       err,
		"pvc":       params.pvcName,
		"namespace": params.pvcNamespace,
	}).Debug("Result of getAssociatedVolumeGroupAndVolumesForPVC")
	if err != nil {
		return nil, err
	}
	if lvg != nil && len(lvg.Name) > 0 {
		// check if pvc is in the lvg, if not, add it
		for _, vol := range lvg.Spec.Volumes {
			if vol.PersistentVolumeClaimName == params.pvcName {
				// case 1: in the LVG
				return lvg, nil
			}
		}
		// case 2: has the LVG, but pvc is not in it. Add pvc into
		lvg.Spec.Volumes = append(lvg.Spec.Volumes, apisv1alpha1.VolumeInfo{PersistentVolumeClaimName: params.pvcName})
		p.logger.WithFields(log.Fields{"lvg": lvg.Name, "pvc": params.pvcName}).Debug("Adding a new PVC into the LVG")
		return lvg, p.apiClient.Update(context.TODO(), lvg)
	}

	// case 3: not found the local volume group, create it
	p.logger.WithFields(log.Fields{"pvc": params.pvcName, "namespace": params.pvcNamespace}).Debug("Not found the associated LocalVolumeGroup or LocalVolumes")

	var selectedNodes []string
	// for snapshot restore, volume topology must keep same with source volume
	if len(params.snapshot) > 0 {
		sourceVolume, err := getSourceVolumeFromSnapshot(params.snapshot, p.apiClient)
		if err != nil {
			p.logger.WithField("snapshot", params.snapshot).WithError(err).Error("failed to get source volume from snapshot")
			return nil, err
		}
		selectedNodes = sourceVolume.Spec.Accessibility.Nodes
	} else {
		candidateNodes := p.storageMember.Controller().VolumeScheduler().GetNodeCandidates(lvs)
		foundThisNode := false
		for _, nn := range candidateNodes {
			if len(selectedNodes) == int(params.replicaNumber) {
				break
			}
			if nn.Name == requiredNodeName {
				foundThisNode = true
				selectedNodes = append(selectedNodes, nn.Name)
			} else {
				if len(selectedNodes) == int(params.replicaNumber-1) {
					if foundThisNode {
						selectedNodes = append(selectedNodes, nn.Name)
					}
				} else {
					selectedNodes = append(selectedNodes, nn.Name)
				}
			}
		}
		if !foundThisNode {
			p.logger.WithField("requireNode", requiredNodeName).Errorf("requireNode is not exist in candidateNodes")
			return nil, fmt.Errorf("requireNode %s is not ready", requiredNodeName)
		}
		if len(selectedNodes) < int(params.replicaNumber) {
			p.logger.WithFields(log.Fields{"nodes": selectedNodes, "replica": params.replicaNumber}).Error("No enough nodes")
			return nil, fmt.Errorf("no enough nodes")
		}
	}

	lvg = &apisv1alpha1.LocalVolumeGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: genUUID(),
		},
		Spec: apisv1alpha1.LocalVolumeGroupSpec{
			Namespace: params.pvcNamespace,
			Volumes:   []apisv1alpha1.VolumeInfo{},
			Accessibility: apisv1alpha1.AccessibilityTopology{
				Nodes: selectedNodes,
			},
		},
	}
	for _, lv := range lvs {
		lvg.Spec.Volumes = append(lvg.Spec.Volumes, apisv1alpha1.VolumeInfo{PersistentVolumeClaimName: lv.Spec.PersistentVolumeClaimName})
	}
	p.logger.WithFields(log.Fields{"lvg": lvg.Name, "spec": lvg.Spec}).Debug("Creating a new LVG ...")
	if err := p.apiClient.Create(context.TODO(), lvg); err != nil {
		p.logger.WithField("lvg", lvg.Name).WithError(err).Error("Failed to create LVG")
		return nil, err
	}

	return lvg, nil
}

func (p *plugin) getAssociatedVolumeGroupAndVolumesForPVC(pvcNamespace string, pvcName string) (*apisv1alpha1.LocalVolumeGroup, []*apisv1alpha1.LocalVolume, error) {
	lvs, err := p.getAssociatedVolumes(pvcNamespace, pvcName)
	if err != nil {
		p.logger.WithFields(log.Fields{"pvc": pvcName, "namespace": pvcNamespace}).WithError(err).Error("Not found associated volumes")
		return nil, lvs, fmt.Errorf("not found associated volumes")
	}

	lvgList := apisv1alpha1.LocalVolumeGroupList{}
	if err := p.apiClient.List(context.TODO(), &lvgList); err != nil {
		return nil, lvs, err
	}
	for i, lvg := range lvgList.Items {
		if lvg.Spec.Namespace != pvcNamespace {
			continue
		}
		for _, vol := range lvg.Spec.Volumes {
			for _, lv := range lvs {
				if vol.PersistentVolumeClaimName == lv.Spec.PersistentVolumeClaimName {
					return &lvgList.Items[i], lvs, nil
				}
			}
		}
	}
	return nil, lvs, nil
}

func (p *plugin) getAssociatedVolumes(namespace string, pvcName string) ([]*apisv1alpha1.LocalVolume, error) {
	podList := corev1.PodList{}
	if err := p.apiClient.List(context.TODO(), &podList, &client.ListOptions{Namespace: namespace}); err != nil {
		p.logger.WithError(err).Error("Failed to list Pods")
		return []*apisv1alpha1.LocalVolume{}, err
	}
	for i, pod := range podList.Items {
		for _, vol := range pod.Spec.Volumes {
			if vol.PersistentVolumeClaim == nil {
				continue
			}
			if vol.PersistentVolumeClaim.ClaimName == pvcName {
				return p.getHwameiStorPVCs(&podList.Items[i])
			}
		}
	}
	log.Debug("getAssociatedVolumes return empty")
	return []*apisv1alpha1.LocalVolume{}, nil

}

func (p *plugin) getHwameiStorPVCs(pod *corev1.Pod) ([]*apisv1alpha1.LocalVolume, error) {
	lvs := []*apisv1alpha1.LocalVolume{}
	p.logger.WithField("pod", pod.Name).Debug("Query hwameistor PVCs")

	ctx := context.Background()
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim == nil {
			continue
		}
		pvc := &corev1.PersistentVolumeClaim{}
		if err := p.apiClient.Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: vol.PersistentVolumeClaim.ClaimName}, pvc); err != nil {
			// if pvc can't be found in the cluster, the pod should not be able to be scheduled
			p.logger.WithField("pvc", vol.PersistentVolumeClaim.ClaimName).WithError(err).Error("Failed to get PVC")
			return lvs, err
		}
		if pvc.Spec.StorageClassName == nil {
			// should not be the CSI pvc, ignore
			continue
		}
		sc := &storagev1.StorageClass{}
		if err := p.apiClient.Get(ctx, types.NamespacedName{Name: *pvc.Spec.StorageClassName}, sc); err != nil {
			// can't found storageclass in the cluster, the pod should not be able to be scheduled
			p.logger.WithField("storageclass", *pvc.Spec.StorageClassName).WithError(err).Error("Failed to get StorageClass")
			return lvs, err
		}
		if sc.Provisioner == apisv1alpha1.CSIDriverName {
			if lv, err := constructLocalVolumeForPVC(pvc, sc); err == nil {
				lvs = append(lvs, lv)
			}
		}
	}
	return lvs, nil
}

func constructLocalVolumeForPVC(pvc *corev1.PersistentVolumeClaim, sc *storagev1.StorageClass) (*apisv1alpha1.LocalVolume, error) {

	lv := apisv1alpha1.LocalVolume{}
	poolName, err := buildStoragePoolName(
		sc.Parameters[apisv1alpha1.VolumeParameterPoolClassKey],
	)
	if err != nil {
		return nil, err
	}

	//	lv.Name = pvc.Name
	lv.Spec.PersistentVolumeClaimNamespace = pvc.Namespace
	lv.Spec.PersistentVolumeClaimName = pvc.Name
	lv.Spec.PoolName = poolName
	storage := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	lv.Spec.RequiredCapacityBytes = storage.Value()
	replica, _ := strconv.Atoi(sc.Parameters[apisv1alpha1.VolumeParameterReplicaNumberKey])
	lv.Spec.ReplicaNumber = int64(replica)
	return &lv, nil
}

func buildStoragePoolName(poolClass string) (string, error) {

	if poolClass == apisv1alpha1.DiskClassNameHDD {
		return apisv1alpha1.PoolNameForHDD, nil
	}
	if poolClass == apisv1alpha1.DiskClassNameSSD {
		return apisv1alpha1.PoolNameForSSD, nil
	}
	if poolClass == apisv1alpha1.DiskClassNameNVMe {
		return apisv1alpha1.PoolNameForNVMe, nil
	}

	return "", fmt.Errorf("invalid pool info")
}

// restoreVolumeFromSnapshot creates a new volume from the snapshot
// Main Steps:
//  1. Create a new empty LocalVolume
//  2. Fill the new LocalVolume with the contents from the snapshot
func (p *plugin) restoreVolumeFromSnapshot(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	logCtx := p.logger.WithFields(log.Fields{
		"volume":           req.Name,
		"snapshot":         req.VolumeContentSource.GetSnapshot().SnapshotId,
		"requiredCapacity": req.CapacityRange.RequiredBytes,
		"limitedCapacity":  req.CapacityRange.LimitBytes,
		"parameters":       req.Parameters,
		"topology":         req.AccessibilityRequirements,
		"capabilities":     req.VolumeCapabilities,
	})
	logCtx.Debug("restoreVolumeFromSnapshot")

	volume := apisv1alpha1.LocalVolume{}
	resp := &csi.CreateVolumeResponse{Volume: &csi.Volume{
		ContentSource: &csi.VolumeContentSource{
			Type: &csi.VolumeContentSource_Snapshot{
				Snapshot: &csi.VolumeContentSource_SnapshotSource{
					SnapshotId: req.VolumeContentSource.GetSnapshot().SnapshotId,
				},
			},
		},
	}}
	snapshotRestoreName := utils.GetSnapshotRestoreNameByVolume(req.Name)
	volumeSnapshotRestore := apisv1alpha1.LocalVolumeSnapshotRestore{}
	releaseRestore := func() error {
		if err := p.apiClient.Get(ctx, client.ObjectKey{Name: snapshotRestoreName}, &volumeSnapshotRestore); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			p.logger.WithError(err).Errorf("failed to get volumeSnapshotRestore %v", err)
			return err
		}
		// LocalVolumeSnapshotRestore is ready, remove the protection finalizer from the object
		volumeSnapshotRestore.SetFinalizers(utils.RemoveStringItem(volumeSnapshotRestore.Finalizers, apisv1alpha1.SnapshotRestoringFinalizer))
		return p.apiClient.Update(ctx, &volumeSnapshotRestore)
	}

	// 0. finishes directly if snapshot restore has already been completed
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.Name}, &volume); err != nil {
		if !errors.IsNotFound(err) {
			return resp, err
		}
	}
	if volume.GetAnnotations() != nil {
		if _, ok := volume.GetAnnotations()[apisv1alpha1.VolumeSnapshotRestoreCompletedAnnoKey]; ok {
			if err := releaseRestore(); err != nil {
				p.logger.WithError(err).Errorf("failed to release snapshot restore")
				return resp, err
			}
			resp.Volume.VolumeId = volume.Name
			resp.Volume.CapacityBytes = volume.Status.AllocatedCapacityBytes
			resp.Volume.VolumeContext = req.Parameters
			resp.Volume.VolumeContext[apisv1alpha1.SourceVolumeSnapshotAnnoKey] = req.VolumeContentSource.GetSnapshot().SnapshotId
			return resp, nil
		}
	}

	// 1. create new empty LocalVolume
	if err := p.validateVolumeCreateRequestForSnapshot(ctx, req); err != nil {
		logCtx.WithError(err).Error("failed to validate CreateVolumeRequest for snapshot")
		return resp, status.Errorf(codes.InvalidArgument, "failed to validate CreateVolumeRequest for snapshot: %v", err)
	}

	logCtx.Debugf("Step1: Start creating LocalVolume %s", req.GetName())
	if err := p.createEmptyVolumeFromRequest(ctx, req); err != nil {
		logCtx.WithError(err).Error("failed to create LocalVolume")
		return resp, status.Errorf(codes.Internal, "failed to create LocalVolume")
	}

	// 2. waits for LocalVolume ready to use
	logCtx.Debugf("Step2: Checking if LocalVolume %s ready to use", req.Name)
	tryCount := 0
	if err := wait.PollUntil(RetryInterval, func() (done bool, err error) {
		tryCount++
		// don't retry if err happens when fetch volume
		if err = p.apiClient.Get(ctx, client.ObjectKey{Name: req.Name}, &volume); err != nil {
			p.logger.WithError(err).Errorf("failed to get volume %v", err)
			return true, err
		}
		if volume.Status.State != apisv1alpha1.VolumeStateReady {
			if tryCount >= 10 {
				// stop retrying, return error immediately
				return true, status.Errorf(codes.Internal, "LocalVolume %s is NotReady after %d retries", volume.Name, tryCount)
			}
			p.logger.WithField("tryCount", tryCount).Debugf("LocalVolume %s is NotReady", volume.Name)
			return false, nil
		}
		// Volume is ready and prepare to restore snapshot, return immediately
		return true, nil
	}, ctx.Done()); err != nil {
		logCtx.WithError(err).Errorf("LocalVolume %s is NotReady", req.Name)
		return resp, status.Errorf(codes.Unavailable, "failed to check LocalVolume %s Ready or Not: %v", volume.Name, err)
	}

	// 3. create LocalVolumeSnapshotRestore instance
	logCtx.Debugf("Step3: Start creating LocalVolumeSnapshotRestore %s", snapshotRestoreName)
	volumeSnapshotRestore.Name = snapshotRestoreName
	volumeSnapshotRestore.Spec.TargetVolume = req.GetName()
	volumeSnapshotRestore.Spec.TargetPoolName = volume.Spec.PoolName
	// protection finalizer to prevent objects to be deleted
	volumeSnapshotRestore.SetFinalizers([]string{apisv1alpha1.SnapshotRestoringFinalizer})
	volumeSnapshotRestore.Spec.RestoreType = apisv1alpha1.RestoreTypeCreate
	volumeSnapshotRestore.Spec.SourceVolumeSnapshot = req.VolumeContentSource.GetSnapshot().SnapshotId
	if err := p.apiClient.Create(ctx, &volumeSnapshotRestore); err != nil {
		if !errors.IsAlreadyExists(err) {
			logCtx.WithError(err).Errorf("failed to create LocalVolumeSnapshotRestore %s", snapshotRestoreName)
			return resp, status.Errorf(codes.Unavailable, "failed to create LocalVolumeSnapshotRestore: %v", err)
		}
	}

	// 4. waits for LocalVolumeSnapshotRestore completed
	//
	// The LocalVolumeSnapshotRestore is an operation on the VolumeSnapshot, so it will be deleted when the operation is completed.
	// Thus, we need to hold the delete operation before we confirm that the operation is completed.
	logCtx.Debugf("Step4: Checking if LocalVolumeSnapshotRestore %s ready to use", snapshotRestoreName)
	tryCount = 0
	if err := wait.PollUntil(RetryInterval, func() (done bool, err error) {
		if err = p.apiClient.Get(ctx, client.ObjectKey{Name: snapshotRestoreName}, &volumeSnapshotRestore); err != nil {
			p.logger.WithError(err).Errorf("failed to get volumeSnapshotRestore %v", err)
			return true, err
		}
		if volumeSnapshotRestore.Status.State != apisv1alpha1.OperationStateCompleted {
			if tryCount >= 10 {
				return true, status.Errorf(codes.Unavailable, "LocalVolumeSnapshotRestore %s is not completed after %d retries", volumeSnapshotRestore.Name, tryCount)
			}
			p.logger.WithField("tryCount", tryCount).Debugf("LocalVolumeSnapshotRestore %s is not completed", volume.Name)
			return false, nil
		}
		// LocalVolumeSnapshotRestore is ready, remove the protection finalizer from the object
		volumeSnapshotRestore.SetFinalizers(utils.RemoveStringItem(volumeSnapshotRestore.Finalizers, apisv1alpha1.SnapshotRestoringFinalizer))
		return true, p.apiClient.Update(ctx, &volumeSnapshotRestore)
	}, ctx.Done()); err != nil {
		logCtx.WithError(err).Errorf("LocalVolumeSnapshotRestore %s is not completed", snapshotRestoreName)
		return resp, err
	}

	resp.Volume.VolumeId = volume.Name
	resp.Volume.CapacityBytes = volume.Status.AllocatedCapacityBytes
	resp.Volume.VolumeContext = req.Parameters
	resp.Volume.VolumeContext[apisv1alpha1.SourceVolumeSnapshotAnnoKey] = req.VolumeContentSource.GetSnapshot().SnapshotId
	return resp, nil
}

func (p *plugin) validateVolumeCreateRequestForSnapshot(ctx context.Context, req *csi.CreateVolumeRequest) error {
	if req.VolumeContentSource == nil || req.VolumeContentSource.GetSnapshot() == nil {
		return fmt.Errorf("snapshot must be provided")
	}
	if req.CapacityRange == nil || req.CapacityRange.RequiredBytes <= 0 {
		return fmt.Errorf("required capacity must be provided")
	}
	volumeSnapshotId := req.VolumeContentSource.GetSnapshot().SnapshotId
	requiredCapacityBytes := req.CapacityRange.RequiredBytes

	// fetch source volume by snapshotId
	volumeSnapshot := apisv1alpha1.LocalVolumeSnapshot{}
	if err := p.apiClient.Get(ctx, client.ObjectKey{Name: volumeSnapshotId}, &volumeSnapshot); err != nil {
		return err
	}
	sourceVolume := apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, client.ObjectKey{Name: volumeSnapshot.Spec.SourceVolume}, &sourceVolume); err != nil {
		return err
	}

	if requiredCapacityBytes < sourceVolume.Spec.RequiredCapacityBytes {
		return fmt.Errorf("the new volume required capacity must be greater than the existing volume required capacity")
	}

	return nil
}

// cloneVolume creates a new volume from the given volume
// Main steps
// 1. take a snapshot of the given volume
// 2. create a new volume with the given StorageClass
// 3. restore the snapshot(from step1) to the new volume(from step2)
func (p *plugin) cloneVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	logCtx := p.logger.WithFields(log.Fields{
		"volume":           req.Name,
		"sourceVolume":     req.VolumeContentSource.GetVolume().VolumeId,
		"requiredCapacity": req.CapacityRange.RequiredBytes,
		"limitedCapacity":  req.CapacityRange.LimitBytes,
		"parameters":       req.Parameters,
		"topology":         req.AccessibilityRequirements,
		"capabilities":     req.VolumeCapabilities,
	})
	logCtx.Debug("cloneVolume")

	resp := &csi.CreateVolumeResponse{Volume: &csi.Volume{VolumeId: req.Name, VolumeContext: req.Parameters}}
	sourceVolumeCopy := *req.VolumeContentSource
	snapshotName := fmt.Sprintf("%v-%v", "snapshot-for", req.Name)
	params := map[string]string{apisv1alpha1.SnapshotParameterSizeKey: SnapshotSize}
	snapshotRestoreRequest := req
	snapshotDeleteRequest := &csi.DeleteSnapshotRequest{}
	snapshotCreateRequest := &csi.CreateSnapshotRequest{
		Name:           snapshotName,
		SourceVolumeId: req.VolumeContentSource.GetVolume().VolumeId,
		Parameters:     params,
	}

	// Step1: take a snapshot for the source volume
	logCtx.Debug("Step1: Creating snapshot")
	if snapshotResp, err := p.CreateSnapshot(ctx, snapshotCreateRequest); err != nil {
		return resp, status.Errorf(codes.Internal, "Failed to create snapshot %v", err)
	} else {
		// construct volume restore request with the actual snapshot in response
		snapshotRestoreRequest.VolumeContentSource = &csi.VolumeContentSource{
			Type: &csi.VolumeContentSource_Snapshot{
				Snapshot: &csi.VolumeContentSource_SnapshotSource{
					SnapshotId: snapshotResp.Snapshot.SnapshotId,
				},
			},
		}
	}

	// Step2: restore volume from snapshot
	logCtx.Debug("Step2: Restoring snapshot")
	if snapResp, err := p.restoreVolumeFromSnapshot(ctx, snapshotRestoreRequest); err != nil {
		return resp, status.Errorf(codes.Unavailable, "Failed to restore volume from snapshot %v", err)
	} else {
		snapshotDeleteRequest.SnapshotId = snapResp.GetVolume().GetVolumeContext()[apisv1alpha1.SourceVolumeSnapshotAnnoKey]
		resp.Volume.ContentSource = &sourceVolumeCopy
	}

	// Step3: clean up the snapshot
	logCtx.Debug("Step3: Cleaning up snapshot")
	if _, err := p.DeleteSnapshot(ctx, snapshotDeleteRequest); err != nil {
		return resp, status.Errorf(codes.Unavailable, "Failed to delete snapshot %v", err)
	}
	logCtx.Debug("volume clone successful")
	return resp, nil
}

// ControllerGetVolume implementation, idempotent
func (p *plugin) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	p.logger.WithFields(log.Fields{"volume": req.VolumeId}).Debug("ControllerGetVolume")

	// get volume from apiClient
	resp := &csi.ControllerGetVolumeResponse{}
	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		if !errors.IsNotFound(err) {
			p.logger.WithFields(log.Fields{"volName": req.VolumeId, "error": err.Error()}).Error("Failed to query volume")
			return resp, err
		}
		// not found volume, should be deleted already
		return resp, nil
	}

	resp.Volume = &csi.Volume{
		CapacityBytes:      vol.Status.AllocatedCapacityBytes,
		VolumeId:           req.VolumeId,
		AccessibleTopology: []*csi.Topology{},
	}
	resp.Status = &csi.ControllerGetVolumeResponse_VolumeStatus{
		PublishedNodeIds: []string{},
		VolumeCondition:  &csi.VolumeCondition{},
	}
	if vol.Status.PublishedNodeName != "" {
		resp.Status.PublishedNodeIds = append(resp.Status.PublishedNodeIds, vol.Status.PublishedNodeName)
	}
	volReplica := &apisv1alpha1.LocalVolumeReplica{}
	for _, replicaID := range vol.Status.Replicas {
		if err := p.apiClient.Get(ctx, types.NamespacedName{Name: replicaID}, volReplica); err != nil {
			p.logger.WithFields(log.Fields{"replica": replicaID, "error": err.Error()}).Error("Failed to query volume replica")
			return resp, err
		}
		if volReplica.Spec.NodeName == "" {
			continue
		}
		resp.Volume.AccessibleTopology = append(resp.Volume.AccessibleTopology,
			&csi.Topology{
				Segments: map[string]string{
					apis.TopologyNodeKey: volReplica.Spec.NodeName,
				},
			},
		)
	}

	if vol.Status.State != apisv1alpha1.VolumeStateReady {
		resp.Status.VolumeCondition.Abnormal = true
		resp.Status.VolumeCondition.Message = "The volume is not ready"
	} else {
		resp.Status.VolumeCondition.Abnormal = false
		resp.Status.VolumeCondition.Message = "The volume is ready"
	}

	return resp, nil
}

// DeleteVolume implementation, idempotent
func (p *plugin) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	p.logger.WithFields(log.Fields{
		"volume":  req.VolumeId,
		"secrets": req.Secrets,
	}).Debug("DeleteVolume")

	resp := &csi.DeleteVolumeResponse{}
	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		if !errors.IsNotFound(err) {
			p.logger.WithFields(log.Fields{"volName": req.VolumeId, "error": err.Error()}).Error("Failed to query volume")
			return resp, err
		}
		// not found volume, should be deleted already
		return resp, nil
	}

	// abort snapshot restore operation if needed
	if err := p.abortVolumeSnapshotRestoreIfNeeded(vol); err != nil {
		p.logger.WithError(err).WithField("volName", req.VolumeId).Error("Failed to abort volume restore operation")
		return nil, err
	}

	if vol.Status.State == apisv1alpha1.VolumeStateDeleted {
		return resp, nil
	}
	if vol.Status.State == apisv1alpha1.VolumeStateToBeDeleted {
		// volume will be deleted after all snapshots removed, return directly here
		if vss, _ := listVolumeSnapshots(vol.Name, p.apiClient); len(vss) > 0 {
			return resp, nil
		}
		return resp, fmt.Errorf("volume in deleting")
	}
	vol.Spec.Delete = true
	if err := p.apiClient.Update(ctx, vol); err != nil {
		p.logger.WithFields(log.Fields{"volName": vol.Name, "state": vol.Status.State, "error": err.Error()}).Error("Failed to set volume status")
		return resp, err
	}
	return resp, fmt.Errorf("volume in deleting")
}

func (p *plugin) abortVolumeSnapshotRestoreIfNeeded(volume *apisv1alpha1.LocalVolume) error {
	need, err := p.needAbortVolumeSnapshotRestore(volume)
	if err != nil {
		p.logger.WithError(err).WithFields(log.Fields{"volName": volume.Name}).Error("Failed to get snapshotRestore")
		return err
	} else if !need {
		return nil
	}

	if err = p.abortVolumeSnapshotRestore(volume); err != nil {
		p.logger.WithError(err).WithFields(log.Fields{"volName": volume.Name}).Error("Failed to abort snapshotRestore")
	}
	return err
}

func (p *plugin) abortVolumeSnapshotRestore(volume *apisv1alpha1.LocalVolume) error {
	snapRestoreName := utils.GetSnapshotRestoreNameByVolume(volume.Name)
	snapRestore := apisv1alpha1.LocalVolumeSnapshotRestore{}
	if err := p.apiClient.Get(context.Background(), client.ObjectKey{Name: snapRestoreName}, &snapRestore); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		p.logger.WithError(err).WithFields(log.Fields{"snapRestore": snapRestoreName, "volume": volume.Name}).Error("Failed to get snapshotRestore")
		return err
	}

	snapRestore.Spec.Abort = true
	snapRestore.Finalizers = utils.RemoveStringItem(snapRestore.GetFinalizers(), apisv1alpha1.SnapshotRestoringFinalizer)
	return p.apiClient.Update(context.Background(), &snapRestore)
}

func (p *plugin) needAbortVolumeSnapshotRestore(volume *apisv1alpha1.LocalVolume) (bool, error) {
	snapRestoreName := utils.GetSnapshotRestoreNameByVolume(volume.Name)
	snapRestore := apisv1alpha1.LocalVolumeSnapshotRestore{}
	if err := p.apiClient.Get(context.Background(), client.ObjectKey{Name: snapRestoreName}, &snapRestore); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		p.logger.WithError(err).WithFields(log.Fields{"snapRestore": snapRestoreName, "volume": volume.Name}).Error("Failed to get snapshotRestore")
		return true, err
	}

	return snapRestore.Spec.Abort == false || isStringInArray(apisv1alpha1.SnapshotRestoringFinalizer, snapRestore.GetFinalizers()), nil
}

// ControllerPublishVolume implementation, idempotent
func (p *plugin) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	p.logger.WithFields(log.Fields{
		"volume":        req.VolumeId,
		"node":          req.NodeId,
		"volumeContext": req.VolumeContext,
		"readonly":      req.Readonly,
		"secrets":       req.Secrets,
		"AccessMode":    req.VolumeCapability.AccessMode.Mode,
		"AccessType":    req.VolumeCapability.AccessType,
	}).Debug("ControllerPublishVolume")

	resp := &csi.ControllerPublishVolumeResponse{PublishContext: map[string]string{}}

	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		p.logger.WithFields(log.Fields{"volName": req.VolumeId, "error": err.Error()}).Error("Failed to query volume")
		return resp, err
	}
	volReplica := &apisv1alpha1.LocalVolumeReplica{}
	for _, replicaName := range vol.Status.Replicas {
		if err := p.apiClient.Get(ctx, types.NamespacedName{Name: replicaName}, volReplica); err != nil {
			p.logger.WithFields(log.Fields{"replica": replicaName, "error": err.Error()}).Error("Failed to query volume replica")
			return resp, err
		}
		if volReplica.Spec.NodeName != req.NodeId {
			continue
		}
		if volReplica.Status.State != apisv1alpha1.VolumeReplicaStateReady {
			p.logger.WithFields(log.Fields{"replica": replicaName, "state": volReplica.Status.State}).Error("volume replica is not ready")
			return resp, fmt.Errorf("replica not ready")
		}
		if !volReplica.Status.Synced {
			p.logger.WithFields(log.Fields{"replica": replicaName, "synced": volReplica.Status.Synced}).Error("volume replica is out of date")
			return resp, fmt.Errorf("replica out of date")
		}
		if volReplica.Status.DevicePath == "" {
			p.logger.WithFields(log.Fields{"replica": replicaName, "devicePath": volReplica.Status.DevicePath}).Error("invalid volume replica device path")
			return resp, fmt.Errorf("invalid device path of volume replica")

		}
		if req.VolumeCapability.GetBlock() != nil {
			vol.Status.PublishedRawBlock = true
		} else if req.VolumeCapability.GetMount() != nil {
			if len(req.VolumeCapability.GetMount().FsType) == 0 {
				vol.Status.PublishedFSType = "ext4"
			} else {
				vol.Status.PublishedFSType = req.VolumeCapability.GetMount().FsType
			}
		}
		vol.Status.PublishedNodeName = req.NodeId
		if err := p.apiClient.Status().Update(ctx, vol); err != nil {
			p.logger.WithFields(log.Fields{"volume": vol.Name, "node": req.NodeId}).Error("Failed to update volume with published node info")
			return resp, err
		}
		p.logger.WithFields(log.Fields{"volume": req.VolumeId, "node": req.NodeId, "devicePath": volReplica.Status.DevicePath}).Debug("Found valid volume replica")
		resp.PublishContext[VolumeReplicaDevicePathKey] = volReplica.Status.DevicePath
		resp.PublishContext[VolumeReplicaNameKey] = volReplica.Name
		return resp, nil
	}

	p.logger.WithFields(log.Fields{"volume": req.VolumeId, "node": req.NodeId}).Error("not found valid volume replica")
	return resp, fmt.Errorf("not found valid volume replica")
}

// ControllerUnpublishVolume implementation, idempotent
func (p *plugin) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	p.logger.WithFields(log.Fields{
		"volume":  req.VolumeId,
		"node":    req.NodeId,
		"secrets": req.Secrets,
	}).Debug("ControllerUnpublishVolume")

	resp := &csi.ControllerUnpublishVolumeResponse{}

	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		p.logger.WithFields(log.Fields{"volume": req.VolumeId, "error": err.Error()}).Error("Failed to query volume")
		return resp, err
	}

	// when publish node is empty, return directly
	// https://github.com/hwameistor/hwameistor/issues/296
	if vol.Status.PublishedNodeName == "" {
		p.logger.WithFields(log.Fields{"volume": req.VolumeId, "node": req.NodeId}).Debug("Volume has already unpublished")
		return resp, nil
	}

	if vol.Status.PublishedNodeName != req.NodeId {
		p.logger.WithFields(log.Fields{"volume": req.VolumeId, "node": req.NodeId}).Error("Wrong published node in request")
		return resp, fmt.Errorf("wrong published node")
	}
	vol.Status.PublishedNodeName = ""
	if err := p.apiClient.Status().Update(ctx, vol); err != nil {
		p.logger.WithFields(log.Fields{"volume": vol.Name, "node": req.NodeId}).Error("Failed to update volume to unpublish node")
		return resp, err
	}

	return resp, nil
}

// ValidateVolumeCapabilities implementation
func (p *plugin) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	p.logger.WithFields(log.Fields{
		"volume":             req.VolumeId,
		"parameters":         req.Parameters,
		"secrets":            req.Secrets,
		"volumeCapabilities": req.VolumeCapabilities,
	}).Debug("ValidateVolumeCapabilities")

	resp := &csi.ValidateVolumeCapabilitiesResponse{}

	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		if errors.IsNotFound(err) {
			p.logger.WithFields(log.Fields{"volName": req.VolumeId, "error": err.Error()}).Error("not found volume")

		} else {
			p.logger.WithFields(log.Fields{"volName": req.VolumeId, "error": err.Error()}).Error("Failed to query volume")
		}
		return resp, err
	}

	for _, reqCap := range req.VolumeCapabilities {
		if reqCap.AccessMode != nil && reqCap.AccessMode.Mode != csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER {
			return resp, nil
		}
	}
	resp.Confirmed = &csi.ValidateVolumeCapabilitiesResponse_Confirmed{VolumeCapabilities: p.vCaps}
	return resp, nil

}

// ListVolumes implementation, consider the case of multiple pages
func (p *plugin) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	p.logger.WithFields(log.Fields{
		"maxEntries": req.MaxEntries,
		"token":      req.StartingToken,
	}).Debug("ListVolumes")

	// Limit: not support token/multi-pages implementation currently

	resp := &csi.ListVolumesResponse{
		Entries: []*csi.ListVolumesResponse_Entry{},
	}
	volumeList := &apisv1alpha1.LocalVolumeList{}
	if err := p.apiClient.List(ctx, volumeList); err != nil {
		p.logger.WithFields(log.Fields{"error": err}).Error("Failed to list volumes")
		return resp, err
	}

	replicaList := &apisv1alpha1.LocalVolumeReplicaList{}
	if err := p.apiClient.List(ctx, replicaList); err != nil {
		p.logger.WithFields(log.Fields{"error": err}).Error("Failed to list volume replicas")
		return resp, err
	}
	replicas := map[string]*apisv1alpha1.LocalVolumeReplica{}
	for i, replica := range replicaList.Items {
		replicas[replica.Name] = &replicaList.Items[i]
	}

	for i, vol := range volumeList.Items {
		// return all volumes when MaxEntries == 0, otherwise limited by MaxEntries
		if req.MaxEntries > 0 && int32(i) >= req.MaxEntries {
			break
		}
		entry := &csi.ListVolumesResponse_Entry{
			Volume: &csi.Volume{
				CapacityBytes:      vol.Status.AllocatedCapacityBytes,
				VolumeId:           vol.Name,
				AccessibleTopology: []*csi.Topology{},
			},
			Status: &csi.ListVolumesResponse_VolumeStatus{
				PublishedNodeIds: []string{},
				VolumeCondition:  &csi.VolumeCondition{},
			},
		}
		if vol.Status.PublishedNodeName != "" {
			entry.Status.PublishedNodeIds = append(entry.Status.PublishedNodeIds, vol.Status.PublishedNodeName)
		}
		if vol.Status.State != apisv1alpha1.VolumeStateReady {
			entry.Status.VolumeCondition.Abnormal = true
			entry.Status.VolumeCondition.Message = "The volume is not ready"
		} else {
			entry.Status.VolumeCondition.Abnormal = false
			entry.Status.VolumeCondition.Message = "The volume is ready"
		}
		for _, replicaName := range vol.Status.Replicas {
			if replica, ok := replicas[replicaName]; ok {
				if replica.Spec.NodeName != "" {
					entry.Volume.AccessibleTopology = append(entry.Volume.AccessibleTopology,
						&csi.Topology{
							Segments: map[string]string{
								apis.TopologyNodeKey: replica.Spec.NodeName,
							},
						},
					)
				}
			}
		}
		resp.Entries = append(resp.Entries, entry)
	}

	return resp, nil
}

// GetCapacity implementation
func (p *plugin) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	p.logger.WithFields(log.Fields{
		"volumeCapabilities": req.VolumeCapabilities,
		"parameters":         req.Parameters,
		"AccessibleTopology": req.AccessibleTopology.Segments,
	}).Debug("GetCapacity")

	resp := &csi.GetCapacityResponse{}

	// DO NOT SUPPORT: The scheduler should take it into account

	return resp, fmt.Errorf("not supported")
}

// CreateSnapshot implementation, idempotent
func (p *plugin) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	logCtx := p.logger.WithFields(log.Fields{
		"name":         req.Name,
		"sourceVolume": req.SourceVolumeId,
		"parameters":   req.Parameters,
		"secrets":      req.Secrets,
	})
	logCtx.Debug("CreateSnapshot")
	err := validateSnapshotRequest(req)
	if err != nil {
		return nil, err
	}

	resp := &csi.CreateSnapshotResponse{Snapshot: &csi.Snapshot{}}
	resp.Snapshot.SourceVolumeId = req.SourceVolumeId
	// the underlying snapshot name is consistent with snapshotcontent name
	snapshotID := strings.Replace(req.Name, "snapshot", "snapcontent", 1)
	resp.Snapshot.SnapshotId = snapshotID

	// 1. create snapshot if not exist
	snapshot := &apisv1alpha1.LocalVolumeSnapshot{}
	if err = p.apiClient.Get(ctx, types.NamespacedName{Name: snapshotID}, snapshot); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get LocalVolumeSnapshot")
			return nil, status.Errorf(codes.Internal, "Failed to get LocalVolumeSnapshot: %v", err)
		}

		snapsize, err := getSnapshotSize(req.SourceVolumeId, req.Parameters, p.apiClient)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get snapshot size")
			return nil, status.Errorf(codes.Internal, "Failed to get snapshot size: %v", err)
		}

		// NOTE: We only take snapshots on the volume replica that exist at the moment!
		// For those volume replicas that created later but belong to this volume won't take snapshots.

		accessTopology, err := getVolumeAccessibility(req.SourceVolumeId, p.apiClient)
		if err != nil {
			logCtx.WithError(err).Error("Failed to get volume access topology")
			return nil, status.Errorf(codes.Internal, "Failed to get volume access topology: %v", err)
		}

		// for now, we only support take snapshot on single replica volume
		if len(accessTopology.Nodes) > 1 {
			logCtx.WithField("topology", accessTopology.Nodes).Error("Haven't support take snapshot on HA-Volume")
			return nil, status.Errorf(codes.Internal, "Haven't support take snapshot on HA-Volume")
		}

		snapshot.Name = snapshotID
		snapshot.Spec.Accessibility = accessTopology
		snapshot.Spec.RequiredCapacityBytes = snapsize
		snapshot.Spec.SourceVolume = req.SourceVolumeId

		if err = p.apiClient.Create(ctx, snapshot); err != nil {
			logCtx.WithError(err).Error("Failed to create LocalVolumeSnapshot")
			return nil, status.Errorf(codes.Internal, "Failed to create LocalVolumeSnapshot: %v", err)
		}
	}

	// 2. checks if snapshot ready to use per 3 seconds
	return resp, wait.PollUntil(RetryInterval, func() (bool, error) {
		logCtx.Debug("Checking if snapshot ready to use")
		if err = p.apiClient.Get(ctx, types.NamespacedName{Name: snapshotID}, snapshot); err != nil {
			logCtx.WithError(err).Error("Failed to get LocalVolumeSnapshot")
			return false, status.Errorf(codes.Internal, "Failed to get LocalVolumeSnapshot: %v", err)
		}

		if snapshot.Status.State != apisv1alpha1.VolumeStateReady {
			return false, status.Errorf(codes.Internal, "LocalVolumeSnapshot %s is NotReady", snapshot.Name)
		}

		resp.Snapshot.ReadyToUse = true
		resp.Snapshot.SizeBytes = snapshot.Status.AllocatedCapacityBytes
		resp.Snapshot.CreationTime = &timestamp.Timestamp{
			Seconds: snapshot.Status.CreationTime.Unix(),
			Nanos:   int32(snapshot.Status.CreationTime.Nanosecond()),
		}
		return true, nil
	}, ctx.Done())
}

// DeleteSnapshot implementation, idempotent
func (p *plugin) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	logCtx := p.logger.WithFields(log.Fields{
		"id":      req.SnapshotId,
		"secrets": req.Secrets,
	})
	logCtx.Debug("DeleteSnapshot")

	resp := &csi.DeleteSnapshotResponse{}
	snapshot := &apisv1alpha1.LocalVolumeSnapshot{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.SnapshotId}, snapshot); err != nil {
		if errors.IsNotFound(err) {
			logCtx.Debug("Snapshot already deleted")
			return resp, nil
		}
		logCtx.WithError(err).Error("Failed to get LocalVolumeSnapshot")
		return resp, status.Errorf(codes.Internal, "Failed to get LocalVolumeSnapshot: %v", err)
	}

	// delete snapshot by setting delete true in snapshot's spec
	snapshot.Spec.Delete = true
	return resp, p.apiClient.Update(ctx, snapshot)
}

// ListSnapshots implementation, idempotent
func (p *plugin) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	logCtx := p.logger.WithFields(log.Fields{
		"sourceVolume": req.SourceVolumeId,
		"snapshotID":   req.SnapshotId,
		"secrets":      req.Secrets,
		"maxEntries":   req.MaxEntries,
		"token":        req.StartingToken,
	})

	logCtx.Debug("ListSnapshots")

	resp := &csi.ListSnapshotsResponse{}
	snapshots := apisv1alpha1.LocalVolumeSnapshotList{}
	err := p.apiClient.List(ctx, &snapshots)
	if err != nil {
		logCtx.WithError(err).Error("Failed to list LocalVolumeSnapshot")
		return nil, status.Errorf(codes.Internal, "Failed to list LocalVolumeSnapshot: %v", err)
	}

	for _, snap := range snapshots.Items {
		resp.Entries = append(resp.Entries, &csi.ListSnapshotsResponse_Entry{
			Snapshot: &csi.Snapshot{
				SnapshotId:     snap.Name,
				SourceVolumeId: snap.Spec.SourceVolume,
				SizeBytes:      snap.Status.AllocatedCapacityBytes,
				ReadyToUse:     snap.Status.State == apisv1alpha1.VolumeStateReady,
				CreationTime: &timestamp.Timestamp{
					Seconds: snap.Status.CreationTime.Unix(),
					Nanos:   int32(snap.Status.CreationTime.Nanosecond()),
				},
			},
		})
	}

	return resp, nil
}

// ControllerExpandVolume - it will expand volume size in storage pool, idempotent
func (p *plugin) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	logCtx := p.logger.WithFields(log.Fields{"volume": req.VolumeId})
	logCtx.WithFields(log.Fields{
		"RequiredCapacity": req.CapacityRange.RequiredBytes,
		"AccessMode":       req.VolumeCapability.AccessMode.Mode,
		"AccessType":       req.VolumeCapability.AccessType,
	}).Debug("ControllerExpandVolume")

	resp := &csi.ControllerExpandVolumeResponse{}

	vol := &apisv1alpha1.LocalVolume{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, vol); err != nil {
		logCtx.WithError(err).Error("Failed to query volume")
		return resp, err
	}

	if req.CapacityRange == nil {
		logCtx.Error("Invalid volume capacity request")
		return resp, fmt.Errorf("invalid capacity request")
	}

	// new capacity is less than the current, disallow it
	if (req.CapacityRange.RequiredBytes + apisv1alpha1.VolumeExpansionCapacityBytesMin) < vol.Status.AllocatedCapacityBytes {
		logCtx.WithFields(log.Fields{"currentCapacity": vol.Status.AllocatedCapacityBytes, "newCapacity": req.CapacityRange.RequiredBytes}).Error("Can't reduce volume capacity")
		return resp, fmt.Errorf("can't reduce capacity")
	}

	// new capacity is close to the current (diff < 10MB), treat it as a successful expansion
	if math.Abs(float64(req.CapacityRange.RequiredBytes-vol.Status.AllocatedCapacityBytes)) <= float64(apisv1alpha1.VolumeExpansionCapacityBytesMin) {
		logCtx.WithFields(log.Fields{"currentCapacity": vol.Status.AllocatedCapacityBytes, "newCapacity": req.CapacityRange.RequiredBytes}).Info("Volume capacity expand completed")
		resp.CapacityBytes = vol.Status.AllocatedCapacityBytes
		if req.GetVolumeCapability().GetBlock() != nil {
			// there is no need to NodeExpansion if volumeMode is block
			resp.NodeExpansionRequired = false
		} else {
			resp.NodeExpansionRequired = true
		}
		return resp, nil
	}

	// new capacity is bigger than the current (diff > 10MB), expand it
	expand := &apisv1alpha1.LocalVolumeExpand{}
	if err := p.apiClient.Get(ctx, types.NamespacedName{Name: req.VolumeId}, expand); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume expansion action")
			return resp, err
		}
		// no expansion in progress, create a new one
		expand.Name = req.VolumeId
		expand.Spec.VolumeName = req.VolumeId
		expand.Spec.RequiredCapacityBytes = req.CapacityRange.RequiredBytes
		if err := p.apiClient.Create(ctx, expand); err != nil {
			logCtx.WithError(err).Error("Failed to submit volume expansion request")
			return resp, err
		}
		logCtx.WithField("newCapacity", expand.Spec.RequiredCapacityBytes).Info("Submitted volume expansion request")
		// still return error, will check volume size at next visit
		return resp, fmt.Errorf("volume expansion not completed yet")
	}

	logCtx.WithFields(log.Fields{"spec": expand.Spec, "status": expand.Status}).Debug("Volume expansion is still in progress")
	return resp, fmt.Errorf("volume expansion in progress")
}

// validateSnapshotRequest is used to validate a snapshot request against the volume specification and validate
func validateSnapshotRequest(req *csi.CreateSnapshotRequest) error {
	if len(req.Name) == 0 {
		return status.Errorf(
			codes.InvalidArgument,
			"snapshot name must be provided",
		)
	}
	if len(req.SourceVolumeId) == 0 {
		return status.Errorf(
			codes.InvalidArgument,
			"source volumeID must be provided",
		)
	}
	return nil
}

// getSnapshotSize is used to get the size from the snapshot specification if exists. return the source volume size if not specified.
func getSnapshotSize(sourceVolume string, params map[string]string, apiClient client.Client) (int64, error) {
	// 1. get snapshot size from the volume specification
	snapSize, ok := params[apisv1alpha1.SnapshotParameterSizeKey]
	if ok {
		snapSizeInt, err := strconv.Atoi(snapSize)
		return int64(snapSizeInt), err
	}

	// 2. find the source volume size
	return getVolumeAllocatedCapacity(sourceVolume, apiClient)
}

func getVolumeAllocatedCapacity(volumeName string, apiClient client.Client) (int64, error) {
	volume := apisv1alpha1.LocalVolume{}
	if err := apiClient.Get(context.Background(), types.NamespacedName{Name: volumeName}, &volume); err != nil {
		return 0, err
	}

	// use the source volume allocated size in status rather than the capacity in spec
	return volume.Status.AllocatedCapacityBytes, nil
}

// getVolumeAccessibility returns the access topology from the given volume
func getVolumeAccessibility(volumeName string, apiClient client.Client) (apisv1alpha1.AccessibilityTopology, error) {
	vol := apisv1alpha1.LocalVolume{}
	if err := apiClient.Get(context.Background(), types.NamespacedName{Name: volumeName}, &vol); err != nil {
		return apisv1alpha1.AccessibilityTopology{}, err
	}

	return vol.Spec.Accessibility, nil
}

// getVolumeSnapshotAccessibility returns the access topology from the given volume
func getSourceVolumeFromSnapshot(volumeSnapshotName string, apiClient client.Client) (*apisv1alpha1.LocalVolume, error) {
	volumeSnapshot := apisv1alpha1.LocalVolumeSnapshot{}
	if err := apiClient.Get(context.Background(), types.NamespacedName{Name: volumeSnapshotName}, &volumeSnapshot); err != nil {
		return nil, err
	}

	volume := apisv1alpha1.LocalVolume{}
	if err := apiClient.Get(context.Background(), types.NamespacedName{Name: volumeSnapshot.Spec.SourceVolume}, &volume); err != nil {
		return nil, err
	}

	return &volume, nil
}

func listVolumeSnapshots(volumeName string, apiClient client.Client) ([]apisv1alpha1.LocalVolumeSnapshot, error) {
	snapList := apisv1alpha1.LocalVolumeSnapshotList{}
	if err := apiClient.List(context.TODO(), &snapList, &client.ListOptions{
		FieldSelector: fields.ParseSelectorOrDie("spec.sourceVolume=" + volumeName)}); err != nil {
		return nil, err
	}
	return snapList.Items, nil
}
