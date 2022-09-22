package controller

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/wxnacy/wgo/arrays"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

const (
	imageRegistry          = "daocloud.io/daocloud"
	rcloneImageVersion     = "v1.1.2"
	rcloneImageName        = imageRegistry + "/" + "hwameistor-migrate-rclone" + ":" + rcloneImageVersion
	rcloneConfigMapName    = "migrate-rclone-config"
	rcloneConfigMapKey     = "rclone.conf"
	rcloneCertKey          = "rclonemerged"
	rcloneKeyConfigMapName = "rclone-key-config"
)

func (m *manager) startVolumeMigrateTaskWorker(stopCh <-chan struct{}) {
	m.logger.Debug("VolumeMigrate Worker is working now")
	go func() {
		for {
			task, shutdown := m.volumeMigrateTaskQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the VolumeMigrate worker")
				break
			}
			if err := m.processVolumeMigrate(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "attempts": m.volumeMigrateTaskQueue.NumRequeues(task), "error": err.Error()}).Error("Failed to process VolumeMigrate task, retry later")
				m.volumeMigrateTaskQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a VolumeMigrate task.")
				m.volumeMigrateTaskQueue.Forget(task)
			}
			m.volumeMigrateTaskQueue.Done(task)
		}
	}()

	<-stopCh
	m.volumeMigrateTaskQueue.Shutdown()
}

func (m *manager) processVolumeMigrate(vmNamespacedName string) error {
	logCtx := m.logger.WithFields(log.Fields{"VolumeMigrate": vmNamespacedName})
	logCtx.Debug("Working on a VolumeMigrate task")

	splitRes := strings.Split(vmNamespacedName, "/")
	var ns, vmName string
	if len(splitRes) >= 2 {
		ns = splitRes[0]
		vmName = splitRes[1]
	}

	migrate := &apisv1alpha1.LocalVolumeMigrate{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: vmName}, migrate); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to get VolumeMigrate from cache")
			return err
		}
		logCtx.Info("Not found the VolumeMigrate from cache, should be deleted already.")
		return nil
	}

	if migrate.Spec.Abort &&
		migrate.Status.State != apisv1alpha1.OperationStateToBeAborted &&
		migrate.Status.State != apisv1alpha1.OperationStateAborting &&
		migrate.Status.State != apisv1alpha1.OperationStateAborted &&
		migrate.Status.State != apisv1alpha1.OperationStateCompleted {

		migrate.Status.State = apisv1alpha1.OperationStateToBeAborted
		return m.apiClient.Status().Update(context.TODO(), migrate)
	}

	ctx := context.TODO()

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: migrate.Spec.VolumeName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume")
		} else {
			logCtx.Info("Not found the volume")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: localVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeMigrateSubmit: Failed to query lvg")
		} else {
			logCtx.Info("VolumeMigrateSubmit: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	var convertible = true

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			volList := &apisv1alpha1.LocalVolumeList{}
			if err := m.apiClient.List(context.TODO(), volList); err != nil {
				m.logger.WithError(err).Fatal("VolumeMigrateSubmit: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if !vol.Spec.Convertible {
						convertible = false
						break
					}
				}
			}
		}
	}

	// state chain: (empty) -> Submitted -> Start -> InProgress -> Completed

	// log with namespace/name is enough
	logCtx = m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Starting to process a VolumeMigrate task")
	switch migrate.Status.State {
	case "":
		if convertible == true {
			return m.volumeMigrateSubmit(migrate)
		} else {
			return m.nonConvertibleVolumeMigrateSubmit(migrate)
		}
	case apisv1alpha1.OperationStateSubmitted:
		if convertible == true {
			return m.volumeMigrateStart(migrate)
		} else {
			return m.nonConvertibleVolumeMigrateStart(migrate)
		}
	case apisv1alpha1.OperationStateInProgress:
		if convertible == true {
			return m.volumeMigrateInProgress(migrate)
		} else {
			return m.nonConvertibleVolumeMigrateInProgress(migrate)
		}
	case apisv1alpha1.OperationStateCompleted:
		if convertible == true {
			return m.volumeMigrateCleanup(migrate)
		} else {
			return m.nonConvertibleVolumeMigrateCleanup(migrate)
		}
	case apisv1alpha1.OperationStateToBeAborted:
		if convertible == true {
			return m.volumeMigrateAbort(migrate)
		} else {
			return m.nonConvertibleVolumeMigrateAbort(migrate)
		}
	case apisv1alpha1.OperationStateAborted:
		if convertible == true {
			return m.volumeMigrateCleanup(migrate)
		} else {
			return m.nonConvertibleVolumeMigrateCleanup(migrate)
		}
	default:
		logCtx.Error("Invalid state/phase")
	}
	return fmt.Errorf("invalid state/phase")
}

