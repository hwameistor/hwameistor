package volumegroup

import (
	"context"
	"fmt"
	"strings"
	"sync"

	apisv1alpha1 "github.com/hwameistor/local-storage/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/common"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	volumeGroupFinalizer = "hwameistor.io/localvolumegroup-protection"
)

type manager struct {
	apiClient      client.Client
	informersCache runtimecache.Cache

	logger *log.Entry

	lock                  sync.Mutex
	localVolumeGroupQueue *common.TaskQueue
	localVolumeQueue      *common.TaskQueue
	pvcQueue              *common.TaskQueue
	podQueue              *common.TaskQueue

	// lv -> volumegroup
	localVolumeToVolumeGroups map[string]string
	// pvc[namespace/name] -> volumegroup
	pvcToVolumeGroups map[string]string
	// pod[namespace/name] -> volumegroup
	podToVolumeGroups map[string]string
}

func namespacedName(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

// return: namespace, name
func parseNamespacedName(nn string) (string, string) {
	items := strings.Split(nn, "/")
	if len(items) == 0 {
		return "", ""
	}
	if len(items) == 1 {
		return "", items[0]
	}
	return items[0], items[1]
}

func NewManager(cli client.Client, informersCache runtimecache.Cache) apisv1alpha1.VolumeGroupManager {
	return &manager{
		apiClient:                 cli,
		informersCache:            informersCache,
		localVolumeGroupQueue:     common.NewTaskQueue("VolumeGroupQueue", 0),
		pvcQueue:                  common.NewTaskQueue("PVCQueue", 0),
		localVolumeQueue:          common.NewTaskQueue("LocalVolumeQueue", 0),
		podQueue:                  common.NewTaskQueue("PodQueue", 0),
		pvcToVolumeGroups:         make(map[string]string),
		localVolumeToVolumeGroups: make(map[string]string),
		podToVolumeGroups:         make(map[string]string),
		logger:                    log.WithField("Module", "LocalVolumeGroupManager"),
	}
}

func (m *manager) debug() {
	m.logger.WithFields(log.Fields{
		"pvcToVg": m.pvcToVolumeGroups,
		"lvToVg":  m.localVolumeToVolumeGroups,
		"podToVg": m.podToVolumeGroups,
	}).Debug(" === DUMP ===")
}

func (m *manager) Init(stopCh <-chan struct{}) {

	lvInformer, err := m.informersCache.GetInformer(context.TODO(), &apisv1alpha1.LocalVolume{})
	if err != nil {
		m.logger.WithError(err).Fatal("Failed to initiate informer for LocalVolume")
	}
	lvInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handleLocalVolumeEventAdd,
		DeleteFunc: m.handleLocalVolumeEventDelete,
		UpdateFunc: m.handleLocalVolumeEventUpdate,
	})

	pvcInformer, err := m.informersCache.GetInformer(context.TODO(), &corev1.PersistentVolumeClaim{})
	if err != nil {
		m.logger.WithError(err).Fatal("Failed to initiate informer for PersistentVolumeClaim")
	}
	pvcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handlePVCEventAdd,
		DeleteFunc: m.handlePVCEventDelete,
	})

	podInformer, err := m.informersCache.GetInformer(context.TODO(), &corev1.Pod{})
	if err != nil {
		m.logger.WithError(err).Fatal("Failed to initiate informer for Pod")
	}

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    m.handlePodEventAdd,
		DeleteFunc: m.handlePodEventDelete,
	})

	go m.startLocalVolumeGroupWorker(stopCh)
	go m.startLocalVolumeWorker(stopCh)
	go m.startPVCWorker(stopCh)
	go m.startPodWorker(stopCh)

}

func (m *manager) GetLocalVolumeGroupByName(lvgName string) (*apisv1alpha1.LocalVolumeGroup, error) {
	lvg := &apisv1alpha1.LocalVolumeGroup{}
	err := m.apiClient.Get(
		context.TODO(),
		types.NamespacedName{Name: lvgName},
		lvg)
	return lvg, err
}

func (m *manager) GetLocalVolumeGroupByLocalVolume(lvName string) (*apisv1alpha1.LocalVolumeGroup, error) {
	lvg := &apisv1alpha1.LocalVolumeGroup{}
	err := m.apiClient.Get(
		context.TODO(),
		types.NamespacedName{Name: m.localVolumeToVolumeGroups[lvName]},
		lvg)
	return lvg, err
}

