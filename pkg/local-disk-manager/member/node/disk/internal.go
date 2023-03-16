package disk

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/member/types"
	"reflect"
	"sync"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/builder/localdisknode"
	localdisk2 "github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/kubernetes"
)

var (
	once sync.Once
	ldn  *localDiskNodesManager
)

const (
	ReservedPVCKey = "disk.hwameistor.io/pvc"
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

func (ldn *localDiskNodesManager) ReleaseDisk(disk string) error {
	if disk == "" {
		log.Debug("ReleaseDisk skipped due to disk needs to release is empty")
		return nil
	}
	ld, err := ldn.DiskHandler.GetLocalDisk(client.ObjectKey{Name: disk})
	if err != nil {
		return err
	}
	ldn.DiskHandler.For(ld)
	ldn.DiskHandler.RemoveLabel(labels.Set{ReservedPVCKey: ""})
	ldn.DiskHandler.SetPartition(false)
	return ldn.DiskHandler.Update()
}

func (ldn *localDiskNodesManager) UnReserveDiskForPVC(pvc string) error {
	label := labels.Set{ReservedPVCKey: pvc}
	list, err := ldn.DiskHandler.GetLocalDiskWithLabels(label)
	if err != nil {
		return err
	}

	for _, disk := range list.Items {
		ldn.DiskHandler.For(&disk)
		ldn.DiskHandler.RemoveLabel(label)
		if err = ldn.DiskHandler.Update(); err != nil {
			return err
		}
	}

	return err
}

func New() Manager {
	once.Do(func() {
		ldn = &localDiskNodesManager{}
		ldn.GetClient = localdisknode.NewKubeclient
		cli, _ := kubernetes.NewClient()
		recoder, _ := kubernetes.NewRecorderFor("localdisknodemanager")
		ldn.DiskHandler = localdisk2.NewLocalDiskHandler(cli, recoder)
	})

	return ldn
}

// GetClusterDisks
// Here is just a simple implementation
func (ldn *localDiskNodesManager) GetClusterDisks() (map[string][]*types.Disk, error) {
	cli, err := ldn.GetClient()
	if err != nil {
		return nil, err
	}

	// fixme: should do more check
	var clusterDisks = make(map[string][]*types.Disk)

	var nodes *v1alpha1.LocalDiskNodeList

	nodes, err = cli.List()
	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		var nodeDisks []*types.Disk
		for name, disk := range node.Status.Disks {
			nodeDisks = append(nodeDisks, convertToDisk(node.Spec.AttachNode, name, disk))
		}

		clusterDisks[node.Spec.AttachNode] = nodeDisks
	}

	return clusterDisks, nil
}

// GetNodeDisks get disks which attached on the node
func (ldn *localDiskNodesManager) GetNodeDisks(node string) ([]*types.Disk, error) {
	cli, err := ldn.GetClient()
	if err != nil {
		return nil, err
	}

	var diskNode *v1alpha1.LocalDiskNode
	diskNode, err = cli.Get(node)
	if err != nil {
		return nil, err
	}

	var nodeDisks []*types.Disk
	for name, disk := range diskNode.Status.Disks {
		nodeDisks = append(nodeDisks, convertToDisk(node, name, disk))
	}

	return nodeDisks, nil
}

func (ldn *localDiskNodesManager) filterDisk(reqDisk, existDisk types.Disk) bool {
	if existDisk.Status != types.DiskStatusAvailable {
		return false
	}
	if existDisk.DiskType == reqDisk.DiskType &&
		existDisk.Capacity >= reqDisk.Capacity {
		return true
	}
	return false
}

func (ldn *localDiskNodesManager) diskScoreMax(reqDisk types.Disk, existDisks []*types.Disk) *types.Disk {
	if len(existDisks) == 0 {
		return nil
	}

	selDisk := existDisks[0]
	for _, existDisk := range existDisks {
		if existDisk.Capacity < selDisk.Capacity {
			selDisk = existDisk
		}
	}

	return selDisk
}

// GetReservedDiskByPVC get disk by use pvc as a label selector
// Return err if reserved disk is more than 1
func (ldn *localDiskNodesManager) GetReservedDiskByPVC(pvc string) (*types.Disk, error) {
	list, err := ldn.DiskHandler.GetLocalDiskWithLabels(labels.Set{ReservedPVCKey: pvc})
	if err != nil {
		return nil, err
	}

	// Want only one disk reserved by the pvc
	if len(list.Items) > 1 || (list.GetRemainingItemCount() != nil && *list.GetRemainingItemCount() > 0) {
		return nil, fmt.Errorf("there are more one disks reserved by pvc %s", pvc)
	}

	var reservedDisk *types.Disk
	for _, disk := range list.Items {
		reservedDisk = &types.Disk{
			AttachNode: disk.Spec.NodeName,
			Name:       disk.Name,
			DevPath:    disk.Spec.DevicePath,
			Capacity:   disk.Spec.Capacity,
			DiskType:   disk.Spec.DiskAttributes.Type,
		}
		break
	}
	return reservedDisk, nil
}

// ClaimDisk claim a localDisk by update localDisk spec.hasPartition true
func (ldn *localDiskNodesManager) ClaimDisk(name string) error {
	if name == "" {
		return fmt.Errorf("disk is empty")
	}

	ld, err := ldn.DiskHandler.GetLocalDisk(client.ObjectKey{Name: name})
	if err != nil {
		log.Errorf("failed to get localDisk %s", err.Error())
		return err
	}
	ldn.DiskHandler.For(ld)
	ldn.DiskHandler.SetPartition(true)

	return ldn.DiskHandler.Update()
}

