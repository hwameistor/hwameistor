package faultmanagement

import log "github.com/sirupsen/logrus"

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
	return nil
}