func (m *manager) volumeMigrateSubmit(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec})
	logCtx.Debug("Submit a VolumeMigrate")

	ctx := context.TODO()

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: migrate.Spec.VolumeName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume")
		} else {
			logCtx.Info("Not found the volume")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: localVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeMigrateSubmit: Failed to query lvg")
		} else {
			logCtx.Info("VolumeMigrateSubmit: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			m.lock.Lock()
			defer m.lock.Unlock()

			volList := &apisv1alpha1.LocalVolumeList{}
			if err := m.apiClient.List(context.TODO(), volList); err != nil {
				m.logger.WithError(err).Fatal("VolumeMigrateSubmit: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Name != migrate.Spec.VolumeName && migrate.Spec.MigrateAllVols == false {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "migrateAllVols": migrate.Spec.MigrateAllVols}).Error("VolumeMigrateSubmit: Can't migrate false migrateAllVols flag volume")
						migrate.Status.Message = "VolumeMigrateSubmit: Can't migrate volume whose localVolumeGroup has other volumes, meantime migrate's migrateAllVols flag is false; If you want migrateAllVols, modify migrateAllVols flag into true"
						return m.apiClient.Status().Update(context.TODO(), migrate)
					}
				}
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Spec.ReplicaNumber == 1 && !vol.Spec.Convertible {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "replicaNumber": vol.Spec.ReplicaNumber}).Error("VolumeMigrateSubmit: Can't migrate inconvertible non-HA volume")
						migrate.Status.Message = "VolumeMigrateSubmit: Can't migrate inconvertible non-HA volume"
						return m.apiClient.Status().Update(context.TODO(), migrate)
					}

					if vol.Status.State != apisv1alpha1.VolumeStateReady {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "state": vol.Status.State}).Error("Volume is not ready")
						replicas, err := m.getReplicasForVolume(vol.Name)
						if err != nil {
							logCtx.Error("VolumeMigrateSubmit: Failed to list VolumeReplica")
							return err
						}
						if len(replicas) != int(vol.Spec.ReplicaNumber) {
							logCtx.Info("VolumeMigrateSubmit: Not all VolumeReplicas are created")
							return fmt.Errorf("VolumeMigrateSubmit: volume not ready")
						}
						var needMigrate bool
						for _, replica := range replicas {
							if replica.Status.State == apisv1alpha1.VolumeReplicaStateReady {
								needMigrate = true
								break
							}
						}
						if needMigrate == false {
							migrate.Status.Message = "VolumeMigrateSubmit: Volume And VolumeReplica both not ready"
							return m.apiClient.Status().Update(context.TODO(), migrate)
						}
					}

					migrate.Status.ReplicaNumber = vol.Spec.ReplicaNumber
					migrate.Status.State = apisv1alpha1.OperationStateSubmitted
					m.apiClient.Status().Update(context.TODO(), migrate)
				}
			}
		}
	}
	return nil
}

