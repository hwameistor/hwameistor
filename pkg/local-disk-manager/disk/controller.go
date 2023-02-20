package disk

import (
	"context"
	log "github.com/sirupsen/logrus"
	crmanager "sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/localdisk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/lsblk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/smart"
	_ "github.com/hwameistor/hwameistor/pkg/local-disk-manager/udev"
)

// Controller
type Controller struct {
	// diskManager Represents how to discover and manage disks
	diskManager manager.Manager

	// diskQueue disk events queue
	diskQueue chan manager.Event

	// localDiskController
	localDiskController localdisk.Controller
}

// NewController
func NewController(mgr crmanager.Manager) *Controller {
	return &Controller{
		diskManager:         manager.NewManager(),
		localDiskController: localdisk.NewController(mgr),
		diskQueue:           make(chan manager.Event),
	}
}

// StartMonitor
func (ctr *Controller) StartMonitor() {
	// Wait cache synced
	ctr.localDiskController.Mgr.GetCache().WaitForCacheSync(context.TODO())

	// Start event handler
	go ctr.HandleEvent()

	// Start list disk exist
	for _, disk := range ctr.diskManager.ListExist() {
		ctr.Push(disk)
	}

	// Start monitor disk event
	diskEventChan := make(chan manager.Event)
	go ctr.diskManager.Monitor(diskEventChan)

	// Start trigger disk event
	go ctr.diskManager.StartTimerTrigger(diskEventChan)

	// Start push disk event to controller event chan
	for disk := range diskEventChan {
		ctr.Push(disk)
	}
}

// HandleEvent
func (ctr *Controller) HandleEvent() {
	var DiskParser = defaultDiskParser()
	for {
		event := ctr.Pop()
		log.Infof("Receive disk event %+v", event)
		DiskParser.For(*manager.NewDiskIdentifyWithName(event.DevPath, event.DevName))

		switch event.Type {
		case manager.ADD:
			fallthrough
		case manager.EXIST:
			// Get disk basic info
			newDisk := DiskParser.ParseDisk()
			// log.Debugf("Disk %v basicinfo: %v", event.DevPath, newDisk)
			// Convert disk resource to localDisk
			localDisk := ctr.localDiskController.ConvertDiskToLocalDisk(newDisk)

			// Judge whether the disk is completely new
			if ctr.localDiskController.IsAlreadyExist(localDisk) {
				// log.Debugf("Disk %+v has been already exist", newDisk.DevName)
				// If the disk already exists, try to update
				if err := ctr.localDiskController.UpdateLocalDisk(localDisk); err != nil {
					log.WithError(err).Errorf("Update localDisk fail for disk %v", newDisk)
				}
				continue
			}

			// Create disk resource
			if err := ctr.localDiskController.CreateLocalDisk(localDisk); err != nil {
				log.WithError(err).Errorf("Create localDisk fail for disk %v", newDisk)
				continue
			}

		default:
			log.Infof("UNKNOWN event %v, skip it", event)
		}
	}
}

// defaultDiskParser
func defaultDiskParser() *manager.DiskParser {
	diskBase := &manager.DiskIdentify{}
	return manager.NewDiskParser(
		diskBase,
		lsblk.NewPartitionParser(diskBase),
		&manager.RaidParser{},
		lsblk.NewAttributeParser(diskBase),
		smart.NewSMARTParser(diskBase))
}
