package hwameistor

import (
	"context"
	hoapisv1alpha1 "github.com/hwameistor/hwameistor-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
)

const (
	drbdJobPrefix = "drbd-adapter"
	drbdVersion   = "drbd-version"
)

type SettingController struct {
	client.Client
	record.EventRecorder

	clientset *kubernetes.Clientset
}

func NewSettingController(client client.Client, clientset *kubernetes.Clientset, recorder record.EventRecorder) *SettingController {
	return &SettingController{
		Client:        client,
		EventRecorder: recorder,
		clientset:     clientset,
	}
}

func (settingController *SettingController) EnableHighAvailability() (*hwameistorapi.DrbdEnableSettingRspBody, error) {
	var RspBody = &hwameistorapi.DrbdEnableSettingRspBody{}
	clusterList := &hoapisv1alpha1.ClusterList{}
	if err := settingController.Client.List(context.TODO(), clusterList); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to list clusterList")
		} else {
			log.Info("Not found the clusterList")
		}
		return RspBody, err
	}

	// for _, cluster := range clusterList.Items {
	// 	if cluster.Name == OperatorClusterName {
	// 		drbdSpec := &hoapisv1alpha1.DRBDSpec{}
	// 		drbdSpec.Enable = true
	// 		cluster.Spec.DRBD = drbdSpec

	// 		if err := settingController.Client.Update(context.TODO(), &cluster); err != nil {
	// 			return RspBody, err
	// 		}
	// 		var drbdEnableSetting = &hwameistorapi.DrbdEnableSetting{}
	// 		drbdEnableSetting.Enable = true
	// 		drbdEnableSetting.State = hwameistorapi.DrbdModuleStatusEnabled
	// 		drbdEnableSetting.Version = "v0.0.1"
	// 		RspBody.DrbdEnableSetting = drbdEnableSetting
	// 		break
	// 	}
	// }

	return RspBody, nil
}

func (settingController *SettingController) GetDRBDSetting() (*hwameistorapi.DrbdEnableSetting, error) {
	jobs, err := settingController.getDrbdJobListByNS()
	if err != nil {
		log.WithError(err).Error("Failed to getJobListByNS")
		return nil, err
	}

	var drbdSetting = &hwameistorapi.DrbdEnableSetting{}
	for _, job := range jobs {
		if label, exists := job.Labels[drbdVersion]; exists {
			log.Infof("GetDRBDSetting label = %v", label)
			drbdSetting.Version = label
		}
		log.Infof("job.Status.Succeeded = %v, job.Status.Active = %v", job.Status.Succeeded, job.Status.Active)
		if job.Status.Succeeded != 0 && (job.Status.Active == job.Status.Succeeded) {
			drbdSetting.State = hwameistorapi.DrbdModuleStatusEnabled
			drbdSetting.Enable = true
		} else {
			drbdSetting.State = hwameistorapi.DrbdModuleStatusDisabled
			drbdSetting.Enable = false
		}
	}

	return drbdSetting, nil
}

func (settingController *SettingController) getDrbdJobListByNS() ([]v1.Job, error) {
	var jobList, err = settingController.clientset.BatchV1().Jobs("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Filter jobs by prefix
	var items []v1.Job
	for _, v := range jobList.Items {
		log.Infof("getDrbdJobListByNS v.Name = %v", v.Name)
		if strings.HasPrefix(v.Name, drbdJobPrefix) {
			items = append(items, v)
		}
	}
	return items, nil
}