func (m *manager) nonConvertibleVolumeMigrateSubmit(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec})
	logCtx.Debug("nonConvertibleVolumeMigrateSubmit Submit a VolumeMigrate")

	ctx := context.TODO()

	rcl := m.dataCopyManager.UseRclone(rcloneImageName, rcloneConfigMapName, rcloneKeyConfigMapName, migrate.Namespace, rcloneConfigMapKey, rcloneCertKey)
	if err := rcl.EnsureRcloneConfigMapToTargetNamespace(migrate.Namespace); err != nil {
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}
	m.rclone = rcl

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: migrate.Spec.VolumeName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume")
		} else {
			logCtx.Info("Not found the volume")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: localVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeMigrateSubmit: Failed to query lvg")
		} else {
			logCtx.Info("VolumeMigrateSubmit: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	logCtx.Debug("nonConvertibleVolumeMigrateSubmit migrate.Namespace = %v", migrate.Namespace)
	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			m.lock.Lock()
			defer m.lock.Unlock()

			tmpRcloneKeyConfigMap := &corev1.ConfigMap{}
			err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: migrate.Namespace, Name: rcloneKeyConfigMapName}, tmpRcloneKeyConfigMap)
			if err == nil {
				if delErr := m.apiClient.Delete(context.TODO(), tmpRcloneKeyConfigMap); delErr != nil {
					m.logger.WithError(err).Error("Failed to delete Configmap")
					return delErr
				}
			}

			rcloneKeyCM := m.rclone.GenerateRcloneKeyConfigMap(migrate.Namespace)
			logCtx.Debug("nonConvertibleVolumeMigrateSubmit rcloneKeyCM = %v", rcloneKeyCM)

			err = m.apiClient.Create(ctx, rcloneKeyCM)
			if err != nil {
				if errors.IsAlreadyExists(err) {
					logCtx.WithError(err).Error("Failed to create RcloneKeyConfigmap, already exists")
				} else {
					logCtx.WithError(err).Error("Failed to create RcloneKeyConfigmap")
					migrate.Status.Message = err.Error()
					m.apiClient.Status().Update(ctx, migrate)
					return err
				}
			}

			err = m.makeMigrateRcloneConfigmap(migrate.Spec.TargetNodesNames[0], migrate.Spec.SourceNodesNames[0], migrate.Namespace, migrate.Spec.VolumeName)
			if err != nil {
				if errors.IsAlreadyExists(err) {
					logCtx.WithError(err).Error("Failed to makeMigrateRcloneConfigmap, already exists")
				} else {
					logCtx.WithError(err).Error("Failed to makeMigrateRcloneConfigmap")
					migrate.Status.Message = err.Error()
					m.apiClient.Status().Update(ctx, migrate)
					return err
				}
			}

			// cannot depend on vol.Spec.ReplicaNumber
			migrate.Status.ReplicaNumber = 1
			migrate.Status.State = apisv1alpha1.OperationStateSubmitted
			m.apiClient.Status().Update(context.TODO(), migrate)
		}
	}
	return nil
}

func (m *manager) volumeMigrateStart(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start a VolumeMigrate")

	ctx := context.TODO()

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: migrate.Spec.VolumeName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume")
		} else {
			logCtx.Info("Not found the volume")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: localVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeMigrateStart: Failed to query lvg")
		} else {
			logCtx.Info("VolumeMigrateStart: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			m.lock.Lock()
			defer m.lock.Unlock()

			volList := &apisv1alpha1.LocalVolumeList{}
			if err := m.apiClient.List(context.TODO(), volList); err != nil {
				m.logger.WithError(err).Fatal("VolumeMigrateStart: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Spec.Config == nil {
						migrate.Status.Message = "VolumeMigrateStart: Volume to be migrated is not ready yet"
						m.apiClient.Status().Update(ctx, migrate)
						return fmt.Errorf("VolumeMigrateStart: volume not ready")
					}

					if vol.Spec.ReplicaNumber > migrate.Status.ReplicaNumber {
						migrate.Status.State = apisv1alpha1.OperationStateInProgress
						return m.apiClient.Status().Update(ctx, migrate)
					}

					if vol.Status.State != apisv1alpha1.VolumeStateReady {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "state": vol.Status.State}).Error("VolumeMigrateStart: The volume is not ready")
						replicas, err := m.getReplicasForVolume(vol.Name)
						if err != nil {
							logCtx.Error("VolumeMigrateStart: Failed to list VolumeReplica")
							return err
						}
						if len(replicas) != int(vol.Spec.ReplicaNumber) {
							logCtx.Info("VolumeMigrateStart: Not all VolumeReplicas are created")
							return fmt.Errorf("VolumeMigrateStart: volume not ready")
						}
						var needMigrate bool
						for _, replica := range replicas {
							if replica.Status.State == apisv1alpha1.VolumeReplicaStateReady {
								needMigrate = true
								break
							}
						}
						if needMigrate == false {
							migrate.Status.Message = "VolumeMigrateStart: Volume And VolumeReplica both not ready"
							return m.apiClient.Status().Update(context.TODO(), migrate)
						}
					}

					replicas, err := m.getReplicasForVolume(vol.Name)
					if err != nil {
						logCtx.Error("VolumeMigrateStart: Failed to list VolumeReplica")
						return err
					}

					var needMigrateNum int
					if len(replicas) > len(migrate.Spec.TargetNodesNames) {
						needMigrateNum = len(migrate.Spec.TargetNodesNames)
					} else {
						needMigrateNum = len(replicas)
					}

					for i := 0; i < needMigrateNum; i++ {
						// start the migrate by adding a new replica which will be scheduled on a new node
						// trigger the migration
						vol.Spec.ReplicaNumber++
					}
					if err := m.apiClient.Update(ctx, &vol); err != nil {
						logCtx.WithFields(log.Fields{"volume": vol.Name}).WithError(err).Error("VolumeMigrateStart: Failed to add a new replica")
						migrate.Status.Message = err.Error()
						return m.apiClient.Status().Update(ctx, migrate)
					}
					migrate.Status.State = apisv1alpha1.OperationStateInProgress
					m.apiClient.Status().Update(ctx, migrate)
				}
			}
		}
	}

	return nil
}

