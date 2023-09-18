package autoresizer

import (
	hwameistorv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type SelectionType string

const (
	PVCSelector SelectionType = "pvcSelector"
	NamespaceSelector SelectionType = "namespaceSelector"
	StorageClassSelector SelectionType = "storageClassSelector"
)

type Selection struct {
	Selected bool
	Type SelectionType
}

func matchStorageClass(rp *hwameistorv1alpha1.ResizePolicy, sc *storagev1.StorageClass) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(rp.Spec.StorageClassSelector)
	if err != nil {
		log.Errorf("convert LabelSelector to Selector err: %v", err)
		return false, err
	}
	return selector.Matches(labels.Set(sc.Labels)), nil
}

func matchNamespace(rp *hwameistorv1alpha1.ResizePolicy, namespace *corev1.Namespace) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(rp.Spec.NamespaceSelector)
	if err != nil {
		log.Errorf("convert LabelSelector to Selector err: %v", err)
		return false, err
	}
	return selector.Matches(labels.Set(namespace.Labels)), nil
}

func matchPVC(rp *hwameistorv1alpha1.ResizePolicy, pvc *corev1.PersistentVolumeClaim) (bool, error) {
	selector, err := metav1.LabelSelectorAsSelector(rp.Spec.PVCSelector)
	if err != nil {
		log.Errorf("convert LabelSelector to Selector err: %v", err)
		return false, err
	}
	return selector.Matches(labels.Set(pvc.Labels)), nil
}

func determineResizePolicyForPVC(pvc *corev1.PersistentVolumeClaim, namespace *corev1.Namespace, sc *storagev1.StorageClass, resizePolicies []hwameistorv1alpha1.ResizePolicy) (*hwameistorv1alpha1.ResizePolicy, error) {
	var pvcSelector *hwameistorv1alpha1.ResizePolicy
	var namespaceSelector *hwameistorv1alpha1.ResizePolicy
	var storageClassSelector *hwameistorv1alpha1.ResizePolicy
	for _, v := range resizePolicies {
		resizePolicy := v
		matched, err := matchPVC(&resizePolicy, pvc)
		if err != nil {
			log.Errorf("match pvc %v:%v err: %v", pvc.Namespace, pvc.Name, err)
			return nil, err
		}
		if matched {
			log.Infof("pvc %v:%v matched resizepolicy: %v", pvc.Namespace, pvc.Name, resizePolicy.Name)
			pvcSelector = comparePriorityBetweenSameSelector(pvcSelector, &resizePolicy)
			log.Debugf("chosen: %v", pvcSelector.Name)
			continue
		}
		matched, err = matchNamespace(&resizePolicy, namespace)
		if err != nil {
			log.Errorf("match namespace %v err: %v", namespace.Name, err)
			return nil, err
		}
		if matched {
			log.Infof("namespace %v matched resizepolicy: %v", namespace.Name, resizePolicy.Name)
			namespaceSelector = comparePriorityBetweenSameSelector(namespaceSelector, &resizePolicy)
			continue
		}
		matched, err = matchStorageClass(&resizePolicy, sc)
		if err != nil {
			log.Errorf("match storageclass %v err: %v", sc.Name, err)
			return nil, err
		}
		if matched {
			log.Infof("sc %v matched resizepolicy: %v", sc.Name, resizePolicy.Name)
			storageClassSelector = comparePriorityBetweenSameSelector(storageClassSelector, &resizePolicy)
		}
	}
	if pvcSelector != nil {
		log.Infof("to return pvcSelector: %v", pvcSelector.Name)
		return pvcSelector, nil
	}
	if namespaceSelector != nil {
		log.Infof("to return namespaceSelector: %v", namespaceSelector.Name)
		return namespaceSelector, nil
	}
	if storageClassSelector != nil {
		return storageClassSelector, nil
	}

	return nil, nil
}

func comparePriorityBetweenSameSelector(a,b *hwameistorv1alpha1.ResizePolicy) *hwameistorv1alpha1.ResizePolicy {
	if (a == nil) || a.CreationTimestamp.Before(&b.CreationTimestamp) {
		return b
	}
	return a
}