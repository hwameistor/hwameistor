package scheduler

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	admission "k8s.io/api/admission/v1beta1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	mykube "github.com/hwameistor/hwameistor/pkg/utils/kubernetes"
	"github.com/hwameistor/hwameistor/pkg/webhook"
	"github.com/hwameistor/hwameistor/pkg/webhook/config"
)

var (
	once                 = sync.Once{}
	hwameiStorSuffix     = "hwameistor.io"
	name                 = "hwameistor-scheduler-webhooks"
	defaultSchedulerName = "hwameistor-scheduler"
)

type patchSchedulerName struct {
	client        *kubernetes.Clientset
	schedulerName string
}

func NewPatchSchedulerWebHook() *patchSchedulerName {
	return &patchSchedulerName{}
}

func (p *patchSchedulerName) Init(s webhook.ServerOption) {
	once.Do(func() {
		client, err := mykube.NewClientSet()
		if err != nil {
			os.Exit(1)
		}
		p.client = client
		if s.SchedulerName == "" {
			p.schedulerName = defaultSchedulerName
		} else {
			p.schedulerName = s.SchedulerName
		}

	})
}

// Mutate checkout if scheduler name is set, patch it if not when use hwameistor volume
func (p *patchSchedulerName) Mutate(review admission.AdmissionReview) ([]webhook.PatchOperation, error) {
	// Parse the Pod object.
	raw := review.Request.Object.Raw
	pod := corev1.Pod{}
	if _, _, err := webhook.UniversalDeserializer.Decode(raw, nil, &pod); err != nil {
		logrus.WithError(err).Errorf("could not deserialize pod object: %v", err)
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}
	logrus.Infof("mutating resource %s/%s", pod.Namespace, pod.GetGenerateName())

	var patches []webhook.PatchOperation
	scheduler := pod.Spec.SchedulerName
	switch scheduler {
	case "":
		logrus.Infof("pod %s use hwameistor volume but not set hwameistor scheduler, "+
			"auto patch scheduler name to this pod", pod.Name)
		patches = append(patches, webhook.PatchOperation{
			Operation: "add",
			Path:      "/spec/schedulerName",
			Value:     p.schedulerName,
		})
	case p.schedulerName:
		logrus.Infof("pod %s scheduler name is same with configured %s in webhook config", pod.GetGenerateName(), p.schedulerName)
	default:
		logrus.Infof("change pod %s scheduler name from %s to %s", pod.GetGenerateName(), scheduler, p.schedulerName)
		patches = append(patches, webhook.PatchOperation{
			Operation: "replace",
			Path:      "/spec/schedulerName",
			Value:     p.schedulerName,
		})
	}
	return patches, nil
}

func (p *patchSchedulerName) Name() string {
	return name
}

func (p *patchSchedulerName) ResourceNeedHandle(req admission.AdmissionReview) (bool, error) {
	if req.Request.Resource != webhook.PodResource {
		logrus.Infof("skip handle resource %v, only handler pod resource", req.Request.Resource)
		return false, nil
	}

	// Parse the Pod object.
	raw := req.Request.Object.Raw
	pod := corev1.Pod{}
	if _, _, err := webhook.UniversalDeserializer.Decode(raw, nil, &pod); err != nil {
		logrus.WithError(err).Errorf("could not deserialize pod object: %v", err)
		return false, fmt.Errorf("could not deserialize pod object: %v", err)
	}
	pod.SetNamespace(req.Request.Namespace)
	logCtx := logrus.Fields{"NameSpace": pod.GetNamespace(), "Pod": pod.GetGenerateName()}

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		ok, err := p.IsHwameiStorVolume(pod.GetNamespace(), volume.PersistentVolumeClaim.ClaimName)
		if err != nil {
			// case1: if FailurePolicy is Ignore and the StorageClass is not found, skip this volume
			if errors.IsNotFound(err) && *config.GetFailurePolicy() == admissionregistrationv1.Ignore {
				logrus.WithFields(logCtx).Infof("ignore volume %s because of pvc or storageclass is not found"+
					" and FailurePolicy is %s", volume.PersistentVolumeClaim.ClaimName, admissionregistrationv1.Ignore)
				continue
			}

			// default: reject!
			logrus.WithFields(logCtx).WithError(err).Error("failed to judge volume is hwameistor volume or not")
			return false, err
		}

		// return directly if one hwameistor volume found
		if ok {
			logrus.WithFields(logCtx).Infof("found hwameistor volume %s", volume.PersistentVolumeClaim.ClaimName)
			return ok, nil
		}
	}

	return false, nil
}

func (p *patchSchedulerName) IsHwameiStorVolume(ns, pvc string) (bool, error) {
	logCtx := logrus.Fields{"NameSpace": ns, "PVC": pvc}

	// Step 1: get storageclass by pvc
	sc, err := p.getStorageClassByPVC(ns, pvc)
	if err != nil {
		logrus.WithFields(logCtx).WithError(err).Error("failed to get storageclass")
		return false, err
	}

	// skip static volume
	if sc == "" {
		return false, nil
	}

	// Step 2: compare provisioner name with hwameistor.io
	provisioner, err := p.getProvisionerByStorageClass(sc)
	if err != nil {
		logrus.WithFields(logCtx).WithError(err).Error("failed to get provisioner")
		return false, err
	}
	return strings.HasSuffix(provisioner, hwameiStorSuffix), nil
}

// getStorageClassByPVC return sc name if set, else return empty if it is a static volume
func (p *patchSchedulerName) getStorageClassByPVC(ns, pvcName string) (string, error) {
	pvc, err := p.client.CoreV1().PersistentVolumeClaims(ns).Get(context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if pvc.Spec.StorageClassName == nil {
		return "", nil
	}
	return *pvc.Spec.StorageClassName, nil
}

func (p *patchSchedulerName) getProvisionerByStorageClass(scName string) (string, error) {
	sc, err := p.client.StorageV1().StorageClasses().Get(context.Background(), scName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return sc.Provisioner, nil
}
