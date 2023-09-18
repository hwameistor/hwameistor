package autoresizer

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"

	hwameistorv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"

	// corev1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	runtimecache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	log.Debugf("to init")
	pvcSelectorChain.Name = "pvcSelectorChain"
	namespaceSelectorChain.Name = "namespaceSelectorChain"
	storageClassSelectorChain.Name = "storageClassSelectorChain"
	clusterResizePolicyChain.Name = "clusterResizePolicyChain"
}

type ResizePolicyChain struct {
	Name string
	Lock sync.Mutex
	Chain []*hwameistorv1alpha1.ResizePolicy
}

var pvcSelectorChain ResizePolicyChain
var namespaceSelectorChain ResizePolicyChain
var storageClassSelectorChain ResizePolicyChain
var clusterResizePolicyChain ResizePolicyChain

func isPVCSelector(rp *hwameistorv1alpha1.ResizePolicy) bool {
	return (rp.Spec.PVCSelector != nil)
}

func isNamespaceSelector(rp *hwameistorv1alpha1.ResizePolicy) bool {
	return (rp.Spec.NamespaceSelector != nil)
}

func isStorageClassSelector(rp *hwameistorv1alpha1.ResizePolicy) bool {
	return (rp.Spec.StorageClassSelector != nil)
}

func (c *ResizePolicyChain) insert(rp *hwameistorv1alpha1.ResizePolicy) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	chain := append(c.Chain, rp)
	sort.SliceStable(chain, func(i, j int) bool {
		return chain[i].CreationTimestamp.After(chain[j].CreationTimestamp.Time)
	})
	c.Chain = chain
}

func (c *ResizePolicyChain) delete(rp *hwameistorv1alpha1.ResizePolicy) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	for i, v := range c.Chain {
		if v.Name == rp.Name {
			c.Chain = append(c.Chain[:i], c.Chain[i+1:]...)
		}
	}
}

func (c *ResizePolicyChain) have(rp *hwameistorv1alpha1.ResizePolicy) bool {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	for _, v := range c.Chain {
		if v.Name == rp.Name {
			return true
		}
	}
	return false
}

func (c *ResizePolicyChain) update(rp *hwameistorv1alpha1.ResizePolicy) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	for i, v := range c.Chain {
		if v.Name == rp.Name {
			c.Chain[i] = rp
		}
	}
}

func (c *ResizePolicyChain) print() {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	toPrint := ""
	for i, v := range c.Chain {
		toPrint = fmt.Sprintf("%v %v:%v ", toPrint, i, v.Name)
	}
	log.Infof("resizepolicy chain %v has %v resizepolicies now: %v", c.Name, len(c.Chain), toPrint)
}

func addFunc(obj interface{}) {
	rp := obj.(*hwameistorv1alpha1.ResizePolicy)
	log.Infof("resizepolicy %v added", rp.Name)
	defer requeuePVC(cliVar, pvcWorkQueue)
	if isPVCSelector(rp) {
		pvcSelectorChain.insert(rp)
		log.Infof("resizepolicy %v added into chain %v", rp.Name, pvcSelectorChain.Name)
		pvcSelectorChain.print()
		return
	}
	if isNamespaceSelector(rp) {
		namespaceSelectorChain.insert(rp)
		log.Infof("resizepolicy %v added into chain %v", rp.Name, namespaceSelectorChain.Name)
		namespaceSelectorChain.print()
		return
	}
	if isStorageClassSelector(rp) {
		storageClassSelectorChain.insert(rp)
		log.Infof("resizepolicy %v added into chain %v", rp.Name, storageClassSelectorChain.Name)
		storageClassSelectorChain.print()
		return
	}
	clusterResizePolicyChain.insert(rp)
	log.Infof("resizepolicy %v added into chain %v", rp.Name, clusterResizePolicyChain.Name)
	clusterResizePolicyChain.print()
}

