package localdiskvolume

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdiskvolume"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new LocalDiskVolume Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLocalDiskVolume{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("localdiskvolume-controller"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("localdiskvolume-controller", mgr, controller.Options{
		Reconciler:              r,
		MaxConcurrentReconciles: 1,
	})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource LocalDiskVolume
	err = c.Watch(&source.Kind{Type: &v1alpha1.LocalDiskVolume{}}, &handler.EnqueueRequestForObject{}, withCurrentNode())
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner LocalDiskVolume
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.LocalDiskVolume{},
	})
	if err != nil {
		return err
	}

	return nil
}

// withCurrentNode filter volume request for this node
func withCurrentNode() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			volume, _ := event.Object.DeepCopyObject().(*v1alpha1.LocalDiskVolume)
			return len(volume.Spec.Accessibility.Nodes) > 0 && volume.Spec.Accessibility.Nodes[0] == utils.GetNodeName()
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			volume, _ := deleteEvent.Object.DeepCopyObject().(*v1alpha1.LocalDiskVolume)
			return len(volume.Spec.Accessibility.Nodes) > 0 && volume.Spec.Accessibility.Nodes[0] == utils.GetNodeName()
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			volume, _ := updateEvent.ObjectNew.DeepCopyObject().(*v1alpha1.LocalDiskVolume)
			return len(volume.Spec.Accessibility.Nodes) > 0 && volume.Spec.Accessibility.Nodes[0] == utils.GetNodeName()
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			volume, _ := genericEvent.Object.DeepCopyObject().(*v1alpha1.LocalDiskVolume)
			return len(volume.Spec.Accessibility.Nodes) > 0 && volume.Spec.Accessibility.Nodes[0] == utils.GetNodeName()
		},
	}
}

// blank assignment to verify that ReconcileLocalDiskVolume implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLocalDiskVolume{}

// ReconcileLocalDiskVolume reconciles a LocalDiskVolume object
type ReconcileLocalDiskVolume struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile
func (r *ReconcileLocalDiskVolume) Reconcile(_ context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.WithField("LocalDiskVolume", request.Name).Info("Reconciling LocalDiskVolume")
	var result reconcile.Result
	v, err := r.reconcileForVolume(request.NamespacedName)
	if err != nil {
		if errors.IsNotFound(err) {
			return result, nil
		}
		log.WithError(err).Errorf("Failed to new handler for LocalDiskVolume %s", request.Name)

		return result, err
	}

	// check finalizers
	if err := v.CheckFinalizers(); err != nil {
		log.WithError(err).Errorf("Failed to check finalizers for LocalDiskVolume %s", request.Name)
	}

	switch v.VolumeState() {
	// Mount Volumes
	case v1alpha1.VolumeStateNotReady, v1alpha1.VolumeStateEmpty:
		return v.ReconcileMount()

	// Unmount Volumes
	case v1alpha1.VolumeStateToBeUnmount:
		return v.ReconcileUnmount()

	// ToDelete Volume
	case v1alpha1.VolumeStateToBeDeleted:
		return v.ReconcileToBeDeleted()

	// Delete Volume
	case v1alpha1.VolumeStateDeleted:
		return v.ReconcileDeleted()

	// Volume Ready/Creating/... do nothing
	default:
		log.Infof("Volume state %s , no handling now", v.VolumeState())
	}

	return result, nil
}

func (r *ReconcileLocalDiskVolume) reconcileForVolume(name client.ObjectKey) (*localdiskvolume.DiskVolumeHandler, error) {
	volumeHandler := localdiskvolume.NewLocalDiskVolumeHandler(r.client, r.Recorder)
	volume, err := volumeHandler.GetLocalDiskVolume(name)
	if err != nil {
		return nil, err
	}

	volumeHandler.For(volume)
	return volumeHandler, nil
}
