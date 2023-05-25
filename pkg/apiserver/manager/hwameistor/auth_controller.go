package hwameistor

import (
	"context"
	"encoding/json"
	"errors"
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

func NewAuthController(client client.Client, recorder record.EventRecorder) *AuthController {
	return &AuthController{
		Client:        client,
		EventRecorder: recorder,
		tm:            newTokenManager(client),
	}
}

func (authController *AuthController) Auth(req api.AuthReqBody) (string, int64, error) {
	accessId, ok1 := os.LookupEnv(api.AuthAccessIdEnvName)
	secretKey, ok2 := os.LookupEnv(api.AuthSecretKeyEnvName)
	if ok1 && ok2 {
		if len(accessId) == 0 || len(secretKey) == 0 {
			return "", 0, errors.New("accessId and secretKey env cant be nil")
		}
		if accessId == req.AccessId && secretKey == req.SecretKey {
			// authorization success
			token, expireAt := authController.tm.generateToken()
			return token, expireAt, nil
		} else {
			return "", 0, errors.New("wrong accessId or secretKey")
		}
	} else {
		return "", 0, errors.New("there is no set auth env")
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

type tokenParameter struct {
	ExpireAt int64 `json:"ExpireAt"`
}

type tokenManager struct {
	client       client.Client
	tokens       map[string]tokenParameter
	tokensSecret *v1.Secret
}

// init tokenManager, get the tokens from secret, in each change, it will save "tokens" to "tokensSecret" in kubernetes
func newTokenManager(c client.Client) *tokenManager {
	tm := &tokenManager{
		client:       c,
		tokens:       map[string]tokenParameter{},
		tokensSecret: &v1.Secret{},
	}
	tm.load()
	go tm.checkTokenExpire()
	return tm
}

// generate a new token with expire time
func (tm *tokenManager) generateToken() (string, int64) {
	token := uuid.New().String()
	tm.tokens[token] = tokenParameter{
		ExpireAt: time.Now().Add(api.AuthTokenExpireTime).Unix(),
	}
	tm.save()
	log.Infof("Generate a new token, token count:%v", len(tm.tokensSecret.Data))
	return token, tm.tokens[token].ExpireAt
}

func (tm *tokenManager) verifyToken(token string) bool {
	parameter, in := tm.tokens[token]
	if in {
		expireTime := time.Unix(parameter.ExpireAt, 0)
		if time.Now().After(expireTime) {
			// token expired
			tm.removeToken(token)
			return false
		}
	}
	return in
}

func (tm *tokenManager) removeToken(token string) {
	delete(tm.tokens, token)
	tm.save()
	log.Infof("Remove token:%v, token count:%v", token, len(tm.tokens))
}

func (tm *tokenManager) checkTokenExpire() {
	time.Sleep(time.Second)
	for {
		log.Infof("Start to check tokens expire status")
		for token, parameter := range tm.tokens {
			expireTime := time.Unix(parameter.ExpireAt, 0)
			if time.Now().After(expireTime) {
				// this token expired, delete it
				tm.removeToken(token)
			}
		}
		time.Sleep(api.CheckTokenExpireTime)
	}
}

// load tokensSecret from kubernetes
func (tm *tokenManager) load() {
	authTokensObjKey := client.ObjectKey{
		Namespace: utils.GetNamespace(),
		Name:      api.AuthTokenSecretName,
	}

	// get the kubernetes secret object, create a new one if its nil
	err := tm.client.Get(context.Background(), authTokensObjKey, tm.tokensSecret)
	if err != nil {
		log.Infof("Fail to get auth token secret:%v in nameSpace:%v, create the secret now", api.AuthTokenSecretName, utils.GetNamespace())
		err = tm.client.Create(context.Background(), &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: authTokensObjKey.Namespace,
				Name:      authTokensObjKey.Name,
			},
		})
		if err != nil {
			log.Errorf("Fail to create auth token secret:%v, err:%v", api.AuthTokenSecretName, err)
			return
		}
		// get the new tokenSecret object
		err = tm.client.Get(context.Background(), authTokensObjKey, tm.tokensSecret)
		if err != nil {
			log.Errorf("Fail to get new auth token secret:%v, err:%v", api.AuthTokenSecretName, err)
			return
		}
	}

	// load tokens data
	if tm.tokensSecret.Data != nil {
		for token, parameterData := range tm.tokensSecret.Data {
			parameter := tokenParameter{}
			err = json.Unmarshal(parameterData, &parameter)
			if err != nil {
				log.Errorf("Fail to unmarshal token parameter data, err:%v", err)
				return
			}
			tm.tokens[token] = parameter
		}
	}
}

// save tokensSecret to kubernetes
func (tm *tokenManager) save() {
	tm.tokensSecret.Data = map[string][]byte{}
	for token, parameter := range tm.tokens {
		parameterData, err := json.Marshal(parameter)
		if err != nil {
			log.Errorf("Fail to marshal token parameter to json, err:%v", err)
			return
		}
		tm.tokensSecret.Data[token] = parameterData
	}

	err := tm.client.Update(context.Background(), tm.tokensSecret)
	if err != nil {
		log.Errorf("Fail to save token secret, err:%v", err)
		return
	}
}
