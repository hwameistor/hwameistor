package scheduler

import (
	"fmt"
	"os"
	"time"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	lvmscheduler "github.com/hwameistor/hwameistor/pkg/local-storage/member/controller/scheduler"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	storagev1lister "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// VolumeScheduler is to scheduler hwameistor volume
type Scheduler struct {
	lvmScheduler  VolumeScheduler
	diskScheduler VolumeScheduler

	pvLister  corev1lister.PersistentVolumeLister
	pvcLister corev1lister.PersistentVolumeClaimLister
	scLister  storagev1lister.StorageClassLister
}

// NewDataCache creates a cache instance
func NewScheduler(f framework.Handle) *Scheduler {

	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.WithError(err).Fatal("Failed to construct the cluster config")
	}

	options := manager.Options{
		MetricsBindAddress: "0", // disable metrics
	}

	mgr, err := manager.New(cfg, options)
	if err != nil {
		klog.V(1).Info(err)
		os.Exit(1)
	}

	// Setup Scheme for all resources of Local Storage
	if err := v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.WithError(err).Fatal("Failed to setup scheme for local-storage resources")
	}

	apiClient := mgr.GetClient()
	hwameiStorCache := mgr.GetCache()

	sche := Scheduler{
		pvLister:  f.SharedInformerFactory().Core().V1().PersistentVolumes().Lister(),
		pvcLister: f.SharedInformerFactory().Core().V1().PersistentVolumeClaims().Lister(),
		scLister:  f.SharedInformerFactory().Storage().V1().StorageClasses().Lister(),
	}

	ctx := signals.SetupSignalHandler()
	go func() {
		mgr.Start(ctx)
	}()
	replicaScheduler := lvmscheduler.New(apiClient, hwameiStorCache, 1000)
	// wait for cache synced
	for {
		if hwameiStorCache.WaitForCacheSync(ctx) {
			break
		}
		time.Sleep(time.Second * 1)
	}
	replicaScheduler.Init()

	sche.lvmScheduler = NewLVMVolumeScheduler(f, replicaScheduler, hwameiStorCache, apiClient)
	sche.diskScheduler = NewDiskVolumeScheduler(f)

	return &sche
}

func (s *Scheduler) Reserve(pod *corev1.Pod, node string) error {
	_, _, _, diskNewPVCs, err := s.getHwameiStorPVCs(pod)
	if err != nil {
		return err
	}

	return s.diskScheduler.Reserve(diskNewPVCs, node)
}

func (s *Scheduler) Unreserve(pod *corev1.Pod, node string) error {
	_, _, _, diskNewPVCs, err := s.getHwameiStorPVCs(pod)
	if err != nil {
		return err
	}

	return s.diskScheduler.Unreserve(diskNewPVCs, node)
}

func (s *Scheduler) Filter(pod *corev1.Pod, node *corev1.Node) (bool, error) {
	lvmProvisionedPVCs, lvmNewPVCs, diskProvisionedPVCs, diskNewPVCs, err := s.getHwameiStorPVCs(pod)
	if err != nil {
		return false, err
	}
	// figure out the existing local volume associated to the PVC, and send it to the scheduler's filter
	existingLocalVolumes := []string{}
	for _, pvc := range lvmProvisionedPVCs {
		pv, err := s.pvLister.Get(pvc.Spec.VolumeName)
		if err != nil {
			log.WithFields(log.Fields{"pvc": pvc.Name, "namespace": pvc.Namespace, "pv": pvc.Spec.VolumeName}).WithError(err).Error("Failed to get a Provisioned PVC's PV")
			return false, err
		}
		if pv.Spec.CSI == nil || len(pv.Spec.CSI.VolumeHandle) == 0 {
			log.WithFields(log.Fields{"pvc": pvc.Name, "namespace": pvc.Namespace, "pv": pvc.Spec.VolumeName}).Error("Wrong PV status of a Provisioned PVC")
			return false, fmt.Errorf("wrong pv")
		}
		existingLocalVolumes = append(existingLocalVolumes, pv.Spec.CSI.VolumeHandle)
	}
	canSchedule, err := s.lvmScheduler.Filter(existingLocalVolumes, lvmNewPVCs, node)
	if err != nil {
		return false, err
	}
	if !canSchedule {
		return false, fmt.Errorf("can't schedule the LVM volume to node %s", node.Name)
	}

	// figure out the existing local volume associated to the PVC, and send it to the scheduler's filter
	existingLocalVolumes = []string{}
	for _, pvc := range diskProvisionedPVCs {
		pv, err := s.pvLister.Get(pvc.Spec.VolumeName)
		if err != nil {
			log.WithFields(log.Fields{"pvc": pvc.Name, "namespace": pvc.Namespace, "pv": pvc.Spec.VolumeName}).WithError(err).Error("Failed to get a Provisioned PVC's PV")
			return false, err
		}
		if pv.Spec.CSI == nil || len(pv.Spec.CSI.VolumeHandle) == 0 {
			log.WithFields(log.Fields{"pvc": pvc.Name, "namespace": pvc.Namespace, "pv": pvc.Spec.VolumeName}).Error("Wrong PV status of a Provisioned PVC")
			return false, fmt.Errorf("wrong pv")
		}
		existingLocalVolumes = append(existingLocalVolumes, pv.Spec.CSI.VolumeHandle)
	}

	return s.diskScheduler.Filter(existingLocalVolumes, diskNewPVCs, node)
}

