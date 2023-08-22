package disk

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"sync"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/builder/localdisknode"
	localdisk2 "github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/kubernetes"
)

var (
	once sync.Once
	ldn  *localDiskNodesManager
)

// localDiskNodesManager manage all disks in the cluster by interacting with localDisk resources
type localDiskNodesManager struct {
	// GetClient for query LocalDiskNode resources from k8s
	GetClient func() (*localdisknode.Kubeclient, error)

	// distributed lock or mutex lock(controller already has distributed lock )
	mutex sync.Mutex

	// DiskHandler manage LD resources in cluster
	DiskHandler *localdisk2.Handler
}

func New() Manager {
	once.Do(func() {
		ldn = &localDiskNodesManager{}
		ldn.GetClient = localdisknode.NewKubeclient
		cli, _ := kubernetes.NewClientWithCache()
		recoder, _ := kubernetes.NewRecorderFor("LocalDiskNodeManager")
		ldn.DiskHandler = localdisk2.NewLocalDiskHandler(cli, recoder)
	})

	return ldn
}

// GetNodeDisks get disks which attached on the node
func (ldn *localDiskNodesManager) GetNodeDisks(node string) ([]types.Disk, error) {
	cli, err := ldn.GetClient()
	if err != nil {
		return nil, err
	}

	var diskNode *v1alpha1.LocalDiskNode
	diskNode, err = cli.Get(node)
	if err != nil {
		return nil, err
	}

	var nodeDisks []types.Disk
	for _, pool := range diskNode.Status.Pools {
		for _, disk := range pool.Disks {
			nodeDisks = append(nodeDisks, *convertToDisk(node, disk))
		}
	}

	return nodeDisks, nil
}

// GetNodeAvailableDisks get available disks which attached on the node
func (ldn *localDiskNodesManager) GetNodeAvailableDisks(node string) ([]types.Disk, error) {
	cli, err := ldn.GetClient()
	if err != nil {
		return nil, err
	}

	var diskNode *v1alpha1.LocalDiskNode
	diskNode, err = cli.Get(node)
	if err != nil {
		return nil, err
	}

	var nodeAvailableDisks []types.Disk
	for _, pool := range diskNode.Status.Pools {
		for _, disk := range pool.Disks {
			if disk.State == v1alpha1.DiskStateAvailable {
				nodeAvailableDisks = append(nodeAvailableDisks, *convertToDisk(node, disk))
			}
		}
	}

	return nodeAvailableDisks, nil
}

func (ldn *localDiskNodesManager) MarkNodeDiskInuse(node string, disk *types.Disk) error {
	cli, err := ldn.GetClient()
	if err != nil {
		return err
	}

	var diskNode *v1alpha1.LocalDiskNode
	diskNode, err = cli.Get(node)
	if err != nil {
		return err
	}

	diskNodeOld := diskNode.DeepCopy()
	poolName := types.GetLocalDiskPoolName(disk.DiskType)
	for i, poolDisk := range diskNode.Status.Pools[poolName].Disks {
		if poolDisk.DevPath == disk.DevPath {
			diskNode.Status.Pools[poolName].Disks[i].State = v1alpha1.DiskStateInUse
			break
		}
	}

	return cli.Patch(diskNodeOld, diskNode)
}

func (ldn *localDiskNodesManager) MarkNodeDiskAvailable(node string, disk *types.Disk) error {
	cli, err := ldn.GetClient()
	if err != nil {
		return err
	}

	var diskNode *v1alpha1.LocalDiskNode
	diskNode, err = cli.Get(node)
	if err != nil {
		return err
	}

	diskNodeOld := diskNode.DeepCopy()
	poolName := types.GetLocalDiskPoolName(disk.DiskType)
	for i, poolDisk := range diskNode.Status.Pools[poolName].Disks {
		if poolDisk.DevPath == disk.DevPath {
			diskNode.Status.Pools[poolName].Disks[i].State = v1alpha1.DiskStateAvailable
			break
		}
	}

	return cli.Patch(diskNodeOld, diskNode)
}

func (ldn *localDiskNodesManager) NodeIsReady(node string) (bool, error) {
	cli, err := ldn.GetClient()
	if err != nil {
		return false, err
	}

	var diskNode *v1alpha1.LocalDiskNode
	diskNode, err = cli.Get(node)
	if err != nil {
		return false, err
	}

	return diskNode.Status.State == v1alpha1.NodeStateReady, nil
}

func (ldn *localDiskNodesManager) ListLocalDiskByNodeDevicePath(nodeName, devicePath string) ([]v1alpha1.LocalDisk, error) {
	return ldn.DiskHandler.ListLocalDiskByNodeDevicePath(nodeName, devicePath)
}

func convertToDisk(diskNode string, disk v1alpha1.LocalDevice) *types.Disk {
	return &types.Disk{
		AttachNode: diskNode,
		Name:       disk.DevPath,
		DevPath:    disk.DevPath,
		Capacity:   disk.CapacityBytes,
		DiskType:   disk.Class,
		Status:     types.DiskStatus(disk.State),
	}
}
