package hwameistor

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	utils "github.com/hwameistor/hwameistor/pkg/apiserver/util"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

type AuthController struct {
	client.Client
	record.EventRecorder
	tm *tokenManager
}

func NewAuthController(c client.Client, recorder record.EventRecorder) *AuthController {
	return &AuthController{
		Client:        c,
		EventRecorder: recorder,
		tm:            newTokenManager(c),
	}
}

func (authController *AuthController) Auth(req api.AuthReqBody) (string, int64, error) {
	objKey := client.ObjectKey{
		Namespace: utils.GetNamespace(),
		Name:      api.AuthSecretName,
	}
	secretObj := &v1.Secret{}
	// get the secret
	if err := authController.Client.Get(context.Background(), objKey, secretObj); err != nil {
		return "", 0, err
	}
	accessId, ok1 := secretObj.Data[api.AuthAccessIdName]
	secretKey, ok2 := secretObj.Data[api.AuthSecretKeyName]
	if ok1 && ok2 {
		if string(accessId) == req.AccessId && string(secretKey) == req.SecretKey {
			// authorization success
			token, expireAt := authController.tm.generateToken()
			return token, expireAt, nil
		} else {
			return "", 0, errors.New("wrong accessId or secretKey")
		}
	} else {
		return "", 0, errors.New("wrong auth secret")
	}
}

func (authController *AuthController) Logout(token string) error {
	if authController.tm.verifyToken(token) {
		authController.tm.removeToken(token)
		return nil
	}
	return errors.New("token verify failed")
}

// IsEnableAuth return if enable Auth
func (authController *AuthController) IsEnableAuth() bool {
	isEnable, ok := os.LookupEnv(api.AuthEnableEnvName)
	return ok && strings.ToLower(isEnable) == "true"
}

func (authController *AuthController) VerifyToken(token string) bool {
	return authController.tm.verifyToken(token)
}

type tokenManager struct {
	client.Client
	tokensSecret *v1.Secret
}

// init tokenManager, get the tokens from secret
func newTokenManager(c client.Client) *tokenManager {
	tm := &tokenManager{
		Client:       c,
		tokensSecret: &v1.Secret{},
	}
	objKey := client.ObjectKey{
		Namespace: utils.GetNamespace(),
		Name:      api.AuthTokenSecretName,
	}
	if err := c.Get(context.Background(), objKey, tm.tokensSecret); err != nil {
		log.Infof("Fail to get auth token secret:%v in nameSpace:%v, create the secret now", api.AuthTokenSecretName, utils.GetNamespace())
		err := c.Create(context.Background(), &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: objKey.Namespace,
				Name:      objKey.Name,
			},
		})
		if err != nil {
			log.Errorf("Fail to create auth token secret:%v, err:%v", api.AuthTokenSecretName, err)
			return nil
		}
		c.Get(context.Background(), objKey, tm.tokensSecret)
	}
	go tm.checkTokenExpire()
	return tm
}

// generate a new token with expire time
func (tm *tokenManager) generateToken() (string, int64) {
	token := uuid.New().String()
	expireAt := time.Now().Add(api.AuthTokenExpireTime)
	if tm.tokensSecret.Data == nil {
		tm.tokensSecret.Data = map[string][]byte{}
	}
	tm.tokensSecret.Data[token] = []byte(fmt.Sprintf("%v", expireAt.Unix()))
	tm.save()
	log.Infof("Generate a new token, token count:%v", len(tm.tokensSecret.Data))
	return token, expireAt.Unix()
}

func (tm *tokenManager) verifyToken(token string) bool {
	expireAt, in := tm.tokensSecret.Data[token]
	if in {
		expireTime := time.Unix(utils.ConvertByteToInt64(expireAt), 0)
		if time.Now().After(expireTime) {
			// token expired
			tm.removeToken(token)
			return false
		}
	}
	return in
}

func (tm *tokenManager) removeToken(token string) {
	delete(tm.tokensSecret.Data, token)
	tm.save()
	log.Infof("Remove token:%v", token)
}

func (tm *tokenManager) checkTokenExpire() {
	time.Sleep(time.Second)
	for {
		log.Infof("Start to check tokens expire status")
		for token, expireAt := range tm.tokensSecret.Data {
			expireTime := time.Unix(utils.ConvertByteToInt64(expireAt), 0)
			if time.Now().After(expireTime) {
				// this token expired, delete it
				tm.removeToken(token)
			}
		}
		time.Sleep(api.CheckTokenExpireTime)
	}
}

// save to kubernetes secret
func (tm *tokenManager) save() {
	err := tm.Client.Update(context.Background(), tm.tokensSecret)
	if err != nil {
		log.Errorf("Fail to save token secret, err:%v", err)
	}
}
