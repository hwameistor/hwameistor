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
	events, err := getExistDeviceEvents(GenRuleForBlock())
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
	setNextTriggerDuration := func(duration *time.Duration) time.Duration {
		if *duration >= (time.Minute * 8) {
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
		log.Debugf("Trigger disk event")
		for _, diskEvent := range dm.ListExist() {
			c <- diskEvent
		}

		// Reset trigger time
		setNextTriggerDuration(&duration)
		timer.Reset(duration)
		log.Debugf("Next disk event wil be triggered at: %v", time.Now().Add(duration))
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
		case deviceEvt, empty := <-eventChan:
			if !empty {
				return fmt.Errorf("EventChan has been closed when monitor udev event")
			}

			event := manager.Event{
				Type:    string(deviceEvt.Action),
				DevPath: addSysPrefix(deviceEvt.KObj),
				DevType: deviceEvt.Env["DEVTYPE"],
				DevName: deviceEvt.Env["DEVNAME"],
			}

			switch string(deviceEvt.Action) {
			case manager.REMOVE:
				// push chan directly
				c <- event
			default:
				device := NewDevice(deviceEvt.KObj)
				if err := device.ParseDeviceInfo(); err != nil {
					log.WithError(err).WithField("Device", deviceEvt.KObj).Error("Failed to parse device, drop it")
					continue
				}

				if !device.FilterDisk() {
					log.Debugf("Device:%+v is drop", deviceEvt)
					continue
				}
				log.Debugf("Device:%+v is keep", deviceEvt)
				c <- event
			}

		case err := <-errChan:
			log.WithError(err).Errorf("Monitor udev event error")
			return err

		case <-quit:
			return fmt.Errorf("receive quit signal when monitor udev event")
		}
	}
}

// getExistDeviceEvents
func getExistDeviceEvents(matchRule netlink.Matcher) (events []manager.Event, err error) {
	deviceEvent := make(chan crawler.Device)
	errors := make(chan error)
	crawler.ExistingDevices(deviceEvent, errors, matchRule)

	for {
		select {
		case deviceEvt, empty := <-deviceEvent:
			if !empty {
				return
			}

			// Filter non disk events
			device := NewDevice(deviceEvt.KObj)
			if err = device.ParseDeviceInfo(); err != nil {
				log.WithError(err).WithField("Device", deviceEvt.KObj).Error("Failed to parse device, drop it")
				continue
			}

			if !device.FilterDisk() {
				log.Debugf("Device:%+v is drop", deviceEvt)
				continue
			}
			log.Debugf("Device:%+v is keep", deviceEvt)

			events = append(events, manager.Event{
				Type:    manager.EXIST,
				DevPath: deviceEvt.KObj,
				DevType: deviceEvt.Env["DEVTYPE"],
				DevName: deviceEvt.Env["DEVNAME"],
			})

		case err = <-errors:
			close(errors)
			return
		}
	}
}

// ListAllBlockDevices list all block devices by udev
func ListAllBlockDevices() ([]manager.Attribute, error) {
	return getExistDevice(GenRuleForBlock())
}

// getExistDevice
func getExistDevice(matchRule netlink.Matcher) (devices []manager.Attribute, err error) {
	deviceEvent := make(chan crawler.Device)
	errors := make(chan error)
	crawler.ExistingDevices(deviceEvent, errors, matchRule)

	for {
		select {
		case deviceEvt, empty := <-deviceEvent:
			if !empty {
				return
			}

			// Filter non disk events
			device := NewDevice(deviceEvt.KObj)
			if err = device.ParseDeviceInfo(); err != nil {
				log.WithError(err).WithField("Device", deviceEvt.KObj).Error("failed to parse device, drop it")
				continue
			}
			if !device.FilterDisk() {
				log.Debugf("Device:%+v is drop", deviceEvt)
				continue
			}
			log.Debugf("Device:%+v is keep", deviceEvt)

			devices = append(devices, manager.Attribute{
				DevPath:  device.DevPath,
				DevType:  device.DevType,
				DevName:  device.DevName,
				Serial:   device.Serial,
				DevLinks: device.DevLinks,
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
