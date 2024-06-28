package dataloader

import (
	"context"
	"fmt"
	mykube "github.com/hwameistor/hwameistor/pkg/utils/kubernetes"
	"github.com/hwameistor/hwameistor/pkg/webhook"
	"github.com/hwameistor/hwameistor/pkg/webhook/config"
	"github.com/sirupsen/logrus"
	admission "k8s.io/api/admission/v1beta1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"sync"
)

const (
	webhookName   = "hwameistor-dataloader-webhook"
	containerName = "hwameistor-dataloader"
)

var defaultContainerTemplate = corev1.Container{
	Name:            containerName,
	ImagePullPolicy: corev1.PullIfNotPresent,
	Env: []corev1.EnvVar{
		{
			Name:  "NAMESPACE",
			Value: "",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  "MY_NODENAME",
			Value: "",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "spec.nodeName",
				},
			},
		},
	},
}

type patchDataLoaderContainer struct {
	client         *kubernetes.Clientset
	once           sync.Once
	containerImage string
}

func NewDataLoaderContainerWebHook() webhook.MutateAdmissionWebhook {
	return &patchDataLoaderContainer{}
}

func (p *patchDataLoaderContainer) ResourceNeedHandle(req admission.AdmissionReview) (bool, error) {
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

	datasets, err := p.getDatasetVolumesFromPod(pod)
	if err != nil {
		// case1: if FailurePolicy is Ignore and the bounded pv is not found, skip this pod
		if errors.IsNotFound(err) && *config.GetFailurePolicy() == admissionregistrationv1.Ignore {
			logrus.WithFields(logCtx).Infof("ignore pod %s because of pvc or bounded pv is not found"+
				" and FailurePolicy is %s", pod.Name, admissionregistrationv1.Ignore)
			return false, nil
		}

		// default: reject!
		logrus.WithFields(logCtx).WithError(err).Error("failed to judge pod use hwameistor dataset volume or not")
		return false, err
	}

	// return directly if one hwameistor dataset volume found
	if len(datasets) > 0 {
		logrus.WithFields(logCtx).Infof("found hwameistor dataset volume %v", datasets)
		return true, nil
	}

	return false, nil
}

func (p *patchDataLoaderContainer) Name() string {
	return webhookName
}

func (p *patchDataLoaderContainer) Init(option webhook.ServerOption) {
	p.once.Do(func() {
		client, err := mykube.NewClientSet()
		if err != nil {
			logrus.WithError(err).Fatal("failed to create client for kubernetes")
		}
		p.client = client
		image, ok := os.LookupEnv("DATALOADER_IMAGE")
		if !ok {
			logrus.Info("not found DATALOADER_IMAGE")
		}
		p.containerImage = image
	})
}

func (p *patchDataLoaderContainer) getPersistentVolumeNameByPVC(ns, pvcName string) (string, error) {
	pvc, err := p.client.CoreV1().PersistentVolumeClaims(ns).Get(context.Background(), pvcName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return pvc.Spec.VolumeName, nil
}

func (p *patchDataLoaderContainer) getPersistentVolumeByName(pvName string) (*corev1.PersistentVolume, error) {
	pv, err := p.client.CoreV1().PersistentVolumes().Get(context.Background(), pvName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pv, nil
}

func (p *patchDataLoaderContainer) Mutate(review admission.AdmissionReview) ([]webhook.PatchOperation, error) {
	// Parse the Pod object.
	raw := review.Request.Object.Raw
	pod := corev1.Pod{}
	if _, _, err := webhook.UniversalDeserializer.Decode(raw, nil, &pod); err != nil {
		logrus.WithError(err).Error("could not deserialize pod object")
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}
	logrus.Infof("mutating resource %s/%s", pod.Namespace, pod.GetGenerateName())

	patchContainer, err := p.genDataloaderPatchForPod(pod)
	if err != nil {
		logrus.WithError(err).Errorf("failed to generate dataloader container patch value")
		return nil, err
	}
	orderedInitContainers := reorderPodInitContainers(pod.Spec.InitContainers, patchContainer)

	logrus.Debugf("patch value for dataloader: %+v", orderedInitContainers)

	var patches []webhook.PatchOperation
	patches = append(patches, webhook.PatchOperation{
		Operation: "add",
		Path:      "/spec/initContainers",
		Value:     orderedInitContainers,
	})
	return patches, nil
}

func (p *patchDataLoaderContainer) genDataloaderPatchForPod(pod corev1.Pod) (*corev1.Container, error) {
	datasets, err := p.getDatasetVolumesFromPod(pod)
	if err != nil {
		return nil, err
	}

	if len(datasets) == 0 {
		// this shouldn't happen!
		logrus.Warnf("no hwameistor dataset volume found in pod %s", pod.Name)
		return nil, fmt.Errorf("no hwameistor dataset volume found in pod %s", pod.Name)
	}

	initContainer := defaultContainerTemplate.DeepCopy()
	initContainer.Image = p.containerImage
	initContainer.Env = append(initContainer.Env, corev1.EnvVar{Name: "DATASET_NAME", Value: datasets[0]})

	return initContainer, nil
}

func (p *patchDataLoaderContainer) getDatasetVolumesFromPod(pod corev1.Pod) ([]string, error) {
	var datasetVolumes []string
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}

		pvName, err := p.getPersistentVolumeNameByPVC(pod.Namespace, volume.PersistentVolumeClaim.ClaimName)
		if err != nil {
			return nil, err
		}

		if pvName == "" {
			continue
		}

		pv, err := p.getPersistentVolumeByName(pvName)
		if err != nil {
			return nil, err
		}

		if pv.Annotations["hwameistor.io/acceleration-dataset"] == "true" {
			datasetVolumes = append(datasetVolumes, pvName)
		}
	}

	return datasetVolumes, nil
}

func reorderPodInitContainers(originInitContainers []corev1.Container, dataloaderContainer *corev1.Container) []corev1.Container {
	if len(originInitContainers) == 0 {
		return []corev1.Container{*dataloaderContainer}
	}

	// avoid patching twice
	var orderedInitContainers []corev1.Container
	for _, container := range originInitContainers {
		if container.Name == dataloaderContainer.Name {
			return originInitContainers
		}
	}

	orderedInitContainers = append(orderedInitContainers, *dataloaderContainer)
	return orderedInitContainers
}
