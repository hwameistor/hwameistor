package disk

import (
	"context"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crmanager "sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"

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
	// Start push disk event to controller event chan
	for disk := range diskEventChan {
		ctr.Push(disk)
	}
}

// HandleEvent
func (ctr *Controller) HandleEvent() {
	var p = defaultDiskParser()
	for {
		event := ctr.Pop()
		log.Infof("Receive disk event %+v", event)
		p.For(*manager.NewDiskIdentifyWithName(event.DevPath, event.DevName))

		switch event.Type {
		case manager.ADD:
			fallthrough
		case manager.EXIST, manager.CHANGE:
			// Get disk basic info
			newDisk := p.ParseDisk()
			log.Debugf("Disk %v basicinfo: %v", event.DevPath, newDisk)
			// Convert disk resource to localDisk
			localDisk := ctr.localDiskController.ConvertDiskToLocalDisk(newDisk)

			// Judge whether the disk is completely new
			if ctr.localDiskController.IsAlreadyExist(localDisk) {
				// If the disk already exists, try to update
				if err := ctr.localDiskController.UpdateLocalDiskAttr(localDisk); err != nil {
					log.WithError(err).Errorf("Update localDisk fail for disk %v", newDisk)
				}
				continue
			}

			// Create disk resource
			if err := ctr.localDiskController.CreateLocalDisk(localDisk); err != nil {
				log.WithError(err).Errorf("Create localDisk fail for disk %v", newDisk)
				continue
			}

		case manager.REMOVE:
			log.WithField("devPath", event.DevName).Info("Detect disk removed")

			localDiskName := fmt.Sprintf("%s-%s", utils.GetNodeName(), strings.TrimPrefix(event.DevName, "/dev/"))
			localDisk, err := ctr.localDiskController.GetLocalDisk(client.ObjectKey{Name: localDiskName})
			if err != nil {
				log.WithField("devPath", event.DevName).WithError(err).Error("Failed to get localdisk")
				continue
			}
			if localDisk.Name == "" {
				log.WithField("localdisk", localDiskName).Info("Ignore unmanaged disk")
				continue
			}

			// mark disk state inactive
			// NOTES: currently we are not doing anything about the event that the disk goes offline, just mark it here
			localDisk.Spec.State = v1alpha1.LocalDiskInactive
			if err = ctr.localDiskController.UpdateLocalDiskAttr(localDisk); err != nil {
				log.WithError(err).Errorf("Failed to mark localDisk state %v to inactive", localDisk)
			}
			continue

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
