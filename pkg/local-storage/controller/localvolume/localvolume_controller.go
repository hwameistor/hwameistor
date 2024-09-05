package localvolume

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	apis "github.com/hwameistor/hwameistor/pkg/apis/hwameistor"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new LocalVolume Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLocalVolume{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		// storageMember is a global variable
		storageMember: member.Member(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	replicaToVolumeRequestFunc := handler.EnqueueRequestsFromMapFunc(
		func(_ context.Context, a client.Object) []reconcile.Request {
			replica, ok := a.(*apisv1alpha1.LocalVolumeReplica)
			if !ok {
				return []reconcile.Request{}
			}

			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{Name: replica.Spec.VolumeName},
				},
			}
		})

	// Create a new controller
	c, err := controller.New("localvolume-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource LocalVolume
	err = c.Watch(source.Kind(mgr.GetCache(), &apisv1alpha1.LocalVolume{}), &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(source.Kind(mgr.GetCache(), &apisv1alpha1.LocalVolumeReplica{}), replicaToVolumeRequestFunc)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLocalVolume implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLocalVolume{}

// ReconcileLocalVolume reconciles a LocalVolume object
type ReconcileLocalVolume struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme

	storageMember apis.LocalStorageMember
}

// Reconcile reads that state of the cluster for a LocalVolume object and makes changes based on the state read
// and what is in the LocalVolume.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLocalVolume) Reconcile(_ context.Context, request reconcile.Request) (reconcile.Result, error) {

	// Fetch the LocalVolume instance
	instance := &apisv1alpha1.LocalVolume{}
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

	r.storageMember.Controller().ReconcileVolume(instance)

	// each Node should take the assignment from the Volume if has
	r.storageMember.Node().TakeVolumeReplicaTaskAssignment(instance)

	return reconcile.Result{}, nil
}
