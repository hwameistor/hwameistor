package thinpoolclaim

import (
	"context"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/utils"
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
	return &ReconcileThinPoolClaim{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("thinpoolclaim-controller"),
	}
}

// add a new Controller to mgr with r as reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("thinpoolclaim-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource thinpoolclaim
	err = c.Watch(&source.Kind{Type: &v1alpha1.ThinPoolClaim{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileThinPoolClaim{}

// ReconcileThinPoolClaim reconciles a ThinPoolClaim object
type ReconcileThinPoolClaim struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *ReconcileThinPoolClaim) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log.WithField("ThinPoolClaim", req.Name).Infof("Start Reconcile ThinPoolClaim")

	res := reconcile.Result{}
	tpc := &v1alpha1.ThinPoolClaim{}
	err := r.Client.Get(ctx, req.NamespacedName, tpc)
	if err != nil {
		return res, client.IgnoreNotFound(err)
	}

	switch tpc.Status.Status {
	case v1alpha1.ThinPoolClaimPhaseEmpty:
		err = r.processThinPoolClaimEmpty(ctx, tpc)
	case v1alpha1.ThinPoolClaimPhasePending:
		err = r.processThinPoolClaimPending(ctx, tpc)
	case v1alpha1.ThinPoolClaimPhaseToBeConsumed:
		err = r.processThinPoolClaimToBeConsumed(ctx, tpc)
	case v1alpha1.ThinPoolClaimPhaseConsumed:
		err = r.processThinPoolClaimConsumed(ctx, tpc)
	case v1alpha1.ThinPoolClaimPhaseToBeDeleted:
		err = r.processThinPoolClaimToBeDeleted(ctx, tpc)
	case v1alpha1.ThinPoolClaimPhaseDeleted:
		err = r.processThinPoolClaimDeleted(ctx, tpc)
	default:
		log.Warningf("ThinPoolClaim %s status %v is UNKNOWN", tpc.Name, tpc.Status.Status)
	}

	if err != nil {
		res.RequeueAfter = RequeueInterval
		if apierrors.IsConflict(err) {
			return res, nil
		}
		r.Recorder.Eventf(tpc, corev1.EventTypeWarning, v1alpha1.ThinPoolClaimEventFailed, err.Error())
		log.WithError(err).Errorf("Failed to reconcile ThinPoolClaim %v", tpc.Name)
	}

	return res, err
}

func (r *ReconcileThinPoolClaim) processThinPoolClaimEmpty(ctx context.Context, tpc *v1alpha1.ThinPoolClaim) error {
	log.WithField("name", tpc.Name).Info("Start processing Empty ThinPoolClaim")

	if tpc.Spec.Description.OverProvisionRatio != nil {
		overProvisionRatio, err := strconv.ParseFloat(*tpc.Spec.Description.OverProvisionRatio, 64)
		if err != nil {
			err = fmt.Errorf("fail to parse .spec.description.overProvisionRatio: %w", err)
			return err
		}

		if overProvisionRatio < 1.0 {
			err = fmt.Errorf(".spec.description.overProvisionRatio must be greater than or equal to 1.0")
			return err
		}
	}

	tpc.Status.Status = v1alpha1.ThinPoolClaimPhasePending
	return r.Client.Status().Update(ctx, tpc)
}

func (r *ReconcileThinPoolClaim) processThinPoolClaimPending(ctx context.Context, tpc *v1alpha1.ThinPoolClaim) error {
	log.WithField("name", tpc.Name).Info("Start processing Pending ThinPoolClaim")

	// check whether thin pool exists
	lsn := v1alpha1.LocalStorageNode{}
	err := r.Get(ctx, client.ObjectKey{Name: tpc.Spec.NodeName}, &lsn)
	if err != nil {
		return fmt.Errorf("failed to get LocalStorageNode %s: %w", tpc.Spec.NodeName, err)
	}
	pool, ok := lsn.Status.Pools[tpc.Spec.Description.PoolName]
	if !ok {
		return fmt.Errorf("storage pool %s in node %s not found", tpc.Spec.NodeName, tpc.Spec.Description.PoolName)
	}

	// check whether there are enough free capacities
	metadataSize := int64(1)
	if tpc.Spec.Description.PoolMetadataSize != nil {
		metadataSize = int64(*tpc.Spec.Description.PoolMetadataSize)
	}

	// for new thin pool
	// data size + metadata size + pmspare size
	requiredSize := (tpc.Spec.Description.Capacity + metadataSize*2) * utils.Gi

	// for existing thin pool
	if pool.ThinPool != nil {
		if pool.ThinPool.Size > tpc.Spec.Description.Capacity*utils.Gi {
			return fmt.Errorf("thin pool %s size %d is larger than requested size %d", pool.ThinPool.Name, pool.ThinPool.Size, tpc.Spec.Description.Capacity)
		}

		if pool.ThinPool.MetadataSize > metadataSize*utils.Gi {
			return fmt.Errorf("thin pool %s metadata size %d is larger than requested size %d", pool.ThinPool.Name, pool.ThinPool.MetadataSize, metadataSize)
		}

		metaDataExtendSize := metadataSize*utils.Gi - pool.ThinPool.MetadataSize
		dataPoolExtendSize := tpc.Spec.Description.Capacity*utils.Gi - pool.ThinPool.Size

		// data size + metadata size + pmspare size
		requiredSize = dataPoolExtendSize + metaDataExtendSize*2
	}

	if requiredSize > pool.FreeCapacityBytes {
		return fmt.Errorf("not enough free space on storage pool %s on node %s. Required %d bytes but only have %d bytes",
			tpc.Spec.Description.PoolName,
			tpc.Spec.NodeName,
			requiredSize,
			pool.FreeCapacityBytes)
	}

	tpc.Status.Status = v1alpha1.ThinPoolClaimPhaseToBeConsumed
	return r.Client.Status().Update(ctx, tpc)
}

// processThinPoolClaimToBeConsumed does nothing here
func (r *ReconcileThinPoolClaim) processThinPoolClaimToBeConsumed(_ context.Context, _ *v1alpha1.ThinPoolClaim) error {
	return nil
}

func (r *ReconcileThinPoolClaim) processThinPoolClaimConsumed(ctx context.Context, tpc *v1alpha1.ThinPoolClaim) error {
	log.WithField("name", tpc.Name).Info("Start processing Consumed ThinPoolClaim")

	tpc.Status.Status = v1alpha1.ThinPoolClaimPhaseToBeDeleted
	return r.Client.Status().Update(ctx, tpc)
}

func (r *ReconcileThinPoolClaim) processThinPoolClaimToBeDeleted(ctx context.Context, tpc *v1alpha1.ThinPoolClaim) error {
	log.WithField("name", tpc.Name).Info("Start processing ToBeDeleted ThinPoolClaim")

	tpc.Status.Status = v1alpha1.ThinPoolClaimPhaseDeleted
	return r.Client.Status().Update(ctx, tpc)
}

func (r *ReconcileThinPoolClaim) processThinPoolClaimDeleted(ctx context.Context, tpc *v1alpha1.ThinPoolClaim) error {
	log.WithField("name", tpc.Name).Info("Start processing Deleted ThinPoolClaim")

	// Delete this claim
	return r.Client.Delete(ctx, tpc)
}
