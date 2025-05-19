package node

import (
	"context"

	v1 "k8s.io/api/core/v1"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

func (m *manager) startThinPoolClaimTaskWorker(stopCh <-chan struct{}) {

	m.logger.Debug("ThinPoolClaim Worker is working now")
	go func() {
		for {
			task, shutdown := m.thinPoolClaimTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the ThinPoolClaim worker")
				break
			}
			if err := m.processThinPoolClaim(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process ThinPoolClaim task, retry later")
				m.thinPoolClaimTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a ThinPoolClaim task.")
				m.thinPoolClaimTaskQueue.Forget(task)
			}
			m.thinPoolClaimTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.thinPoolClaimTaskQueue.Shutdown()
}

func (m *manager) processThinPoolClaim(thinPoolClaimName string) error {
	logCtx := m.logger.WithField("ThinPoolClaim", thinPoolClaimName)
	logCtx.Debug("start processing ThinPoolClaim")

	tpc := &apisv1alpha1.ThinPoolClaim{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: thinPoolClaimName}, tpc); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		return nil
	}

	var err error
	switch tpc.Status.Status {
	case apisv1alpha1.ThinPoolClaimPhaseToBeConsumed:
		err = m.processThinPoolClaimToBeConsumed(tpc)
	default:
		logCtx.Error("Invalid ThinPoolClaim state")
	}

	return err
}

func (m *manager) processThinPoolClaimToBeConsumed(tpc *apisv1alpha1.ThinPoolClaim) error {
	logCtx := m.logger.WithField("ThinPoolClaim", tpc.GetName())
	logCtx.Debug("start processing ToBeConsumed ThinPoolClaim")

	// 1. create or extend thin pool
	if err := m.storageMgr.PoolManager().ExtendThinPool(tpc); err != nil {
		logCtx.WithError(err).Error("Failed to ExtendThinPool")
		m.recorder.Eventf(tpc, v1.EventTypeWarning, apisv1alpha1.ThinPoolClaimEventFailed,
			"Failed to ExtendThinPool, due to error: %v", err)
		return err
	}

	// 2. update thin pool extend record
	if err := m.storageMgr.Registry().UpdateThinPoolExtendRecord(tpc); err != nil {
		logCtx.WithError(err).Error("Failed to UpdatePoolExtendRecord")
		m.recorder.Eventf(tpc, v1.EventTypeWarning, apisv1alpha1.ThinPoolClaimEventFailed,
			"Failed to UpdatePoolExtendRecord, due to error: %v", err)
		return err
	}

	// 3. rebuild Node resource
	if err := m.storageMgr.Registry().SyncNodeResources(); err != nil {
		logCtx.WithError(err).Error("Failed to SyncNodeResources for updating thin pool info")
		m.recorder.Eventf(tpc, v1.EventTypeWarning, apisv1alpha1.ThinPoolClaimEventFailed,
			"Failed to SyncNodeResources for updating thin pool info, due to error: %v", err)
		return err
	}

	m.recorder.Eventf(tpc, v1.EventTypeNormal, apisv1alpha1.ThinPoolClaimEventSucceed,
		"Consumed ThinPoolClaim %v succeed, ThinPoolClaim will be deleted later", tpc.Name)

	// 4. update thin pool claim status
	if err := m.updateThinPoolClaimConsumed(tpc); err != nil {
		logCtx.WithError(err).Error("Failed to update ThinPoolClaim status")
		return err
	}
	return nil
}

func (m *manager) updateThinPoolClaimConsumed(tpc *apisv1alpha1.ThinPoolClaim) error {
	tpc.Status.Status = apisv1alpha1.ThinPoolClaimPhaseConsumed
	return m.apiClient.Status().Update(context.Background(), tpc)
}
