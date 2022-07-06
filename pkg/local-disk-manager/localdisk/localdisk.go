package localdisk

import (
	"context"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/builder/localdisk"

	ldm "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-disk-manager/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crmanager "sigs.k8s.io/controller-runtime/pkg/manager"
)

// Controller The smallest unit to be processed here should be the disk.
// The main thing to do is how to convert the local disk into resources in the cluster
type Controller struct {
	// Mgr k8s runtime controller
	Mgr crmanager.Manager

	// Namespace is the namespace in which ldm is installed
	NameSpace string

	// NodeName is the node in which ldm is installed
	NodeName string
}

// NewController
func NewController(mgr crmanager.Manager) Controller {
	return Controller{
		Mgr:       mgr,
		NameSpace: utils.GetNamespace(),
		NodeName:  utils.GetNodeName(),
	}
}

// CreateLocalDisk
func (ctr Controller) CreateLocalDisk(ld ldm.LocalDisk) error {
	log.Debugf("Create LocalDisk for %+v", ld)
	return ctr.Mgr.GetClient().Create(context.Background(), &ld)
}

// CreateLocalDisk
func (ctr Controller) UpdateLocalDisk(ld ldm.LocalDisk) error {
	newLd := ld.DeepCopy()
	key := client.ObjectKey{Name: ld.GetName(), Namespace: ""}

	oldLd, err := ctr.GetLocalDisk(key)
	if err != nil {
		return err
	}

	// TODO: merge old disk and new disk
	ctr.mergerLocalDisk(oldLd, newLd)
	return ctr.Mgr.GetClient().Update(context.Background(), newLd)
}

// IsAlreadyExist
func (ctr Controller) IsAlreadyExist(ld ldm.LocalDisk) bool {
	key := client.ObjectKey{Name: ld.GetName(), Namespace: ""}
	if lookLd, err := ctr.GetLocalDisk(key); err != nil {
		return false
	} else {
		return ld.GetName() == lookLd.GetName()
	}
}

// GetLocalDisk
func (ctr Controller) GetLocalDisk(key client.ObjectKey) (ldm.LocalDisk, error) {
	ld := ldm.LocalDisk{}
	if err := ctr.Mgr.GetClient().Get(context.Background(), key, &ld); err != nil {
		if errors.IsNotFound(err) {
			return ld, nil
		}
		return ld, err
	}

	return ld, nil
}

// ConvertDiskToLocalDisk
func (ctr Controller) ConvertDiskToLocalDisk(disk manager.DiskInfo) (ld ldm.LocalDisk) {
	ld, _ = localdisk.NewBuilder().WithName(ctr.GenLocalDiskName(disk)).
		SetupState().
		SetupRaidInfo(disk.Raid).
		SetupUUID(disk.GenerateUUID()).
		SetupAttribute(disk.Attribute).
		SetupPartitionInfo(disk.Partitions).
		SetupNodeName(utils.ConvertNodeName(ctr.NodeName)).
		GenerateStatus().
		Build()
	return
}

func (ctr Controller) mergerLocalDisk(oldLd ldm.LocalDisk, newLd *ldm.LocalDisk) {
	newLd.Status = oldLd.Status
	newLd.TypeMeta = oldLd.TypeMeta
	newLd.ObjectMeta = oldLd.ObjectMeta
}

func (ctr Controller) GenLocalDiskName(disk manager.DiskInfo) string {
	return utils.ConvertNodeName(ctr.NodeName) + "-" + disk.Name
}