func deleteFunc(obj interface{}) {
	rp := obj.(*hwameistorv1alpha1.ResizePolicy)
	log.Infof("resizepolicy %v deleted", rp.Name)
	if pvcSelectorChain.have(rp) {
		pvcSelectorChain.delete(rp)
		log.Infof("resizepolicy %v deleted from chain %v", rp.Name, pvcSelectorChain.Name)
		pvcSelectorChain.print()
		return
	}
	if namespaceSelectorChain.have(rp) {
		namespaceSelectorChain.delete(rp)
		log.Infof("resizepolicy %v deleted from chain %v", rp.Name, namespaceSelectorChain.Name)
		namespaceSelectorChain.print()
		return
	}
	if storageClassSelectorChain.have(rp) {
		storageClassSelectorChain.delete(rp)
		log.Infof("resizepolicy %v deleted from chain %v", rp.Name, storageClassSelectorChain.Name)
		storageClassSelectorChain.print()
		return
	}
	if clusterResizePolicyChain.have(rp) {
		clusterResizePolicyChain.delete(rp)
		log.Infof("resizepolicy %v deleted from chain %v", rp.Name, clusterResizePolicyChain.Name)
		clusterResizePolicyChain.print()
		return
	}
}

func findChainResizePolicyIn(rp *hwameistorv1alpha1.ResizePolicy) *ResizePolicyChain {
	if pvcSelectorChain.have(rp) {
		return &pvcSelectorChain
	}
	if namespaceSelectorChain.have(rp) {
		return &namespaceSelectorChain
	}
	if storageClassSelectorChain.have(rp) {
		return &storageClassSelectorChain
	}
	if clusterResizePolicyChain.have(rp) {
		return &clusterResizePolicyChain
	}
	return nil
}

func decideChainForResizePolicy(rp *hwameistorv1alpha1.ResizePolicy) *ResizePolicyChain {
	if isPVCSelector(rp) {
		return &pvcSelectorChain
	}
	if isNamespaceSelector(rp) {
		return &namespaceSelectorChain
	}
	if isStorageClassSelector(rp) {
		return &storageClassSelectorChain
	}
	return &clusterResizePolicyChain
}

func updateFunc(olbObj, newObj interface{}) {
	rp := newObj.(*hwameistorv1alpha1.ResizePolicy)
	log.Infof("resizepolicy %v updated", rp.Name)
	rpOld := olbObj.(*hwameistorv1alpha1.ResizePolicy)
	if selectorUpdated(rp, rpOld) {
		log.Infof("resizepolicy %v updated selector, will requeue pvc", rp.Name)
		defer requeuePVC(cliVar, pvcWorkQueue)
	}
	chainLocated := findChainResizePolicyIn(rp)
	chainShoudlBeIn := decideChainForResizePolicy(rp)
	if chainLocated != nil {
		chainLocated.update(rp)
		log.Infof("resizepolicy chain %v updated resizepolicy %v in it", chainLocated.Name, rp.Name)
		if chainLocated != chainShoudlBeIn {
			chainLocated.delete(rp)
			log.Infof("resizepolicy %v deleted from chain %v", rp.Name, chainLocated.Name)
			chainLocated.print()
			chainShoudlBeIn.insert(rp)
			log.Infof("resizepolicy %v added into chain %v", rp.Name, chainShoudlBeIn.Name)
			chainShoudlBeIn.print()
		}
		return
	}
	chainShoudlBeIn.insert(rp)
	log.Infof("resizepolicy %v added into chain %v", rp.Name, chainShoudlBeIn.Name)
	chainShoudlBeIn.print()
}


var cliVar client.Client
var pvcWorkQueue workqueue.RateLimitingInterface

func StartResizePolicyEventHandlerV2(cli client.Client, q workqueue.RateLimitingInterface, runtimeCache runtimecache.Cache) {
	cliVar = cli
	pvcWorkQueue = q

	handlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc: addFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
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

func selectorUpdated(new, old *hwameistorv1alpha1.ResizePolicy) bool {
	if !reflect.DeepEqual(new.Spec.PVCSelector, old.Spec.PVCSelector) {
		return true
	}
	if !reflect.DeepEqual(new.Spec.NamespaceSelector, old.Spec.NamespaceSelector) {
		return true
	}
	if !reflect.DeepEqual(new.Spec.StorageClassSelector, old.Spec.StorageClassSelector) {
		return true
	}
	return false
}

func requeuePVC(cli client.Client, q workqueue.RateLimitingInterface) {
	log.Infof("to requeue pvc")
	pvcList := &corev1.PersistentVolumeClaimList{}
	if err := cli.List(context.TODO(), pvcList); err != nil {
		log.Errorf("list pvc err: %v", err)
		return
	}
	for _, pvc := range pvcList.Items {
		q.Add(&pvc)
	}
}
