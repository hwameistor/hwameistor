package evictor

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	EvictRecordStateInProgress = "InProgress"
	EvictRecordStateCompleted  = "Completed"
)

type evictNodeRecord struct {
	NodeName           string
	EvictVolumeRecords map[string]*evictVolumeRecord
	State              string
}

type evictVolumeRecord struct {
	VolumeName string
	SourceNode string
	State      string
}

type evictRecordManager struct {
	// key: nodeName
	evictNodeRecords map[string]*evictNodeRecord
	// key: volumeName
	evictVolumeRecords map[string]*evictVolumeRecord

	lock sync.Locker
}

func newEvictRecordManager() *evictRecordManager {
	return &evictRecordManager{
		evictNodeRecords:   map[string]*evictNodeRecord{},
		evictVolumeRecords: map[string]*evictVolumeRecord{},
		lock:               &sync.Mutex{},
	}
}

func (rm *evictRecordManager) run(stopCh <-chan struct{}) {
	go rm.cleanup(stopCh)

}
func (rm *evictRecordManager) isNodeEvictionCompleted(nodeName string) bool {
	record, exists := rm.evictNodeRecords[nodeName]
	return exists && record.State == EvictRecordStateCompleted
}

func (rm *evictRecordManager) hasNodeEvictionRecord(nodeName string) bool {
	_, exists := rm.evictNodeRecords[nodeName]
	return exists
}

func (rm *evictRecordManager) isVolumeEvictionCompleted(lvName string) bool {
	record, exists := rm.evictVolumeRecords[lvName]
	return exists && record.State == EvictRecordStateCompleted
}

func (rm *evictRecordManager) submitVolumeEviction(lvName string, srcNodeName string) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	if _, exists := rm.evictVolumeRecords[lvName]; exists {
		return
	}
	rm.evictVolumeRecords[lvName] = &evictVolumeRecord{VolumeName: lvName, SourceNode: srcNodeName, State: EvictRecordStateInProgress}
	nodeRecord, exists := rm.evictNodeRecords[srcNodeName]
	if !exists {
		nodeRecord = &evictNodeRecord{
			NodeName:           srcNodeName,
			EvictVolumeRecords: map[string]*evictVolumeRecord{},
			State:              EvictRecordStateInProgress,
		}
		rm.evictNodeRecords[srcNodeName] = nodeRecord
	}
	nodeRecord.EvictVolumeRecords[lvName] = rm.evictVolumeRecords[lvName]
	nodeRecord.State = EvictRecordStateInProgress

	rm._updateNodeRecordState(srcNodeName)

}

func (rm *evictRecordManager) completeVolumeEviction(lvName string) {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	record, exists := rm.evictVolumeRecords[lvName]
	if !exists {
		return
	}
	record.State = EvictRecordStateCompleted

	rm._updateNodeRecordState(record.SourceNode)
}

func (rm *evictRecordManager) _updateNodeRecordState(nodeName string) {
	if record, exists := rm.evictNodeRecords[nodeName]; exists {
		for _, vRec := range record.EvictVolumeRecords {
			if vRec.State != EvictRecordStateCompleted {
				record.State = EvictRecordStateInProgress
				return
			}
		}
		record.State = EvictRecordStateCompleted
	}
}

func (rm *evictRecordManager) cleanup(stopCh <-chan struct{}) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			log.Debug("Terminated cleanup processor")
			return
		case <-ticker.C:
			rm._cleanupRecords()
		}
	}
}

func (rm *evictRecordManager) _cleanupRecords() {
	rm.lock.Lock()
	defer rm.lock.Unlock()

	newEvictNodeRecords := map[string]*evictNodeRecord{}
	for nodeName := range rm.evictNodeRecords {
		if rm.evictNodeRecords[nodeName].State != EvictRecordStateCompleted {
			newEvictNodeRecords[nodeName] = rm.evictNodeRecords[nodeName]
			continue
		}
		for volName := range rm.evictNodeRecords[nodeName].EvictVolumeRecords {
			delete(rm.evictVolumeRecords, volName)
		}
	}

	rm.evictNodeRecords = newEvictNodeRecords
}