func (ldn *localDiskNodesManager) reserve(disk *types.Disk, pvc string) error {
	if disk == nil {
		return fmt.Errorf("disk is nil")
	}

	ld, err := ldn.DiskHandler.GetLocalDisk(client.ObjectKey{Name: disk.Name})
	if err != nil {
		log.Errorf("failed to get localDisk %s", err.Error())
		return err
	}
	ldn.DiskHandler.For(ld)
	ldn.DiskHandler.SetupLabel(labels.Set{ReservedPVCKey: pvc})
	ldn.DiskHandler.ReserveDisk()

	return ldn.DiskHandler.Update()
}

// ReserveDiskForVolume reserve a localDisk by update localDisk spec.reserved and label this disk for the volume
func (ldn *localDiskNodesManager) ReserveDiskForVolume(reqDisk types.Disk, pvc string) error {
	ldn.mutex.Lock()
	defer ldn.mutex.Unlock()
	var finalSelectDisk *types.Disk
	var err error

	// lookup if a disk was reserved by the pvc
	if finalSelectDisk, err = ldn.GetReservedDiskByPVC(pvc); err != nil {
		log.WithError(err).Errorf("failed to get reserved disk for pvc %s", pvc)
		return err
	}

	// select a new disk for the pvc
	if finalSelectDisk == nil {
		finalSelectDisk, err = ldn.SelectDisk(reqDisk)
		if err != nil {
			log.WithError(err).Errorf("failed to select disk for pvc %s", pvc)
			return err
		}
	}

	// update disk status to Reserved
	if err = ldn.reserve(finalSelectDisk, pvc); err != nil {
		log.WithError(err).Errorf("failed to reserve disk %s", finalSelectDisk.Name)
		return err
	}

	return nil
}

func (ldn *localDiskNodesManager) SelectDisk(reqDisk types.Disk) (*types.Disk, error) {
	// get all disks attached on this node
	existDisks, err := ldn.GetNodeDisks(reqDisk.AttachNode)
	if err != nil {
		log.WithError(err).Errorf("failed to get node %s disks", reqDisk.AttachNode)
		return nil, err
	}

	// find out all matchable disks
	var matchDisks []*types.Disk
	for _, existDisk := range existDisks {
		if ldn.filterDisk(reqDisk, *existDisk) {
			matchDisks = append(matchDisks, existDisk)
		}
	}
	if len(matchDisks) == 0 {
		return nil, fmt.Errorf("no available disk for request: %+v", reqDisk)
	}

	// reserve one most matchable disk
	return ldn.diskScoreMax(reqDisk, matchDisks), nil
}

func (ldn *localDiskNodesManager) FilterFreeDisks(reqDisks []types.Disk) (bool, error) {
	if len(reqDisks) == 0 {
		return true, nil
	}

	// get all disks attached on this node
	existDisks, err := ldn.GetNodeDisks(reqDisks[0].AttachNode)
	if err != nil {
		log.WithError(err).Errorf("failed to get node %s disks", reqDisks[0].AttachNode)
		return false, err
	}

	for _, reqDisk := range reqDisks {
		// find out all matchable disks
		var matchDisks []*types.Disk
		for _, existDisk := range existDisks {
			if ldn.filterDisk(reqDisk, *existDisk) {
				matchDisks = append(matchDisks, existDisk)
			}
		}
		if len(matchDisks) == 0 {
			return false, fmt.Errorf("no available disk for request: %+v", reqDisk)
		}

		// Attention: if a pod claim more than one volume, filter should filter for all the volumes and find more than one disk.
		// remove disk already match for one volume.
		scoreMaxDisk := ldn.diskScoreMax(reqDisk, matchDisks)
		for i, existDisk := range existDisks {
			if existDisk.Name == scoreMaxDisk.Name {
				existDisks = append(existDisks[:i], existDisks[i+1:]...)
			}
		}
	}

	return true, nil
}

func convertToDisk(diskNode, diskName string, disk v1alpha1.Disk) *types.Disk {
	return &types.Disk{
		AttachNode: diskNode,
		Name:       diskName,
		DevPath:    disk.DevPath,
		Capacity:   disk.Capacity,
		DiskType:   disk.DiskType,
		Status:     disk.Status,
	}
}

func isSameDisk(d1, d2 types.Disk) bool {
	return reflect.DeepEqual(d1, d2)
}

//func init() {
//	// create LocalDiskNode Resource first when this module is imported
//	cli, err := New().GetClient()
//	if err != nil {
//		log.Errorf("failed to get cli %s.", err.Error())
//		return
//	}
//
//	// LocalDiskNode will be created if not exist
//	ldn, err := cli.Get(utils.GetNodeName())
//	if ldn.GetName() != "" {
//		log.Infof("LocalDiskNode %s is already exist.", ldn.GetName())
//		return
//	}
//
//	ldn, _ = localdisknode.NewBuilder().WithName(utils.GetNodeName()).
//		SetupAttachNode(utils.GetNodeName()).Build()
//	if _, err = cli.Create(ldn); err != nil {
//		log.Errorf("failed to create LocalDiskNode instance %s.", err.Error())
//		return
//	}
//
//	log.Infof("LocalDiskNode %s create successfully.", ldn.GetName())
//}
