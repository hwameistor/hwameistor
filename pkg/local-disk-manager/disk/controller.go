package disk

import (
	"context"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/common"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/localdisk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/lsblk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/smart"
	_ "github.com/hwameistor/hwameistor/pkg/local-disk-manager/udev"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	log "github.com/sirupsen/logrus"
	crmanager "sigs.k8s.io/controller-runtime/pkg/manager"
)

var diskParser = defaultDiskParser()

// Controller
type Controller struct {
	nodeName string

	// diskManager Represents how to discover and manage disks
	diskManager manager.Manager

	// diskQueue disk events queue
	diskQueue *common.TaskQueue

	// localDiskController
	localDiskController localdisk.Controller
}

// NewController
func NewController(mgr crmanager.Manager) *Controller {
	return &Controller{
		nodeName:            utils.GetNodeName(),
		diskManager:         manager.NewManager(),
		localDiskController: localdisk.NewController(mgr),
		diskQueue:           common.NewTaskQueue("DiskDiscoveryController", 10),
	}
}

// StartMonitor sets up and starts monitoring of local disks
func (ctr *Controller) StartMonitor() {
	// Wait cache synced
	ctr.localDiskController.Mgr.GetCache().WaitForCacheSync(context.TODO())

	// Start event handler
	go ctr.HandleEvent()

	// Check existing disks and handle any that are no longer exist
	existDisks := ctr.diskManager.ListExist()
	ctr.handleStaleDisks(existDisks)

	for _, disk := range existDisks {
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

// handleStaleDisks inactivates localdisks that were once recognized and registered in the system
// but no longer exist on the node
func (ctr *Controller) handleStaleDisks(existDisks []manager.Event) {
	existDiskAttrs := make([]manager.Attribute, 0, len(existDisks))
	for _, e := range existDisks {
		diskParser.For(*manager.NewDiskIdentifyWithName(e.DevPath, e.DevName))
		existDiskAttrs = append(existDiskAttrs, diskParser.AttributeParser.ParseDiskAttr())
	}
	// list all localDisks in current node
	lastLocalDisks, _ := ctr.localDiskController.ListLocalDisksByNode(utils.GetNodeName())
	// search stale localdisks
	for _, ld := range lastLocalDisks {
		if ld.Spec.State == v1alpha1.LocalDiskInactive {
			continue
		}

		exist := false
		// search by serial number
		if ld.Spec.DiskAttributes.SerialNumber != "" {
			for _, attr := range existDiskAttrs {
				if ld.Spec.DiskAttributes.SerialNumber == attr.Serial {
					log.WithField("serialNumber", attr.Serial).WithField("ldName", ld.Name).Info("Found existing disk serial number")
					exist = true
					break
				}
			}
		} else {
			devLinkSet := make(map[string]struct{})
			for _, devLink := range ld.Spec.DevLinks {
				devLinkSet[devLink] = struct{}{}
			}
			// search by dev link
			for _, attr := range existDiskAttrs {
				existDevLinkSet := make(map[string]struct{})
				for _, existDevLink := range attr.DevLinks {
					existDevLinkSet[existDevLink] = struct{}{}
				}
				for key, _ := range devLinkSet {
					if _, ok := existDevLinkSet[key]; ok {
						log.WithField("devLink", attr).WithField("ldName", ld.Name).Info("Found existing disk IDPath")
						exist = true
						break
					}
				}
				if exist {
					break
				}
			}
		}
		// can't find device in exist disks, it must be removed from this node
		if !exist {
			log.WithField("ldName", ld.Name).Info("Stale disk found, mark it inactive")
			_ = ctr.markLocalDiskInactive(ld)
		}
	}
}

// HandleEvent processes disk events in the work queue
func (ctr *Controller) HandleEvent() {
	for {
		event := ctr.Pop()
		if err := ctr.processSingleEvent(event); err != nil {
			log.WithError(err).WithFields(log.Fields{"DevName": event.DevName, "EventType": event.Type}).Error("Failed to process device udev event")
			ctr.PushRateLimited(event)
		} else {
			log.WithError(err).WithFields(log.Fields{"DevName": event.DevName, "EventType": event.Type}).Error("Succeed to process device udev event")
			ctr.Forget(event)
		}
		ctr.Done(event)
	}
}

// processSingleEvent processes a single event, applying different treatments based on the event type.
// Primarily, it handles two main scenarios: creating or updating LocalDisk resources;
// and for remove events, identifying and marking the corresponding LocalDisk as inactive.
func (ctr *Controller) processSingleEvent(event manager.Event) error {
	log.Infof("Receive disk event %+v", event)
	diskParser.For(*manager.NewDiskIdentifyWithName(event.DevPath, event.DevName))

	switch event.Type {
	case manager.EXIST, manager.CHANGE, manager.ADD:
		// Get disk basic info
		newDisk := diskParser.ParseDisk()
		log.Debugf("Disk %v basicinfo: %v", event.DevPath, newDisk)
		// Convert disk resource to localDisk
		localDisk := ctr.localDiskController.ConvertDiskToLocalDisk(newDisk)

		// Judge whether the disk is completely new
		if ctr.localDiskController.IsAlreadyExist(localDisk) {
			// If the disk already exists, try to update
			if err := ctr.localDiskController.UpdateLocalDiskAttr(localDisk); err != nil {
				log.WithError(err).Errorf("Update localDisk fail for disk %v", newDisk)
				return err
			}
			return nil
		}

		// Create disk resource
		if err := ctr.localDiskController.CreateLocalDisk(localDisk); err != nil {
			log.WithError(err).Errorf("Create localDisk fail for disk %v", newDisk)
			return err
		}

	case manager.REMOVE:
		log.WithField("devPath", event.DevName).Info("Detect disk removed")

		// for remove events, no serial can be found, so we need to find the disk by node device path
		localDisks, err := ctr.localDiskController.ListLocalDiskByNodeDevicePath(ctr.nodeName, event.DevName)
		if err != nil {
			log.WithError(err).Errorf("Failed to list LocalDisk by node device path %v/%v", ctr.nodeName, event.DevName)
			return err
		}

		// no local disk can be found by node device path, must be removed already
		if len(localDisks) == 0 {
			log.Infof("No LocalDisk found by node device path %v/%v", ctr.nodeName, event.DevName)
			return nil
		} else if len(localDisks) > 1 {
			log.Warningf("Multiple LocalDisk(%d) found by node device path %v/%v", len(localDisks), ctr.nodeName, event.DevName)
			return nil
		}
		localDisk := localDisks[0]

		// NOTES: currently we are not doing anything about the event that the disk goes offline, just mark it as inactive here
		if err := ctr.markLocalDiskInactive(localDisk); err != nil {
			return err
		}

	default:
		log.Infof("UNKNOWN event %v, skip it", event)
	}
	return nil
}

// markLocalDiskInactive marks the LocalDisk as inactive, updating its status and other attributes to reflect
// that it is no longer active or exist on the node
func (ctr *Controller) markLocalDiskInactive(localDisk v1alpha1.LocalDisk) error {
	// NOTES: currently we are not doing anything about the event that the disk goes offline, just mark it as inactive here
	localDisk.Spec.State = v1alpha1.LocalDiskInactive
	localDisk.Spec.PreDevicePath = localDisk.Spec.DevicePath
	localDisk.Spec.PreNodeName = localDisk.Spec.NodeName
	localDisk.Spec.DevicePath = ""
	localDisk.Spec.NodeName = ""
	localDisk.Spec.Major = ""
	localDisk.Spec.Minor = ""
	if err := ctr.localDiskController.UpdateLocalDiskAttr(localDisk); err != nil {
		log.WithError(err).Errorf("Failed to mark localDisk state %v to inactive", localDisk)
		return err
	}
	return nil
}

// defaultDiskParser initializes and returns a default disk parser for parsing disk information
func defaultDiskParser() *manager.DiskParser {
	diskBase := &manager.DiskIdentify{}
	return manager.NewDiskParser(
		diskBase,
		lsblk.NewPartitionParser(diskBase),
		&manager.RaidParser{},
		lsblk.NewAttributeParser(diskBase),
		smart.NewSMARTParser(diskBase))
}