func (m *manager) GetLocalVolumeGroupByPVC(pvcNamespace string, pvcName string) (*apisv1alpha1.LocalVolumeGroup, error) {
	lvg := apisv1alpha1.LocalVolumeGroup{}
	lvgName := m.pvcToVolumeGroups[namespacedName(pvcNamespace, pvcName)]
	if err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: lvgName}, &lvg); err != nil {
		return nil, err
	}
	return &lvg, nil
}

func (m *manager) handleLocalVolumeEventAdd(obj interface{}) {
	instance := obj.(*apisv1alpha1.LocalVolume)

	if err := m.addLocalVolume(instance); err != nil {
		m.localVolumeQueue.Add(instance.Name)
	}
}

func (m *manager) handleLocalVolumeEventDelete(obj interface{}) {
	instance := obj.(*apisv1alpha1.LocalVolume)

	if err := m.deleteLocalVolume(instance.Name); err != nil {
		m.localVolumeQueue.Add(instance.Name)
	}
}

func (m *manager) handleLocalVolumeEventUpdate(oldObj, newObj interface{}) {
	instance := newObj.(*apisv1alpha1.LocalVolume)
	if err := m.addLocalVolume(instance); err != nil {
		m.localVolumeQueue.Add(instance.Name)
	}
}

func (m *manager) handlePVCEventAdd(obj interface{}) {
	instance := obj.(*corev1.PersistentVolumeClaim)
	if !m.isHwameiStorPVC(instance) {
		return
	}
	if err := m.addPVC(instance); err != nil {
		m.pvcQueue.Add(namespacedName(instance.Namespace, instance.Name))
	}
}

func (m *manager) handlePVCEventDelete(obj interface{}) {
	instance := obj.(*corev1.PersistentVolumeClaim)
	if !m.isHwameiStorPVC(instance) {
		return
	}
	if err := m.deletePVC(instance.Namespace, instance.Name); err != nil {
		m.pvcQueue.Add(namespacedName(instance.Namespace, instance.Name))
	}
}

func (m *manager) handlePodEventAdd(obj interface{}) {
	instance := obj.(*corev1.Pod)
	if !m.isHwameiStorPod(instance) {
		return
	}
	if err := m.addPod(instance); err != nil {
		m.podQueue.Add(namespacedName(instance.Namespace, instance.Name))
	}
}

func (m *manager) handlePodEventDelete(obj interface{}) {
	instance := obj.(*corev1.Pod)
	if !m.isHwameiStorPod(instance) {
		return
	}
	if err := m.deletePod(instance.Namespace, instance.Name); err != nil {
		m.podQueue.Add(namespacedName(instance.Namespace, instance.Name))
	}
}

func (m *manager) isHwameiStorPVC(pvc *corev1.PersistentVolumeClaim) bool {
	if pvc.Spec.StorageClassName == nil {
		return false
	}
	sc := &storagev1.StorageClass{}
	err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: *pvc.Spec.StorageClassName}, sc)
	if err != nil {
		m.logger.WithFields(log.Fields{"namespace": pvc.Namespace, "pvc": pvc.Name, "storageclass": *pvc.Spec.StorageClassName}).WithError(err).Error("Failed to fetch storageclass")
		return false
	}
	return sc.Provisioner == apisv1alpha1.CSIDriverName
}

func (m *manager) isHwameiStorPod(pod *corev1.Pod) bool {
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim == nil {
			continue
		}
		pvc := &corev1.PersistentVolumeClaim{}
		err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: pod.Namespace, Name: vol.PersistentVolumeClaim.ClaimName}, pvc)
		if err != nil {
			m.logger.WithFields(log.Fields{"namespace": pod.Namespace, "pvc": vol.PersistentVolumeClaim.ClaimName}).WithError(err).Error("Failed to fetch PVC")
			continue
		}
		if m.isHwameiStorPVC(pvc) {
			return true
		}
	}
	return false
}

func (m *manager) startLocalVolumeGroupWorker(stopCh <-chan struct{}) {
	m.logger.Debug("LocalVolumeGroup Worker is working now")
	go func() {
		for {
			task, shutdown := m.localVolumeGroupQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the LocalVolumeGroup task worker")
				break
			}
			if err := m.processLocalVolumeGroup(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process LocalVolumeGroup task, retry later ...")
				m.localVolumeGroupQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a LocalVolumeGroup task.")
				m.localVolumeGroupQueue.Forget(task)
			}
			m.localVolumeGroupQueue.Done(task)
		}
	}()

	<-stopCh
	m.localVolumeGroupQueue.Shutdown()
}

