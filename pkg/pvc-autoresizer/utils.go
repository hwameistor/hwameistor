package autoresizer

import (
	"context"

	hwameistorv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ListResizePolicies(cli client.Client) ([]hwameistorv1alpha1.ResizePolicy, error) {
	resizepolicyList := &hwameistorv1alpha1.ResizePolicyList{}
	err := cli.List(context.TODO(), resizepolicyList)

	return resizepolicyList.Items, err
}