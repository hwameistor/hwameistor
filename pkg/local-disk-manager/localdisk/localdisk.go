package localdisk

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/fields"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crmanager "sigs.k8s.io/controller-runtime/pkg/manager"

	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/builder/localdisk"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

// Controller The smallest unit to be processed here should be the disk.
// The main thing to do is how to convert the local disk into resources in the cluster
type Controller struct {
	// Mgr k8s runtime controller
	Mgr crmanager.Manager

	// Namespace is the namespace in which v1alpha1 is installed
	NameSpace string

	// NodeName is the node in which v1alpha1 is installed
	NodeName string
}

func NewController(mgr crmanager.Manager) Controller {
	return Controller{
		Mgr:       mgr,
		NameSpace: utils.GetNamespace(),
		NodeName:  utils.GetNodeName(),
	}
}

func (ctr Controller) CreateLocalDisk(ld v1alpha1.LocalDisk) error {
	log.Debugf("Create localDisk for %+v", ld)
	// Setup disk.spec.Reserved if found filesystem or partitions on it already
	if ld.Spec.HasPartition {
		ld.Spec.Reserved = true
	}
	return ctr.Mgr.GetClient().Create(context.Background(), &ld)
}

func (ctr Controller) UpdateLocalDiskAttr(newLocalDisk v1alpha1.LocalDisk) error {
	remote, err := ctr.GetLocalDisk(client.ObjectKey{Name: newLocalDisk.GetName()})
	if err != nil {
		return err
	}
	remoteOrigin := remote.DeepCopy()
	ctr.mergeLocalDiskAttr(&remote, newLocalDisk)

	// user may modify disk type for some reasons(#1130)
	// don't update disk type here if serial number exists
	if remoteOrigin.Spec.DiskAttributes.SerialNumber != "" && remoteOrigin.Spec.DiskAttributes.Type != "" {
		remote.Spec.DiskAttributes.Type = remoteOrigin.Spec.DiskAttributes.Type
	}

	return ctr.Mgr.GetClient().Patch(context.Background(), &remote, client.MergeFrom(remoteOrigin))
}

func (ctr Controller) IsAlreadyExist(ld v1alpha1.LocalDisk) bool {
	key := client.ObjectKey{Name: ld.GetName(), Namespace: ""}
	if lookLd, err := ctr.GetLocalDisk(key); err != nil {
		return false
	} else {
		return ld.GetName() == lookLd.GetName()
	}
}

func (ctr Controller) GetLocalDisk(key client.ObjectKey) (v1alpha1.LocalDisk, error) {
	ld := v1alpha1.LocalDisk{}
	if err := ctr.Mgr.GetClient().Get(context.Background(), key, &ld); err != nil {
		if errors.IsNotFound(err) {
			return ld, nil
		}
		return ld, err
	}

	return ld, nil
}

func (ctr Controller) ListLocalDisksByNode(nodeName string) ([]v1alpha1.LocalDisk, error) {
	lds := &v1alpha1.LocalDiskList{}

	if err := ctr.Mgr.GetClient().List(context.Background(), lds, &client.ListOptions{
		FieldSelector: fields.ParseSelectorOrDie("spec.nodeName=" + nodeName),
	}); err != nil {
		return nil, err
	}
	return lds.Items, nil
}

// ListLocalDiskByNodeDevicePath returns LocalDisks by given node device path
// This is should only be used when disk serial cannot be found(e.g. trigger by disk remove events)
func (ctr Controller) ListLocalDiskByNodeDevicePath(nodeName, devPath string) ([]v1alpha1.LocalDisk, error) {
	var ldList v1alpha1.LocalDiskList
	if err := ctr.Mgr.GetClient().List(context.Background(), &ldList, client.MatchingFields{"spec.nodeName/devicePath": nodeName + "/" + devPath}); err != nil {
		return nil, err
	}
	// NOTES: this logic applies only to scenarios that upgrade after an older version(<=v0.11.2) was installed
	var matchedLocalDisks []v1alpha1.LocalDisk
	for _, item := range ldList.Items {
		if strings.HasPrefix(item.Name, v1alpha1.LocalDiskObjectPrefix) {
			matchedLocalDisks = append(matchedLocalDisks, *item.DeepCopy())
		}
	}
	return matchedLocalDisks, nil
}

func (ctr Controller) ConvertDiskToLocalDisk(disk manager.DiskInfo) (ld v1alpha1.LocalDisk) {
	ld, _ = localdisk.NewBuilder().WithName(ctr.GenLocalDiskName(disk)).
		SetupState().
		SetupRaidInfo(disk.Raid).
		SetupSmartInfo(disk.Smart).
		SetupUUID(disk.GenerateUUID()).
		SetupAttribute(disk.Attribute).
		SetupPartitionInfo(disk.Partitions).
		SetupNodeName(ctr.NodeName).
		Build()
	return
}

// mergeLocalDiskAttr only merge disk self attrs(e.g., capacity, partition, attributes, etc.)
func (ctr Controller) mergeLocalDiskAttr(oldLd *v1alpha1.LocalDisk, newLd v1alpha1.LocalDisk) {
	oldLd.Spec.DiskAttributes = newLd.Spec.DiskAttributes
	oldLd.Spec.Capacity = newLd.Spec.Capacity
	oldLd.Spec.HasRAID = newLd.Spec.HasRAID
	oldLd.Spec.HasSmartInfo = newLd.Spec.HasSmartInfo
	oldLd.Spec.SmartInfo = newLd.Spec.SmartInfo
	oldLd.Spec.HasPartition = newLd.Spec.HasPartition
	oldLd.Spec.PartitionInfo = newLd.Spec.PartitionInfo
	oldLd.Spec.UUID = newLd.Spec.UUID
	oldLd.Spec.State = newLd.Spec.State
	oldLd.Spec.Major = newLd.Spec.Major
	oldLd.Spec.Minor = newLd.Spec.Minor
	oldLd.Spec.DevLinks = newLd.Spec.DevLinks

	// record historical information about where the disk was attached and the os path
	// see issue #982 for more details
	if oldLd.Spec.DevicePath != "" && oldLd.Spec.DevicePath != newLd.Spec.DevicePath {
		oldLd.Spec.PreDevicePath = oldLd.Spec.DevicePath
	}
	oldLd.Spec.DevicePath = newLd.Spec.DevicePath

	if oldLd.Spec.NodeName != "" && oldLd.Spec.NodeName != newLd.Spec.NodeName {
		oldLd.Spec.PreNodeName = oldLd.Spec.NodeName
	}
	oldLd.Spec.NodeName = newLd.Spec.NodeName
}

func (ctr Controller) GenLocalDiskName(disk manager.DiskInfo) string {
	return v1alpha1.LocalDiskObjectPrefix + disk.GenerateUUID()
}
