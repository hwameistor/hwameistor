package autoresizer

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	// "k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	log "github.com/sirupsen/logrus"

	hwameistorclient "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned"
	hwameistorinformer "github.com/hwameistor/hwameistor/pkg/apis/client/informers/externalversions/hwameistor/v1alpha1"
	hwameistorv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type process struct {
	cli client.Client
	ctx context.Context
	localVolume localVolumeWrapper
}

type localVolumeWrapper struct {
	hwameistorv1alpha1.LocalVolume
	replicas []localVolumeReplicaWrapper
	pvc corev1.PersistentVolumeClaim
	resizePolicy hwameistorv1alpha1.ResizePolicy
}

type localVolumeReplicaWrapper struct {
	hwameistorv1alpha1.LocalVolumeReplica
	node hwameistorv1alpha1.LocalStorageNode
	pool hwameistorv1alpha1.LocalPool
}

func newProcess(cli client.Client, ctx context.Context, lv hwameistorv1alpha1.LocalVolume) *process {
	return &process{
		cli: cli,
		ctx: ctx,
		localVolume: localVolumeWrapper{
			LocalVolume: lv,
		},
	}
}

type AutoResizer struct {
	Client client.Client
	Context context.Context
}

func NewAutoResizer(cli client.Client, ctx context.Context) *AutoResizer {
	return &AutoResizer{
		Client: cli,
		Context: ctx,
	}
}

func (r *AutoResizer) Start() {
	handlerFuncs := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Infof("LocalVolume added: %+v", obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			localVolume := newObj.(*hwameistorv1alpha1.LocalVolume)
			p := newProcess(r.Client, r.Context, *localVolume)
			if err := p.getPVC(); err != nil {
				log.Errorf("get pvc err: %v", err)
				return
			}

			requestStorageCapacity := p.localVolume.pvc.Spec.Resources.Requests.Storage().Value()
			log.Infof("requestStorageCapacity: %v", requestStorageCapacity)
			usedCapacityBytes := p.localVolume.Status.UsedCapacityBytes
			log.Infof("UsedCapacityBytes: %v", usedCapacityBytes)
			usagePercentage := computeUsedPercentage(usedCapacityBytes, requestStorageCapacity)
			log.Infof("usagePercentage: %v", usagePercentage)

			resizePolicyName, ok := p.localVolume.pvc.Annotations[PVCResizePolicyAnnotationKey]
			if ok {
				resizePolicy := hwameistorv1alpha1.ResizePolicy{}
				if err := p.cli.Get(p.ctx, types.NamespacedName{Name: resizePolicyName}, &resizePolicy); err != nil {
					log.Errorf("get ResizePolicy err: %v", err)
					return
				}
				p.localVolume.resizePolicy = resizePolicy
			} else {
				// expect resizepolicy related only when annotation appointed
				log.Infof("pvc %v:%v has no resizepolicy annotation", p.localVolume.pvc.Namespace, p.localVolume.pvc.Name)
				return

				resizePolicyList := &hwameistorv1alpha1.ResizePolicyList{}
				labelSelector, err := metav1.LabelSelectorAsSelector(&defaultResizePolicyLabelSelector)
				if err != nil {
					log.Errorf("convert labelSelector err: %v", err)
					return
				}
				if err := p.cli.List(p.ctx, resizePolicyList, &client.ListOptions{LabelSelector: labelSelector}); err != nil {
					log.Errorf("list resizepolicy err: %v", err)
					return
				}

				if len(resizePolicyList.Items) > 0 {
					p.localVolume.resizePolicy = resizePolicyList.Items[0]
				} else {
					log.Infof("no default resizepolicy in cluster")
					return
				}
			}

			if p.reachedResizeThreshold(usagePercentage) {
				if err := p.getLocalVolumeReplicaAndStorageNodePool(); err != nil {
					log.Errorf("getLocalVolumeReplicaAndStorageNodePool err: %v", err)
					return
				}
			} else {
				log.Infof("have not reached resizeThreshold, don't resize")
				return
			}

			if p.reachedNodePoolUsageLimit() {
				log.Infof("reached nodePoolUsageLimit, don't resize")
				return
			}

			bytesToResize := computeBytesToResize(usedCapacityBytes, p.localVolume.resizePolicy.Spec.WarningThreshold)
			log.Infof("bytesToResize: %v", bytesToResize)
			if p.checkPoolCapacityEnough(bytesToResize) {
				quantityToResize := resource.NewQuantity(bytesToResize, resource.BinarySI)
				pvcToUpdate := p.localVolume.pvc.DeepCopy()
				pvcToUpdate.Spec.Resources.Requests[corev1.ResourceStorage] = quantityToResize.DeepCopy()
				if err := r.Client.Update(r.Context, pvcToUpdate); err != nil {
					log.Errorf("pvc updated err: %v", err)
					return
				}
				log.Infof("pvc requested storage updated to %+v", quantityToResize)
			} else {
				log.Infof("pool capacity not enough")
			}
		},
		DeleteFunc: func(obj interface{}) {
			log.Infof("LocalVolume deleted: %+v", obj)
		},
	}

	config, err := rest.InClusterConfig()
	// config, err := clientcmd.BuildConfigFromFlags("", "/Users/home/.kube/config")
	if err != nil {
		log.WithError(err).Error("Failed to build kubernetes config")
		return
	}
	clientset, err := hwameistorclient.NewForConfig(config)
	if err != nil {
		log.WithError(err).Error("Failed to build clientset")
		return
	}

	localVolumeInformer := hwameistorinformer.NewLocalVolumeInformer(clientset, 0, cache.Indexers{})
	localVolumeInformer.AddEventHandler(handlerFuncs)
	log.Infof("Going to run localVolumeInformer")
	localVolumeInformer.Run(r.Context.Done())
	log.Infof("localVolumeInformer exited")
}

