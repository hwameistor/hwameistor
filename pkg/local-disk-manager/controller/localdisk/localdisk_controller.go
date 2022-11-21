package localdisk

import (
	"context"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
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

// add a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("localdisk-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource localDisk
	err = c.Watch(&source.Kind{Type: &v1alpha1.LocalDisk{}}, &handler.EnqueueRequestForObject{}, withCurrentNode())
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
	case v1alpha1.LocalDiskAvailable:
		err = r.processDiskAvailable(localDisk)
	case v1alpha1.LocalDiskBound:
		err = r.processDiskBound(localDisk)
	case v1alpha1.LocalDiskReserved:
		err = r.processDiskReserved(localDisk)
	default:
		err = fmt.Errorf("invalid disk state: %v", localDisk.Status.State)
	}

	if err != nil {
		log.WithError(err).WithField("name", localDisk.Name).Error("Failed to reconcile disk, retry later")
	}

	return reconcile.Result{}, err
}

// processDiskEmpty update disk status to Bound or Available
// according attributes, partitions, filesystem on it
func (r *ReconcileLocalDisk) processDiskEmpty(disk *v1alpha1.LocalDisk) error {
	logCtx := log.Fields{"name": disk.Name}
	log.WithFields(logCtx).Info("Start to processing Empty localdisk")

	if disk.Spec.HasPartition || disk.Spec.ClaimRef != nil {
		return r.updateDiskStatusBound(disk)
	}

	return r.updateDiskStatusAvailable(disk)
}

// processDiskEmpty update disk status to Bound or Reserved
func (r *ReconcileLocalDisk) processDiskAvailable(disk *v1alpha1.LocalDisk) error {
	logCtx := log.Fields{"name": disk.Name}
	log.WithFields(logCtx).Info("Start to processing Available localdisk")

	var (
		diskBound = false
		err       error
	)

	if disk.Spec.ClaimRef != nil || disk.Spec.HasPartition {
		diskBound = true
	}

	// Update status to Bound immediately, if found filesystem or partitions or claimRef on it
	if diskBound {
		err = r.updateDiskStatusBound(disk)
	} else if disk.Spec.Reserved {
		err = r.updateDiskStatusReserved(disk)
	}

	return err
}

// processDiskEmpty update disk status to Available
func (r *ReconcileLocalDisk) processDiskReserved(disk *v1alpha1.LocalDisk) error {
	logCtx := log.Fields{"name": disk.Name}
	log.WithFields(logCtx).Info("Start to processing Reserved localdisk")

	var (
		err error
	)

	// Update to Bound if detect already be in used
	if disk.Spec.ClaimRef != nil || disk.Spec.HasPartition {
		r.diskHandler.SetupStatus(v1alpha1.LocalDiskBound)
		if err = r.diskHandler.UpdateStatus(); err == nil {
			log.WithFields(logCtx).Info("Update a Reserved disk state to Bound")
			r.Recorder.Eventf(disk, v1.EventTypeNormal, v1alpha1.LocalDiskEventReasonBound, "Update disk %v state from Reserved to Bound, "+
				"due to detect partition or filesystem on it", disk.Name)
		}
	}

	return err
}

// processDiskEmpty update disk status to Available
func (r *ReconcileLocalDisk) processDiskBound(disk *v1alpha1.LocalDisk) error {
	logCtx := log.Fields{"name": disk.Name}
	log.WithFields(logCtx).Info("Start to processing Bound localdisk")

	var (
		err error
	)

	// Check if disk can be released
	if disk.Spec.ClaimRef == nil && !disk.Spec.HasPartition {
		if err = r.updateDiskStatusAvailable(disk); err != nil {
			log.WithError(err).WithFields(logCtx).Error("Failed to release disk")
			r.Recorder.Eventf(disk, v1.EventTypeWarning, v1alpha1.LocalDiskEventReasonReleaseFail, "Failed to release disk %v, "+
				"due to error: %v", disk.Name, err)
		} else {
			log.WithFields(logCtx).Info("Succeed to release disk")
			r.Recorder.Eventf(disk, v1.EventTypeNormal, v1alpha1.LocalDiskEventReasonRelease, "Succeed to release disk %v", disk.Name)
		}
	}

	return err
}

// reservedDisk update disk status to Reserved
func (r *ReconcileLocalDisk) updateDiskStatusReserved(disk *v1alpha1.LocalDisk) error {
	var (
		eventReason  = v1alpha1.LocalDiskEventReasonReserved
		eventType    = v1.EventTypeNormal
		eventMessage = fmt.Sprintf("Reserve Disk %v succeed", disk.GetName())
	)

	r.diskHandler.SetupStatus(v1alpha1.LocalDiskReserved)
	err := r.diskHandler.UpdateStatus()
	if err != nil {
		eventReason = v1alpha1.LocalDiskEventReasonReservedFail
		eventType = v1.EventTypeWarning
		eventMessage = fmt.Sprintf("Reserve Disk %v failed", disk.GetName())
	}

	r.Recorder.Eventf(disk, eventType, eventReason, eventMessage)
	return err
}

// reservedDisk update disk status to Bound
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

// reservedDisk update disk status to Available
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
