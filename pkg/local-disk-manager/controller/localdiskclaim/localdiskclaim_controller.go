package localdiskclaim

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"time"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdiskclaim"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
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
	log.WithFields(logCtx).Info("Start to processing Empty localdiskclaim")

	r.diskClaimHandler.SetupClaimStatus(v1alpha1.LocalDiskClaimStatusPending)
	return r.diskClaimHandler.UpdateClaimStatus()
}

// processDiskClaimPending assign free disks for this request according claim.spec.description
func (r *ReconcileLocalDiskClaim) processDiskClaimPending(diskClaim *v1alpha1.LocalDiskClaim) error {
	logCtx := log.Fields{"name": diskClaim.Name}
	log.WithFields(logCtx).Info("Start to processing Pending localdiskclaim")
	var (
		err error
	)

	if err = r.diskClaimHandler.AssignFreeDisk(); err != nil {
		r.Recorder.Eventf(diskClaim, v1.EventTypeWarning, v1alpha1.LocalDiskClaimEventReasonAssignFail,
			"Assign free disk fail, due to error: %v", err)
		log.WithError(err).WithFields(logCtx).Errorf("Assign free disk for locadiskclaim %v/%v fail, "+
			"will try after %v", diskClaim.GetNamespace(), diskClaim.GetName(), RequeueInterval)
		return err
	}

	// Update claim.spec.diskRefs according to disk status
	if err = r.diskClaimHandler.UpdateBoundDiskRef(); err != nil {
		log.WithError(err).Errorf("Failed to extend for locadiskclaim %v fail, will try after %v",
			diskClaim.GetName(), RequeueInterval)
		return err
	}
	r.Recorder.Eventf(diskClaim, v1.EventTypeNormal, v1alpha1.LocalDiskClaimEventReasonExtend,
		"Success to extend for localdiskclaim %v", diskClaim.GetName())

	r.diskClaimHandler.SetupClaimStatus(v1alpha1.LocalDiskClaimStatusBound)
	return r.diskClaimHandler.UpdateClaimStatus()
}

// processDiskClaimBound check need to assign new disk or not
func (r *ReconcileLocalDiskClaim) processDiskClaimBound(diskClaim *v1alpha1.LocalDiskClaim) error {
	logCtx := log.Fields{"name": diskClaim.Name}
	log.WithFields(logCtx).Info("Start to processing Bound localdiskclaim")

	var (
		err error
	)

	// issue: https://github.com/hwameistor/hwameistor/issues/517
	// Update claim.spec.diskRefs according to disk status
	if err = r.diskClaimHandler.UpdateBoundDiskRef(); err != nil {
		log.WithError(err).Errorf("Failed to extend for locadiskclaim %v fail, will try after %v",
			diskClaim.GetName(), RequeueInterval)
		return err
	}

	return nil
}
