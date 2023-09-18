package autoresizer

import (
	"context"
	"encoding/json"
	"time"

	hwameistorv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	// "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
)

type ResizePolicyWorker struct {
	cli client.Client
	resizePolicy *hwameistorv1alpha1.ResizePolicy
	queue workqueue.RateLimitingInterface
}

func (w *ResizePolicyWorker) Select(pvc *corev1.PersistentVolumeClaim) (*Selection, error) {
	matched, err := matchPVC(w.resizePolicy, pvc)
	if err != nil {
		log.Errorf("resizepolicy %v match pvc %v:%v err: %v", w.resizePolicy.Name, pvc.Namespace, pvc.Name, err)
		return nil, err
	}
	if matched {
		return &Selection{
			Selected: true,
			Type: PVCSelector,
		}, nil
	}

	var ns *corev1.Namespace
	if err := w.cli.Get(context.TODO(), types.NamespacedName{Name: pvc.Namespace}, ns); err != nil {
		log.Errorf("get namespace err: ")
	}
	matched, err = matchNamespace(w.resizePolicy, ns)
	if err != nil {
		log.Errorf("resizepolicy %v match ns %v err: %v", w.resizePolicy.Name, ns.Name, err)
		return nil, err
	}
	if matched {
		return &Selection{
			Selected: true,
			Type: PVCSelector,
		}, nil
	}

	var sc * storagev1.StorageClass
	if err := w.cli.Get(context.TODO(), types.NamespacedName{Name: *pvc.Spec.StorageClassName}, sc); err != nil {
		log.Errorf("get storageclass err: ")
		return nil, err
	}
	matched, err = matchStorageClass(w.resizePolicy, sc)
	if err != nil {
		log.Errorf("resizepolicy %v match storageclass %v err: %v", w.resizePolicy.Name, sc.Name, err)
		return nil, err
	}
	if matched {
		return &Selection{
			Selected: true,
			Type: StorageClassSelector,
		}, nil
	}

	return &Selection{
		Selected: false,
	}, err
}

func NewResizePolicyWorker(cli client.Client, resizePolicy *hwameistorv1alpha1.ResizePolicy) *ResizePolicyWorker {
	return &ResizePolicyWorker{
		cli: cli,
		resizePolicy: resizePolicy,
		queue: workqueue.NewNamedRateLimitingQueue(
			workqueue.NewItemExponentialFailureRateLimiter(time.Second, 16*time.Second),
			resizePolicy.Name,
		),
	}
}

func (w *ResizePolicyWorker) pushPVC(pvc *corev1.PersistentVolumeClaim) error {
	bytes, err := json.Marshal(pvc)
	if err != nil {
		log.Errorf("json marshal err: %v", err)
		return err
	}
	w.addString(string(bytes))
	return nil
}

func (w *ResizePolicyWorker) pushPVCRateLimited(pvc *corev1.PersistentVolumeClaim) error {
	bytes, err := json.Marshal(pvc)
	if err != nil {
		log.Errorf("json marshal err: %v", err)
		return err
	}
	w.addStringRateLimited(string(bytes))
	return nil
}

func (w *ResizePolicyWorker) popPVC() (*corev1.PersistentVolumeClaim, bool) {
	pvcString, shutdown := w.getString()
	if pvcString == "" {
		return &corev1.PersistentVolumeClaim{}, true
	}
	pvc := &corev1.PersistentVolumeClaim{}
	_ = json.Unmarshal([]byte(pvcString), pvc)
	return pvc, shutdown
}

func (w *ResizePolicyWorker) addString(s string) {
	w.queue.Add(s)
}

func (w *ResizePolicyWorker) addStringRateLimited(s string) {
	w.queue.AddRateLimited(s)
}

func (w *ResizePolicyWorker) getString() (string, bool) {
	item, shutdown := w.queue.Get()
	if item == nil {
		return "", true
	}
	return item.(string), shutdown
}