func computeBytesToResize(usedCapacityBytes int64, warningThreshold int8) int64 {
	// expect to lower usagePercentage to under warningThreshold
	targetPercentage := warningThreshold - 1

	// bytesToResize := usedCapacityBytes / ( int64(targetPercentage) / 100 )
	bytesToResize := usedCapacityBytes * 100 / int64(targetPercentage)
	log.Infof("bytesToResize: %v", bytesToResize)
	bytesToResize = NumericToLVMBytes(bytesToResize)
	log.Infof("aligned bytesToResize: %v", bytesToResize)
	for {
		if checkUsageLowerThanThreshold(usedCapacityBytes, bytesToResize, warningThreshold) {
			break
		}
		log.Infof("Usage not under waringThreshold yet, go on increasing bytesToResize")
		bytesToResize = increaseBytesToResize(bytesToResize)
	}
	return bytesToResize
}

func NumericToLVMBytes(bytes int64) int64 {
	peSize := int64(4 * 1024 * 1024)
	if bytes <= peSize {
		return peSize
	}
	if bytes%peSize == 0 {
		return bytes
	}
	return (bytes/peSize + 1) * peSize
}

func increaseBytesToResize (bytesToResize int64) int64 {
	increment := int64(4 * 1024 * 1024)
	return bytesToResize + increment
}

func computeUsedPercentage(usedBytes, totalBytes int64) float64 {
	usedPercentage := 100 * float64(usedBytes) / float64(totalBytes)
	log.Infof("usedPercentage: %v", usedPercentage)
	return usedPercentage
}

func checkUsageLowerThanThreshold(usedBytes, totalBytes int64, threshold int8) bool {
	usage := computeUsedPercentage(usedBytes, totalBytes)
	log.Infof("usage: %v, threshold: %v", usage, threshold)
	return ( usage < float64(threshold) )
}

