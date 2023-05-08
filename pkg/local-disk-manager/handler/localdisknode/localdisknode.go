package localdisknode

import (
	"context"
	"reflect"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	localdisk2 "github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
)

type DiskNodeHandler struct {
	client.Client
	record.EventRecorder
	diskNode    *v1alpha1.LocalDiskNode
	diskHandler *localdisk2.Handler
}

func NewDiskNodeHelper(cli client.Client, recorder record.EventRecorder) *DiskNodeHandler {
	return &DiskNodeHandler{
		Client:        cli,
		EventRecorder: recorder,
		diskHandler:   localdisk2.NewLocalDiskHandler(cli, recorder),
	}
}

func (n *DiskNodeHandler) For(name types.NamespacedName) error {
	ldn := &v1alpha1.LocalDiskNode{}
	err := n.Get(context.Background(), name, ldn)
	if err != nil {
		return err
	}

	n.diskNode = ldn
	return nil
}

func (n *DiskNodeHandler) UpdateStatus() error {
	err := n.Update(context.Background(), n.diskNode)
	if err != nil {
		log.WithError(err).Errorf("failed to update disks")
	} else {
		log.Infof("Update disks successfully")
	}

	return err
}

func (n *DiskNodeHandler) UpdateDiskStats() {
	n.diskNode.Status.TotalDisk = 0
	n.diskNode.Status.FreeDisk = 0
	for _, disk := range n.Disks() {
		n.diskNode.Status.TotalDisk++
		if disk.Status == string(v1alpha1.LocalDiskUnclaimed) ||
			disk.Status == string(v1alpha1.LocalDiskReleased) {
			n.diskNode.Status.FreeDisk++
		}
	}
}

func (n *DiskNodeHandler) Disks() map[string]v1alpha1.Disk {
	var disks map[string]v1alpha1.Disk
	for _, pool := range n.diskNode.Status.Pools {
		for _, disk := range pool.Disks {
			disks[disk.DevPath] = v1alpha1.Disk{
				DevPath:  disk.DevPath,
				Capacity: disk.CapacityBytes,
				DiskType: disk.Class,
				Status:   string(disk.State),
			}
		}
	}
	return disks
}

func (n *DiskNodeHandler) ListNodeDisks() (map[string]v1alpha1.Disk, error) {
	lds, err := n.diskHandler.ListNodeLocalDisk(n.diskNode.Spec.NodeName)
	if err != nil {
		return nil, err
	}

	disks := map[string]v1alpha1.Disk{}
	for _, ld := range lds.Items {
		disks[ld.GetName()] = convertToDisk(ld)
	}
	return disks, nil
}

// IsSameDisk judge the disk in LocalDiskNode is same as disk in localDisk
func (n *DiskNodeHandler) IsSameDisk(name string, newDisk v1alpha1.Disk) bool {
	oldDisk := n.Disks()[name]

	return reflect.DeepEqual(&oldDisk, &newDisk)
}

func convertToDisk(ld v1alpha1.LocalDisk) v1alpha1.Disk {
	return v1alpha1.Disk{
		DevPath:  ld.Spec.DevicePath,
		Capacity: ld.Spec.Capacity,
		DiskType: ld.Spec.DiskAttributes.Type,
		Status:   string(ld.Status.State),
	}
}