func (m *manager) startLocalVolumeWorker(stopCh <-chan struct{}) {
	m.logger.Debug("LocalVolume Worker is working now")
	go func() {
		for {
			task, shutdown := m.localVolumeQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the LocalVolume task worker")
				break
			}
			if err := m.processLocalVolume(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process LocalVolume task, retry later ...")
				m.localVolumeQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a LocalVolume task.")
				m.localVolumeQueue.Forget(task)
			}
			m.localVolumeQueue.Done(task)
		}
	}()

	<-stopCh
	m.localVolumeQueue.Shutdown()
}

func (m *manager) startPVCWorker(stopCh <-chan struct{}) {
	m.logger.Debug("PVC Worker is working now")
	go func() {
		for {
			task, shutdown := m.pvcQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the PVC task worker")
				break
			}
			if err := m.processPVC(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process PVC task, retry later ...")
				m.pvcQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a PVC task.")
				m.pvcQueue.Forget(task)
			}
			m.pvcQueue.Done(task)
		}
	}()

	<-stopCh
	m.pvcQueue.Shutdown()
}

func (m *manager) startPodWorker(stopCh <-chan struct{}) {
	m.logger.Debug("Pod Worker is working now")
	go func() {
		for {
			task, shutdown := m.podQueue.Get()
			if shutdown {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Stop the Pod task worker")
				break
			}
			if err := m.processPod(task); err != nil {
				m.logger.WithFields(log.Fields{"task": task, "error": err.Error()}).Error("Failed to process Pod task, retry later ...")
				m.podQueue.AddRateLimited(task)
			} else {
				m.logger.WithFields(log.Fields{"task": task}).Debug("Completed a Pod task.")
				m.podQueue.Forget(task)
			}
			m.podQueue.Done(task)
		}
	}()

	<-stopCh
	m.podQueue.Shutdown()
}

func (m *manager) ReconcileVolumeGroup(lvg *apisv1alpha1.LocalVolumeGroup) {
	m.logger.WithField("lvg", lvg.Name).Debug("Reconciling a VolumeGroup")

	if lvg.DeletionTimestamp != nil {
		// lvg is in Deleting state, waiting for finalizer to be clean up
		if err := m.releaseLocalVolumeGroup(lvg); err != nil {
			m.localVolumeGroupQueue.Add(lvg.Name)
		}
	} else if len(lvg.Spec.Volumes) == 0 {
		// no pvc/lv associated with LVG, should delete it
		if err := m.deleteLocalVolumeGroup(lvg); err != nil {
			m.localVolumeGroupQueue.Add(lvg.Name)
		}
	} else {
		// add or update LVG
		if err := m.addLocalVolumeGroup(lvg); err != nil {
			m.localVolumeGroupQueue.Add(lvg.Name)
		}
	}
}

func (m *manager) processLocalVolumeGroup(lvgName string) error {

	lvg, err := m.GetLocalVolumeGroupByName(lvgName)
	if err != nil {
		if errors.IsNotFound(err) {
			m.cleanCacheForLocalVolumeGroup(lvgName)
			return nil
		}
		return err
	}

	if lvg.DeletionTimestamp != nil {
		return m.releaseLocalVolumeGroup(lvg)
	}
	if len(lvg.Spec.Volumes) == 0 {
		return m.deleteLocalVolumeGroup(lvg)
	}
	return m.addLocalVolumeGroup(lvg)
}

func (m *manager) addLocalVolumeGroup(lvg *apisv1alpha1.LocalVolumeGroup) error {
	defer m.debug()

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			m.lock.Lock()
			for _, vol := range lvg.Spec.Volumes {
				if len(vol.PersistentVolumeClaimName) > 0 {
					m.pvcToVolumeGroups[namespacedName(lvg.Spec.Namespace, vol.PersistentVolumeClaimName)] = lvg.Name
				}
				if len(vol.LocalVolumeName) > 0 {
					m.localVolumeToVolumeGroups[vol.LocalVolumeName] = lvg.Name
				}
			}
			if len(lvg.Spec.Pods) > 0 {
				for _, podName := range lvg.Spec.Pods {
					m.podToVolumeGroups[namespacedName(lvg.Spec.Namespace, podName)] = lvg.Name
				}
			}
			m.lock.Unlock()
			return nil
		}
	}
	nLvg := lvg.DeepCopy()
	nLvg.Finalizers = append(nLvg.Finalizers, volumeGroupFinalizer)
	patch := client.MergeFrom(lvg)
	return m.apiClient.Patch(context.TODO(), nLvg, patch)
}

