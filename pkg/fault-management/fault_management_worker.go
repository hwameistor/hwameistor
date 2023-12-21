package faultmanagement

import (
	"context"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	faultTickerProtectionFinalizer = "hwameistor.io/faultticket-protection"
)

func (m *manager) startFaultTicketTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("FaultTicket Worker is working now")
	go func() {
		for {
			task, shutdown := m.faultTicketTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the FaultTicket worker")
				break
			}
			if err := m.processFaultTicket(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process FaultTicket task, retry later")
				m.faultTicketTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a FaultTicket task.")
				m.faultTicketTaskQueue.Forget(task)
			}
			m.faultTicketTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.faultTicketTaskQueue.Shutdown()
}

func (m *manager) processFaultTicket(faultTicketName string) error {
	logger := m.logger.WithField("faultTicketName", faultTicketName)

	faultTicket, err := m.hmClient.HwameistorV1alpha1().FaultTickets().Get(context.Background(), faultTicketName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("faultTicket not found from cache, may be deleted already")
			return nil
		}
		logger.WithError(err).Error("failed to get faultTicket")
		return err
	}

	// stop everything
	if faultTicket.Spec.Suspend {
		logger.Debug("faultTicket is suspended, ignore this fault")
		return nil
	}

	// proceed with faultticket deletion and remove finalizers when needed
	if faultTicket.ObjectMeta.DeletionTimestamp != nil {
		return m.processFaultTicketDeletion(faultTicket)
	}

	// proceed with faultticket according to status.phase
	switch faultTicket.Status.Phase {
	case v1alpha1.TicketPhaseEmpty:
		err = m.processFaultTicketEmpty(faultTicket)

	// evaluate first and find out resources effected by the fault
	case v1alpha1.TicketPhaseEvaluating:
		err = m.processFaultTicketEvaluating(faultTicket)

	// recovery the fault if necessary
	case v1alpha1.TicketPhaseRecovering:
		err = m.processFaultTicketRecovering(faultTicket)

	// recovery action is done
	case v1alpha1.TicketPhaseCompleted:
		err = m.processFaultTicketCompleted(faultTicket)

	default:
		err = fmt.Errorf("unexpected phase for faultTicket")
	}

	if err != nil {
		m.logger.WithError(err).Error("failed to process faultTicker")
	}
	return err
}

func (m *manager) processFaultTicketCompleted(faultTicket *v1alpha1.FaultTicket) error {
	return nil
}

// processFaultTicketEmpty start the evaluation of the fault
func (m *manager) processFaultTicketEmpty(faultTicket *v1alpha1.FaultTicket) error {
	faultTicket.Status.Phase = v1alpha1.TicketPhaseEvaluating
	_, err := m.hmClient.HwameistorV1alpha1().FaultTickets().UpdateStatus(context.Background(), faultTicket, metav1.UpdateOptions{})
	if err != nil {
		m.logger.WithField("faultTicket", faultTicket.Name).Errorf("failed to update ticket status to %s", v1alpha1.TicketPhaseEvaluating)
	}
	return err
}

// processFaultTicketDeletion processes finalizers and stop the actions backend the ticket(e.g., recovery, analyze).
// It has the following steps:
// 1. List all actions backend this ticket
// 2. Stop the action obtained from step 1
// 3. Remove the finalizers from the ticker when all actions are done
func (m *manager) processFaultTicketDeletion(faultTicket *v1alpha1.FaultTicket) error {
	logger := m.logger.WithField("faultTicket", faultTicket.Name)
	logger.Debug("Starting processFaultTicketDeletion")
	if !checkFinalizer(faultTicket, faultTickerProtectionFinalizer) {
		return nil
	}

	// remove the finalizers from the ticker directly(do nothing now)
	removedFinalizers, _ := removeFinalizer(faultTicket.ObjectMeta.Finalizers, faultTickerProtectionFinalizer)
	faultTicket.SetFinalizers(removedFinalizers)

	_, err := m.hmClient.HwameistorV1alpha1().FaultTickets().Update(context.Background(), faultTicket, metav1.UpdateOptions{})
	if err != nil {
		m.logger.WithError(err).Error("failed to remove faultTickerProtectionFinalizer")
	}

	return err
}

func checkFinalizer(obj metav1.Object, finalizer string) bool {
	for _, f := range obj.GetFinalizers() {
		if f == finalizer {
			return true
		}
	}
	return false
}

func removeFinalizer(finalizers []string, finalizerToRemove string) ([]string, bool) {
	modified := false
	modifiedFinalizers := make([]string, 0)
	for _, finalizer := range finalizers {
		if finalizer != finalizerToRemove {
			modifiedFinalizers = append(modifiedFinalizers, finalizer)
		}
	}
	if len(modifiedFinalizers) == 0 {
		modifiedFinalizers = nil
	}
	if len(modifiedFinalizers) != len(finalizers) {
		modified = true
	}
	return modifiedFinalizers, modified
}

func addFinalizer(finalizers []string, finalizerToAdd string) ([]string, bool) {
	modifiedFinalizers := make([]string, 0)
	for _, finalizer := range finalizers {
		if finalizer == finalizerToAdd {
			// finalizer already exists
			return finalizers, false
		}
	}
	modifiedFinalizers = append(modifiedFinalizers, finalizers...)
	modifiedFinalizers = append(modifiedFinalizers, finalizerToAdd)
	return modifiedFinalizers, true
}