func (w *ResizePolicyWorker) handlePVCSelectorMatched(pvc *corev1.PersistentVolumeClaim) error {
	log.Infof("pvc %v:%v matched resizepolicy %v", pvc.Namespace, pvc.Name, w.resizePolicy.Name)
	if pvc.Annotations[PVCResizePolicyAnnotationKey] == "" {
		pvc.Annotations[PVCResizePolicyAnnotationKey] = w.resizePolicy.Name
		if err := w.cli.Update(context.TODO(), pvc); err != nil {
			log.Errorf("update pvc err: %v", err)
			return err
		}
		return nil
	} else {
		if pvc.Annotations[PVCResizePolicyAnnotationKey] == w.resizePolicy.Name {
			return nil
		} else {
			var rpFormer hwameistorv1alpha1.ResizePolicy
			if err := w.cli.Get(context.TODO(), types.NamespacedName{Name: pvc.Annotations[PVCResizePolicyAnnotationKey]}, &rpFormer); err != nil {
				log.Errorf("get resizepolicy %v err: %v", pvc.Annotations[PVCResizePolicyAnnotationKey], err)
				return err
			}
			var namespace *corev1.Namespace
			if err := w.cli.Get(context.TODO(), types.NamespacedName{Name: pvc.Namespace}, namespace); err != nil {
				log.Errorf("get namespace %v err: %v", pvc.Namespace, err)
				return err
			}
			var sc *storagev1.StorageClass
			if err := w.cli.Get(context.TODO(), types.NamespacedName{Name: *pvc.Spec.StorageClassName}, sc); err != nil {
				log.Errorf("get storageclass %v err: %v", pvc.Spec.StorageClassName, err)
				return err
			}
			determinedResizePolicy, err := determineResizePolicyForPVC(pvc, namespace, sc, []hwameistorv1alpha1.ResizePolicy{rpFormer, *w.resizePolicy})
			if err != nil {
				log.Errorf("determine resizepolicy in %v and %v err: %v", rpFormer, w.resizePolicy, err)
				return err
			}
			log.Infof("determined resizepolicy %v for pvc %v:%v between %v and %v", determinedResizePolicy.Name, pvc.Namespace, pvc.Name, rpFormer, w.resizePolicy)
			if determinedResizePolicy.Name == w.resizePolicy.Name {
				pvc.Annotations[PVCResizePolicyAnnotationKey] = w.resizePolicy.Name
				if err = w.cli.Update(context.TODO(), pvc); err != nil {
					log.Errorf("update pvc err: %v", err)
					return err
				}
				return nil
			}
		}
	}

	return nil
}

func (w *ResizePolicyWorker) handleNamespaceSelectorMatched(ns *corev1.Namespace, pvc *corev1.PersistentVolumeClaim) error {
	log.Infof("namespace %v matched resizepolicy %v", ns.Name, w.resizePolicy.Name)
	if pvc.Annotations[PVCResizePolicyAnnotationKey] == "" {
		pvc.Annotations[PVCResizePolicyAnnotationKey] = w.resizePolicy.Name
		if err := w.cli.Update(context.TODO(), pvc); err != nil {
			log.Errorf("update pvc err: %v", err)
			return err
		}
		return nil
	} else {
		if pvc.Annotations[PVCResizePolicyAnnotationKey] == w.resizePolicy.Name {
			return nil
		} else {
			var rpFormer hwameistorv1alpha1.ResizePolicy
			if err := w.cli.Get(context.TODO(), types.NamespacedName{Name: pvc.Annotations[PVCResizePolicyAnnotationKey]}, &rpFormer); err != nil {
				log.Errorf("get resizepolicy %v err: %v", pvc.Annotations[PVCResizePolicyAnnotationKey], err)
				return err
			}

			matched, err := matchPVC(&rpFormer, pvc)
			if err != nil {
				log.Errorf("resizepolicy %v match pvc %v:%v err: %v", rpFormer.Name, pvc.Namespace, pvc.Name, err)
				return err
			}
			if matched {
				return nil
			}

			matched, err = matchNamespace(&rpFormer, ns)
			if err != nil {
				log.Errorf("resizepolicy %v match namespace %v err: %v", rpFormer.Name, ns.Name, err)
				return err
			}
			if matched {
				if rpFormer.CreationTimestamp.Before(&w.resizePolicy.CreationTimestamp) {
					pvc.Annotations[PVCResizePolicyAnnotationKey] = w.resizePolicy.Name
					if err = w.cli.Update(context.TODO(), pvc); err != nil {
						log.Errorf("update pvc err: %v", err)
						return err
					}
					return nil
				}
				return nil
			} else {
				pvc.Annotations[PVCResizePolicyAnnotationKey] = w.resizePolicy.Name
				if err = w.cli.Update(context.TODO(), pvc); err != nil {
					log.Errorf("update pvc err: %v", err)
					return err
				}
				return nil
			}
		}
	}
}