func (m *manager) deleteLocalVolumeGroup(lvg *apisv1alpha1.LocalVolumeGroup) error {
	defer m.debug()

	if len(lvg.Spec.Volumes) > 0 {
		return fmt.Errorf("volumes not empty")
	}

	m.cleanCacheForLocalVolumeGroup(lvg.Name)

	return m.apiClient.Delete(context.TODO(), lvg)
}

func (m *manager) releaseLocalVolumeGroup(lvg *apisv1alpha1.LocalVolumeGroup) error {
	defer m.debug()

	m.cleanCacheForLocalVolumeGroup(lvg.Name)

	for _, fnlr := range lvg.Finalizers {
		if fnlr == volumeGroupFinalizer {
			nLvg := lvg.DeepCopy()
			nLvg.Finalizers = []string{}
			patch := client.MergeFrom(lvg)
			return m.apiClient.Patch(context.TODO(), nLvg, patch)
		}
	}
	return fmt.Errorf("not found finalizer")
}

func (m *manager) cleanCacheForLocalVolumeGroup(name string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	newLvToLvgMap := map[string]string{}
	for key, value := range m.localVolumeToVolumeGroups {
		if value != name {
			newLvToLvgMap[key] = value
		}
	}
	m.localVolumeToVolumeGroups = newLvToLvgMap

	newPvcToLvgMap := map[string]string{}
	for key, value := range m.pvcToVolumeGroups {
		if value != name {
			newPvcToLvgMap[key] = value
		}
	}
	m.pvcToVolumeGroups = newPvcToLvgMap

	newPodToLvgMap := map[string]string{}
	for key, value := range m.podToVolumeGroups {
		if value != name {
			newPodToLvgMap[key] = value
		}
	}
	m.podToVolumeGroups = newPodToLvgMap
}

func (m *manager) cleanCacheForLocalVolume(name string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.localVolumeToVolumeGroups, name)
}

func (m *manager) cleanCacheForPVC(namespace string, name string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.pvcToVolumeGroups, namespacedName(namespace, name))
}

func (m *manager) cleanCacheForPod(namespace string, name string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.podToVolumeGroups, namespacedName(namespace, name))
}

func (m *manager) processLocalVolume(lvName string) error {
	lv := &apisv1alpha1.LocalVolume{}
	err := m.apiClient.Get(context.TODO(), types.NamespacedName{Name: lvName}, lv)
	if err != nil {
		if errors.IsNotFound(err) {
			return m.deleteLocalVolume(lv.Name)
		}
		return err
	}

	if lv.Status.State == apisv1alpha1.VolumeStateDeleted {
		return m.deleteLocalVolume(lv.Name)
	}

	return m.addLocalVolume(lv)
}

func (m *manager) addLocalVolume(lv *apisv1alpha1.LocalVolume) error {
	defer m.debug()

	m.lock.Lock()
	defer m.lock.Unlock()

	lvg, err := m.GetLocalVolumeGroupByName(lv.Spec.VolumeGroup)
	if err != nil {
		return err
	}

	for i, vol := range lvg.Spec.Volumes {
		if vol.LocalVolumeName == lv.Name {
			if len(lvg.Spec.Accessibility.Nodes) == 0 && len(lv.Spec.Accessibility.Nodes) > 0 {
				lv.Spec.Accessibility.DeepCopyInto(&lvg.Spec.Accessibility)
				return m.apiClient.Update(context.TODO(), lvg, &client.UpdateOptions{})
			}
			return nil
		}
		if vol.PersistentVolumeClaimName == lv.Spec.PersistentVolumeClaimName && lvg.Spec.Namespace == lv.Spec.PersistentVolumeClaimNamespace {
			// localvolume is just created to serve PVC
			lvg.Spec.Volumes[i].LocalVolumeName = lv.Name
			m.localVolumeToVolumeGroups[lv.Name] = lv.Spec.VolumeGroup
			return m.apiClient.Update(context.TODO(), lvg, &client.UpdateOptions{})
		}
	}

	m.logger.WithFields(log.Fields{"localvolume": lv.Name, "localvolumegroup": lvg.Name}).Error("Not found the matched PVC")
	return fmt.Errorf("not found matched PVC")
}

