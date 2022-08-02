package controller

import (
	"context"
	"fmt"
	"github.com/wxnacy/wgo/arrays"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage/v1alpha1"
)

const (
	imageRegistry                = "daocloud.io/daocloud"
	rcloneImageVersion           = "v1.1.2"
	rcloneImageName              = imageRegistry + "/" + "hwameistor-migrate-rclone" + ":" + rcloneImageVersion
	rcloneConfigMapName          = "migrate-rclone-config"
	rcloneConfigMapKey           = "rclone.conf"
	migrateDstPodName            = "nonconvertible-dst-migrate-pod"
	imagePullSecrets             = "docker-secret"
	migrateDstPodContainerName   = "migrate"
	migrateDstPodImageName       = imageRegistry + "/" + "nginx:v0.0.1"
	migrateVolumeMountPathPrefix = "/var/data/"
	migrateVolumePrefix          = "migrate-data-"
	migratePvcSuffix             = "-migrate"
	volumeSelectedNodeKey        = "volume.kubernetes.io/selected-node"
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
			logCtx.WithError(err).Error("VolumeGroupMigrateSubmit: Failed to query lvg")
		} else {
			logCtx.Info("VolumeGroupMigrateSubmit: Not found the lvg")
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
				m.logger.WithError(err).Fatal("VolumeGroupMigrateSubmit: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Spec.ReplicaNumber == 1 && !vol.Spec.Convertible {
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
			logCtx.WithError(err).Error("VolumeGroupMigrateSubmit: Failed to query lvg")
		} else {
			logCtx.Info("VolumeGroupMigrateSubmit: Not found the lvg")
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
				m.logger.WithError(err).Fatal("VolumeGroupMigrateSubmit: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Name != migrate.Spec.VolumeName && migrate.Spec.MigrateAllVols == false {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "migrateAllVols": migrate.Spec.MigrateAllVols}).Error("VolumeGroupMigrateSubmit: Can't migrate false migrateAllVols flag volume")
						migrate.Status.Message = "VolumeGroupMigrateSubmit: Can't migrate volume whose localVolumeGroup has other volumes, meantime migrate's migrateAllVols flag is false; If you want migrateAllVols, modify migrateAllVols flag into true"
						return m.apiClient.Status().Update(context.TODO(), migrate)
					}
				}
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Spec.ReplicaNumber == 1 && !vol.Spec.Convertible {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "replicaNumber": vol.Spec.ReplicaNumber}).Error("VolumeGroupMigrateSubmit: Can't migrate inconvertible non-HA volume")
						migrate.Status.Message = "VolumeGroupMigrateSubmit: Can't migrate inconvertible non-HA volume"
						return m.apiClient.Status().Update(context.TODO(), migrate)
					}

					if vol.Status.State != apisv1alpha1.VolumeStateReady {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "state": vol.Status.State}).Error("Volume is not ready")
						replicas, err := m.getReplicasForVolume(vol.Name)
						if err != nil {
							logCtx.Error("VolumeGroupMigrateSubmit: Failed to list VolumeReplica")
							return err
						}
						if len(replicas) != int(vol.Spec.ReplicaNumber) {
							logCtx.Info("VolumeGroupMigrateSubmit: Not all VolumeReplicas are created")
							return fmt.Errorf("VolumeGroupMigrateSubmit: volume not ready")
						}
						var needMigrate bool
						for _, replica := range replicas {
							if replica.Status.State == apisv1alpha1.VolumeReplicaStateReady {
								needMigrate = true
								break
							}
						}
						if needMigrate == false {
							migrate.Status.Message = "VolumeGroupMigrateSubmit: Volume And VolumeReplica both not ready"
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
			logCtx.WithError(err).Error("VolumeGroupMigrateSubmit: Failed to query lvg")
		} else {
			logCtx.Info("VolumeGroupMigrateSubmit: Not found the lvg")
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
				m.logger.WithError(err).Fatal("VolumeGroupMigrateSubmit: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Name != migrate.Spec.VolumeName && migrate.Spec.MigrateAllVols == false {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "migrateAllVols": migrate.Spec.MigrateAllVols}).Error("VolumeGroupMigrateSubmit: Can't migrate false migrateAllVols flag volume")
						migrate.Status.Message = "VolumeGroupMigrateSubmit: Can't migrate volume whose localVolumeGroup has other volumes, meantime migrate's migrateAllVols flag is false; If you want migrateAllVols, modify migrateAllVols flag into true"
						return m.apiClient.Status().Update(context.TODO(), migrate)
					}
				}
			}

			var lvs = []apisv1alpha1.LocalVolume{}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					pvcName := vol.Spec.PersistentVolumeClaimName
					pvc := &corev1.PersistentVolumeClaim{}
					if err := m.apiClient.Get(ctx, types.NamespacedName{Name: pvcName, Namespace: vol.Spec.PersistentVolumeClaimNamespace}, pvc); err != nil {
						if !errors.IsNotFound(err) {
							logCtx.WithError(err).Error("Failed to query pvc")
						} else {
							logCtx.Info("Not found the pvc")
						}
						migrate.Status.Message = err.Error()
						m.apiClient.Status().Update(ctx, migrate)
						return err
					}

					migratePvc, err := m.makeMigratePVC(pvc, migrate.Spec.TargetNodesNames)
					if err != nil {
						logCtx.WithError(err).Error("Failed to makeMigratePVC")
						migrate.Status.Message = err.Error()
						m.apiClient.Status().Update(ctx, migrate)
						return err
					}

					if err := m.apiClient.Create(ctx, migratePvc); err != nil {
						if errors.IsAlreadyExists(err) {
							logCtx.WithError(err).Error("Failed to create MigratePVC, persistentvolumeclaims already exists")
						} else {
							logCtx.WithError(err).Error("Failed to create MigratePVC")
							migrate.Status.Message = err.Error()
							m.apiClient.Status().Update(ctx, migrate)
							return err
						}
					}

					lvs = append(lvs, vol)
				}
			}

			migratePod, err := m.makeMigratePod(lvs, migrate.Namespace)
			if err != nil {
				logCtx.WithError(err).Error("Failed to makeMigratePod")
				migrate.Status.Message = err.Error()
				m.apiClient.Status().Update(ctx, migrate)
				return err
			}

			if err := m.apiClient.Create(ctx, migratePod); err != nil {
				if errors.IsAlreadyExists(err) {
					logCtx.WithError(err).Error("Failed to create MigratePod, pod already exists")
				} else {
					logCtx.WithError(err).Error("Failed to create MigratePod")
					migrate.Status.Message = err.Error()
					m.apiClient.Status().Update(ctx, migrate)
					return err
				}
			}

			var runningTargetPod = &corev1.Pod{}
			var podIp string
			for {
				if err := m.apiClient.Get(ctx, types.NamespacedName{Namespace: migrate.Namespace, Name: migratePod.Name}, runningTargetPod); err != nil {
					if !errors.IsNotFound(err) {
						logCtx.WithError(err).Error("nonConvertibleVolumeMigrateSubmit: Failed to query pod")
					} else {
						logCtx.Info("nonConvertibleVolumeMigrateSubmit: Not found the pod")
					}
					migrate.Status.Message = err.Error()
					m.apiClient.Status().Update(ctx, migrate)
					return err
				}
				if runningTargetPod.Status.PodIP != "" {
					podIp = runningTargetPod.Status.PodIP
					break
				}
			}

			err = m.makeMigrateRcloneConfigmap(podIp, migrate.Namespace)
			if err != nil {
				if errors.IsAlreadyExists(err) {
					logCtx.WithError(err).Error("Failed to create MigrateRcloneConfigmap, MigrateRcloneConfigmap already exists")
				} else {
					logCtx.WithError(err).Error("Failed to create MigrateRcloneConfigmap")
					migrate.Status.Message = err.Error()
					m.apiClient.Status().Update(ctx, migrate)
					return err
				}
			}

			migrate.Status.ReplicaNumber = vol.Spec.ReplicaNumber
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
			logCtx.WithError(err).Error("VolumeGroupMigrateStart: Failed to query lvg")
		} else {
			logCtx.Info("VolumeGroupMigrateStart: Not found the lvg")
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
				m.logger.WithError(err).Fatal("VolumeGroupMigrateStart: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Spec.Config == nil {
						migrate.Status.Message = "VolumeGroupMigrateStart: Volume to be migrated is not ready yet"
						m.apiClient.Status().Update(ctx, migrate)
						return fmt.Errorf("VolumeGroupMigrateStart: volume not ready")
					}

					if vol.Spec.ReplicaNumber > migrate.Status.ReplicaNumber {
						migrate.Status.State = apisv1alpha1.OperationStateInProgress
						return m.apiClient.Status().Update(ctx, migrate)
					}

					if vol.Status.State != apisv1alpha1.VolumeStateReady {
						logCtx.WithFields(log.Fields{"volume": vol.Name, "state": vol.Status.State}).Error("VolumeGroupMigrateStart: The volume is not ready")
						replicas, err := m.getReplicasForVolume(vol.Name)
						if err != nil {
							logCtx.Error("VolumeGroupMigrateStart: Failed to list VolumeReplica")
							return err
						}
						if len(replicas) != int(vol.Spec.ReplicaNumber) {
							logCtx.Info("VolumeGroupMigrateStart: Not all VolumeReplicas are created")
							return fmt.Errorf("VolumeGroupMigrateStart: volume not ready")
						}
						var needMigrate bool
						for _, replica := range replicas {
							if replica.Status.State == apisv1alpha1.VolumeReplicaStateReady {
								needMigrate = true
								break
							}
						}
						if needMigrate == false {
							migrate.Status.Message = "VolumeGroupMigrateStart: Volume And VolumeReplica both not ready"
							return m.apiClient.Status().Update(context.TODO(), migrate)
						}
						return m.apiClient.Status().Update(context.TODO(), migrate)
					}

					// start the migrate by adding a new replica which will be scheduled on a new node
					// trigger the migration
					vol.Spec.ReplicaNumber++
					if err := m.apiClient.Update(ctx, &vol); err != nil {
						logCtx.WithFields(log.Fields{"volume": vol.Name}).WithError(err).Error("VolumeGroupMigrateStart: Failed to add a new replica")
						migrate.Status.Message = err.Error()
					} else {
						migrate.Status.State = apisv1alpha1.OperationStateInProgress
					}

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

	rcl := m.dataCopyManager.UseRclone(rcloneImageName, rcloneConfigMapName, migrate.Namespace, rcloneConfigMapKey)
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
			logCtx.WithError(err).Error("VolumeGroupMigrateStart: Failed to query lvg")
		} else {
			logCtx.Info("VolumeGroupMigrateStart: Not found the lvg")
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
				m.logger.WithError(err).Fatal("VolumeGroupMigrateStart: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					if vol.Spec.Config == nil {
						migrate.Status.Message = "VolumeGroupMigrateStart: Volume to be migrated is not ready yet"
						m.apiClient.Status().Update(ctx, migrate)
						return fmt.Errorf("VolumeGroupMigrateStart: volume not ready")
					}
					//jobName := utils.GenerateResourceName([]string{"hwameistor", migrate.Name}, true, true, 63)
					jobName := migrate.Name + "-job-" + vol.Spec.PersistentVolumeClaimName
					if err := rcl.PVCToRemotePVC(jobName, vol.Spec.PersistentVolumeClaimName, "", vol.Spec.PersistentVolumeClaimName+migratePvcSuffix, "", "", vol.Spec.PersistentVolumeClaimNamespace, true, 0); err != nil {
						logCtx.WithError(err).Error("VolumeGroupMigrateStart: Job PVCToPVC failed")
						migrate.Status.Message = "VolumeGroupMigrateStart: Job PVCToRemotePVC failed"
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
			logCtx.WithError(err).Error("VolumeGroupMigrateInProgress: Failed to query lvg")
		} else {
			logCtx.WithError(err).Error("VolumeGroupMigrateInProgress: Not found the lvg")
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
				m.logger.WithError(err).Fatal("VolumeGroupMigrateInProgress: Failed to list LocalVolumes")
			}

			for _, vol := range volList.Items {
				if vol.Spec.VolumeGroup == lvg.Name {
					// firstly, make sure all the replicas are ready
					if int(vol.Spec.ReplicaNumber) != len(vol.Spec.Config.Replicas) {
						logCtx.Debug("VolumeGroupMigrateInProgress: Volume is still not configured")
						return fmt.Errorf("VolumeGroupMigrateInProgress: volume not ready")

					}
					replicas, err := m.getReplicasForVolume(vol.Name)
					if err != nil {
						logCtx.Error("VolumeGroupMigrateInProgress: Failed to list VolumeReplica")
						return err
					}
					if len(replicas) != int(vol.Spec.ReplicaNumber) {
						logCtx.Info("VolumeGroupMigrateInProgress: Not all VolumeReplicas are created")
						return fmt.Errorf("VolumeGroupMigrateInProgress: volume not ready")
					}
					hasOldReplica := false
					for _, replica := range replicas {
						if replica.Status.State != apisv1alpha1.VolumeReplicaStateReady {
							logCtx.Info("VolumeGroupMigrateInProgress: Not all VolumeReplicas are ready")
							return fmt.Errorf("VolumeGroupMigrateInProgress: volume not ready")
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
						for i := range vol.Spec.Config.Replicas {
							for _, nodeName := range migrate.Spec.SourceNodesNames {
								if vol.Spec.Config.Replicas[i].Hostname != nodeName {
									replicas = append(replicas, vol.Spec.Config.Replicas[i])
								}
							}
						}
						vol.Spec.Config.Replicas = replicas

						if err := m.apiClient.Update(ctx, &vol); err != nil {
							logCtx.WithError(err).Error("VolumeGroupMigrateInProgress: Failed to re-configure Volume")
							migrate.Status.Message = err.Error()
							m.apiClient.Status().Update(ctx, migrate)
							return err
						}

						return fmt.Errorf("VolumeGroupMigrateInProgress: wait old replica deleted")
					}

					if hasOldReplica {
						logCtx.Info("VolumeGroupMigrateInProgress: The old replica has not been cleanup")
						return fmt.Errorf("VolumeGroupMigrateInProgress: not cleanup")
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
		log.WithError(err).Error("VolumeGroupMigrateInProgress Reconcile : Failed to re-configure Volume")
		return err
	}

	return nil
}

func (m *manager) nonConvertibleVolumeMigrateInProgress(migrate *apisv1alpha1.LocalVolumeMigrate) error {
	logCtx := m.logger.WithFields(log.Fields{"migration": migrate.Name, "spec": migrate.Spec, "status": migrate.Status})
	logCtx.Debug("Start a nonConvertibleVolumeMigrateInProgress")

	ctx := context.TODO()

	rcl := m.dataCopyManager.UseRclone(rcloneImageName, rcloneConfigMapName, migrate.Namespace, rcloneConfigMapKey)
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
						if err := rcl.WaitMigrateJobTaskDone(jobName, vol.Spec.PersistentVolumeClaimName, vol.Spec.PersistentVolumeClaimName+migratePvcSuffix, true, 0); err != nil {
							logCtx.WithError(err).Error("nonConvertibleVolumeMigrateInProgress: Job PVCToPVC failed")
							migrate.Status.Message = "nonConvertibleVolumeMigrateInProgress: Job PVCToRemotePVC failed"
							m.apiClient.Status().Update(ctx, migrate)
							return err
						}
					}

					err = m.updatedMigrateSrcLocalVolume(vol)
					if err != nil {
						log.WithError(err).Error("nonConvertibleVolumeMigrateInProgress updatedMigrateSrcLocalVolume failed")
						migrate.Status.Message = err.Error()
						m.apiClient.Status().Update(ctx, migrate)
						return err
					}
				}
			}

			migrate.Status.State = apisv1alpha1.OperationStateCompleted
			m.apiClient.Status().Update(ctx, migrate)
		}
	}

	if err := m.apiClient.Update(context.TODO(), lvg); err != nil {
		log.WithError(err).Error("VolumeGroupMigrateInProgress Reconcile : Failed to re-configure Volume")
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

func (m *manager) delMigratePod(ns string) error {
	var migratePod = &corev1.Pod{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: migrateDstPodName}, migratePod); err != nil {
		if !errors.IsNotFound(err) {
			m.logger.WithError(err).Error("delMigratePod: Failed to query pod")
		} else {
			m.logger.Info("delMigratePod: Not found the pod")
		}
		return err
	}
	return m.apiClient.Delete(context.TODO(), migratePod)
}

