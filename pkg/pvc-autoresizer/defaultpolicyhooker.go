package autoresizer

import (
	"context"
	"time"

	hwameistorv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	// apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	// "k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var defaultResizePolicyName = "default"
var defaultResizePolicy = "default"

var defaultResizePolicyLabelSelector = metav1.LabelSelector{
	MatchLabels: map[string]string{
		"hwameistor.io/is-default-resizepolicy": "true",
	},
}

type Hooker struct {
	Client client.Client
	Context context.Context
}

func NewHooker(cli client.Client, ctx context.Context) *Hooker {
	return &Hooker{
		Client: cli,
		Context: ctx,
	}
}

func (h *Hooker) Start() {
	handlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pvc := obj.(*corev1.PersistentVolumeClaim)
			log.Infof("New pvc: %v in namespace %v", pvc.Name, pvc.Namespace)
			if policyName, exist := pvc.Annotations[PVCResizePolicyAnnotationName]; exist {
				log.Infof("pvc %v in namespace %v already related to resizePolicy %v, don't add default resizepolicy annotation", pvc.Name, pvc.Namespace, policyName)
				return
			}
			resizePolicyList := &hwameistorv1alpha1.ResizePolicyList{}
			labelSelector, err := metav1.LabelSelectorAsSelector(&defaultResizePolicyLabelSelector)
			if err != nil {
				log.Errorf("convert labelSelector err: %v", err)
				return
			}
			if err := h.Client.List(h.Context, resizePolicyList, &client.ListOptions{LabelSelector: labelSelector}); err != nil {
				log.Errorf("list resizepolicy err: %v", err)
				return
			}
			if len(resizePolicyList.Items) > 0 {
				pvc.Annotations[PVCResizePolicyAnnotationKey] = defaultResizePolicy
				if err := h.Client.Update(h.Context, pvc); err != nil {
					log.Errorf("add annotation err for pvc %v in namespace %v, err: %v", pvc.Name, pvc.Namespace, err)
					return
				}
				log.Infof("Added default resizepolicy annotation for pvc %v in namespace %v", pvc.Name, pvc.Namespace)
			}
			// defaultPolicy := hwameistorv1alpha1.ResizePolicy{}
			// if err := h.Client.Get(h.Context, types.NamespacedName{Name: defaultResizePolicyName}, &defaultPolicy); err != nil {
			// 	if apierrors.IsNotFound(err) {
			// 		log.Infof("No default resizepolicy exist, pvc %v in namespace %v not added annotation", pvc.Name, pvc.Namespace)
			// 		return
			// 	} else {
			// 		log.Errorf("get ResizePolicy err: %v", err)
			// 		return
			// 	}
			// }
			// pvc.Annotations[PVCResizePolicyAnnotationName] = defaultResizePolicyName
			// if err := h.Client.Update(h.Context, pvc); err != nil {
			// 	log.Errorf("add annotation err for pvc %v in namespace %v, err: %v", pvc.Name, pvc.Namespace, err)
			// 	return
			// }
			// log.Infof("Added default resizepolicy annotation for pvc %v in namespace %v", pvc.Name, pvc.Namespace)
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

	resyncPeriod := 0 * time.Minute
	pvcInformer := cache.NewSharedIndexInformer(
        &cache.ListWatch{ListFunc: func(lo metav1.ListOptions) (runtime.Object, error) {
            return clientset.CoreV1().PersistentVolumeClaims("").List(context.Background(), lo)
        }, WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
            return clientset.CoreV1().PersistentVolumeClaims("").Watch(context.Background(), lo)
        }},
        &corev1.PersistentVolumeClaim{},
        resyncPeriod,
        cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
    )

	pvcInformer.AddEventHandler(handlerFuncs)
	log.Infof("Going to run pvcInformer")
	pvcInformer.Run(h.Context.Done())
	log.Infof("pvcInformer exited")
}