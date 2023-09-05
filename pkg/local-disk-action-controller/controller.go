package localdiskactioncontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/google/uuid"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/gobwas/glob"
	clientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	informers "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	listers "github.com/hwameistor/hwameistor/pkg/apis/client/listers/hwameistor/v1alpha1"
	"github.com/sirupsen/logrus"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	LatestMatchedLength = 10
	DefaultLogKey       = "lda-controller"
)

func NewLocalDiskActionController(
	clientSet clientset.Interface,
	localDiskInformer informers.LocalDiskInformer,
	localDiskActionInformer informers.LocalDiskActionInformer) *LocalDiskActionController {

	c := &LocalDiskActionController{
		clientSet:                clientSet,
		localDisksLister:         localDiskInformer.Lister(),
		localDisksSynced:         localDiskInformer.Informer().HasSynced,
		localDiskActionsLister:   localDiskActionInformer.Lister(),
		localDiskActionsSynced:   localDiskActionInformer.Informer().HasSynced,
		localDiskWorkqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "localDisks"),
		localDiskActionWorkqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "localDiskActions"),
	}

	localDiskInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueLocalDisk,
		UpdateFunc: func(old, new interface{}) {
			c.enqueueLocalDisk(new)
		},
	})

	localDiskActionInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueLocalDiskAction,
		UpdateFunc: func(old, new interface{}) {
			c.enqueueLocalDiskAction(new)
		},
	})

	return c
}

type LocalDiskActionController struct {
	clientSet                clientset.Interface
	localDisksLister         listers.LocalDiskLister
	localDisksSynced         cache.InformerSynced
	localDiskActionsLister   listers.LocalDiskActionLister
	localDiskActionsSynced   cache.InformerSynced
	localDiskWorkqueue       workqueue.RateLimitingInterface
	localDiskActionWorkqueue workqueue.RateLimitingInterface
}

func (c *LocalDiskActionController) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	defer utilruntime.HandleCrash()
	defer func() {
		if waitTimeout(&wg, 30*time.Second) {
			logrus.Info("timeout: processors exit")
		} else {
			logrus.Info("all processors exit")
		}
	}()
	defer c.localDiskWorkqueue.ShutDown()
	defer c.localDiskActionWorkqueue.ShutDown()

	logrus.Info("starting localdiskaction controller")

	// Wait for the caches to be synced before starting processors
	logrus.Info("waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), c.localDisksSynced, c.localDiskActionsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	wg.Add(2)
	go func() {
		defer wg.Done()
		logrus.Info("start LDA processors")
		wait.UntilWithContext(ctx, c.runLDAProcessor, time.Second)
		logrus.Info("shutdown LDA processors")
	}()
	go func() {
		defer wg.Done()
		logrus.Info("start LD processors")
		wait.UntilWithContext(ctx, c.runLDProcessor, time.Second)
		logrus.Info("shutdown LD processors")
	}()

	<-ctx.Done()

	return nil
}

func (c *LocalDiskActionController) enqueueLocalDisk(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		logrus.Error(err)
		return
	}
	if ld, ok := obj.(*v1alpha1.LocalDisk); !ok {
		logrus.Error("error decoding object, invalid type")
		return
	} else if ld.Spec.Reserved {
		return
	}
	c.localDiskWorkqueue.Add(key)
}

func (c *LocalDiskActionController) enqueueLocalDiskAction(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		logrus.Error(err)
		return
	}
	c.localDiskActionWorkqueue.Add(key)
}

func (c *LocalDiskActionController) runLDProcessor(ctx context.Context) {
	for c.processNextLDWorkItem(ctx) {
	}
}

func (c *LocalDiskActionController) runLDAProcessor(ctx context.Context) {
	for c.processNextLDAWorkItem(ctx) {
	}
}

