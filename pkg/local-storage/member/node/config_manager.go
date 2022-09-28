package node

import (
	"context"
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/common"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/configer"
)

type configManager struct {
	// this node hostname
	hostname               string
	systemConfig           apisv1alpha1.SystemConfig
	apiClient              client.Client
	configer               configer.Configer
	logger                 *log.Entry
	syncReplicaStatusQueue *common.TaskQueue
}

func NewConfigManager(hostname string, config apisv1alpha1.SystemConfig, apiClient client.Client) (*configManager, error) {

	m := &configManager{
		hostname:               hostname,
		systemConfig:           config,
		apiClient:              apiClient,
		logger:                 log.WithField("Module", "NodeConfigManager"),
		syncReplicaStatusQueue: common.NewTaskQueue("syncReplicaStatusQueue", 0),
	}
	configer, err := configer.ConfigerFactory(hostname, config, apiClient, m.syncReplicaStatus)
	if err != nil {
		return nil, err
	}
	m.configer = configer
	return m, nil
}

func (m *configManager) Run(stopCh <-chan struct{}) error {

	go m.startReplicaStatusSyncWorker(stopCh)

	m.configer.Run(stopCh)

	m.ConsistencyCheck()

	return nil
}

func (m *configManager) getCurrentNodeReplicas() (replicas []apisv1alpha1.LocalVolumeReplica, err error) {
	var replicaList apisv1alpha1.LocalVolumeReplicaList
	if err = m.apiClient.List(context.TODO(), &replicaList); err != nil {
		return nil, err
	}

	for _, replica := range replicaList.Items {
		if replica.Spec.NodeName == m.hostname {
			replicas = append(replicas, replica)
		}
	}

	return replicas, nil
}

func (m *configManager) ConsistencyCheck() {
	m.logger.Info("do replica config ConsistencyCheck")
	currentNodeReplicas, err := m.getCurrentNodeReplicas()
	if err != nil {
		m.logger.Errorf("list current node replica err: %s", err)
		return
	}

	m.configer.ConsistencyCheck(currentNodeReplicas)

	m.logger.Info("ConsistencyCheck completed")
}

func (m *configManager) TestVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error {
	_, isHA, err := m.getConfig(replica)
	if err != nil {
		return err
	}
	if !isHA {
		return nil
	}

	haState, err := m.configer.GetReplicaHAState(replica)
	if err != nil {
		return fmt.Errorf("get configed replica status err: %s", err)
	}
	replica.Status.HAState = &haState
	replica.Status.State = m.genReplicaStateFromHAState(haState)
	return nil
}

// EnsureConfig should make sure the device is available at replica.Status.DevicePath, may be symbolic
func (m *configManager) EnsureConfig(replica *apisv1alpha1.LocalVolumeReplica) error {

	config, isHA, err := m.getConfig(replica)
	if err != nil {
		return fmt.Errorf("get replica config err: %s", err)
	}
	if !isHA {
		// no config will be applied to non-HA volume's replica
		return m.ensureConfigForNonHA(replica, config)
	}

	return m.ensureConfigForHA(replica, config)
}

func (m *configManager) ensureConfigForNonHA(replica *apisv1alpha1.LocalVolumeReplica, config *apisv1alpha1.VolumeConfig) error {
	logCtx := m.logger.WithFields(log.Fields{"replica": replica.Name})
	logCtx.Debug("Ensuring config for non-HA volume replica")

	replica.Status.DevicePath = replica.Status.StoragePath
	return nil
}

func (m *configManager) ensureConfigForHA(replica *apisv1alpha1.LocalVolumeReplica, config *apisv1alpha1.VolumeConfig) error {
	logCtx := m.logger.WithFields(log.Fields{"replica": replica.Name})
	logCtx.Debug("Ensuring config for HA volume replica")

	if replica.Status.DevicePath == replica.Status.StoragePath {
		replica.Status.DevicePath = m.genDevicePath(replica)
	}

	if err := m.ensureDirectory(replica.Status.DevicePath); err != nil {
		logCtx.WithField("device", replica.Status.DevicePath).WithError(err).Error("Failed to create HA device directory")
		return fmt.Errorf("ensure device directory err: %s", err)
	}

	if err := m.configer.ApplyConfig(replica, *config); err != nil {
		logCtx.WithError(err).Error("Failed to apply the config")
		return err
	}

	// do volume initialize
	if !config.Initialized && config.ReadyToInitialize && m.isThisNodePrimary(*config) {
		if err := m.configer.Initialize(replica, *config); err != nil {
			return fmt.Errorf("initialize volume err: %s", err)
		}

		config.Initialized = true
		return m.updateConfig(replica, config)
	}

	return nil
}

