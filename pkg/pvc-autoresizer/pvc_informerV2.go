package autoresizer

import (
	"context"

	hwameistorv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type pvcWorker struct {
	Namespace *corev1.Namespace
	StorageClass *storagev1.StorageClass
	cli client.Client
}

func newPVCWorker(cli client.Client) *pvcWorker {
	return &pvcWorker{
		cli: cli,
	}
}

func (w *pvcWorker) conformPVCAgainstResizePolicy(pvc *corev1.PersistentVolumeClaim, rp *hwameistorv1alpha1.ResizePolicy) (bool, error) {
	if rp.Spec.PVCSelector != nil {
		matched, err := matchPVC(rp, pvc)
		if err != nil {
			log.Errorf("match pvc err: %v", err)
			return false, err
		}
		if !matched {
			log.Debugf("pvc %v:%v unmatched resizepolicy %v", pvc.Namespace, pvc.Name, rp.Name)
			return false, nil
		}
	}
	if rp.Spec.NamespaceSelector != nil {
		if w.Namespace == nil {
			ns := &corev1.Namespace{}
			if err := w.cli.Get(context.TODO(), types.NamespacedName{Name: pvc.Namespace}, ns); err != nil {
				log.Errorf("get namespace err: %v", err)
				return false, err
			}
			w.Namespace = ns
		}
		matched, err := matchNamespace(rp, w.Namespace)
		if err != nil {
			log.Errorf("match namespace err: %v", err)
			return false, err
		}
		if !matched {
			log.Debugf("namespace %v unmatched resizepolicy %v", w.Namespace.Name, rp.Name)
			return false, nil
		}
	}
	if rp.Spec.StorageClassSelector != nil {
		if w.StorageClass == nil {
			if pvc.Spec.StorageClassName == nil {
				log.Infof("pvc %v:%v has no storageclass, return now", pvc.Namespace, pvc.Name)
				return false, nil
			}
			if *pvc.Spec.StorageClassName == "" {
				log.Infof("storageclassname of pvc %v:%v is empty string, return now", pvc.Namespace, pvc.Name)
				return false, nil
			}
			sc := &storagev1.StorageClass{}
			if err := w.cli.Get(context.TODO(), types.NamespacedName{Name: *pvc.Spec.StorageClassName}, sc); err != nil {
				log.Errorf("get storageclass err: %v", err)
				return false, err
			}
			w.StorageClass = sc
		}
		matched, err := matchStorageClass(rp, w.StorageClass)
		if err != nil {
			log.Errorf("match storageclass err: %v", err)
			return false, err
		}
		if !matched {
			log.Debugf("storageclass %v unmatched resizepolicy %v", w.StorageClass.Name, rp.Name)
			return false, nil
		}
		log.Debugf("storageclass %v matched resizepolicy %v", w.StorageClass.Name, rp.Name)
	}
	return true, nil
}

func (a *PVCAttacher) processV2(pvc *corev1.PersistentVolumeClaim) error {
	log.Infof("to process pvc %v:%v", pvc.Namespace, pvc.Name)
	worker := newPVCWorker(a.cli)

	pvcSelectorChain.Lock.Lock()
	defer pvcSelectorChain.Lock.Unlock()
	for i := 0; i < len(pvcSelectorChain.Chain); i = i + 1 {
		rp := pvcSelectorChain.Chain[i]
		log.Debugf("index: %v, resizepolicy: %v", i, rp.Name)
		conformed, err := worker.conformPVCAgainstResizePolicy(pvc, rp)
		if err != nil {
			log.Errorf("conform pvc against resizepolicy err: %v", err)
			return err
		}
		if conformed {
			newPVC := pvc.DeepCopy()
			newPVC.Annotations[PVCResizePolicyAnnotationKey] = rp.Name
			if err := a.cli.Patch(context.TODO(), newPVC, client.MergeFrom(pvc)); err != nil {
				log.Errorf("patch pvc err: %v", err)
				return err
			}
			return nil
		}
	}

	namespaceSelectorChain.Lock.Lock()
	defer namespaceSelectorChain.Lock.Unlock()
	for i := 0; i < len(namespaceSelectorChain.Chain); i = i + 1 {
		rp := namespaceSelectorChain.Chain[i]
		conformed, err := worker.conformPVCAgainstResizePolicy(pvc, rp)
		if err != nil {
			log.Errorf("conform pvc against resizepolicy err: %v", err)
			return err
		}
		if conformed {
			newPVC := pvc.DeepCopy()
			newPVC.Annotations[PVCResizePolicyAnnotationKey] = rp.Name
			if err := a.cli.Patch(context.TODO(), newPVC, client.MergeFrom(pvc)); err != nil {
				log.Errorf("patch pvc err: %v", err)
				return err
			}
			return nil
		}
	}

	storageClassSelectorChain.Lock.Lock()
	defer storageClassSelectorChain.Lock.Unlock()
	for i := 0; i < len(storageClassSelectorChain.Chain); i = i + 1 {
		rp := storageClassSelectorChain.Chain[i]
		conformed, err := worker.conformPVCAgainstResizePolicy(pvc, rp)
		if err != nil {
			log.Errorf("conform pvc against resizepolicy err: %v", err)
			return err
		}
		if conformed {
			newPVC := pvc.DeepCopy()
			newPVC.Annotations[PVCResizePolicyAnnotationKey] = rp.Name
			if err := a.cli.Patch(context.TODO(), newPVC, client.MergeFrom(pvc)); err != nil {
				log.Errorf("patch pvc err: %v", err)
				return err
			}
			return nil
		}
	}

	clusterResizePolicyChain.Lock.Lock()
	defer clusterResizePolicyChain.Lock.Unlock()
	for i := 0; i < len(clusterResizePolicyChain.Chain); i = i + 1 {
		rp := clusterResizePolicyChain.Chain[i]
		conformed, err := worker.conformPVCAgainstResizePolicy(pvc, rp)
		if err != nil {
			log.Errorf("conform pvc against resizepolicy err: %v", err)
			return err
		}
		if conformed {
			newPVC := pvc.DeepCopy()
			newPVC.Annotations[PVCResizePolicyAnnotationKey] = rp.Name
			if err := a.cli.Patch(context.TODO(), newPVC, client.MergeFrom(pvc)); err != nil {
				log.Errorf("patch pvc err: %v", err)
				return err
			}
			return nil
		}
	}

	return nil
}