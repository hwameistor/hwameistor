package localdisknode

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisknode"
	"k8s.io/apimachinery/pkg/api/errors"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new LocalDiskNode Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileLocalDiskNode{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("localdisknode-controller"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("localdisknode-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource LocalDiskNode
	err = c.Watch(&source.Kind{Type: &v1alpha1.LocalDiskNode{}}, &handler.EnqueueRequestForObject{}, withCurrentNode())
	if err != nil {
		return err
	}

	localDiskToLocalDiskNodeRequestFunc := handler.EnqueueRequestsFromMapFunc(
		func(a client.Object) []reconcile.Request {
			ld, ok := a.(*v1alpha1.LocalDisk)
			if !ok || ld.Spec.NodeName != utils.GetNodeName() {
				return []reconcile.Request{}
			}

			return []reconcile.Request{
				reconcile.Request{
					NamespacedName: types.NamespacedName{Name: ld.Spec.NodeName},
				},
			}
		})

	// Watch for changes for resource LocalDisk on this node
	err = c.Watch(&source.Kind{Type: &v1alpha1.LocalDisk{}}, localDiskToLocalDiskNodeRequestFunc)
	if err != nil {
		return err
	}

	return nil
}

// withCurrentNode filter volume request for this node
func withCurrentNode() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			node, _ := event.Object.DeepCopyObject().(*v1alpha1.LocalDiskNode)
			return node.Spec.AttachNode == utils.GetNodeName()
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			node, _ := deleteEvent.Object.DeepCopyObject().(*v1alpha1.LocalDiskNode)
			return node.Spec.AttachNode == utils.GetNodeName()
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			node, _ := updateEvent.ObjectNew.DeepCopyObject().(*v1alpha1.LocalDiskNode)
			return node.Spec.AttachNode == utils.GetNodeName() &&
				updateEvent.ObjectNew.GetGeneration() != updateEvent.ObjectOld.GetGeneration()
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			node, _ := genericEvent.Object.DeepCopyObject().(*v1alpha1.LocalDiskNode)
			return node.Spec.AttachNode == utils.GetNodeName()
		},
	}
}

// blank assignment to verify that ReconcileLocalDiskNode implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileLocalDiskNode{}

// ReconcileLocalDiskNode reconciles a LocalDiskNode object
type ReconcileLocalDiskNode struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a LocalDiskNode object and makes changes based on the state read
func (r *ReconcileLocalDiskNode) Reconcile(_ context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.WithField("LocalDiskNode", request.NamespacedName).Info("Reconciling LocalDiskNode")
	ldnHandler := localdisknode.NewDiskNodeHelper(r.client, r.Recorder)
	err := ldnHandler.For(request.NamespacedName)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		log.WithError(err).Errorf("failed to get LocalDiskNode %s", request.NamespacedName)
		return reconcile.Result{}, err
	}

	// The main task here is to update the resource status of
	// the localDisk of this node to the latest status.
	newDisks, err := ldnHandler.ListNodeDisks()
	if err != nil {
		log.WithError(err).Errorf("failed to list disks on node %s", request.Name)
		return reconcile.Result{}, err
	}

	// find out new disks
	needUpdateDisks := map[string]v1alpha1.Disk{}
	for name, newDisk := range newDisks {
		if !ldnHandler.IsSameDisk(name, newDisk) {
			needUpdateDisks[name] = *newDisk.DeepCopy()
		}
	}

	// find out disk which is removed already
	needRemoveDisks := map[string]v1alpha1.Disk{}
	for name, disk := range ldnHandler.Disks() {
		if _, exist := newDisks[name]; !exist {
			needRemoveDisks[name] = *disk.DeepCopy()
		}
	}

	// update or remove disk
	ldnHandler.UpdateDiskLists(needUpdateDisks, needRemoveDisks)
	ldnHandler.UpdateDiskStats()

	return reconcile.Result{}, ldnHandler.UpdateStatus()
}