// DeleteConfig configer should make sure the device is deleted.
func (m *configManager) DeleteConfig(replica *apisv1alpha1.LocalVolumeReplica) error {
	return m.configer.DeleteConfig(replica)
}

func (m *configManager) ensureDirectory(filepath string) error {
	dir := path.Dir(filepath)
	if _, err := os.Stat(dir); err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func (m *configManager) genDevicePath(replica *apisv1alpha1.LocalVolumeReplica) string {
	return fmt.Sprintf("/dev/%s-HA/%s", replica.Spec.PoolName, replica.Spec.VolumeName)
}

// return: config, isHA, error
func (m *configManager) getConfig(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.VolumeConfig, bool, error) {
	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: replica.Spec.VolumeName}, vol); err != nil {
		return nil, false, err
	}
	if vol.Spec.Config == nil {
		return nil, false, fmt.Errorf("not found")
	}
	//return vol.Spec.Config, len(vol.Spec.Config.Replicas) > 1 || vol.Spec.Config.Convertible, nil
	return vol.Spec.Config, vol.Spec.Config.Convertible, nil
}

func (m *configManager) updateConfig(replica *apisv1alpha1.LocalVolumeReplica, config *apisv1alpha1.VolumeConfig) error {
	m.logger.WithField("Replica", replica.Name).Debug("update replica volume, set Initialized=true")
	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: replica.Spec.VolumeName}, vol); err != nil {
		return err
	}
	oldVol := vol.DeepCopy()
	patch := client.MergeFrom(oldVol)
	vol.Spec.Config = config
	return m.apiClient.Patch(context.TODO(), vol, patch)
}

func (m *configManager) isThisNodePrimary(config apisv1alpha1.VolumeConfig) bool {
	for _, peer := range config.Replicas {
		if peer.Hostname == m.hostname && peer.Primary {
			return true
		}
	}
	return false
}

func (m *configManager) genReplicaStateFromHAState(haState apisv1alpha1.HAState) apisv1alpha1.State {
	if haState.State == apisv1alpha1.HAVolumeReplicaStateConsistent {
		return apisv1alpha1.VolumeReplicaStateReady
	}
	return apisv1alpha1.VolumeReplicaStateNotReady
}

// configer will sync replica status by this func
func (m *configManager) syncReplicaStatus(replicaName string) {
	m.syncReplicaStatusQueue.Add(replicaName)
}

func (m *configManager) startReplicaStatusSyncWorker(stopCh <-chan struct{}) {
	go func() {
		for {
			task, shutdown := m.syncReplicaStatusQueue.Get()
			if shutdown {
				m.logger.Info("replica sync worker shutdown")
				return
			}

			if err := m.processReplicaStatusUpdate(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process VolumeReplica status update, retry later")
				m.syncReplicaStatusQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a VolumeReplica status update.")
				m.syncReplicaStatusQueue.Forget(task)
			}
			m.syncReplicaStatusQueue.Done(task)
		}
	}()

	<-stopCh
	m.syncReplicaStatusQueue.Shutdown()

}

func (m *configManager) processReplicaStatusUpdate(replicaName string) error {
	var replica apisv1alpha1.LocalVolumeReplica
	if err := m.apiClient.Get(context.TODO(), client.ObjectKey{Name: replicaName}, &replica); err != nil {
		if errors.IsNotFound(err) {
			m.logger.Debugf("ignore replica status update, replica %s not found", replicaName)
			return nil
		}
		return err
	}

	haState, err := m.configer.GetReplicaHAState(&replica)
	if err != nil {
		return err
	}

	if replica.Status.HAState != nil && *replica.Status.HAState == haState {
		return nil
	}

	newReplica := replica.DeepCopy()
	newReplica.Status.HAState = &haState
	newReplica.Status.State = m.genReplicaStateFromHAState(haState)
	patch := client.MergeFrom(&replica)
	if err := m.apiClient.Status().Patch(context.TODO(), newReplica, patch); err != nil {
		return fmt.Errorf("update replica %s status err: %s", replica.Name, err)
	}

	return nil
}
