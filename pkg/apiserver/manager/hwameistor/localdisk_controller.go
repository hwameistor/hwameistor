package hwameistor

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
)

type LocalDiskController struct {
	client.Client
	record.EventRecorder
}

func NewLocalDiskController(client client.Client, recorder record.EventRecorder) *LocalDiskController {
	return &LocalDiskController{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (ldController *LocalDiskController) ListLocalDisk() (*apisv1alpha1.LocalDiskList, error) {
	diskList := &apisv1alpha1.LocalDiskList{}
	if err := ldController.Client.List(context.TODO(), diskList); err != nil {
		log.WithError(err).Error("Failed to list LocalDisks")
	}
	return diskList, nil
}

func (ldController *LocalDiskController) GetLocalDisk(key client.ObjectKey) (*apisv1alpha1.LocalDisk, error) {
	disk := &apisv1alpha1.LocalDisk{}
	if err := ldController.Client.Get(context.TODO(), key, disk); err != nil {
		if !errors.IsNotFound(err) {
			log.WithError(err).Error("Failed to query diskume")
		} else {
			log.Info("Not found the diskume")
		}
		return nil, err
	}
	return disk, nil
}

func (ldController *LocalDiskController) AddLocalDiskClaim(node, diskType, owner string) error {
	claim := &apisv1alpha1.LocalDiskClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(fmt.Sprintf("%s-%s-claim", node, diskType)),
		},
		Spec: apisv1alpha1.LocalDiskClaimSpec{
			Owner:    owner,
			NodeName: node,
			Description: apisv1alpha1.DiskClaimDescription{
				DiskType: diskType,
			},
		},
	}

	if err := ldController.Create(context.TODO(), claim); err != nil {
		log.WithError(err).Error("Fail to create LocalDiskClaim")
		return err
	}

	return nil
}