func (w *ResizePolicyWorker) handleStorageClassSelectorMatched(sc *storagev1.StorageClass, ns *corev1.Namespace, pvc *corev1.PersistentVolumeClaim) error {
	log.Infof("storageclass %v matched resizepolicy %v", sc.Name, w.resizePolicy.Name)
	if pvc.Annotations[PVCResizePolicyAnnotationKey] == "" {
		pvc.Annotations[PVCResizePolicyAnnotationKey] = w.resizePolicy.Name
		if err := w.cli.Update(context.TODO(), pvc); err != nil {
			log.Errorf("update pvc err: %v", err)
			return err
		}
		return nil
	} else {
		if pvc.Annotations[PVCResizePolicyAnnotationKey] == w.resizePolicy.Name {
			return nil
		} else {
			var rpFormer hwameistorv1alpha1.ResizePolicy
			if err := w.cli.Get(context.TODO(), types.NamespacedName{Name: pvc.Annotations[PVCResizePolicyAnnotationKey]}, &rpFormer); err != nil {
				log.Errorf("get resizepolicy %v err: %v", pvc.Annotations[PVCResizePolicyAnnotationKey], err)
				return err
			}

			matched, err := matchPVC(&rpFormer, pvc)
			if err != nil {
				log.Errorf("resizepolicy %v match pvc %v:%v err: %v", rpFormer.Name, pvc.Namespace, pvc.Name, err)
				return err
			}
			if matched {
				return nil
			}

			matched, err = matchNamespace(&rpFormer, ns)
			if err != nil {
				log.Errorf("resizepolicy %v match namespace %v:%v err: %v", rpFormer.Name, pvc.Namespace, pvc.Name, err)
				return err
			}
			if matched {
				return nil
			}

			matched, err = matchStorageClass(&rpFormer, sc)
			if err != nil {
				log.Errorf("resizepolicy %v match storageclass %v err: %v", rpFormer.Name, sc.Name, err)
				return err
			}
			if matched {
				if rpFormer.CreationTimestamp.Before(&w.resizePolicy.CreationTimestamp) {
					pvc.Annotations[PVCResizePolicyAnnotationKey] = w.resizePolicy.Name
					if err = w.cli.Update(context.TODO(), pvc); err != nil {
						log.Errorf("update pvc err: %v", err)
						return err
					}
					return nil
				}
				return nil
			} else {
				pvc.Annotations[PVCResizePolicyAnnotationKey] = w.resizePolicy.Name
				if err = w.cli.Update(context.TODO(), pvc); err != nil {
					log.Errorf("update pvc err: %v", err)
					return err
				}
				return nil
			}
		}
	}
}

func (w *ResizePolicyWorker) processPVC(pvc *corev1.PersistentVolumeClaim) error {
	matched, err := matchPVC(w.resizePolicy, pvc)
	if err != nil {
		log.Errorf("resizepolicy %v match pvc %v:%v err: %v", w.resizePolicy.Name, pvc.Namespace, pvc.Name, err)
		return err
	}
	if matched {
		if err := w.handlePVCSelectorMatched(pvc); err != nil {
			log.Errorf("handle pvc selector matched err: %v", err)
			return err
		}
		return nil
	} 
	
	ns := &corev1.Namespace{}
	if err := w.cli.Get(context.TODO(), types.NamespacedName{Name: pvc.Namespace}, ns); err != nil {
		log.Errorf("get namespace err: %v", err)
		return err
	}
	matched, err = matchNamespace(w.resizePolicy, ns)
	if err != nil {
		log.Errorf("resizepolicy %v match namespace %v err: %v", w.resizePolicy.Name, ns.Name, err)
		return err
	}
	if matched {
		if err := w.handleNamespaceSelectorMatched(ns, pvc); err != nil {
			log.Errorf("handle namespace selector matched err: %v", err)
			return err
		}
		return nil
	}

	sc := &storagev1.StorageClass{}
	if err := w.cli.Get(context.TODO(), types.NamespacedName{Name: *pvc.Spec.StorageClassName}, sc); err != nil {
		log.Errorf("get storageclass err: %v", err)
		return err
	}
	matched, err = matchStorageClass(w.resizePolicy, sc)
	if err != nil {
		log.Errorf("resizepolicy %v match storageclass %v err: %v", w.resizePolicy.Name, sc.Name, err)
		return err
	}

	if matched {
		if err := w.handleStorageClassSelectorMatched(sc, ns, pvc); err != nil {
			log.Errorf("handle storageclass selector matched err: %v", err)
			return err
		}
		return nil
	}
	
	if pvc.Annotations[PVCResizePolicyAnnotationKey] == w.resizePolicy.Name {
		delete(pvc.Annotations, PVCResizePolicyAnnotationKey)
		if err = w.cli.Update(context.TODO(), pvc); err != nil {
			log.Errorf("update pvc err: %v", err)
			return err
		}
	}

	return nil
}