func (c *LocalDiskActionController) processNextLDWorkItem(ctx context.Context) bool {
	obj, shutdown := c.localDiskWorkqueue.Get()
	if shutdown {
		return false
	}
	log := logrus.WithFields(logrus.Fields{
		"processor": "localdisk",
		"traceId":   uuid.New().String(),
	})

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		defer c.localDiskWorkqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.localDiskWorkqueue.Forget(obj)
			log.Errorf("expected string in workqueue but got %#v", obj)
			return nil
		}
		log = log.WithField("name", key)
		ctx = logIntoContext(ctx, log)
		if err := c.syncLDHandler(ctx, key); err != nil {
			c.localDiskWorkqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.localDiskWorkqueue.Forget(obj)
		log.Info("successfully synced")
		return nil
	}(obj)

	if err != nil {
		log.Error(err)
		return true
	}
	return true
}

func (c *LocalDiskActionController) processNextLDAWorkItem(ctx context.Context) bool {
	obj, shutdown := c.localDiskActionWorkqueue.Get()
	if shutdown {
		return false
	}
	log := logrus.WithFields(logrus.Fields{
		"processor": "localdiskaction",
		"traceId":   uuid.New().String(),
	})

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		defer c.localDiskActionWorkqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.localDiskActionWorkqueue.Forget(obj)
			log.Errorf("expected string in workqueue but got %#v", obj)
			return nil
		}
		log = log.WithField("name", key)
		ctx = logIntoContext(ctx, log)
		if err := c.syncLDAHandler(ctx, key); err != nil {
			c.localDiskActionWorkqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.localDiskActionWorkqueue.Forget(obj)
		log.Info("successfully synced")
		return nil
	}(obj)

	if err != nil {
		log.Error(err)
		return true
	}
	return true
}

func (c *LocalDiskActionController) syncLDHandler(ctx context.Context, key string) error {
	log := logFromContext(ctx)
	log.Info("start sync localdisk")
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Errorf("invalid resource key: %s", key)
		return nil
	}
	ld, err := c.localDisksLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Error("localdisk in work queue no longer exists")
			return nil
		}
		return err
	}
	ldas, err := c.localDiskActionsLister.List(labels.Everything())
	if err != nil {
		return err
	}
	if !ld.Spec.Reserved {
		_, err = c.doReserveByLDAs(ctx, ld, ldas)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *LocalDiskActionController) syncLDAHandler(ctx context.Context, key string) error {
	log := logFromContext(ctx)
	log.Info("start sync localdiskaction")
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Errorf("invalid resource key: %s", key)
		return nil
	}
	lda, err := c.localDiskActionsLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Error("localdiskaction in work queue no longer exists")
			return nil
		}
		return err
	}
	localDisks, err := c.localDisksLister.List(labels.Everything())
	if err != nil {
		return err
	}
	if lda.Spec.Action == v1alpha1.LocalDiskActionReserve {
		_, err = c.doReserveByLDs(ctx, lda, localDisks)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown action type")
	}

	return nil
}

func (c *LocalDiskActionController) doReserveByLDs(ctx context.Context, lda *v1alpha1.LocalDiskAction, localDisks []*v1alpha1.LocalDisk) (*v1alpha1.LocalDiskAction, error) {
	log := logFromContext(ctx)
	log.WithField("rule", lda.Spec.Rule).Info("do reserve localdisks by localdiskaction")
	for _, ld := range localDisks {
		if c.shouldFilterLDByRule(ctx, &lda.Spec.Rule, ld) && !ld.Spec.Reserved {
			log.WithField("localdiskName", ld.Name).Info("localdisk should be reserved")
			_, err := c.patchReserve(ctx, ld)
			if err != nil {
				return nil, err
			}
			log.Info("refresh localdiskaction")
			lda, err = c.refreshLatestLd(ctx, lda, ld.Name)
			if err != nil {
				return nil, err
			}
		}
	}
	return lda, nil
}

