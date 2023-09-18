package autoresizer

import (
	"context"
	"reflect"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	// "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PVCAttacher struct {
	cli client.Client
	queue workqueue.RateLimitingInterface
}

func NewPVCAttacher(cli client.Client, q workqueue.RateLimitingInterface) *PVCAttacher {
	return &PVCAttacher{
		cli: cli,
		queue: q,
	}
}

func (a *PVCAttacher) Start (stopCh <-chan struct{}) {
	log.Infof("pvc-attacher is working now")
	go func() {
		for {
			pvc, shutdown := a.getPVC()
			if shutdown {
				log.Infof("queue shutdown, pvc: %v:%v", pvc.Namespace, pvc.Name)
				break
			}
			if err := a.processV2(pvc); err != nil {
			// if err := a.process(pvc); err != nil {
				log.Errorf("pvc: %v:%v, attempts: %v, err: %v, failed to process pvc, retry later", pvc.Namespace, pvc.Name, a.queue.NumRequeues(pvc), err)
				a.queue.AddRateLimited(pvc)
			} else {
				log.Infof("processed pvc: %v:%v", pvc.Namespace, pvc.Name)
				a.queue.Forget(pvc)
			}
			a.queue.Done(pvc)
		}
	}()

	<-stopCh
	a.queue.ShutDown()
}

func (a *PVCAttacher) getPVC() (*corev1.PersistentVolumeClaim, bool) {
	item, shutdown := a.queue.Get()
	if item == nil {
		return &corev1.PersistentVolumeClaim{}, true
	}
	return item.(*corev1.PersistentVolumeClaim), shutdown
}

func (a *PVCAttacher) process(pvc *corev1.PersistentVolumeClaim) error {
    sc := &storagev1.StorageClass{}
	if err := a.cli.Get(context.TODO(), types.NamespacedName{Name: *pvc.Spec.StorageClassName}, sc); err != nil {
		log.Errorf("get storageclass %v of pvc %v:%v err: %v", pvc.Spec.StorageClassName, pvc.Namespace, pvc.Name, err)
		return err
	}
	namespace := &corev1.Namespace{}
	if err := a.cli.Get(context.TODO(), types.NamespacedName{Name: pvc.Namespace}, namespace); err != nil {
		log.Errorf("get namespace %v of pvc %v err: %v", pvc.Namespace, pvc.Name, err)
		return err
	}
	resizePolicies, err := ListResizePolicies(a.cli)
	if err != nil {
		log.Errorf("list resizepolicy err: %v", err)
		return err
	}
	resizePolicy, err := determineResizePolicyForPVC(pvc, namespace, sc, resizePolicies)
	if err != nil {
		log.Errorf("determine resizepolicy for pvc %v:%v err: %v", pvc.Namespace, pvc.Name, err)
		return err
	}
	if resizePolicy != nil {
		log.Infof("determined resizepolicy:%v for pvc: %v:%v", resizePolicy.Name, pvc.Namespace, pvc.Name)
		newPVC := pvc.DeepCopy()
		newPVC.Annotations[PVCResizePolicyAnnotationKey] = resizePolicy.Name
		if err := a.cli.Patch(context.TODO(), newPVC, client.MergeFrom(pvc)); err != nil {
			log.Errorf("patch pvc err: %v", err)
			return err
		}
	}

	return nil
}

func (a *PVCAttacher) StartPVCInformer(cli client.Client, ctx context.Context) {
	handlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pvc := obj.(*corev1.PersistentVolumeClaim)
			log.Infof("pvc %v:%v added", pvc.Namespace, pvc.Name)
			a.queue.Add(pvc)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pvcOld := oldObj.(*corev1.PersistentVolumeClaim)
			pvcNew := newObj.(*corev1.PersistentVolumeClaim)
			log.Infof("pvc %v:%v updated", pvcNew.Namespace, pvcNew.Name)
			if (pvcNew.Annotations[PVCResizePolicyAnnotationKey] != pvcOld.Annotations[PVCResizePolicyAnnotationKey]) || 
			!reflect.DeepEqual(pvcNew.Labels, pvcOld.Labels) {
				a.queue.Add(pvcNew)
			}
		},
	}

	config, err := rest.InClusterConfig()
	// config, err := clientcmd.BuildConfigFromFlags("", "/Users/home/.kube/config")
	if err != nil {
		log.WithError(err).Error("Failed to build kubernetes config")
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }

	informerFactory := informers.NewSharedInformerFactory(clientset, 0)
	pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()

	pvcInformer.Informer().AddEventHandler(handlerFuncs)
	log.Infof("going to run pvcInformer")
	pvcInformer.Informer().Run(ctx.Done())
	log.Infof("pvcInformer exited")
}