func (m *manager) nonConvertibleVolumeMigrateStart(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("nonConvertibleVolumeMigrateStart Start a VolumeMigrate")

	ctx := context.TODO()
	rcl := m.dataCopyManager.UseRclone(rcloneImageName, rcloneConfigMapName, rcloneKeyConfigMapName, migrate.Namespace, rcloneConfigMapKey, rcloneCertKey)
	if err := rcl.EnsureRcloneConfigMapToTargetNamespace(migrate.Namespace); err != nil {
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: migrate.Spec.VolumeName}, vol); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("Failed to query volume")
		} else {
			logCtx.Info("Not found the volume")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: localVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeMigrateStart: Failed to query lvg")
		} else {
			logCtx.Info("VolumeMigrateStart: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			m.lock.Lock()
			defer m.lock.Unlock()

			volList := &apisv1alpha1.LocalVolumeList{}
			if err := m.apiClient.List(context.TODO(), volList); err != nil {
				m.logger.WithError(err).Fatal("VolumeMigrateStart: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Spec.Config == nil {
						migrate.Status.Message = "VolumeMigrateStart: Volume to be migrated is not ready yet"
						m.apiClient.Status().Update(ctx, migrate)
						return fmt.Errorf("VolumeMigrateStart: volume not ready")
					}

					if vol.Spec.ReplicaNumber == 1 {
						vol.Spec.ReplicaNumber++
					}

					if err := m.apiClient.Update(ctx, &vol); err != nil {
						logCtx.WithFields(log.Fields{"volume": vol.Name}).WithError(err).Error("VolumeMigrateStart: Failed to add a new replica")
						migrate.Status.Message = err.Error()
						return m.apiClient.Status().Update(ctx, migrate)
					}

					jobName := migrate.Name + "-job-" + vol.Spec.PersistentVolumeClaimName
					var migrateSrcName, migrateDstName string
					if len(migrate.Spec.SourceNodesNames) == 1 && migrate.Spec.SourceNodesNames[0] != "" {
						migrateSrcName = migrate.Spec.SourceNodesNames[0]
					}
					if len(migrate.Spec.TargetNodesNames) == 1 && migrate.Spec.TargetNodesNames[0] != "" {
						migrateDstName = migrate.Spec.TargetNodesNames[0]
					}
					if err := rcl.SrcMountPointToRemoteMountPoint(jobName, migrate.Namespace, vol.Spec.PoolName, vol.Name, migrateSrcName, migrateDstName, true, 0); err != nil {
						logCtx.WithError(err).Error("VolumeMigrateStart: Job SrcMountPointToRemoteMountPoint failed")
						migrate.Status.Message = "VolumeMigrateStart: Job SrcMountPointToRemoteMountPoint failed"
						m.apiClient.Status().Update(ctx, migrate)
						return err
					}
				}
			}
			migrate.Status.State = apisv1alpha1.OperationStateInProgress
			m.apiClient.Status().Update(ctx, migrate)
		}
	}

	return nil
}

