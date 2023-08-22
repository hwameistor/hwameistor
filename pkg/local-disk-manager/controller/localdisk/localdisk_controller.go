package localdisk

import (
	"context"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/node/registry"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

// Add creates a new localDisk Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLocalDisk{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("localdisk-controller"),
		diskHandler: localdisk.NewLocalDiskHandler(mgr.GetClient(),
			mgr.GetEventRecorderFor("localdisk-controller")),
	}
}

// withCurrentNode filter volume request for this node
func withCurrentNode() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			disk, _ := event.Object.DeepCopyObject().(*v1alpha1.LocalDisk)
			return disk.Spec.NodeName == utils.GetNodeName()
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			disk, _ := deleteEvent.Object.DeepCopyObject().(*v1alpha1.LocalDisk)
			return disk.Spec.NodeName == utils.GetNodeName()
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			disk, _ := updateEvent.ObjectNew.DeepCopyObject().(*v1alpha1.LocalDisk)
			return disk.Spec.NodeName == utils.GetNodeName()
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			disk, _ := genericEvent.Object.DeepCopyObject().(*v1alpha1.LocalDisk)
			return disk.Spec.NodeName == utils.GetNodeName()
		},
	}
}

// withActiveDisk filter active disk
func withActiveDisk() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			disk, _ := event.Object.DeepCopyObject().(*v1alpha1.LocalDisk)
			return disk.Spec.State == v1alpha1.LocalDiskActive
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			disk, _ := deleteEvent.Object.DeepCopyObject().(*v1alpha1.LocalDisk)
			return disk.Spec.State == v1alpha1.LocalDiskActive
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			disk, _ := updateEvent.ObjectNew.DeepCopyObject().(*v1alpha1.LocalDisk)
			return disk.Spec.State == v1alpha1.LocalDiskActive
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			disk, _ := genericEvent.Object.DeepCopyObject().(*v1alpha1.LocalDisk)
			return disk.Spec.State == v1alpha1.LocalDiskActive
		},
	}
}

// add a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("localdisk-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource localDisk
	err = c.Watch(&source.Kind{Type: &v1alpha1.LocalDisk{}}, &handler.EnqueueRequestForObject{}, withCurrentNode(), withActiveDisk())
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLocalDisk implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLocalDisk{}

// ReconcileLocalDisk reconciles a localDisk object
type ReconcileLocalDisk struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client.Client
	Scheme      *runtime.Scheme
	Recorder    record.EventRecorder
	diskHandler *localdisk.Handler
}

// Reconcile localDisk instance according to disk status
func (r *ReconcileLocalDisk) Reconcile(_ context.Context, req reconcile.Request) (reconcile.Result, error) {
	log.Infof("Reconcile LocalDisk %s", req.Name)

	localDisk, err := r.diskHandler.GetLocalDisk(req.NamespacedName)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		log.WithError(err).Errorf("Failed to get localdisk")
		return reconcile.Result{}, err
	}

	r.diskHandler.For(localDisk)
	// reconcile localdisk according disk status
	switch localDisk.Status.State {
	case v1alpha1.LocalDiskEmpty:
		err = r.processDiskEmpty(localDisk)
	case v1alpha1.LocalDiskPending:
		err = r.processDiskPending(localDisk)
	case v1alpha1.LocalDiskAvailable:
		err = r.processDiskAvailable(localDisk)
	case v1alpha1.LocalDiskBound:
		err = r.processDiskBound(localDisk)
	default:
		err = fmt.Errorf("invalid disk state: %v", localDisk.Status.State)
	}

	if err != nil {
		log.WithError(err).WithField("name", localDisk.Name).Error("Failed to reconcile disk, retry later")
	}

	return reconcile.Result{}, err
}

// processDiskEmpty update disk status to Pending
func (r *ReconcileLocalDisk) processDiskEmpty(disk *v1alpha1.LocalDisk) error {
	logCtx := log.Fields{"name": disk.Name}
	log.WithFields(logCtx).Info("Start to processing Empty localdisk")

	return r.updateDiskStatusPending(disk)
}

