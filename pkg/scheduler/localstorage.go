package scheduler

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	storage "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	corelister "k8s.io/client-go/listers/core/v1"
	storagelister "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	localstorage "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	"github.com/hwameistor/local-storage/pkg/member/controller/scheduler"
	"github.com/hwameistor/local-storage/pkg/utils"
)

const (
	Name = "local-storage-scheduler-plugin"

	defaultMaxVolumeCount = 1000
)

type Args struct {
	CSIDriverName  string `json:"csiDriverName"`
	MaxVolumeCount int    `json:"maxVolumeCount"`
}

type Plugin struct {
	apiClient        client.Client
	replicaScheduler scheduler.Scheduler
	pvLister         corelister.PersistentVolumeLister
	pvcLister        corelister.PersistentVolumeClaimLister
	scLister         storagelister.StorageClassLister
	args             *Args
}

var _ framework.FilterPlugin = &Plugin{}
var _ framework.ScorePlugin = &Plugin{}

// Name returns name of the plugin. It is used in logs, etc.
func (p *Plugin) Name() string {
	return Name
}

func (p *Plugin) getPodMountedClaims(pod *core.Pod) (
	boundClaims []*core.PersistentVolumeClaim, unBoundClaims []*core.PersistentVolumeClaim, err error) {

	for _, v := range pod.Spec.Volumes {
		if v.PersistentVolumeClaim != nil {
			pvc, err := p.pvcLister.PersistentVolumeClaims(pod.Namespace).Get(v.PersistentVolumeClaim.ClaimName)
			if err != nil {
				return nil, nil, err
			}
			if len(pvc.Spec.VolumeName) > 0 {
				boundClaims = append(boundClaims, pvc)
			} else {
				unBoundClaims = append(unBoundClaims, pvc)
			}
		}
	}
	return
}

func (p *Plugin) isLocalStorageVolume(pv *core.PersistentVolume) bool {
	return pv.Spec.CSI != nil && pv.Spec.CSI.Driver == p.args.CSIDriverName
}

func (p *Plugin) getPVLocalVolume(pv *core.PersistentVolume) (*localstorage.LocalVolume, error) {
	volumeName := getVolumeNameFromPV(pv)
	vol := localstorage.LocalVolume{}
	if err := p.apiClient.Get(context.TODO(), client.ObjectKey{Name: volumeName}, &vol); err != nil {
		return nil, err
	}
	return &vol, nil
}

func (p *Plugin) isLocalVolumeAvailableOnNode(localVolume *localstorage.LocalVolume, node *nodeinfo.NodeInfo) (ok bool, err error) {
	replicas, err := p.getVolumeReadyReplicas(localVolume.Name)
	if err != nil {
		return false, err
	}

	for i := range replicas {
		if replicas[i].Spec.NodeName == node.Node().Name {
			return true, nil
		}
	}

	return
}

func (p *Plugin) useLocalStorageProvisioner(sc *storage.StorageClass) bool {
	return sc.Provisioner == p.args.CSIDriverName
}

