package alerter

import (
	"fmt"
	"strings"

	localstorageinformers "github.com/hwameistor/local-storage/pkg/apis/client/informers/externalversions"
	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	log "github.com/sirupsen/logrus"
)

type diskAlerter struct {
	logger *log.Entry

	moduleName string

	queue workqueue.Interface
}

func newDiskAlerter() Alerter {
	return &diskAlerter{
		logger:     log.WithField("Module", ModuleDisk),
		moduleName: ModuleDisk,
		queue:      workqueue.New(),
	}
}

// Run disk alerter
func (alt *diskAlerter) Run(informerFactory localstorageinformers.SharedInformerFactory, stopCh <-chan struct{}) {
	informer := informerFactory.Localstorage().V1alpha1().PhysicalDisks().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    alt.onAdd,
		UpdateFunc: alt.onUpdate,
		DeleteFunc: alt.onDelete,
	})
	go informer.Run(stopCh)
	go alt.process(stopCh)
}

func (alt *diskAlerter) onAdd(obj interface{}) {
	disk, _ := obj.(*localstoragev1alpha1.PhysicalDisk)
	alt.logger.WithFields(log.Fields{"disk": disk.Spec.SerialNumber}).Debug("Checking for a disk just added")

	alt.queue.Add(disk)
}

func (alt *diskAlerter) onUpdate(oldObj, newObj interface{}) {
	disk, _ := newObj.(*localstoragev1alpha1.PhysicalDisk)
	alt.logger.WithFields(log.Fields{"disk": disk.Spec.SerialNumber}).Debug("Checking for a disk just updated")

	alt.queue.Add(disk)
}

func (alt *diskAlerter) onDelete(obj interface{}) {
	disk, _ := obj.(*localstoragev1alpha1.PhysicalDisk)
	alt.logger.WithFields(log.Fields{"disk": disk.Spec.SerialNumber}).Debug("Checking for a disk just deleted")

	alt.queue.Add(disk)
}

func (alt *diskAlerter) process(stopCh <-chan struct{}) {
	alt.logger.Debug("Disk Alerter is working now")

	go func() {
		for {
			obj, shutdown := alt.queue.Get()
			if shutdown {
				alt.logger.Debug("Stop the disk alerter worker")
				break
			}
			disk, ok := obj.(*localstoragev1alpha1.PhysicalDisk)
			if ok && disk != nil {
				alt.predict(disk)
			}

			alt.queue.Done(obj)
		}
	}()

	<-stopCh
	alt.queue.ShutDown()
}

func (alt *diskAlerter) createAlert(severity int, diskName string, event string, details string) {
	createAlert(&localstoragev1alpha1.LocalStorageAlert{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-%s", strings.ToLower(alt.moduleName), diskName, genTimeStampString()),
		},
		Spec: localstoragev1alpha1.LocalStorageAlertSpec{
			Severity: severity,
			Module:   alt.moduleName,
			Resource: diskName,
			Event:    event,
			Details:  details,
		},
	})
}

func (alt *diskAlerter) predict(disk *localstoragev1alpha1.PhysicalDisk) {

	if disk == nil {
		return
	}

	alt.checkForOnline(disk)

	if disk.Status.Online {
		return
	}

	alt.checkForSmartMetrics(disk)
}

func (alt *diskAlerter) checkForOnline(disk *localstoragev1alpha1.PhysicalDisk) {
	if disk.Status.Online {
		return
	}

	alt.createAlert(
		SeverityCritical,
		disk.Name,
		DiskEventOffline,
		"",
	)
}

