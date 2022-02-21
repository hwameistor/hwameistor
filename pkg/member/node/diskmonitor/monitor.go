package diskmonitor

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// EventAction type
type EventAction string

// consts
const (
	ActionAdd    EventAction = "add"
	ActionRemove EventAction = "remove"
	ActionChange EventAction = "change"
)

// DiskEvent struct
type DiskEvent struct {
	DevName   string
	Action    EventAction
	Content   string
	Major     string
	Minor     string
	Subsystem string
	SeqNum    int
}

// DiskMonitor interface
type DiskMonitor interface {
	Run(stopCh <-chan struct{})
}

type diskMonitor struct {
	logger *log.Entry
	events *EventQueue
	cmd    *exec.Cmd
}

// New a disk monitor instance
func New(queue *EventQueue) DiskMonitor {
	return &diskMonitor{
		logger: log.WithField("Module", "DiskMonitor"),
		events: queue,
	}
}

func (dm *diskMonitor) Run(stopCh <-chan struct{}) {

	dm.monitorUdevBlock(stopCh)

}

// monitors udev for block device changes
func (dm *diskMonitor) monitorUdevBlock(stopCh <-chan struct{}) {

	failCh := make(chan error)
	go dm.startUdevMonitor(failCh)

	for {
		select {
		case <-failCh:
			go dm.startUdevMonitor(failCh)
		case <-stopCh:
			dm.stopUdevMonitor()
			return
		}
	}
}

func (dm *diskMonitor) startUdevMonitor(failCh chan error) {
	dm.logger.Debug("Starting to monitor udev block device change events")

	// use kernel event, udev event will duplicate with kernel event,
	// and udev event may has filterd by udev rules in host before
	dm.cmd = exec.Command("stdbuf", "-oL", "udevadm", "monitor", "-p", "-k", "-s", "block")
	stdout, err := dm.cmd.StdoutPipe()
	if err != nil {
		dm.logger.WithError(err).Fatal("Cannot open udevadm stdout")
	}
	err = dm.cmd.Start()
	if err != nil {
		dm.logger.WithError(err).Fatal("Cannot start udevadm monitoring")
	}
	dm.logger.Debug("Starting to scan event")

	// fetch output of udevadm monitor in real time
	scanner := bufio.NewScanner(stdout)
	// event with properties is separated by an empty line
	segments := []string{}
	for scanner.Scan() {
		text := scanner.Text()
		dm.logger.Debugf("udevadm monitor: %s", text)
		// ganther all event's properties together
		if len(strings.TrimSpace(text)) != 0 {
			segments = append(segments, text)
			continue
		}
		if len(segments) == 0 {
			continue
		}

		// filter the udev event
		event, err := dm.filterUdevEvent(segments)
		// reset the segments
		segments = segments[:0]
		if err != nil {
			dm.logger.WithField("event", text).WithError(err).Warning("Failed to filter udevadm event")
			continue
		}
		if dm.events != nil && event != nil {
			dm.events.Add(event)
		}
	}
	if err := scanner.Err(); err != nil {
		dm.logger.WithError(err).Error("Error happened on udevadm monitor scanner")
		dm.stopUdevMonitor()
		failCh <- err
	}
}

func (dm *diskMonitor) stopUdevMonitor() {
	dm.logger.Debug("Stopping udevadm monitor")
	if dm.cmd.ProcessState != nil && !dm.cmd.ProcessState.Exited() && dm.cmd.Process != nil {
		if err := dm.cmd.Process.Kill(); err != nil {
			dm.logger.WithError(err).Error("Failed to stop udevadm monitor")
		} else {
			dm.logger.Debug("Stopped udevadm monitor")
		}
	}
}

// return nil, nil if it's non-disk event
func (dm *diskMonitor) filterUdevEvent(segments []string) (*DiskEvent, error) {
	event := &DiskEvent{}
	for _, segment := range segments {
		parts := strings.SplitN(segment, "=", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "ACTION":
			{
				event.Action = EventAction(parts[1])
			}
		case "DEVNAME":
			{
				event.DevName = parts[1]
			}
		case "DEVTYPE":
			{
				if parts[1] != "disk" {
					return nil, nil
				}
			}
		case "MAJOR":
			{
				event.Major = parts[1]
			}
		case "MINOR":
			{
				event.Minor = parts[1]
			}
		case "SUBSYSTEM":
			{
				event.Subsystem = parts[1]
			}
		case "SEQNUM":
			{
				seqNum, err := strconv.Atoi(parts[1])
				if err != nil {
					dm.logger.Warnf("parse SEQNUM %s to int failed", parts[1])
					continue
				}
				event.SeqNum = seqNum
			}
		}
	}
	event.Content = strings.Join(segments, "\n")
	return event, nil
}