func (c *LocalDiskActionController) doReserveByLDAs(ctx context.Context, ld *v1alpha1.LocalDisk, ldas []*v1alpha1.LocalDiskAction) (*v1alpha1.LocalDisk, error) {
	var err error
	for _, lda := range ldas {
		if lda.Spec.Action == v1alpha1.LocalDiskActionReserve && c.shouldFilterLDByRule(ctx, &lda.Spec.Rule, ld) {
			log := logFromContext(ctx).WithField("localdiskactionName", lda.Name)
			log.WithField("rule", lda.Spec.Rule).Info("localdisk should be reserved")
			ld, err = c.patchReserve(ctx, ld)
			if err != nil {
				return nil, err
			}
			log.Info("refresh localdiskaction")
			_, err = c.refreshLatestLd(ctx, lda, ld.Name)
			if err != nil {
				return nil, err
			}
			// should return if ld has been patched
			return ld, nil
		}
	}
	return ld, nil
}

func (c *LocalDiskActionController) shouldFilterLDByRule(ctx context.Context, rule *v1alpha1.LocalDiskActionRule, ld *v1alpha1.LocalDisk) bool {
	// should not filter if rule is empty
	if rule.MinCapacity == 0 && rule.MaxCapacity == 0 && rule.DevicePath == "" {
		return false
	}
	if rule.MinCapacity != 0 && ld.Spec.Capacity >= rule.MinCapacity {
		return false
	}
	if rule.MaxCapacity != 0 && ld.Spec.Capacity <= rule.MaxCapacity {
		return false
	}
	if rule.DevicePath != "" {
		g, err := glob.Compile(rule.DevicePath)
		if err != nil {
			logFromContext(ctx).WithField("devicePath", rule.DevicePath).Error("can't compile rule DevicePath")
			return false
		}
		if !g.Match(ld.Spec.DevicePath) {
			return false
		}
	}
	return true
}

func (c *LocalDiskActionController) patchReserve(ctx context.Context, oldLocalDisk *v1alpha1.LocalDisk) (*v1alpha1.LocalDisk, error) {
	if oldLocalDisk.Spec.Reserved {
		return oldLocalDisk, nil
	}

	newLocalDisk := oldLocalDisk.DeepCopy()
	newLocalDisk.Spec.Reserved = true

	oldData, _ := json.Marshal(oldLocalDisk)
	newData, _ := json.Marshal(newLocalDisk)

	data, err := jsonpatch.CreateMergePatch(oldData, newData)
	if err != nil {
		return nil, fmt.Errorf("create patch data err: %s", err)
	}
	if res, err := c.clientSet.HwameistorV1alpha1().LocalDisks().Patch(ctx, newLocalDisk.Name, types.MergePatchType, data, metav1.PatchOptions{}); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func (c *LocalDiskActionController) refreshLatestLd(ctx context.Context, oldLda *v1alpha1.LocalDiskAction, ldName string) (*v1alpha1.LocalDiskAction, error) {
	// check if ldName has already exist
	for _, n := range oldLda.Status.LatestMatchedLds {
		if n == ldName {
			return oldLda, nil
		}
	}
	newLda := oldLda.DeepCopy()
	latestMatchedLds := append([]string{ldName}, newLda.Status.LatestMatchedLds...)
	if len(latestMatchedLds) > LatestMatchedLength {
		latestMatchedLds = latestMatchedLds[:LatestMatchedLength]
	}
	newLda.Status.LatestMatchedLds = latestMatchedLds

	oldData, _ := json.Marshal(oldLda)
	newData, _ := json.Marshal(newLda)

	data, err := jsonpatch.CreateMergePatch(oldData, newData)
	if err != nil {
		return nil, fmt.Errorf("create patch data err: %s", err)
	}
	if res, err := c.clientSet.HwameistorV1alpha1().LocalDiskActions().Patch(ctx, newLda.Name, types.MergePatchType, data, metav1.PatchOptions{}, "status"); err != nil {
		return nil, err
	} else {
		return res, nil
	}
}

func logIntoContext(ctx context.Context, log *logrus.Entry) context.Context {
	return context.WithValue(ctx, DefaultLogKey, log)
}

func logFromContext(ctx context.Context) *logrus.Entry {
	if v, ok := ctx.Value(DefaultLogKey).(*logrus.Entry); ok {
		return v
	}
	return logrus.WithFields(nil)
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