// processDiskPending update disk status to Bound or Available
// according attributes, partitions, filesystem on it
func (r *ReconcileLocalDisk) processDiskPending(disk *v1alpha1.LocalDisk) error {
	logCtx := log.Fields{"name": disk.Name}
	log.WithFields(logCtx).Info("Start to processing Empty localdisk")

	// Update disk status if found partition or filesystem or diskRed on it
	if disk.Spec.HasPartition || disk.Spec.ClaimRef != nil {
		return r.updateDiskStatusBound(disk)
	}

	return r.updateDiskStatusAvailable(disk)
}

// processDiskAvailable update disk status to Bound
func (r *ReconcileLocalDisk) processDiskAvailable(disk *v1alpha1.LocalDisk) error {
	logCtx := log.Fields{"name": disk.Name}
	log.WithFields(logCtx).Info("Start to processing Available localdisk")

	// Update disk status if found partition or filesystem or diskRed on it
	if disk.Spec.HasPartition || disk.Spec.ClaimRef != nil {
		return r.updateDiskStatusBound(disk)
	}

	// check and correct spec if needed, return directly if updated or occur error
	if updated, err := r.checkAndCorrectSpec(disk); updated || err != nil {
		return err
	}

	return nil
}

// processDiskBound update disk status to Available
func (r *ReconcileLocalDisk) processDiskBound(disk *v1alpha1.LocalDisk) error {
	logCtx := log.Fields{"name": disk.Name}
	log.WithFields(logCtx).Info("Start to processing Bound localdisk")

	var err error
	// Check if disk can be released
	if disk.Spec.ClaimRef == nil && !disk.Spec.HasPartition {
		if err = r.updateDiskStatusAvailable(disk); err != nil {
			log.WithError(err).WithFields(logCtx).Error("Failed to release disk")
			r.Recorder.Eventf(disk, v1.EventTypeWarning, v1alpha1.LocalDiskEventReasonReleaseFail,
				"Failed to release disk %v due to error: %v", disk.Name, err)
		} else {
			log.WithFields(logCtx).Info("Succeed to release disk")
			r.Recorder.Eventf(disk, v1.EventTypeNormal, v1alpha1.LocalDiskEventReasonRelease,
				"Succeed to release disk %v", disk.Name)
		}
		return err
	}

	// check and correct spec if needed, return directly if updated or occur error
	if updated, err := r.checkAndCorrectSpec(disk); updated || err != nil {
		return err
	}

	return err
}

// updateDiskStatusBound update disk status to Bound
func (r *ReconcileLocalDisk) updateDiskStatusBound(disk *v1alpha1.LocalDisk) error {
	var (
		eventReason  = v1alpha1.LocalDiskEventReasonBound
		eventType    = v1.EventTypeNormal
		eventMessage = fmt.Sprintf("Bound Disk %v succeed", disk.GetName())
	)

	r.diskHandler.SetupStatus(v1alpha1.LocalDiskBound)
	err := r.diskHandler.UpdateStatus()
	if err != nil {
		eventReason = v1alpha1.LocalDiskEventReasonBoundFail
		eventType = v1.EventTypeWarning
		eventMessage = fmt.Sprintf("Bound Disk %v failed", disk.GetName())
	}

	r.Recorder.Eventf(disk, eventType, eventReason, eventMessage)
	return err
}

// updateDiskStatusPending update disk status to Pending
func (r *ReconcileLocalDisk) updateDiskStatusPending(disk *v1alpha1.LocalDisk) error {
	var (
		eventReason  = v1alpha1.LocalDiskEventReasonPending
		eventType    = v1.EventTypeNormal
		eventMessage = fmt.Sprintf("Succeed found a new disk %v", disk.GetName())
	)

	r.diskHandler.SetupStatus(v1alpha1.LocalDiskPending)
	err := r.diskHandler.UpdateStatus()
	if err != nil {
		eventReason = v1alpha1.LocalDiskEventReasonPendingFail
		eventType = v1.EventTypeWarning
		eventMessage = fmt.Sprintf("Failed to update disk %v status to Pending due to err: %v",
			disk.GetName(), err.Error())
	}

	r.Recorder.Eventf(disk, eventType, eventReason, eventMessage)
	return err
}

