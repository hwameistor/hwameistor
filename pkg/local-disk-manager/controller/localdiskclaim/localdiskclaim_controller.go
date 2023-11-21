package localdiskclaim

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdiskclaim"
)

const (
	// RequeueInterval Requeue every 1 seconds
	RequeueInterval = time.Second * 1
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLocalDiskClaim{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("localdiskclaim-controller"),
		diskClaimHandler: localdiskclaim.NewLocalDiskClaimHandler(mgr.GetClient(),
			mgr.GetEventRecorderFor("localdiskclaim-controller")),
	}
}

// add a new Controller to mgr with r as reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("localdiskclaim-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource LocalDiskClaim
	err = c.Watch(&source.Kind{Type: &v1alpha1.LocalDiskClaim{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLocalDiskClaim implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLocalDiskClaim{}

// ReconcileLocalDiskClaim reconciles a LocalDiskClaim object
type ReconcileLocalDiskClaim struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client.Client
	diskClaimHandler *localdiskclaim.Handler
	Scheme           *runtime.Scheme
	Recorder         record.EventRecorder
}

// Reconcile for localdiskclaim instance according to request params
func (r *ReconcileLocalDiskClaim) Reconcile(_ context.Context, req reconcile.Request) (reconcile.Result, error) {
	log.Infof("Reconcile LocalDiskClaim %s", req.Name)
	var (
		err       error
		result    reconcile.Result
		diskClaim *v1alpha1.LocalDiskClaim
	)

	diskClaim, err = r.diskClaimHandler.GetLocalDiskClaim(req.NamespacedName)
	if err != nil {
		log.WithError(err).Errorf("Get localdiskclaim fail, due to error: %v", err)
		return result, err
	} else if diskClaim == nil {
		return reconcile.Result{}, nil
	}

	r.diskClaimHandler.For(diskClaim)
	switch diskClaim.Status.Status {
	case v1alpha1.DiskClaimStatusEmpty:
		err = r.processDiskClaimEmpty(diskClaim)
	case v1alpha1.LocalDiskClaimStatusPending, v1alpha1.LocalDiskClaimStatusExtending:
		err = r.processDiskClaimPending(diskClaim)
	case v1alpha1.LocalDiskClaimStatusBound:
		err = r.processDiskClaimBound(diskClaim)
	case v1alpha1.LocalDiskClaimStatusConsumed:
		err = r.processDiskClaimConsumed(diskClaim)
	case v1alpha1.LocalDiskClaimStatusToBeDeleted:
		err = r.processDiskClaimToBeDeleted(diskClaim)
	case v1alpha1.LocalDiskClaimStatusDeleted:
		err = r.processDiskClaimDeleted(diskClaim)
	default:
		log.Warningf("LocalDiskClaim %s status %v is UNKNOWN", diskClaim.Name, diskClaim.Status.Status)
	}

	if err != nil {
		log.WithError(err).Errorf("Failed to reconcile localdiskclaim %v", diskClaim.GetName())
		result.RequeueAfter = RequeueInterval
	}

	return result, err
}

// processDiskClaimEmpty update status to Pending
func (r *ReconcileLocalDiskClaim) processDiskClaimEmpty(diskClaim *v1alpha1.LocalDiskClaim) error {
	logCtx := log.Fields{"name": diskClaim.Name}
	log.WithFields(logCtx).Info("Start processing Empty localdiskclaim")

	r.diskClaimHandler.SetupClaimStatus(v1alpha1.LocalDiskClaimStatusPending)
	return r.diskClaimHandler.UpdateClaimStatus()
}

// processDiskClaimPending assign free disks for this request according claim.spec.description
func (r *ReconcileLocalDiskClaim) processDiskClaimPending(diskClaim *v1alpha1.LocalDiskClaim) error {
	logCtx := log.Fields{"name": diskClaim.Name}
	log.WithFields(logCtx).Info("Start processing Pending localdiskclaim")

	var err error
	// don't assign disks when owner is not specified
	if len(diskClaim.Spec.Owner) == 0 {
		err = fmt.Errorf("owner must be specified")
		r.Recorder.Eventf(diskClaim, v1.EventTypeWarning, v1alpha1.LocalDiskClaimEventReasonAssignFail,
			"Assign free disk fail, due to error: %v", err)
		return err
	}

	// assign disks or update status bound according to diskRefs
	switch {
	case len(diskClaim.Spec.DiskRefs) == 0:
		if err = r.diskClaimHandler.AssignFreeDisk(); err != nil {
			r.Recorder.Eventf(diskClaim, v1.EventTypeWarning, v1alpha1.LocalDiskClaimEventReasonAssignFail,
				"Assign free disk fail, due to error: %v", err)
			log.WithError(err).WithFields(logCtx).Errorf("Assign free disk for locadiskclaim %v/%v fail, "+
				"will try after %v", diskClaim.GetNamespace(), diskClaim.GetName(), RequeueInterval)
			return err
		}
		log.WithFields(logCtx).Infof("Assign free disk succeed,at least one disk was assigned, "+
			"localdiskclaim %v/%v", diskClaim.GetNamespace(), diskClaim.GetName())
		r.Recorder.Eventf(diskClaim, v1.EventTypeNormal, v1alpha1.LocalDiskClaimEventReasonAssign,
			"Assign free disk succeed,at least one disk was assigned")

		// Update claim.spec.diskRefs according to disk status
		if err = r.diskClaimHandler.PatchBoundDiskRef(); err != nil {
			r.Recorder.Eventf(diskClaim, v1.EventTypeWarning, v1alpha1.LocalDiskClaimEventReasonExtendFail,
				"Failed to extend for localdiskclaim, due to error: %v", err)
			log.WithError(err).Errorf("Failed to extend for localdiskclaim  %v/%v fail, "+
				"will try after %v", diskClaim.GetNamespace(), diskClaim.GetName(), RequeueInterval)
			return err
		}
		log.WithFields(logCtx).Infof("extend for localdiskclaim succeed,"+
			"localdiskclaim %v/%v", diskClaim.GetNamespace(), diskClaim.GetName())
		return err
	default:
		log.WithFields(log.Fields{"diskClaim": diskClaim.GetName(), "assignedDisks": len(diskClaim.Spec.DiskRefs)}).Info("Found Bounded disks")
	}

	return r.diskClaimHandler.UpdateClaimStatusToBound()
}

// processDiskClaimBound check need to assign new disk or not
func (r *ReconcileLocalDiskClaim) processDiskClaimBound(diskClaim *v1alpha1.LocalDiskClaim) error {
	logCtx := log.Fields{"name": diskClaim.Name}
	log.WithFields(logCtx).Info("Start processing Bound localdiskclaim")

	if len(r.diskClaimHandler.DiskRefs()) == 0 {
		log.WithField("diskClaim", diskClaim.GetName()).
			Info("No disk(s) bound, change status to Pending and try to reclaim")
		r.Recorder.Eventf(diskClaim, v1.EventTypeWarning, v1alpha1.LocalDiskClaimEventReasonBoundFail,
			"No disk(s) bound, change status to Pending and try to reclaim")
		r.diskClaimHandler.SetupClaimStatus(v1alpha1.LocalDiskClaimStatusPending)
		return r.diskClaimHandler.UpdateClaimStatus()
	}

	return nil
}

func (r *ReconcileLocalDiskClaim) processDiskClaimConsumed(_ *v1alpha1.LocalDiskClaim) error {
	// Update claim.status to ToBeDeleted if disks backing this claim have been consumed
	r.diskClaimHandler.SetupClaimStatus(v1alpha1.LocalDiskClaimStatusToBeDeleted)
	return r.diskClaimHandler.UpdateClaimStatus()
}

func (r *ReconcileLocalDiskClaim) processDiskClaimToBeDeleted(diskClaim *v1alpha1.LocalDiskClaim) error {
	logCtx := log.Fields{"name": diskClaim.Name}
	log.WithFields(logCtx).Info("Start processing ToBeDeleted localdiskclaim")

	// Nothing to do here
	r.diskClaimHandler.SetupClaimStatus(v1alpha1.LocalDiskClaimStatusDeleted)
	return r.diskClaimHandler.UpdateClaimStatus()
}

func (r *ReconcileLocalDiskClaim) processDiskClaimDeleted(diskClaim *v1alpha1.LocalDiskClaim) error {
	logCtx := log.Fields{"name": diskClaim.Name}
	log.WithFields(logCtx).Info("Start processing Deleted localdiskclaim")

	// Delete this claim
	return r.diskClaimHandler.DeleteLocalDiskClaim()
}