func (p *process) getPVC() error {
	lv := p.localVolume
	relatedPVC := corev1.PersistentVolumeClaim{}
	pvcKey := types.NamespacedName{Namespace: lv.Spec.PersistentVolumeClaimNamespace, Name: lv.Spec.PersistentVolumeClaimName}
	if err := p.cli.Get(p.ctx, pvcKey, &relatedPVC); err != nil {
		log.Errorf("get pvc for localvolume err, localvolume name: %v, pvc key: %v, err: %v", lv.Name, pvcKey, err)
		return err
	}
	p.localVolume.pvc = relatedPVC
	return nil
}

func (p *process) getLocalVolumeReplicaAndStorageNodePool() error {
	lv := p.localVolume
	if len(lv.Status.Replicas) == 0 {
		log.Errorf("replicas is zero by localvolume.status, localvolume: %v", lv.Name)
		return errors.New("no replica")
	}

	for _, replicaName := range lv.Status.Replicas {
		lvr := hwameistorv1alpha1.LocalVolumeReplica{}
		if err := p.cli.Get(p.ctx, types.NamespacedName{Name: replicaName}, &lvr); err != nil {
			log.Errorf("get localvolumereplica %v of localvolume %v err: %v", replicaName, lv.Name, err)
			return err
		}
		lvrWrapper := localVolumeReplicaWrapper{
			LocalVolumeReplica: lvr,
		}

		nodeName := lvr.Spec.NodeName
		lsn := hwameistorv1alpha1.LocalStorageNode{}
		if err := p.cli.Get(p.ctx, types.NamespacedName{Name: nodeName}, &lsn); err != nil {
			log.Errorf("get localstoragenode %v of localvolumereplica %v err: %v", nodeName, lvr.Name, err)
			return err
		}
		lvrWrapper.node = lsn

		poolName := lvr.Spec.PoolName
		pool, ok := lsn.Status.Pools[poolName]
		if !ok {
			log.Errorf("pool %v of localvolumereplica %v not exist on localstoragenode %v", poolName, lvr.Name, lsn.Name)
			return errors.New("pool not exist")
		}
		lvrWrapper.pool = pool

		p.localVolume.replicas = append(p.localVolume.replicas, lvrWrapper)
	}
	return nil
}

func (p *process) reachedResizeThreshold(volumeUsedPercentage float64) bool {
	thresholdPercentage := float64(p.localVolume.resizePolicy.Spec.ResizeThreshold)
	log.Infof("volumeUsedPercentage: %v, thresholdPercentage: %v", volumeUsedPercentage, thresholdPercentage)
	return volumeUsedPercentage >= thresholdPercentage
}

func (p *process) reachedNodePoolUsageLimit() bool {
	resizePolicy := p.localVolume.resizePolicy
	for _, lvrWrapper := range p.localVolume.replicas {
		lvr := lvrWrapper.LocalVolumeReplica
		lsn := lvrWrapper.node
		pool := lvrWrapper.pool
		poolUsage := computeUsedPercentage(pool.UsedCapacityBytes, pool.TotalCapacityBytes)
		log.Infof("pool %v of localstoragenode %v which localvolumereplica %v located on has usage: %v", pool.Name, lsn.Name, lvr.Name, poolUsage)
		log.Infof("nodePoolUsageLimit of resizePolicy %v is %v", resizePolicy.Name, resizePolicy.Spec.NodePoolUsageLimit)
		if poolUsage >= float64(resizePolicy.Spec.NodePoolUsageLimit) {
			log.Infof("poolUsage beyond limit, poolName: %v, localstoragenode: %v, locavolumereplica: %v", pool.Name, lsn.Name, lvr.Name)
			return true
		}
	}

	return false
}

func (p *process) checkPoolCapacityEnough(bytesToResize int64) bool {
	volumeCapacityIncrement := bytesToResize - p.localVolume.pvc.Spec.Resources.Requests.Storage().Value()
	for _, lvrWrapper := range p.localVolume.replicas {
		pool := lvrWrapper.pool
		log.Infof("volumeCapacityIncrement: %v, pool.FreeCapacityBytes: %v", volumeCapacityIncrement, pool.FreeCapacityBytes)
		if volumeCapacityIncrement > pool.FreeCapacityBytes {
			return false
		}
	}
	return true
}