func (m *manager) makeMigratePod(lvs []apisv1alpha1.LocalVolume, ns string) (*corev1.Pod, error) {
	m.logger.Debug("makeMigratePod start")
	var podSpec corev1.PodSpec
	var volumeMounts []corev1.VolumeMount
	var volumes []corev1.Volume

	for _, lv := range lvs {
		volumeMount := corev1.VolumeMount{
			Name: migrateVolumePrefix + lv.Spec.PersistentVolumeClaimName, MountPath: migrateVolumeMountPathPrefix + lv.Spec.PersistentVolumeClaimName,
		}
		volumeMounts = append(volumeMounts, volumeMount)

		volume := corev1.Volume{
			Name: migrateVolumePrefix + lv.Spec.PersistentVolumeClaimName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: lv.Spec.PersistentVolumeClaimName + migratePvcSuffix,
				},
			},
		}
		volumes = append(volumes, volume)
	}

	hostVolumeMount := corev1.VolumeMount{
		Name: migrateVolumePrefix + "hostpath", MountPath: migrateVolumeMountPathPrefix,
	}
	volumeMounts = append(volumeMounts, hostVolumeMount)

	hostVolume := corev1.Volume{
		Name: migrateVolumePrefix + "hostpath",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: migrateVolumeMountPathPrefix,
			},
		},
	}
	volumes = append(volumes, hostVolume)

	podSpec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:         migrateDstPodContainerName,
				Image:        migrateDstPodImageName,
				VolumeMounts: volumeMounts,
			},
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: imagePullSecrets,
			},
		},
		// we decide later whether to use a PVC volume or host volumes for mons, so only populate
		// the base volumes at this point.
		Volumes:     volumes,
		HostNetwork: true,
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      migrateDstPodName,
			Namespace: ns,
		},
		Spec: podSpec,
	}

	pod.Spec.DNSPolicy = corev1.DNSClusterFirstWithHostNet

	return pod, nil
}