// Filter is the functions invoked by the framework at "filter" extension point.
func (p *Plugin) Filter(ctx context.Context, state *framework.CycleState, pod *core.Pod, node *nodeinfo.NodeInfo) *framework.Status {
	boundClaims, unBoundClaims, err := p.getPodMountedClaims(pod)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}

	if len(boundClaims) == 0 && len(unBoundClaims) == 0 {
		return nil
	}

	// check bound local-storage pvc
	for i := range boundClaims {
		pv, err := p.pvLister.Get(boundClaims[i].Spec.VolumeName)
		if err != nil {
			return framework.NewStatus(framework.Error, err.Error())
		}
		if !p.isLocalStorageVolume(pv) {
			continue
		}

		localVolume, err := p.getPVLocalVolume(pv)
		if err != nil {
			return framework.NewStatus(framework.Error, err.Error())
		}

		/* we should schedule pod if the replica is ready, ignore volume state */

		//if localVolume.Status.State != localstorage.VolumeStateReady {
		//	return framework.NewStatus(
		//		framework.Unschedulable,
		//		fmt.Sprintf("pvc %s's backend volume is not ready", pv.Spec.ClaimRef.Name),
		//	)
		//}

		if ok, err := p.isLocalVolumeAvailableOnNode(localVolume, node); err != nil {
			return framework.NewStatus(framework.Error, err.Error())
		} else if !ok {
			return framework.NewStatus(
				framework.Unschedulable,
				fmt.Sprintf("volume is not available on node %s", node.Node().Name))
		}
	}

	scMap := make(map[string]*storage.StorageClass)
	var toProvisionLocalPVCs []*core.PersistentVolumeClaim

	// filter local storage's waitForFirstConsummer pvc
	for i := range unBoundClaims {
		pvc := unBoundClaims[i]
		if pvc.Spec.StorageClassName == nil || len(*pvc.Spec.StorageClassName) == 0 {
			return framework.NewStatus(
				framework.Unschedulable,
				fmt.Sprintf("pvc %s is unbound", pvc.Name))
		}

		sc, err := p.scLister.Get(*pvc.Spec.StorageClassName)
		if err != nil {
			return framework.NewStatus(framework.Error, err.Error())
		}
		if !p.useLocalStorageProvisioner(sc) {
			continue
		}

		if sc.VolumeBindingMode == nil ||
			*sc.VolumeBindingMode != storage.VolumeBindingWaitForFirstConsumer {

			return framework.NewStatus(
				framework.Unschedulable,
				fmt.Sprintf("pvc %s is unbound", pvc.Name))
		}

		scMap[sc.Name] = sc
		toProvisionLocalPVCs = append(toProvisionLocalPVCs, pvc)
	}

	if len(toProvisionLocalPVCs) == 0 {
		return nil
	}

	// check local-storage replica scheduler predicates
	// should check pvc together if they use same storageclass(will use same pool in local storage system)
	// fake a aggrated capacity pvc if they use same storageclass, then schedule them

	// key: storageclass name
	fakedToProvisionLVMPVCs := make(map[string]*core.PersistentVolumeClaim)
	var toProvisionDiskPVCs []*core.PersistentVolumeClaim
	for _, pvc := range toProvisionLocalPVCs {
		sc := scMap[*pvc.Spec.StorageClassName]
		if sc.Parameters[localstorage.VolumeParameterVolumeKindKey] == localstorage.VolumeKindDisk {
			toProvisionDiskPVCs = append(toProvisionDiskPVCs, pvc)
			continue
		}

		fakedPVC, ok := fakedToProvisionLVMPVCs[*pvc.Spec.StorageClassName]
		if !ok {
			fakedPVC = &core.PersistentVolumeClaim{}
			fakedPVC.Spec.StorageClassName = pvc.Spec.StorageClassName
			fakedPVC.Spec.Resources.Requests = make(core.ResourceList)
			fakedToProvisionLVMPVCs[*pvc.Spec.StorageClassName] = fakedPVC
		}
		a := fakedPVC.Spec.Resources.Requests[core.ResourceStorage]
		a.Add(pvc.Spec.Resources.Requests[core.ResourceStorage])

		fakedPVC.Spec.Resources.Requests[core.ResourceStorage] = a
	}

	for _, pvc := range fakedToProvisionLVMPVCs {
		vol, err := generateLocalVolumeForPVC(pvc, scMap[*pvc.Spec.StorageClassName])
		if err != nil {
			return framework.NewStatus(framework.Error, err.Error())
		}
		if err := p.replicaScheduler.Predicate(vol, node.Node().Name); err != nil {
			return framework.NewStatus(framework.Unschedulable, err.Error())
		}
	}

	if len(toProvisionDiskPVCs) == 0 {
		return nil
	}

	var volumes []*localstorage.LocalVolume
	for _, pvc := range toProvisionDiskPVCs {
		vol, err := generateLocalVolumeForPVC(pvc, scMap[*pvc.Spec.StorageClassName])
		if err != nil {
			return framework.NewStatus(framework.Error, err.Error())
		}
		volumes = append(volumes, vol)
	}

	var storageNode localstorage.LocalStorageNode
	if err := p.apiClient.Get(context.TODO(), client.ObjectKey{Name: node.Node().Name}, &storageNode); err != nil {
		if errors.IsNotFound(err) {
			return framework.NewStatus(framework.Unschedulable, "no such storage node %s", node.Node().Name)
		}
		return framework.NewStatus(framework.Error, err.Error())
	}

	if ok, reasons := predicateForDiskVolume(volumes, &storageNode); !ok {
		return framework.NewStatus(framework.Unschedulable, strings.Join(reasons, ", "))
	}

	return nil
}

