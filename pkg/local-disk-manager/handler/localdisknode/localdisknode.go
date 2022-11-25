package localdisknode

import (
	"context"
	"reflect"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	localdisk2 "github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (n *DiskNodeHandler) UpdateDiskLists(updateDisks, removedDisks map[string]v1alpha1.Disk) {
	// remove disk
	for name, removeDisk := range removedDisks {
		delete(n.diskNode.Status.Disks, name)
		if removeDisk.Status != string(v1alpha1.LocalDiskUnclaimed) && removeDisk.Status != string(v1alpha1.LocalDiskReleased) {
			n.EventRecorder.Eventf(n.diskNode, v1.EventTypeWarning, "RemoveDisk", ""+
				"Disk %s is removed but state is %s, disk last info: %+v", removeDisk.DevPath, removeDisk.Status, removeDisk)
		} else {
			n.EventRecorder.Eventf(n.diskNode, v1.EventTypeNormal, "RemoveDisk", "Remove disk %s", removeDisk.DevPath)
		}
	}

	// update disk
	if n.diskNode.Status.Disks == nil {
		n.diskNode.Status.Disks = make(map[string]v1alpha1.Disk, len(updateDisks))
	}
	for name, updateDisk := range updateDisks {
		oldDisk, exist := n.diskNode.Status.Disks[name]
		if !exist {
			n.EventRecorder.Eventf(n.diskNode, v1.EventTypeNormal, "AddDisk", "Add new disk %s", updateDisk.DevPath)
		} else {
			n.EventRecorder.Eventf(n.diskNode, v1.EventTypeNormal, "UpdateDisk", ""+
				"Disk %s old info: %+v", name, oldDisk)
		}

		n.diskNode.Status.Disks[name] = updateDisk
	}
}

func (n *DiskNodeHandler) UpdateDiskStats() {
	n.diskNode.Status.TotalDisk = 0
	n.diskNode.Status.AllocatableDisk = 0
	for _, disk := range n.Disks() {
		n.diskNode.Status.TotalDisk++
		if disk.Status == string(v1alpha1.LocalDiskUnclaimed) ||
			disk.Status == string(v1alpha1.LocalDiskReleased) {
			n.diskNode.Status.AllocatableDisk++
		}
	}
}

func (n *DiskNodeHandler) Disks() map[string]v1alpha1.Disk {
	return n.diskNode.Status.Disks
}

func (n *DiskNodeHandler) ListNodeDisks() (map[string]v1alpha1.Disk, error) {
	lds, err := n.diskHandler.ListNodeLocalDisk(n.diskNode.Spec.AttachNode)
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