func (alt *diskAlerter) checkForSmartMetrics(disk *localstoragev1alpha1.PhysicalDisk) {
	if disk.Status.SmartCheck == nil || disk.Status.SmartCheck.SmartStatus == nil {
		return
	}

	if !disk.Status.SmartCheck.SmartStatus.Passed {
		alt.createAlert(
			SeverityCritical,
			disk.Name,
			DiskEventSmartCheckFailed,
			"",
		)
		return
	}

	if disk.Status.SmartCheck.Temperature != nil {
		if disk.Status.SmartCheck.Temperature.Current > DiskTemperatureThreshold {
			alt.createAlert(
				SeverityWarning,
				disk.Name,
				DiskEventHighTemperature,
				fmt.Sprintf("current temperature is %d", disk.Status.SmartCheck.Temperature.Current),
			)
		}
	}

	if disk.Status.SmartCheck.ATASmartHealthStatus != nil {
		attributes := disk.Status.SmartCheck.ATASmartHealthStatus.AttributesTable
		attributesMap := map[int64]*localstoragev1alpha1.ATASmartHealthAttribute{}
		for i := range attributes {
			attributesMap[attributes[i].ID] = &attributes[i]
		}
		messages := []string{}
		for _, attrToCheck := range ataDiskSmartAttributesToCheck {
			if attr, exists := attributesMap[attrToCheck.id]; exists {
				if attr.WhenFailed != "" {
					messages = append(messages, fmt.Sprintf("%s failed %s, need to check", attr.Name, attr.WhenFailed))
				}
				if attr.Raw != nil {
					if attrToCheck.isSmallerBetter {
						if attr.Raw.Value > attrToCheck.threshold {
							messages = append(messages, fmt.Sprintf("%s value(%d) is abnormal, need to check", attr.Name, attr.Raw.Value))
						}
					} else {
						if attr.Raw.Value < attrToCheck.threshold {
							messages = append(messages, fmt.Sprintf("%s value(%d) is abnormal, need to check", attr.Name, attr.Raw.Value))
						}
					}
				}
			}
		}

		if len(messages) > 0 {
			alt.createAlert(
				SeverityWarning,
				disk.Name,
				DiskEventSmartAttributesAbnormal,
				strings.Join(messages, "\n"),
			)
		}
	}

	if disk.Status.SmartCheck.SCSISmartHealthStatus != nil {
		status := disk.Status.SmartCheck.SCSISmartHealthStatus
		messages := []string{}
		if status.Read != nil {
			if status.Read.TotalErrorsCorrected > scsiDiskSmartAttributeTotalCorrectedErrorsThreshold {
				messages = append(messages, fmt.Sprintf("Too many total corrected Read/IO errors (%d)", status.Read.TotalErrorsCorrected))
			}
			if status.Read.TotalUncorrectedErrors > scsiDiskSmartAttributeTotalUncorrectedErrorsThreshold {
				messages = append(messages, fmt.Sprintf("Too many total uncorrected Read/IO errors (%d)", status.Read.TotalUncorrectedErrors))
			}
		}
		if status.Write != nil {
			if status.Write.TotalErrorsCorrected > scsiDiskSmartAttributeTotalCorrectedErrorsThreshold {
				messages = append(messages, fmt.Sprintf("Too many total corrected Write/IO errors (%d)", status.Write.TotalErrorsCorrected))
			}
			if status.Write.TotalUncorrectedErrors > scsiDiskSmartAttributeTotalUncorrectedErrorsThreshold {
				messages = append(messages, fmt.Sprintf("Too many total uncorrected Write/IO errors (%d)", status.Write.TotalUncorrectedErrors))
			}
		}

		if len(messages) > 0 {
			alt.createAlert(
				SeverityWarning,
				disk.Name,
				DiskEventSmartAttributesAbnormal,
				strings.Join(messages, "\n"),
			)
		}
	}

	if disk.Status.SmartCheck.NVMeSmartHealthStatus != nil {
		status := disk.Status.SmartCheck.NVMeSmartHealthStatus
		messages := []string{}
		if status.CriticalWarning > nvmeDiskSmartAttributeCriticalWarningThreshold {
			messages = append(messages, fmt.Sprintf("Too many critical warnings (%d)", status.CriticalWarning))
		}
		if status.MediaErrors > nvmeDiskSmartAttributeMediaErrorsThreshold {
			messages = append(messages, fmt.Sprintf("Too many media errors (%d)", status.MediaErrors))
		}
		if status.NumErrLogEntries > nvmeDiskSmartAttributeNumErrLogEntriesThreshold {
			messages = append(messages, fmt.Sprintf("Too many error log entries (%d)", status.NumErrLogEntries))
		}
		if status.UnsafeShutdowns > nvmeDiskSmartAttributeUnsafeShutdownsThreshold {
			messages = append(messages, fmt.Sprintf("Too many unsafe shutdowns (%d)", status.UnsafeShutdowns))
		}

		if len(messages) > 0 {
			alt.createAlert(
				SeverityWarning,
				disk.Name,
				DiskEventSmartAttributesAbnormal,
				strings.Join(messages, "\n"),
			)
		}
	}

}
