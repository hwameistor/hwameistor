package udev

import (
	"fmt"
	"github.com/pilebones/go-udev/crawler"
	"github.com/pilebones/go-udev/netlink"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
)

// DiskManager monitor disk by udev
type DiskManager struct {
}

func NewDiskManager() DiskManager {
	return DiskManager{}
}

// ListExist
func (dm DiskManager) ListExist() []manager.Event {
	events, err := getExistDevice(GenRuleForBlock())
	if err != nil {
		log.WithError(err).Errorf("Failed processing existing devices")
		return nil
	}

	log.Info("Finished processing existing devices")
	return events
}

// Monitor
func (dm DiskManager) Monitor(c chan manager.Event) {
	// Monitor udev event in a loop
	for {
		if err := monitorDeviceEvent(c, GenRuleForBlock()); err != nil {
			log.WithError(err).Errorf("Monitor udev event fail, will try to monitor again")
			continue
		}
	}
}

// StartTimerTrigger trigger disk event
func (dm DiskManager) StartTimerTrigger(c chan manager.Event) {
	getNextTriggerTime := func(duration *time.Duration) time.Duration {
		if *duration > (time.Hour * 6) {
			*duration = time.Minute
		}
		*duration = *duration * 2
		return *duration
	}

	duration := time.Minute
	timer := time.NewTimer(duration)
	defer timer.Stop()

	for {
		<-timer.C
		for _, diskEvent := range dm.ListExist() {
			log.Debugf("Trigger disk: %v, time: %v", diskEvent.DevName, time.Now())
			c <- diskEvent
		}

		timer.Reset(getNextTriggerTime(&duration))
	}
}

// monitorDeviceEvent
func monitorDeviceEvent(c chan manager.Event, matchRule netlink.Matcher) error {
	conn := new(netlink.UEventConn)
	if err := conn.Connect(netlink.UdevEvent); err != nil {
		log.WithError(err).Errorf("Failed to connect to Netlink")
		return err
	}

	errChan := make(chan error)
	eventChan := make(chan netlink.UEvent)
	quit := conn.Monitor(eventChan, errChan, matchRule)

	for {
		select {
		case device, empty := <-eventChan:
			if !empty {
				return fmt.Errorf("EventChan has been closed when monitor udev event")
			}

			if !NewCDevice(crawler.Device{KObj: device.KObj, Env: device.Env}).FilterDisk() {
				log.Debugf("Device:%+v is drop", device)
				continue
			}
			log.Debugf("Device:%+v is keep", device)

			c <- manager.Event{
				Type:    string(device.Action),
				DevPath: addSysPrefix(device.KObj),
				DevType: device.Env["DEVTYPE"],
				DevName: device.Env["DEVNAME"],
			}

		case err := <-errChan:
			log.WithError(err).Errorf("Monitor udev event error")
			return err

		case <-quit:
			return fmt.Errorf("receive quit signal when monitor udev event")
		}
	}
}

// getExistDevice
func getExistDevice(matchRule netlink.Matcher) (events []manager.Event, err error) {
	deviceEvent := make(chan crawler.Device)
	errors := make(chan error)
	crawler.ExistingDevices(deviceEvent, errors, matchRule)

	for {
		select {
		case device, empty := <-deviceEvent:
			if !empty {
				return
			}

			// Filter non disk events
			if !NewCDevice(device).FilterDisk() {
				log.Debugf("Device:%+v is drop", device)
				continue
			}
			log.Debugf("Device:%+v is keep", device)

			events = append(events, manager.Event{
				Type:    manager.EXIST,
				DevPath: device.KObj,
				DevType: device.Env["DEVTYPE"],
				DevName: device.Env["DEVNAME"],
			})

		case err = <-errors:
			close(errors)
			return
		}
	}
}

// addSysPrefix
func addSysPrefix(path string) string {
	if strings.HasPrefix(path, "/sys/") {
		return path
	} else {
		return "/sys" + path
	}
}

func init() {
	manager.RegisterManager(DiskManager{})
}