func (m *manager) volumeMigrateInProgress(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start a volumeMigrateInProgress")

	ctx := context.TODO()

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: migrate.Spec.VolumeName}, vol); err != nil {
		logCtx.WithError(err).Error("Failed to query volume")
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: localVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("VolumeMigrateInProgress: Failed to query lvg")
		} else {
			logCtx.WithError(err).Error("VolumeMigrateInProgress: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			m.lock.Lock()
			defer m.lock.Unlock()

			volList := &apisv1alpha1.LocalVolumeList{}
			if err := m.apiClient.List(context.TODO(), volList); err != nil {
				m.logger.WithError(err).Fatal("VolumeMigrateInProgress: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					// firstly, make sure all the replicas are ready
					if int(vol.Spec.ReplicaNumber) != len(vol.Spec.Config.Replicas) {
						logCtx.Debug("VolumeMigrateInProgress: Volume is still not configured")
						return fmt.Errorf("VolumeMigrateInProgress: volume not ready")

					}
					replicas, err := m.getReplicasForVolume(vol.Name)
					if err != nil {
						logCtx.Error("VolumeMigrateInProgress: Failed to list VolumeReplica")
						return err
					}
					if len(replicas) != int(vol.Spec.ReplicaNumber) {
						logCtx.Info("VolumeMigrateInProgress: Not all VolumeReplicas are created")
						return fmt.Errorf("VolumeMigrateInProgress: volume not ready")
					}
					hasOldReplica := false
					for _, replica := range replicas {
						if replica.Status.State != apisv1alpha1.VolumeReplicaStateReady {
							logCtx.Info("VolumeMigrateInProgress: Not all VolumeReplicas are ready")
							return fmt.Errorf("VolumeMigrateInProgress: volume not ready")
						}
						for _, nodeName := range migrate.Spec.SourceNodesNames {
							if replica.Spec.NodeName == nodeName {
								hasOldReplica = true
								break
							}
						}
					}

					// New replica is added and synced successfully, will remove the to-be-migrated replica from Volume's config
					if vol.Spec.ReplicaNumber > migrate.Status.ReplicaNumber {
						// prune the to-be-migrated replica

						vol.Spec.ReplicaNumber = migrate.Status.ReplicaNumber
						replicas := []apisv1alpha1.VolumeReplica{}
						migrateSrcHostNames := []string{}
						for _, nodeName := range migrate.Spec.SourceNodesNames {
							migrateSrcHostNames = append(migrateSrcHostNames, nodeName)
						}
						for i := range vol.Spec.Config.Replicas {
							if arrays.ContainsString(migrateSrcHostNames, vol.Spec.Config.Replicas[i].Hostname) == -1 {
								replicas = append(replicas, vol.Spec.Config.Replicas[i])
							}
						}
						vol.Spec.Config.Replicas = replicas

						if err := m.apiClient.Update(ctx, &vol); err != nil {
							logCtx.WithError(err).Error("VolumeMigrateInProgress: Failed to re-configure Volume")
							migrate.Status.Message = err.Error()
							m.apiClient.Status().Update(ctx, migrate)
							return err
						}

						return fmt.Errorf("VolumeMigrateInProgress: wait old replica deleted")
					}

					if hasOldReplica {
						logCtx.Info("VolumeMigrateInProgress: The old replica has not been cleanup")
						return fmt.Errorf("VolumeMigrateInProgress: not cleanup")
					}

					migrate.Status.State = apisv1alpha1.OperationStateCompleted
					m.apiClient.Status().Update(ctx, migrate)
				}
			}
		}
	}

	for _, nodeName := range migrate.Spec.TargetNodesNames {
		if arrays.ContainsString(lvg.Spec.Accessibility.Nodes, nodeName) == -1 {
			lvg.Spec.Accessibility.Nodes = append(lvg.Spec.Accessibility.Nodes, nodeName)
		}
	}

	if err := m.apiClient.Update(context.TODO(), lvg); err != nil {
		log.WithError(err).Error("VolumeMigrateInProgress Reconcile : Failed to re-configure Volume")
		return err
	}

	return nil
}