func StartResizePolicyEventHandler(cli client.Client, runtimeCache runtimecache.Cache) {
	handlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			rp := obj.(*hwameistorv1alpha1.ResizePolicy)
			pvcList := &corev1.PersistentVolumeClaimList{}
			if err := cli.List(context.TODO(), pvcList); err != nil {
				log.Errorf("list pvc err: %v", err)
				return
			}
			rpWorker := NewResizePolicyWorker(cli, rp)
			for _, pvc := range pvcList.Items {
				rpWorker.pushPVC(&pvc)
			}
			for {
				if rpWorker.queue.Len() == 0 {
					break
				}
				pvc, shutdown := rpWorker.popPVC()
				if shutdown {
					log.Infof("queue shutdown, pvc: %v:%v", pvc.Namespace, pvc.Name)
					break
				}
				if err := rpWorker.processPVC(pvc); err != nil {
					log.Errorf("pvc: %v:%v, attempts: %v, err: %v, failed to process pvc, retry later", pvc.Namespace, pvc.Name, rpWorker.queue.NumRequeues(pvc), err)
					rpWorker.pushPVCRateLimited(pvc)
				} else {
					log.Infof("processed pvc: %v:%v", pvc.Namespace, pvc.Name)
					rpWorker.queue.Forget(pvc)
				}
				rpWorker.queue.Done(pvc)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			rp := newObj.(*hwameistorv1alpha1.ResizePolicy)
			pvcList := &corev1.PersistentVolumeClaimList{}
			if err := cli.List(context.TODO(), pvcList); err != nil {
				log.Errorf("list pvc err: %v", err)
				return
			}
			rpWorker := NewResizePolicyWorker(cli, rp)
			for _, pvc := range pvcList.Items {
				rpWorker.pushPVC(&pvc)
			}
			for {
				if rpWorker.queue.Len() == 0 {
					break
				}
				pvc, shutdown := rpWorker.popPVC()
				if shutdown {
					log.Infof("queue shutdown, pvc: %v:%v", pvc.Namespace, pvc.Name)
					break
				}
				if err := rpWorker.processPVC(pvc); err != nil {
					log.Errorf("pvc: %v:%v, attempts: %v, err: %v, failed to process pvc, retry later", pvc.Namespace, pvc.Name, rpWorker.queue.NumRequeues(pvc), err)
					rpWorker.pushPVCRateLimited(pvc)
				} else {
					log.Infof("processed pvc: %v:%v", pvc.Namespace, pvc.Name)
					rpWorker.queue.Forget(pvc)
				}
				rpWorker.queue.Done(pvc)
			}
		},
		DeleteFunc: func(obj interface{}) {
			rp := obj.(*hwameistorv1alpha1.ResizePolicy)
			pvcList := &corev1.PersistentVolumeClaimList{}
			if err := cli.List(context.TODO(), pvcList); err != nil {
				log.Errorf("list pvc err: %v", err)
				return
			}
			for _, pvc := range pvcList.Items {
				if pvc.Annotations[PVCResizePolicyAnnotationKey] == rp.Name {
					delete(pvc.Annotations, PVCResizePolicyAnnotationKey)
					if err := cli.Update(context.TODO(), &pvc); err != nil {
						log.Errorf("update pvc err: %v", err)
						continue
					}
				}
			}
		},
	}

	rpInformer, err := runtimeCache.GetInformer(context.TODO(), &hwameistorv1alpha1.ResizePolicy{})
	if err != nil {
		log.Errorf("get informer for resizepolicy err: %v", err)
		return
	}

	log.Infof("going to run resizepolicy informer")
	rpInformer.AddEventHandler(handlerFuncs)
	log.Infof("resizepolicy informer started")
}