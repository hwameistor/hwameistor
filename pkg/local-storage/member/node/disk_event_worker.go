package node

import (
	log "github.com/sirupsen/logrus"

	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/diskmonitor"
)

func (m *manager) startDiskEventWorker(stopCh <-chan struct{}) {
	m.logger.Debug("Disk Event Worker is working now")

	go func() {
		for {
			event, shutdown := m.diskEventQueue.Get()
			if shutdown {
				m.logger.Debug("Stop the disk event worker")
				break
			}
			if err := m.processDiskEvent(event); err != nil {
				m.logger.WithFields(log.Fields{"event": event, "error": err.Error()}).Error("Failed to process disk event task, retry later")
				m.diskEventQueue.AddRateLimited(event)
			} else {
				m.logger.WithFields(log.Fields{"event": event}).Debug("Completed a disk event.")
				m.diskEventQueue.Forget(event)
			}
			m.diskEventQueue.Done(event)
		}
	}()

	<-stopCh
	m.diskEventQueue.Shutdown()
}

func (m *manager) processDiskEvent(event *diskmonitor.DiskEvent) error {
	logCtx := m.logger.WithFields(log.Fields{"event": event})
	logCtx.Debug("No further work on the disk event currently")
	return nil
}