func (m *manager) deleteLocalVolume(lvName string) error {
	defer m.debug()

	lvgName, exists := m.localVolumeToVolumeGroups[lvName]
	if !exists {
		return nil
	}
	lvg, err := m.GetLocalVolumeGroupByName(lvgName)
	if err != nil {
		return err
	}

	modified := false
	associatedVolumes := []apisv1alpha1.VolumeInfo{}
	for i := range lvg.Spec.Volumes {
		if lvg.Spec.Volumes[i].LocalVolumeName != lvName {
			associatedVolumes = append(associatedVolumes, lvg.Spec.Volumes[i])
			continue
		}
		if len(lvg.Spec.Volumes[i].PersistentVolumeClaimName) > 0 {
			associatedVolumes = append(associatedVolumes,
				apisv1alpha1.VolumeInfo{
					PersistentVolumeClaimName: lvg.Spec.Volumes[i].PersistentVolumeClaimName,
				})
		}
		modified = true
	}
	if modified {
		lvg.Spec.Volumes = associatedVolumes
		if err := m.apiClient.Update(context.TODO(), lvg, &client.UpdateOptions{}); err != nil {
			return err
		}
	}

	m.cleanCacheForLocalVolume(lvName)

	return nil
}

func (m *manager) processPVC(nn string) error {
	namespace, name := parseNamespacedName(nn)
	if len(namespace) == 0 || len(name) == 0 {
		return fmt.Errorf("invalid PVC namespaced name")
	}
	instance := &corev1.PersistentVolumeClaim{}
	err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return m.deletePVC(namespace, name)
		}
		return err
	}

	if instance.DeletionTimestamp != nil {
		return m.deletePVC(namespace, name)
	}
	return m.addPVC(instance)
}

func (m *manager) addPVC(pvc *corev1.PersistentVolumeClaim) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	return nil
}

func (m *manager) deletePVC(namespace string, name string) error {
	defer m.debug()

	lvgName, exists := m.pvcToVolumeGroups[namespacedName(namespace, name)]
	if !exists {
		return nil
	}
	lvg, err := m.GetLocalVolumeGroupByName(lvgName)
	if err != nil {
		if errors.IsNotFound(err) {
			m.cleanCacheForPVC(namespace, name)
			return nil
		}
		return err
	}

	modified := false
	associatedVolumes := []apisv1alpha1.VolumeInfo{}
	for i := range lvg.Spec.Volumes {
		if lvg.Spec.Volumes[i].PersistentVolumeClaimName != name || lvg.Spec.Namespace != namespace {
			associatedVolumes = append(associatedVolumes, lvg.Spec.Volumes[i])
			continue
		}
		if len(lvg.Spec.Volumes[i].LocalVolumeName) > 0 {
			associatedVolumes = append(associatedVolumes, apisv1alpha1.VolumeInfo{LocalVolumeName: lvg.Spec.Volumes[i].LocalVolumeName})
		}
		modified = true
	}
	if modified {
		lvg.Spec.Volumes = associatedVolumes
		if err := m.apiClient.Update(context.TODO(), lvg, &client.UpdateOptions{}); err != nil {
			return err
		}
	}

	m.cleanCacheForPVC(namespace, name)

	return nil
}

func (m *manager) processPod(nn string) error {
	namespace, name := parseNamespacedName(nn)
	if len(namespace) == 0 || len(name) == 0 {
		return fmt.Errorf("invalid Pod namespaced name")
	}
	instance := &corev1.Pod{}
	err := m.apiClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return m.deletePod(namespace, name)
		}
		return err
	}

	if instance.DeletionTimestamp != nil {
		return m.deletePod(namespace, name)
	}
	return m.addPod(instance)
}

func (m *manager) addPod(pod *corev1.Pod) error {
	// no action

	return nil
}

func (m *manager) deletePod(namespace string, name string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	podKey := namespacedName(namespace, name)
	lvgName, exists := m.podToVolumeGroups[podKey]
	if !exists {
		return nil
	}

	lvg, err := m.GetLocalVolumeGroupByName(lvgName)
	if err != nil {
		if errors.IsNotFound(err) {
			delete(m.podToVolumeGroups, podKey)
			return nil
		}
		return err
	}
	if lvg.Spec.Namespace != namespace {
		delete(m.podToVolumeGroups, podKey)
		return nil
	}

	newPods := []string{}
	for _, podName := range lvg.Spec.Pods {
		if podName != name {
			newPods = append(newPods, podName)
		}
	}
	if len(lvg.Spec.Pods) > len(newPods) {
		lvg.Spec.Pods = newPods
		if err := m.apiClient.Update(context.TODO(), lvg, &client.UpdateOptions{}); err != nil {
			return err
		}
	}

	delete(m.podToVolumeGroups, podKey)
	return nil
}
