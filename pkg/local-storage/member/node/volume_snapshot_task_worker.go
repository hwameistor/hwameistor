package node

import (
	"fmt"
	apisv1alpha "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	errReplicaNotFound         = fmt.Errorf("there is no replica found on the host")
	errReplicaSnapshotNotFound = fmt.Errorf("there is no replica snapshot found on the host")
)

func (m *manager) startVolumeSnapshotTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("VolumeSnapshot Assignment Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeSnapshotTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the Volume Snapshot Task Assignment worker")
				break
			}
			if err := m.processVolumeSnapshotTaskAssignment(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process assignment, retry later")
				m.volumeSnapshotTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed an assignment.")
				m.volumeSnapshotTaskQueue.Forget(task)
			}
			m.volumeSnapshotTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeSnapshotTaskQueue.Shutdown()
}

// processVolumeSnapshotTaskAssignment create or cleanup on-host volume replica snapshot according to the volume snapshot
func (m *manager) processVolumeSnapshotTaskAssignment(volumeSnapshotName string) error {
	logCtx := m.logger.WithField("volumeSnapshot", volumeSnapshotName)
	logCtx.Debug("Processing Volume Snapshot Task Assignment")

	volumeSnapshot := apisv1alpha.LocalVolumeSnapshot{}
	err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: volumeSnapshotName}, &volumeSnapshot)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get volume snapshot")
		return err
	}

	// return directly if the volume is being deleted
	// Tips: the volume replica snapshot normal process is under: pkg/local-storage/member/controller/volume_snapshot_task_worker.go
	if volumeSnapshot.Spec.Delete {
		return nil
	}

	// cleanup replica snapshot according to volume snapshot if the node is removed from accessibility
	if _, exist := utils.StrFind(volumeSnapshot.Spec.Accessibility.Nodes, m.name); !exist {
		return m.cleanupVolumeReplicaSnapshot(volumeSnapshot)
	}

	// create replica snapshot is not exist
	replicaSnapshotName, created := m.getOnHostVolumeReplicaSnapshotFromCache(volumeSnapshot.Name)
	if !created {
		return m.createVolumeReplicaSnapshot(volumeSnapshot)
	}

	logCtx.WithFields(log.Fields{"node": m.name, "replicaSnapshot": replicaSnapshotName}).Debug("VolumeReplicaSnapshot is already exist on the node")
	return nil
}

// createVolumeReplicaSnapshot create on-host volume replica snapshot
func (m *manager) createVolumeReplicaSnapshot(volumeSnapshot apisv1alpha.LocalVolumeSnapshot) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	logCtx := m.logger.WithField("volumeSnapshot", volumeSnapshot.Name)
	logCtx.Debug("Creating Volume Replica Snapshot")

	// find the volume replica
	replica, err := m.getOnHostVolumeReplica(volumeSnapshot.Spec.SourceVolume)
	if err != nil {
		logCtx.WithError(err).Error("Failed to get volume replica")
		return err
	}

	// fill in the volume replica snapshot according to the volume snapshot
	replicaSnapshot := apisv1alpha.LocalVolumeReplicaSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", volumeSnapshot.Name, utilrand.String(6)),
		},
		Spec: apisv1alpha.LocalVolumeReplicaSnapshotSpec{
			NodeName:              m.name,
			SourceVolumeReplica:   replica.Name,
			PoolName:              replica.Spec.PoolName,
			VolumeSnapshotName:    volumeSnapshot.Name,
			Delete:                volumeSnapshot.Spec.Delete,
			SourceVolume:          volumeSnapshot.Spec.SourceVolume,
			RequiredCapacityBytes: volumeSnapshot.Spec.RequiredCapacityBytes,
		},
	}

	logCtx = logCtx.WithField("replicaSnapshot", replicaSnapshot.Name)
	err = m.apiClient.Create(context.Background(), &replicaSnapshot)
	if err != nil {
		logCtx.WithError(err).Error("Failed to create volume replica snapshot")
		return err
	}

	m.replicaSnapshotsRecords[volumeSnapshot.Name] = replicaSnapshot.Name
	logCtx.Debug("Created volume replica snapshot")
	return nil
}

// cleanupVolumeReplicaSnapshot delete on-host volume replica snapshot
func (m *manager) cleanupVolumeReplicaSnapshot(volumeSnapshot apisv1alpha.LocalVolumeSnapshot) error {
	return nil
}

// getOnHostVolumeReplica returns the on-host volume replica according to the given volume
func (m *manager) getOnHostVolumeReplica(volumeName string) (apisv1alpha.LocalVolumeReplica, error) {
	volume := apisv1alpha.LocalVolume{}
	err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: volumeName}, &volume)
	if err != nil {
		return apisv1alpha.LocalVolumeReplica{}, err
	}

	for _, replicaName := range volume.Status.Replicas {
		replica := apisv1alpha.LocalVolumeReplica{}
		err = m.apiClient.Get(context.Background(), client.ObjectKey{Name: replicaName}, &replica)
		if err != nil {
			return apisv1alpha.LocalVolumeReplica{}, err
		}

		if replica.Spec.NodeName == m.name {
			return replica, nil
		}
	}

	return apisv1alpha.LocalVolumeReplica{}, errReplicaNotFound
}

// getOnHostVolumeReplicaSnapshot returns the on-host volume replica snapshot according to the given volume snapshot
func (m *manager) getOnHostVolumeReplicaSnapshot(volumeSnapshotName string) (apisv1alpha.LocalVolumeReplicaSnapshot, error) {
	volumeSnapshot := apisv1alpha.LocalVolumeSnapshot{}
	err := m.apiClient.Get(context.Background(), client.ObjectKey{Name: volumeSnapshotName}, &volumeSnapshot)
	if err != nil {
		return apisv1alpha.LocalVolumeReplicaSnapshot{}, err
	}

	for _, replicaSnapshotName := range volumeSnapshot.Status.ReplicaSnapshots {
		replicaSnapshot := apisv1alpha.LocalVolumeReplicaSnapshot{}
		err = m.apiClient.Get(context.Background(), client.ObjectKey{Name: replicaSnapshotName}, &replicaSnapshot)
		if err != nil {
			return apisv1alpha.LocalVolumeReplicaSnapshot{}, err
		}

		if replicaSnapshot.Spec.NodeName == m.name {
			return replicaSnapshot, nil
		}
	}

	return apisv1alpha.LocalVolumeReplicaSnapshot{}, errReplicaSnapshotNotFound
}

func (m *manager) getOnHostVolumeReplicaSnapshotFromCache(volumeSnapshotName string) (string, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()

	replicaSnapshot, ok := m.replicaSnapshotsRecords[volumeSnapshotName]
	return replicaSnapshot, ok
}
