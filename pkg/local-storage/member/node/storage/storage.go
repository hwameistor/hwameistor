package storage

import (
	"context"
	"fmt"
	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/member/node/encrypt"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type localVolumeReplicaManager struct {
	cmdExec         LocalVolumeExecutor
	volumeValidator *validator
	registry        LocalRegistry
	logger          *log.Entry
	lm              *LocalManager
}

func newLocalVolumeReplicaManager(lm *LocalManager) LocalVolumeReplicaManager {
	return &localVolumeReplicaManager{
		cmdExec:         newLVMExecutor(lm),
		volumeValidator: newValidator(),
		registry:        lm.Registry(),
		lm:              lm,
		logger:          log.WithField("Module", "NodeManager/LocalVolumeReplicaManager"),
	}
}

func (mgr *localVolumeReplicaManager) CreateVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error) {
	mgr.logger.Debugf("Creating VolumeReplica. name:%s, pool:%s, size:%d", replica.Spec.VolumeName, replica.Spec.PoolName, replica.Spec.RequiredCapacityBytes)
	if err := mgr.volumeValidator.canCreateVolumeReplica(replica, mgr.registry); err != nil {
		if err == ErrorReplicaExists {
			mgr.logger.Infof("Replica %s has already exists.", replica.Spec.VolumeName)
			newReplica := replica.DeepCopy()
			currentReplica := mgr.registry.VolumeReplicas()[replica.Spec.VolumeName]
			newReplica.Status.AllocatedCapacityBytes = currentReplica.Status.AllocatedCapacityBytes
			newReplica.Status.StoragePath = currentReplica.Status.StoragePath
			newReplica.Status.DevicePath = currentReplica.Status.DevicePath
			newReplica.Status.Synced = currentReplica.Status.Synced
			return newReplica, nil
		}
		mgr.logger.WithError(err).Errorf("Failed to validate volume replica %s.", replica.Spec.VolumeName)
		return nil, err
	}

	newReplica, err := mgr.cmdExec.CreateVolumeReplica(replica)
	if err != nil {
		mgr.logger.WithError(err).Error("Failed to exec replica create")
		return nil, err
	}

	// encrypt volume if needed
	if replica.Spec.VolumeEncrypt.Enable {
		if err = mgr.EncryptVolumeReplica(newReplica); err != nil {
			mgr.logger.WithError(err).Error("Failed to encrypt volume")
			return nil, err
		}
		mgr.logger.Debugf("Volume %s encrypted successfully", newReplica.Spec.VolumeName)
	}

	return newReplica, nil
}

func (mgr *localVolumeReplicaManager) EncryptVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error {
	key, err := getEncryptSecretKey(mgr.lm.apiClient, replica.Spec.VolumeEncrypt.SecretNamespacedName)
	if err != nil {
		mgr.logger.WithError(err).Errorf("Failed to get key from secret %s", replica.Spec.VolumeEncrypt.SecretNamespacedName)
		return err
	}

	switch replica.Spec.VolumeEncrypt.Type {
	case "LUKS":
		lk := encrypt.NewLUKS()
		return lk.EncryptVolume(replica.Spec.PoolName+"/"+replica.Spec.VolumeName, key)

	default:
		return fmt.Errorf("unsupported encrypt type %s", replica.Spec.VolumeEncrypt.Type)
	}
}

func (mgr *localVolumeReplicaManager) DeleteVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) error {
	mgr.logger.Debugf("Deleting volume replica %s.", replica.Spec.VolumeName)

	if err := mgr.volumeValidator.canDeleteVolumeReplica(replica, mgr.registry); err != nil {
		if err == ErrorReplicaNotFound {
			mgr.logger.Infof("Volume replica %s not found. Skipping.", replica.Spec.VolumeName)
			return nil
		}
		mgr.logger.WithError(err).Errorf("Failed to validate volume replica %s.", replica.Spec.VolumeName)
		return err
	}

	if err := mgr.cmdExec.DeleteVolumeReplica(replica); err != nil {
		mgr.logger.WithError(err).Error("Failed to exec replica delete")
		return err
	}

	return nil
}

func (mgr *localVolumeReplicaManager) ExpandVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica, newCapacityBytes int64) (*apisv1alpha1.LocalVolumeReplica, error) {
	mgr.logger.Debugf("Extending volume replica %s.", replica.Spec.VolumeName)

	if err := mgr.volumeValidator.canExpandVolumeReplica(replica, newCapacityBytes, mgr.registry); err != nil {
		mgr.logger.WithError(err).Errorf("Failed to validate volume replica %s to extend.", replica.Spec.VolumeName)
		return nil, err
	}

	newReplica, err := mgr.cmdExec.ExpandVolumeReplica(replica, newCapacityBytes)
	if err != nil {
		mgr.logger.WithError(err).Error("Failed to exec replica expansion.")
		return nil, err
	}

	return newReplica, nil
}

func (mgr *localVolumeReplicaManager) GetVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error) {

	currentReplica, exists := mgr.registry.VolumeReplicas()[replica.Spec.VolumeName]
	if !exists {
		return nil, fmt.Errorf("not found")
	}
	return currentReplica, nil
}

func (mgr *localVolumeReplicaManager) TestVolumeReplica(replica *apisv1alpha1.LocalVolumeReplica) (*apisv1alpha1.LocalVolumeReplica, error) {
	return mgr.cmdExec.TestVolumeReplica(replica)
}

func (mgr *localVolumeReplicaManager) ConsistencyCheck() {

	mgr.logger.Debug("Consistency Checking ...")

	replicaList := &apisv1alpha1.LocalVolumeReplicaList{}
	if err := mgr.lm.apiClient.List(context.TODO(), replicaList); err != nil {
		mgr.logger.Error("Failed to list volume replicas info from CRDs")
		return
	}
	crdReplicas := map[string]*apisv1alpha1.LocalVolumeReplica{}
	for i, item := range replicaList.Items {
		if item.Spec.NodeName != mgr.lm.nodeConf.Name {
			continue
		}
		crdReplicas[item.Spec.VolumeName] = &replicaList.Items[i]
	}

	mgr.cmdExec.ConsistencyCheck(crdReplicas)

	mgr.logger.Debug("Consistency check completed")
}

func getEncryptSecretKey(cli client.Client, namespacedName string) (string, error) {
	secret := &v1.Secret{}
	ss := strings.Split(namespacedName, "/")
	if len(ss) != 2 {
		return "", fmt.Errorf("invalid secret namespaced name %s", namespacedName)
	}

	key := client.ObjectKey{Name: ss[0], Namespace: ss[1]}
	err := cli.Get(context.Background(), key, secret)
	if err != nil {
		return "", err
	}

	if secret.Data == nil || secret.Data["key"] == nil {
		return "", fmt.Errorf("key is not exist in %s", namespacedName)
	}

	return string(secret.Data["key"]), nil
}
