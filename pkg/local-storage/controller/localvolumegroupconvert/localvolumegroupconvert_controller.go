package localvolumegroupconvert

import (
	"context"
	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage"
    apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

// Add creates a new LocalVolumeConvert Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLocalVolumeGroupConvert{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		// storageMember is a global variable
		storageMember: member.Member(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("LocalVolumeGroupConvert-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource LocalVolumeGroupConvert
	err = c.Watch(&source.Kind{Type: &apisv1alpha1.LocalVolumeGroupConvert{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLocalVolumeGroupConvert implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLocalVolumeGroupConvert{}

// ReconcileLocalVolumeGroupConvert reconciles a LocalVolumeGroupConvert object
type ReconcileLocalVolumeGroupConvert struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme

	storageMember apis.LocalStorageMember
}

// Reconcile reads that state of the cluster for a LocalVolumeGroupConvert object and makes changes based on the state read
// and what is in the LocalVolumeGroupConvert.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLocalVolumeGroupConvert) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the LocalVolumeGroupConvert instance
	instance := &apisv1alpha1.LocalVolumeGroupConvert{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	r.storageMember.Controller().ReconcileVolumeGroupConvert(instance)

	return reconcile.Result{}, nil
}
