package hwameistor

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	utils "github.com/hwameistor/hwameistor/pkg/apiserver/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AuthController struct {
	client.Client
	record.EventRecorder
}

func NewAuthController(client client.Client, recorder record.EventRecorder) *AuthController {
	return &AuthController{
		Client:        client,
		EventRecorder: recorder,
	}
}

func (authController *AuthController) Auth(req api.AuthReqBody) bool {
	secretObj := &v1.Secret{}
	objectKey := client.ObjectKey{
		Namespace: utils.GetNamespace(),
		Name:      api.AuthSecretName,
	}
	authController.Client.Get(context.Background(), objectKey, secretObj)
	// TODO: return a token for authorization
	if accessId, ok := secretObj.Data[api.AuthAccessIdName]; ok {
		if secretKey, ok := secretObj.Data[api.AuthSecretKeyName]; ok {
			return string(accessId) == req.AccessId && string(secretKey) == req.SecretKey
		}
	}
	return false
}
