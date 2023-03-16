package storage

import (
	"context"
	"fmt"

	logr "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ConfigMap struct {
	name       string
	nameSpace  string
	kubeClient *kubernetes.Clientset
}

func (cm *ConfigMap) GetName() string {
	return cm.name
}

func (cm *ConfigMap) GetNameSpace() string {
	return cm.nameSpace
}

func (cm *ConfigMap) SetKubeClient(cli *kubernetes.Clientset) *ConfigMap {
	cm.kubeClient = cli
	return cm
}

func NewConfigMap(name, ns string) *ConfigMap {
	return &ConfigMap{name: name, nameSpace: ns}
}

func (cm *ConfigMap) Read() (map[string]string, error) {
	obj, err := cm.get()
	if err != nil {
		return nil, err
	}

	return obj.Data, nil
}

func (cm *ConfigMap) ReadByKey(key string) (string, error) {
	obj, err := cm.get()
	if err != nil {
		return "", err
	}

	return obj.Data[key], nil
}

func (cm *ConfigMap) SetKV(key, value string) error {
	return cm.updateOrCreate(key, value)
}

func (cm *ConfigMap) get() (*v1.ConfigMap, error) {
	return cm.kubeClient.CoreV1().ConfigMaps(cm.nameSpace).Get(context.TODO(), cm.name, metav1.GetOptions{})
}

func (cm *ConfigMap) updateOrCreate(key, value string) error {
	obj, err := cm.get()
	if err != nil {
		// not found, create it
		if errors.IsNotFound(err) {
			return cm.create(map[string]string{
				key: value,
			})
		}
		return err
	}

	newObj := obj.DeepCopy()
	newObj.Data[key] = value

	logr.WithField("name", fmt.Sprintf("%s/%s", obj.Namespace, obj.Name)).
		Info("Update configmap")
	_, err = cm.kubeClient.CoreV1().ConfigMaps(cm.nameSpace).Update(context.TODO(), newObj, metav1.UpdateOptions{})
	return err
}

func (cm *ConfigMap) create(data map[string]string) error {
	obj := v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind: "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cm.name,
			Namespace: cm.nameSpace,
		},
		Data: data,
	}

	logr.WithField("name", fmt.Sprintf("%s/%s", obj.Namespace, obj.Name)).
		Info("Create configmap")
	_, err := cm.kubeClient.CoreV1().ConfigMaps(obj.Namespace).Create(context.TODO(), &obj, metav1.CreateOptions{})
	return err
}
