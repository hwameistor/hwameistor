package localdiskclaimexpand

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLocalDiskClaimExpand{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("localdiskclaim-controller"),
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
	err = c.Watch(&source.Kind{Type: &v1alpha1.LocalDiskClaimExpand{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLocalDiskClaimExpand implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLocalDiskClaimExpand{}

// ReconcileLocalDiskClaimExpand reconciles a LocalDiskClaim object
type ReconcileLocalDiskClaimExpand struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile for localdiskclaimexpand instance according to request params
func (r *ReconcileLocalDiskClaimExpand) Reconcile(_ context.Context, req reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