func predicateForDiskVolume(volumes []*localstorage.LocalVolume, storageNode *localstorage.LocalStorageNode) (ok bool, reasons []string) {
	if storageNode.Spec.AllowedVolumeKind != localstorage.VolumeKindDisk {
		return false, []string{fmt.Sprintf("node %s does not support volume kind: %s", storageNode.Name, localstorage.VolumeKindDisk)}
	}

	poolRequestedMap := map[string]struct {
		maxRequestedCapacity int64
		volumeCount          int64
	}{}
	for _, replica := range volumes {
		request := poolRequestedMap[replica.Spec.PoolName]
		if replica.Spec.RequiredCapacityBytes > poolRequestedMap[replica.Spec.PoolName].maxRequestedCapacity {
			request.maxRequestedCapacity = replica.Spec.RequiredCapacityBytes
		}
		request.volumeCount++
		poolRequestedMap[replica.Spec.PoolName] = request
	}

	for poolName, request := range poolRequestedMap {
		pool, exists := storageNode.Status.Pools[poolName]
		if !exists {
			reasons = append(reasons, fmt.Sprintf("node %s does't have pool %s", storageNode.Name, poolName))
			continue
		}

		if request.maxRequestedCapacity > pool.VolumeCapacityBytesLimit {
			reasons = append(reasons, fmt.Sprintf(
				"volume requested capacity %d is larger than pool's volume capacity limit %d on node %s",
				request.maxRequestedCapacity,
				pool.VolumeCapacityBytesLimit,
				storageNode.Name))
			continue
		}

		if request.volumeCount > pool.FreeVolumeCount {
			reasons = append(reasons, fmt.Sprintf(
				"requested %d volume, but node %s can only support %d volume",
				request.volumeCount,
				storageNode.Name,
				pool.FreeVolumeCount))
		}
	}

	if len(reasons) > 0 {
		return false, reasons
	}

	return true, nil
}

func (p *Plugin) getVolumeReadyReplicas(volumeName string) ([]localstorage.LocalVolumeReplica, error) {
	replicaList := localstorage.LocalVolumeReplicaList{}
	if err := p.apiClient.List(context.TODO(), &replicaList); err != nil {
		return nil, err
	}

	var replicas []localstorage.LocalVolumeReplica
	for i := range replicaList.Items {
		if replicaList.Items[i].Spec.VolumeName == volumeName {
			replicas = append(replicas, replicaList.Items[i])
		}
	}

	return replicas, nil
}

func getVolumeNameFromPV(pv *core.PersistentVolume) string {
	if pv.Spec.CSI == nil {
		return ""
	}

	return pv.Spec.CSI.VolumeHandle
}

func generateLocalVolumeForPVC(pvc *core.PersistentVolumeClaim, sc *storage.StorageClass) (*localstorage.LocalVolume, error) {
	localVolume := localstorage.LocalVolume{}
	poolName, err := utils.BuildStoragePoolName(
		sc.Parameters[localstorage.VolumeParameterPoolClassKey],
		sc.Parameters[localstorage.VolumeParameterPoolTypeKey])
	if err != nil {
		return nil, err
	}

	localVolume.Spec.PoolName = poolName
	storage := pvc.Spec.Resources.Requests[core.ResourceStorage]
	localVolume.Spec.RequiredCapacityBytes = storage.Value()
	localVolume.Spec.Kind = sc.Parameters[localstorage.VolumeParameterVolumeKindKey]
	replica, _ := strconv.Atoi(sc.Parameters[localstorage.VolumeParameterReplicaNumberKey])
	localVolume.Spec.ReplicaNumber = int64(replica)
	return &localVolume, nil
}

