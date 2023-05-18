package hwameistor

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	utils "github.com/hwameistor/hwameistor/pkg/apiserver/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
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
	// get the secret
	if err := authController.Client.Get(context.Background(), objectKey, secretObj); err != nil {
		return false
	}
	if accessId, ok := secretObj.Data[api.AuthAccessIdName]; ok {
		if secretKey, ok := secretObj.Data[api.AuthSecretKeyName]; ok {
			return string(accessId) == req.AccessId && string(secretKey) == req.SecretKey
		}
	}
	return false
}

func (authController *AuthController) IsEnableAuth() bool {
	isEnable, ok := os.LookupEnv(api.AuthEnableEnvName)
	return ok && strings.ToLower(isEnable) == "true"
}