func (m *manager) nonConvertibleVolumeMigrateInProgress(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start a nonConvertibleVolumeMigrateInProgress")

	ctx := context.TODO()

	rcl := m.dataCopyManager.UseRclone(rcloneImageName, rcloneConfigMapName, rcloneKeyConfigMapName, migrate.Namespace, rcloneConfigMapKey, rcloneCertKey)
	if err := rcl.EnsureRcloneConfigMapToTargetNamespace(migrate.Namespace); err != nil {
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	vol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Name: migrate.Spec.VolumeName}, vol); err != nil {
		logCtx.WithError(err).Error("Failed to query volume")
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	localVolumeGroupName := vol.Spec.VolumeGroup

	lvg := &apisv1alpha1.LocalVolumeGroup{}
	if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: localVolumeGroupName}, lvg); err != nil {
		if !errors.IsNotFound(err) {
			logCtx.WithError(err).Error("nonConvertibleVolumeMigrateInProgress: Failed to query lvg")
		} else {
			logCtx.WithError(err).Error("nonConvertibleVolumeMigrateInProgress: Not found the lvg")
		}
		migrate.Status.Message = err.Error()
		m.apiClient.Status().Update(ctx, migrate)
		return err
	}

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			m.lock.Lock()
			defer m.lock.Unlock()

			volList := &apisv1alpha1.LocalVolumeList{}
			if err := m.apiClient.List(context.TODO(), volList); err != nil {
				m.logger.WithError(err).Fatal("nonConvertibleVolumeMigrateInProgress: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {

					jobName := migrate.Name + "-job-" + vol.Spec.PersistentVolumeClaimName

					migrateJob := &batchv1.Job{}
					err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: vol.Spec.PersistentVolumeClaimNamespace, Name: jobName}, migrateJob)
					if err != nil {
						if !errors.IsNotFound(err) {
							logCtx.WithError(err).Error("Failed to get MigrateJob from cache")
							return err
						}
						logCtx.Info("Not found the MigrateJob from cache, should be deleted already.")
					}

					if err == nil {
						if err := rcl.WaitMigrateJobTaskDone(jobName, vol.Name, true, 0); err != nil {
							logCtx.WithError(err).Error("nonConvertibleVolumeMigrateInProgress: Job SrcMountPointToRemoteMountPoint failed")
							migrate.Status.Message = "nonConvertibleVolumeMigrateInProgress: Job SrcMountPointToRemoteMountPoint failed"
							m.apiClient.Status().Update(ctx, migrate)
							return err
						}
					}

					rclonecm := &corev1.ConfigMap{}
					err = m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: migrate.Namespace, Name: rcloneConfigMapName}, rclonecm)
					if err != nil {
						if !errors.IsNotFound(err) {
							logCtx.WithError(err).Error("Failed to get rclonecm from cache")
							return err
						}
						logCtx.Info("Not found the rclonecm from cache, should be deleted already.")
						return err
					}

					rclonecm.Data["syncDone"] = "True"
					err = m.apiClient.Update(context.TODO(), rclonecm, &client.UpdateOptions{})
					if err != nil {
						logCtx.WithError(err).Error("Failed to update rclonecm")
						return err
					}

					// firstly, make sure all the replicas are ready
					if int(vol.Spec.ReplicaNumber) != len(vol.Spec.Config.Replicas) {
						logCtx.Debug("VolumeMigrateInProgress: Volume is still not configured")
						return fmt.Errorf("VolumeMigrateInProgress: volume not ready")
					}
					replicas, err := m.getReplicasForVolume(vol.Name)
					if err != nil {
						logCtx.Error("VolumeMigrateInProgress: Failed to list VolumeReplica")
						return err
					}
					if len(replicas) != int(vol.Spec.ReplicaNumber) {
						logCtx.Info("VolumeMigrateInProgress: Not all VolumeReplicas are created")
						return fmt.Errorf("VolumeMigrateInProgress: volume not ready")
					}
					hasOldReplica := false
					for _, replica := range replicas {
						if replica.Status.State != apisv1alpha1.VolumeReplicaStateReady {
							logCtx.Info("VolumeMigrateInProgress: Not all VolumeReplicas are ready")
							return fmt.Errorf("VolumeMigrateInProgress: volume not ready")
						}
						for _, nodeName := range migrate.Spec.SourceNodesNames {
							if replica.Spec.NodeName == nodeName {
								hasOldReplica = true
								break
							}
						}
					}

					// New replica is added and synced successfully, will remove the to-be-migrated replica from Volume's config
					if vol.Spec.ReplicaNumber > migrate.Status.ReplicaNumber {
						// prune the to-be-migrated replica

						vol.Spec.ReplicaNumber = migrate.Status.ReplicaNumber
						replicas := []apisv1alpha1.VolumeReplica{}
						migrateSrcHostNames := []string{}
						for _, nodeName := range migrate.Spec.SourceNodesNames {
							migrateSrcHostNames = append(migrateSrcHostNames, nodeName)
						}
						for i := range vol.Spec.Config.Replicas {
							if arrays.ContainsString(migrateSrcHostNames, vol.Spec.Config.Replicas[i].Hostname) == -1 {
								replicas = append(replicas, vol.Spec.Config.Replicas[i])
							}
						}

						logCtx.Debugf("VolumeMigrateInProgress replicas = %v, vol.Spec.ReplicaNumber = %v", replicas, vol.Spec.ReplicaNumber)
						vol.Spec.Config.Replicas = replicas

						if err := m.apiClient.Update(ctx, &vol); err != nil {
							logCtx.WithError(err).Error("VolumeMigrateInProgress: Failed to re-configure Volume")
							migrate.Status.Message = err.Error()
							m.apiClient.Status().Update(ctx, migrate)
							return err
						}

						return fmt.Errorf("VolumeMigrateInProgress: wait old replica deleted")
					}

					if hasOldReplica {
						logCtx.Info("VolumeMigrateInProgress: The old replica has not been cleanup")
						return fmt.Errorf("VolumeMigrateInProgress: not cleanup")
					}

					migrate.Status.State = apisv1alpha1.OperationStateCompleted
					m.apiClient.Status().Update(ctx, migrate)
				}
			}
		}
	}

	for _, nodeName := range migrate.Spec.TargetNodesNames {
		if arrays.ContainsString(lvg.Spec.Accessibility.Nodes, nodeName) == -1 {
			lvg.Spec.Accessibility.Nodes = append(lvg.Spec.Accessibility.Nodes, nodeName)
		}
	}

	if err := m.apiClient.Update(context.TODO(), lvg); err != nil {
		log.WithError(err).Error("VolumeMigrateInProgress Reconcile : Failed to re-configure Volume")
		return err
	}

	return nil
}

