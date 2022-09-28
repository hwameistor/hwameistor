package localdisk

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
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

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new LocalDisk Controller and adds it to the Manager. The Manager will set fields on the Controller
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

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("localdisk-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource LocalDisk
	err = c.Watch(&source.Kind{Type: &v1alpha1.LocalDisk{}}, &handler.EnqueueRequestForObject{}, withCurrentNode())
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileLocalDisk implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLocalDisk{}

// ReconcileLocalDisk reconciles a LocalDisk object
type ReconcileLocalDisk struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a LocalDisk object and makes changes based on the state read
// and what is in the LocalDisk.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileLocalDisk) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Infof("Reconcile LocalDisk %s", req.Name)

	ldHandler := localdisk.NewLocalDiskHandler(r.Client, r.Recorder)
	ld, err := ldHandler.GetLocalDisk(req.NamespacedName)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		log.WithError(err).Errorf("Failed to get localdisk")
		return reconcile.Result{}, err
	}

	if ld != nil {
		ldHandler.For(*ld.DeepCopy())
	} else {
		// Not found
		return reconcile.Result{}, nil
	}

	// NOTE: The control logic of localdisk should only respond to the update events of the disk itself.
	// For example, if the disk smart check fails, it should update its health status at this time.
	// As for the upper layer resources that depend on it, such as LDC, what it should do is to monitor
	// the event changes of LD and adjust the changed contents accordingly.
	// The connection between them should only be related to the state.

	// Update status
	if ldHandler.ClaimRef() != nil && ldHandler.UnClaimed() {
		ldHandler.SetupStatus(v1alpha1.LocalDiskClaimed)
		if err := ldHandler.UpdateStatus(); err != nil {
			r.Recorder.Eventf(&ldHandler.Ld, v1.EventTypeWarning, "UpdateStatusFail", "Update status fail, due to error: %v", err)
			log.WithError(err).Errorf("Update LocalDisk %v status fail", ldHandler.Ld.Name)
			return reconcile.Result{}, err
		}
	}

	// At this stage, we have no relevant inspection data, so we won't do any processing for the time being
	return reconcile.Result{}, nil
}
