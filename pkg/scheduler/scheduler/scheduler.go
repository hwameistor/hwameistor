package scheduler

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"os"
	"strings"
	"time"

	snapshot "github.com/kubernetes-csi/external-snapshotter/client/v6/clientset/versioned/scheme"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	storagev1lister "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	lvmscheduler "github.com/hwameistor/hwameistor/pkg/local-storage/member/controller/scheduler"
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
	// Setup Scheme for all resources of Snapshot
	if err = snapshot.AddToScheme(mgr.GetScheme()); err != nil {
		log.WithError(err).Fatal("Failed to setup scheme for snapshot resources")
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
	// fixme - remove from scheduler config
	log.Debug("Do nothing here(to be removed)")
	return nil
}

func (s *Scheduler) Unreserve(pod *corev1.Pod, node string) error {
	// fixme - remove from scheduler config
	log.Debug("Do nothing here(to be removed)")
	return nil
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

func (s *Scheduler) Score(pod *corev1.Pod, node string) (int64, error) {
	_, lvmNewPVCs, _, diskNewPVCs, err := s.getHwameiStorPVCs(pod)
	if err != nil || len(append(lvmNewPVCs, diskNewPVCs...)) == 0 {
		return 0, err
	}

	var scoreAll int64
	var lvmScore int64
	if len(lvmNewPVCs) > 0 {
		lvmScore, err = s.lvmScheduler.Score(lvmNewPVCs, node)
		if err != nil {
			return 0, err
		}
		scoreAll += framework.MaxNodeScore

		log.WithFields(log.Fields{
			"lvmScore": lvmScore,
			"scoreAll": scoreAll,
			"volumes":  listVolumes(lvmNewPVCs),
			"node":     node,
		}).Debug("node score for lvm-volumes")
	}

	var diskScore int64
	if len(diskNewPVCs) > 0 {
		diskScore, err = s.diskScheduler.Score(diskNewPVCs, node)
		if err != nil {
			return 0, err
		}
		scoreAll += framework.MaxNodeScore

		log.WithFields(log.Fields{
			"diskScore": diskScore,
			"scoreAll":  scoreAll,
			"volumes":   listVolumes(diskNewPVCs),
			"node":      node,
		}).Debug("node score for disk-volumes")
	}

	score := (float64(lvmScore+diskScore) / float64(scoreAll)) * float64(framework.MaxNodeScore)
	log.WithFields(log.Fields{
		"volumes":    listVolumes(append(lvmNewPVCs, diskNewPVCs...)),
		"node":       node,
		"totalScore": score,
		"scoreAll":   scoreAll,
	}).Debug("node score for volumes")
	return int64(score), nil
}

// return: lvmProvisionedClaims, lvmNewClaims, diskProvisionedClaims, diskNewClaims, error
func (s *Scheduler) getHwameiStorPVCs(pod *corev1.Pod) ([]*corev1.PersistentVolumeClaim, []*corev1.PersistentVolumeClaim, []*corev1.PersistentVolumeClaim, []*corev1.PersistentVolumeClaim, error) {
	var lvmProvisionedClaims []*corev1.PersistentVolumeClaim
	var lvmNewClaims []*corev1.PersistentVolumeClaim
	var diskProvisionedClaims []*corev1.PersistentVolumeClaim
	var diskNewClaims []*corev1.PersistentVolumeClaim

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
			// NOTES: in static volume(e.g. fluid volume), StorageClass may not exist even if StorageClassName is not empty
			// treat this volume as a non-hwameistor volume
			if errors.IsNotFound(err) {
				log.WithField("StorageClassName", *pvc.Spec.StorageClassName).Debugf("Ignore volume %s because of StorageClass in not found", pvc.Name)
				continue
			}
			return lvmProvisionedClaims, lvmNewClaims, diskProvisionedClaims, diskNewClaims, err
		}

		switch sc.Provisioner {
		case lvmCSIDriverName:
			if pvc.Status.Phase == corev1.ClaimBound {
				lvmProvisionedClaims = append(lvmProvisionedClaims, pvc)
			} else if pvc.Status.Phase == corev1.ClaimPending {
				lvmNewClaims = append(lvmNewClaims, pvc)
			} else {
				return lvmProvisionedClaims, lvmNewClaims, diskProvisionedClaims, diskNewClaims, fmt.Errorf("unhealthy HwameiStor LVM pvc")
			}

		case diskCSIDriverName:
			if pvc.Status.Phase == corev1.ClaimBound {
				diskProvisionedClaims = append(diskProvisionedClaims, pvc)
			} else if pvc.Status.Phase == corev1.ClaimPending {
				diskNewClaims = append(diskNewClaims, pvc)
			} else {
				return lvmProvisionedClaims, lvmNewClaims, diskProvisionedClaims, diskNewClaims, fmt.Errorf("unhealthy HwameiStor Disk pvc")
			}

		default:
			continue
		}
	}

	return lvmProvisionedClaims, lvmNewClaims, diskProvisionedClaims, diskNewClaims, nil
}

func listVolumes(pvs []*corev1.PersistentVolumeClaim) (s string) {
	for _, pv := range pvs {
		s = s + "," + pv.GetName()
	}
	return strings.TrimPrefix(s, ",")
}