// updateDiskStatusAvailable update disk status to Available
func (r *ReconcileLocalDisk) updateDiskStatusAvailable(disk *v1alpha1.LocalDisk) error {
	var (
		eventReason  = v1alpha1.LocalDiskEventReasonAvailable
		eventType    = v1.EventTypeNormal
		eventMessage = fmt.Sprintf("Succeed found Available disk %v", disk.GetName())
	)

	r.diskHandler.SetupStatus(v1alpha1.LocalDiskAvailable)
	err := r.diskHandler.UpdateStatus()
	if err != nil {
		eventReason = v1alpha1.LocalDiskEventReasonAvailableFail
		eventType = v1.EventTypeWarning
		eventMessage = fmt.Sprintf("Failed to update disk %v to Available", disk.GetName())
	}

	r.Recorder.Eventf(disk, eventType, eventReason, eventMessage)
	return err
}

// checkAndCorrectSpec some fields(e.g. owner info) may change during the usage phase, try to fix them here to be correct before reconcile
// return true if modified.
func (r *ReconcileLocalDisk) checkAndCorrectSpec(disk *v1alpha1.LocalDisk) (bool, error) {
	var updated bool
	if updated, err := r.checkAndCorrectOwner(disk); err != nil {
		return updated, err
	}

	return updated, nil
}

func (r *ReconcileLocalDisk) checkAndCorrectOwner(disk *v1alpha1.LocalDisk) (bool, error) {
	switch disk.Status.State {
	case v1alpha1.LocalDiskBound, v1alpha1.LocalDiskAvailable:
		if disk.Spec.Owner == "" || disk.Spec.Owner == v1alpha1.System {
			log.WithField("disk", disk.GetName()).Info("Start to populate disk owner")
			if actualOwner, err := findDiskOwner(r.Client, disk.Spec.NodeName, disk.Spec.DevicePath); err != nil {
				return false, err
			} else if disk.Status.State == v1alpha1.LocalDiskAvailable {
				if actualOwner != disk.Spec.Owner && actualOwner != v1alpha1.System {
					log.WithField("disk", disk.GetName()).Infof("Try to update owner(actual owner: %s, origin owner: %s)",
						actualOwner, disk.Spec.Owner)
					return true, r.diskHandler.PatchDiskOwner(actualOwner)
				}
			} else if disk.Status.State == v1alpha1.LocalDiskBound {
				if actualOwner != disk.Spec.Owner {
					log.WithField("disk", disk.GetName()).Infof("Try to update owner(actual owner: %s, origin owner: %s)",
						actualOwner, disk.Spec.Owner)
					return true, r.diskHandler.PatchDiskOwner(actualOwner)
				}
			}
		}
	default:
		return false, nil
	}

	return false, nil
}

// findDiskOwner find which system owns the disk(e.g. local-storage, system)
func findDiskOwner(cli client.Client, nodeName, devPath string) (string, error) {
	// we only find disks known by our system, if disks managed by the other system we just show its owner as system
	// list disks managed by local-storage
	lsDisks, err := listDisksOwnedByLocalStorage(cli, nodeName)
	if err != nil {
		return "", err
	}
	if _, found := utils.StrFind(lsDisks, devPath); found {
		return v1alpha1.LocalStorage, nil
	}

	// list disks managed by local-disk-manager
	// discovery from local pool path(/dev/LocalDisk_Pool/...) which managed by local-disk-manager
	ldmDisks, _ := listDisksOwnedByLocalDiskManager()
	if _, found := utils.StrFind(ldmDisks, devPath); found {
		return v1alpha1.LocalDiskManager, nil
	}

	return v1alpha1.System, nil
}

func listDisksOwnedByLocalStorage(cli client.Client, nodeName string) ([]string, error) {
	// ** The best way may be to use lvm to check the list of devices belonging to local-storage **
	lsn := v1alpha1.LocalStorageNode{}
	err := cli.Get(context.Background(), types.NamespacedName{Name: nodeName}, &lsn)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// list all devices existed in storage pool
	var poolDisks []string
	for _, pool := range lsn.Status.Pools {
		for _, disk := range pool.Disks {
			poolDisks = append(poolDisks, disk.DevPath)
		}
	}

	return poolDisks, nil
}

// find disks exist in /dev/LocalDisk_Pool/disks/<disk_type>/<disk_name>
func listDisksOwnedByLocalDiskManager() ([]string, error) {
	var ldmDisks []string
	for _, disk := range registry.New().ListDisks() {
		ldmDisks = append(ldmDisks, disk.DevPath)
	}
	return ldmDisks, nil
}
