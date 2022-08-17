package localvolumemigrate

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/wxnacy/wgo/arrays"
	"k8s.io/apimachinery/pkg/types"

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

// Add creates a new LocalVolumeMigrate Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLocalVolumeMigrate{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		// storageMember is a global variable
		storageMember: member.Member(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("localvolumemigrate-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource LocalVolumeMigrate
	err = c.Watch(&source.Kind{Type: &apisv1alpha1.LocalVolumeMigrate{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLocalVolumeMigrate implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLocalVolumeMigrate{}

// ReconcileLocalVolumeMigrate reconciles a LocalVolumeMigrate object
type ReconcileLocalVolumeMigrate struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme

	storageMember apis.LocalStorageMember
}

// Reconcile reads that state of the cluster for a LocalVolumeMigrate object and makes changes based on the state read
// and what is in the LocalVolumeMigrate.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLocalVolumeMigrate) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &apisv1alpha1.LocalVolumeMigrate{}
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

	vol := &apisv1alpha1.LocalVolume{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.VolumeName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: instance.Namespace, Name: localVolumeGroupName}, lvg)
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

	var accessibilityNodeNames []string
	var errMsg error

	for _, tmpvol := range lvg.Spec.Volumes {
		if tmpvol.LocalVolumeName == "" {
			continue
		}
		vol := &apisv1alpha1.LocalVolume{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: tmpvol.LocalVolumeName}, vol); err != nil {
			if !errors.IsNotFound(err) {
				log.WithFields(log.Fields{"volName": tmpvol.LocalVolumeName, "error": err.Error()}).Error("Failed to query volume")
				return reconcile.Result{}, err
			}
		}
		for _, nodeName := range instance.Spec.TargetNodesNames {
			if arrays.ContainsString(vol.Spec.Accessibility.Nodes, nodeName) == -1 {
				accessibilityNodeNames = append(accessibilityNodeNames, nodeName)
			} else {
				accessibilityNodeNames = vol.Spec.Accessibility.Nodes
			}
		}
		if err := r.client.Update(context.TODO(), vol); err != nil {
			log.WithError(err).Errorf("ReconcileLocalVolumeMigrate Reconcile : Failed to re-configure Volume, vol.Name = %v, tmpvol.LocalVolumeName = %v", vol.Name, tmpvol.LocalVolumeName)
			errMsg = err
		}
	}

	if errMsg != nil {
		return reconcile.Result{}, errMsg
	}

	r.storageMember.Controller().ReconcileVolumeMigrate(instance)

	return reconcile.Result{}, nil
}
