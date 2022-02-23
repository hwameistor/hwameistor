package healths

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// consts
const (
	HealthCheckIntervalDefault = 30 * time.Minute
)

//PhyDisksHealthManager interface
type PhyDisksHealthManager interface {
	Run(stopCh <-chan struct{})
}

type diskHealthManager struct {
	nodeName string

	apiClient client.Client

	checker DiskChecker

	logger *log.Entry
}

// NewDiskHealthManager a disk health manager instance
func NewDiskHealthManager(name string, cli client.Client) PhyDisksHealthManager {
	return &diskHealthManager{
		nodeName:  name,
		apiClient: cli,
		checker:   NewSmartCtl(),
		logger:    log.WithField("Module", "PhyDiskHealthManager"),
	}
}

func (dm *diskHealthManager) Run(stopCh <-chan struct{}) {

	go dm.start(stopCh)
}

func (dm *diskHealthManager) start(stopCh <-chan struct{}) {

	dm.logger.Debug("Starting the physical disk health manager")

	// delay for a random second to avoid disk check storm
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

	dm.checkDiskHealths()
	for {
		select {
		case <-time.After(HealthCheckIntervalDefault):
			dm.checkDiskHealths()
		case <-stopCh:
			dm.logger.Debug("Exit the physical disk health manager")
			return
		}
	}
}

func (dm *diskHealthManager) checkDiskHealths() {
	dm.logger.Debug("Start to check health for all local disks")

	devices, err := dm.checker.GetLocalDisksAll()
	if err != nil {
		dm.logger.WithError(err).Error("Failed to scan disks on node")
		return
	}

	diskCrdList := &localstoragev1alpha1.PhysicalDiskList{}
	if err := dm.apiClient.List(context.TODO(), diskCrdList); err != nil {
		dm.logger.WithError(err).Error("Failed to list PhysicalDisks")
		return
	}
	diskCrds := map[string]*localstoragev1alpha1.PhysicalDisk{}
	for i := range diskCrdList.Items {
		if diskCrdList.Items[i].Spec.NodeName == dm.nodeName {
			diskCrds[diskCrdList.Items[i].Name] = &diskCrdList.Items[i]
		}
	}

	for i := range devices {
		logCtx := dm.logger.WithField("device", devices[i])
		logCtx.Debug("Checking disk health for")
		result, err := dm.checker.CheckHealthForLocalDisk(&devices[i])
		if err != nil {
			logCtx.WithError(err).Error("Failed to check disk health")
			continue
		}
		if result.Device == nil {
			logCtx.Error("No valid device info found")
			continue
		}
		if result.IsVirtualDisk() {
			result.SerailNumber = fmt.Sprintf("%s-%s", dm.nodeName, filepath.Base(result.Device.Name))
		}
		crd, exists := diskCrds[result.SerailNumber]
		if exists {
			crd.Status.Online = true
			crd.Status.SmartCheck = &result.SmartCheck
			crd.Status.SmartCheck.LastTime = &metav1.Time{Time: time.Now()}
			delete(diskCrds, result.SerailNumber)
		} else {
			crd = &localstoragev1alpha1.PhysicalDisk{
				ObjectMeta: metav1.ObjectMeta{Name: result.SerailNumber},
				Spec: localstoragev1alpha1.PhysicalDiskSpec{
					NodeName:     dm.nodeName,
					Vendor:       result.Vendor,
					Product:      result.Product,
					ModelName:    result.ModelName,
					SerialNumber: result.SerailNumber,
					RotationRate: result.RotationRate,
					Type:         result.Device.Type,
					DevicePath:   result.Device.Name,
					Protocol:     result.Device.Protocol,
				},
				Status: localstoragev1alpha1.PhysicalDiskStatus{
					Online:     true,
					SmartCheck: &result.SmartCheck,
				},
			}
			crd.Status.SmartCheck.LastTime = &metav1.Time{Time: time.Now()}
			if result.FormFactor != nil {
				crd.Spec.FormFactor = result.FormFactor.Name
			}
			if result.UserCapacity != nil {
				crd.Spec.Capacity = result.UserCapacity.Bytes
			}
			if result.PCIVendor != nil {
				crd.Spec.PCIVendorID = result.PCIVendor.String()
			}
			if strings.Contains(result.Device.Type, ",") {
				crd.Spec.IsRAID = true
			} else {
				crd.Spec.IsRAID = false
			}
			if result.ATASmartHealthStatus == nil && result.SCSISmartHealthStatus == nil && result.NVMeSmartHealthStatus == nil {
				crd.Spec.SmartSupport = false
			} else {
				crd.Spec.SmartSupport = true
			}

			if err := dm.apiClient.Create(context.TODO(), crd); err != nil {
				logCtx.WithError(err).Error("Failed to create a PhysicalDisk")
			} else {
				logCtx.Debug("Created a PhysicalDisk with health check info")
			}
		}
		if err := dm.apiClient.Status().Update(context.TODO(), crd); err != nil {
			logCtx.WithError(err).Error("Failed to update the health info for PhysicalDisk")
		} else {
			logCtx.Debug("Updated the health info for PhysicalDisk")
		}
	}

	for sn := range diskCrds {
		logCtx := dm.logger.WithField("device", diskCrds[sn].Name)
		diskCrds[sn].Status.Online = false
		if err := dm.apiClient.Status().Update(context.TODO(), diskCrds[sn]); err != nil {
			logCtx.WithError(err).Error("Failed to set a PhysicalDisk to offline")
		} else {
			logCtx.Debug("Updated a PhysicalDisk to offline")
		}
	}

}