// return: lvmProvisionedClaims, lvmNewClaims, diskProvisionedClaims, diskNewClaims, error
func (s *Scheduler) getHwameiStorPVCs(pod *corev1.Pod) ([]*corev1.PersistentVolumeClaim, []*corev1.PersistentVolumeClaim, []*corev1.PersistentVolumeClaim, []*corev1.PersistentVolumeClaim, error) {
	lvmProvisionedClaims := []*corev1.PersistentVolumeClaim{}
	lvmNewClaims := []*corev1.PersistentVolumeClaim{}
	diskProvisionedClaims := []*corev1.PersistentVolumeClaim{}
	diskNewClaims := []*corev1.PersistentVolumeClaim{}

	lvmCSIDriverName := s.lvmScheduler.CSIDriverName()
	diskCSIDriverName := s.diskScheduler.CSIDriverName()

	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim == nil {
			continue
		}
		pvc, err := s.pvcLister.PersistentVolumeClaims(pod.Namespace).Get(vol.PersistentVolumeClaim.ClaimName)
		if err != nil {
			// if pvc can't be found in the cluster, the pod should not be able to be scheduled
			return lvmProvisionedClaims, lvmNewClaims, diskProvisionedClaims, diskNewClaims, err
		}
		if pvc.Spec.StorageClassName == nil {
			// should not be the CSI pvc, ignore
			continue
		}
		sc, err := s.scLister.Get(*pvc.Spec.StorageClassName)
		if err != nil {
			// can't found storageclass in the cluster, the pod should not be able to be scheduled
			return lvmProvisionedClaims, lvmNewClaims, diskProvisionedClaims, diskNewClaims, err
		}
		if sc.Provisioner == lvmCSIDriverName {
			if pvc.Status.Phase == corev1.ClaimBound {
				lvmProvisionedClaims = append(lvmProvisionedClaims, pvc)
			} else if pvc.Status.Phase == corev1.ClaimPending {
				lvmNewClaims = append(lvmNewClaims, pvc)
			} else {
				return lvmProvisionedClaims, lvmNewClaims, diskProvisionedClaims, diskNewClaims, fmt.Errorf("unhealthy HwameiStor LVM pvc")
			}
		}
		if sc.Provisioner == diskCSIDriverName {
			if pvc.Status.Phase == corev1.ClaimBound {
				diskProvisionedClaims = append(diskProvisionedClaims, pvc)
			} else if pvc.Status.Phase == corev1.ClaimPending {
				diskNewClaims = append(diskNewClaims, pvc)
			} else {
				return lvmProvisionedClaims, lvmNewClaims, diskProvisionedClaims, diskNewClaims, fmt.Errorf("unhealthy HwameiStor Disk pvc")
			}
		}
	}

	return lvmProvisionedClaims, lvmNewClaims, diskProvisionedClaims, diskNewClaims, nil
}