func (p *Plugin) Score(ctx context.Context, state *framework.CycleState, pod *core.Pod, nodeName string) (int64, *framework.Status) {
	_, unBoundClaims, err := p.getPodMountedClaims(pod)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}

	if len(unBoundClaims) == 0 {
		return 0, framework.NewStatus(framework.Success, "")
	}

	// filter local storage's waitForFirstConsummer pvc
	var toProvisionLocalPVCs []*core.PersistentVolumeClaim
	scMap := make(map[string]*storage.StorageClass)
	for i := range unBoundClaims {
		pvc := unBoundClaims[i]
		if pvc.Spec.StorageClassName == nil || len(*pvc.Spec.StorageClassName) == 0 {
			return 0, framework.NewStatus(
				framework.Error,
				fmt.Sprintf("pvc %s is unbound", pvc.Name))
		}
		sc, err := p.scLister.Get(*pvc.Spec.StorageClassName)
		if err != nil {
			return 0, framework.NewStatus(framework.Error, err.Error())
		}

		if !p.useLocalStorageProvisioner(sc) {
			continue
		}

		scMap[sc.Name] = sc
		toProvisionLocalPVCs = append(toProvisionLocalPVCs, pvc)
	}

	var score int64
	// scheduler score
	for _, pvc := range toProvisionLocalPVCs {
		tmpReplica, err := generateLocalVolumeForPVC(pvc, scMap[*pvc.Spec.StorageClassName])
		if err != nil {
			return 0, framework.NewStatus(framework.Error, err.Error())
		}
		tmpScore, err := p.replicaScheduler.Score(tmpReplica, nodeName)
		if err != nil {
			return 0, framework.NewStatus(framework.Error, err.Error())
		}
		score += tmpScore / int64(len(toProvisionLocalPVCs))
	}

	if score > framework.MaxNodeScore {
		score = int64(1-float64(framework.MaxNodeScore)/float64(score)) * framework.MaxNodeScore
	}
	return score, framework.NewStatus(framework.Success, "")
}

// just return nil, cause we don't implement ScoreExtensions
func (p *Plugin) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// New initializes a new plugin and returns it.
func New(config *runtime.Unknown, f framework.FrameworkHandle) (framework.Plugin, error) {
	args := &Args{}
	if err := framework.DecodeInto(config, args); err != nil {
		return nil, err
	}

	if len(args.CSIDriverName) == 0 {
		args.CSIDriverName = localstorage.CSIDriverName
	}
	if args.MaxVolumeCount == 0 {
		args.MaxVolumeCount = defaultMaxVolumeCount
	}

	// for LocalStorage scheduler, read --kubeconfig from cmd flag first,
	// then try In-cluster config / $HOME/.kube/config
	var cfg *rest.Config
	kubeconfig, err := GetKubeconfigPath()
	if err != nil {
		if cfg, err = runtimeconfig.GetConfig(); err != nil {
			return nil, fmt.Errorf("get kubeconfig err: %s", err)
		}
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.V(1).Info(err)
			os.Exit(1)
		}
	}

	options := manager.Options{
		MetricsBindAddress: "0", // disable metrics
	}

	mgr, err := manager.New(cfg, options)
	if err != nil {
		klog.V(1).Info(err)
		os.Exit(1)
	}

	// Setup Scheme for all resources of Local Storage
	if err := localstorage.AddToScheme(mgr.GetScheme()); err != nil {
		klog.V(1).Infof("Failed to setup scheme for all resources, %s", err)
		os.Exit(1)
	}

	apiClient := mgr.GetClient()
	cache := mgr.GetCache()
	stopCh := make(chan struct{})
	go func() {
		cache.Start(stopCh)
	}()
	replicaScheduler := scheduler.New(apiClient, cache, args.MaxVolumeCount)
	// wait for cache synced
	for {
		if cache.WaitForCacheSync(stopCh) {
			break
		}
		time.Sleep(time.Second * 1)
	}
	replicaScheduler.Init()

	return &Plugin{
		args:             args,
		apiClient:        mgr.GetClient(),
		replicaScheduler: replicaScheduler,
		pvLister:         f.SharedInformerFactory().Core().V1().PersistentVolumes().Lister(),
		pvcLister:        f.SharedInformerFactory().Core().V1().PersistentVolumeClaims().Lister(),
		scLister:         f.SharedInformerFactory().Storage().V1().StorageClasses().Lister(),
	}, nil
}
