package localdiskclaim

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdiskclaim"
	"time"

	ldmv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-disk-manager/v1alpha1"
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
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new LocalDiskClaim Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.

const (
	// RequeueInterval Requeue every 5 seconds
	RequeueInterval = time.Second * 5
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
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("localdiskclaim-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource LocalDiskClaim
	err = c.Watch(&source.Kind{Type: &ldmv1alpha1.LocalDiskClaim{}}, &handler.EnqueueRequestForObject{})
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
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a LocalDiskClaim object and makes changes based on the state read
// and what is in the LocalDiskClaim.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLocalDiskClaim) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Infof("Reconcile LocalDiskClaim %s", req.Name)
	ldcHandler := localdiskclaim.NewLocalDiskClaimHandler(r.Client, r.Recorder)

	ldc, err := ldcHandler.GetLocalDiskClaim(req.NamespacedName)
	if err != nil {
		log.WithError(err).Errorf("Get localdiskclaim fail, due to error: %v", err)
		return reconcile.Result{}, err
	}

	if ldc != nil {
		ldcHandler = ldcHandler.For(*ldc.DeepCopy())
	} else {
		// Not found
		return reconcile.Result{}, nil
	}

	switch ldcHandler.Phase() {
	case ldmv1alpha1.DiskClaimStatusEmpty:
		fallthrough
	case ldmv1alpha1.LocalDiskClaimStatusPending:
		if err = ldcHandler.AssignFreeDisk(); err != nil {
			r.Recorder.Eventf(ldc, v1.EventTypeWarning, "LocalDiskClaimFail", "Assign free disk fail, due to error: %v", err)
			log.WithError(err).Errorf("Assign free disk for locadiskclaim %v/%v fail, will try after %v", ldc.GetNamespace(), ldc.GetName(), RequeueInterval)
			return reconcile.Result{RequeueAfter: RequeueInterval}, nil
		}

	case ldmv1alpha1.LocalDiskClaimStatusBound:
		// TODO: handle delete events
	default:
		log.Warningf("LocalDiskClaim %s status %v is UNKNOWN", ldc.Name, ldcHandler.Phase())
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}