func (m *manager) makeMigratePVC(pvc *corev1.PersistentVolumeClaim, targetNodesNames []string) (*corev1.PersistentVolumeClaim, error) {
	migratePvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvc.Name + migratePvcSuffix,
			Namespace: pvc.Namespace,
			Labels:    pvc.Labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources:        pvc.Spec.Resources,
			StorageClassName: pvc.Spec.StorageClassName,
			Selector:         pvc.Spec.Selector,
		},
	}

	var annotations = make(map[string]string)

	if len(targetNodesNames) >= 1 {
		annotations[volumeSelectedNodeKey] = targetNodesNames[0]
		migratePvc.Annotations = annotations
	}

	return migratePvc, nil
}

func (m *manager) updatedMigrateSrcLocalVolume(vol apisv1alpha1.LocalVolume) error {
	dstPvc := &corev1.PersistentVolumeClaim{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: vol.Spec.PersistentVolumeClaimName + migratePvcSuffix, Namespace: vol.Spec.PersistentVolumeClaimNamespace}, dstPvc); err != nil {
		if !errors.IsNotFound(err) {
			m.logger.WithError(err).Error("updatedMigrateSrcLocalVolume Failed to query dstPvc")
		} else {
			m.logger.Info("updatedMigrateSrcLocalVolume Not found the dstPvc")
		}
		return err
	}

	dstVol := &apisv1alpha1.LocalVolume{}
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: dstPvc.Spec.VolumeName}, dstVol); err != nil {
		m.logger.WithError(err).Error("updatedMigrateSrcLocalVolume Failed to query volume")
		return err
	}
	vol.Spec.Config.Replicas = dstVol.Spec.Config.Replicas
	vol.Status.Replicas = dstVol.Status.Replicas
	if err := m.apiClient.Update(context.TODO(), &vol); err != nil {
		m.logger.WithFields(log.Fields{"volume": vol.Name}).WithError(err).Error("updatedMigrateSrcLocalVolume: Failed to Update LocalVolume")
		return err
	}

	return nil
}

func (m *manager) makeMigrateRcloneConfigmap(targetPodIp, ns string) error {
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
	typeData := "type = sftp" + "\n"
	hostData := "host = " + targetPodIp + "\n"
	passData := "pass = GFD8J3QqikmnkvOaRiHvAAB8bz6zFdaddg" + "\n"
	shellTypeData := "shell_type = unix" + "\n"
	userData := "user = root" + "\n"
	md5sumCommandData := "md5sum_command = md5sum" + "\n"
	sha1sumCommandData := "sha1sum_command = sha1sum" + "\n"

	data := map[string]string{
		rcloneConfigMapKey: remoteNameData + typeData + hostData + passData + shellTypeData + userData + md5sumCommandData + sha1sumCommandData,
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