func (m *manager) volumeMigrateAbort(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Abort a VolumeMigrate")

	migrate.Status.State = apisv1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), migrate)
}

func (m *manager) nonConvertibleVolumeMigrateAbort(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Abort a nonConvertibleVolumeMigrate")

	migrate.Status.State = apisv1alpha1.OperationStateAborted
	return m.apiClient.Status().Update(context.TODO(), migrate)
}

func (m *manager) volumeMigrateCleanup(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Cleanup a VolumeMigrate")

	return m.apiClient.Delete(context.TODO(), migrate)
}

func (m *manager) nonConvertibleVolumeMigrateCleanup(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Cleanup a nonConvertibleVolumeMigrate")

	return m.apiClient.Delete(context.TODO(), migrate)
}

func (m *manager) makeMigrateRcloneConfigmap(targetNodeName, sourceNodeName, ns, lvname string) error {
	ctx := context.TODO()

	tmpConfigMap := &corev1.ConfigMap{}
	err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: rcloneConfigMapName}, tmpConfigMap)
	if err == nil {
		if delErr := m.apiClient.Delete(context.TODO(), tmpConfigMap); delErr != nil {
			m.logger.WithError(err).Error("Failed to delete Configmap")
			return delErr
		}
	}

	remoteNameData := "[remote]" + "\n"
	sourceNameData := "[source]" + "\n"
	typeData := "type = sftp" + "\n"
	remoteHostData := "host = " + targetNodeName + "\n"
	sourceHostData := "host = " + sourceNodeName + "\n"
	keyFileData := "key_file = /config/rclone/rclonemerged" + "\n"
	shellTypeData := "shell_type = unix" + "\n"
	md5sumCommandData := "md5sum_command = md5sum" + "\n"
	sha1sumCommandData := "sha1sum_command = sha1sum" + "\n"

	data := map[string]string{
		rcloneConfigMapKey: remoteNameData + typeData + remoteHostData + keyFileData + shellTypeData + md5sumCommandData + sha1sumCommandData +
			sourceNameData + typeData + sourceHostData + keyFileData + shellTypeData + md5sumCommandData + sha1sumCommandData,
		"lvname":         lvname,
		"targetNodeName": targetNodeName,
		"sourceNodeName": sourceNodeName,
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rcloneConfigMapName,
			Namespace: ns,
			Labels:    map[string]string{},
		},
		Data: data,
	}

	if err := m.apiClient.Create(ctx, cm); err != nil {
		m.logger.WithError(err).Error("Failed to create MigrateConfigmap")
		return err
	}

	return nil
}
