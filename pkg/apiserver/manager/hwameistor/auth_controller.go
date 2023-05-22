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

func NewAuthController(c client.Client, recorder record.EventRecorder) *AuthController {
	return &AuthController{
		Client:        c,
		EventRecorder: recorder,
		tm:            newTokenManager(c),
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

type tokenManager struct {
	client.Client
	tokens       map[string]int64
	tokensSecret *v1.Secret
}

// init tokenManager, get the tokens from secret
func newTokenManager(c client.Client) *tokenManager {
	tm := &tokenManager{
		Client:       c,
		tokens:       map[string]int64{},
		tokensSecret: &v1.Secret{},
	}
	tm.load()
	go tm.checkTokenExpire()
	return tm
}

// generate a new token with expire time
func (tm *tokenManager) generateToken() (string, int64) {
	token := uuid.New().String()
	expireAt := time.Now().Add(api.AuthTokenExpireTime)
	tm.tokens[token] = expireAt.Unix()
	tm.save()
	log.Infof("Generate a new token, token count:%v", len(tm.tokensSecret.Data))
	return token, expireAt.Unix()
}

func (tm *tokenManager) verifyToken(token string) bool {
	expireAt, in := tm.tokens[token]
	if in {
		expireTime := time.Unix(expireAt, 0)
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
		for token, expireAt := range tm.tokens {
			expireTime := time.Unix(expireAt, 0)
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
	if err := tm.Client.Get(context.Background(), authTokensObjKey, tm.tokensSecret); err != nil {
		log.Infof("Fail to get auth token secret:%v in nameSpace:%v, create the secret now", api.AuthTokenSecretName, utils.GetNamespace())
		err = tm.Client.Create(context.Background(), &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: authTokensObjKey.Namespace,
				Name:      authTokensObjKey.Name,
			},
		})
		if err != nil {
			log.Errorf("Fail to create auth token secret:%v, err:%v", api.AuthTokenSecretName, err)
			return
		}
		if err = tm.Client.Get(context.Background(), authTokensObjKey, tm.tokensSecret); err != nil {
			log.Errorf("Fail to get new auth token secret:%v, err:%v", api.AuthTokenSecretName, err)
			return
		}
	}
	// load tokens data
	if tm.tokensSecret.Data != nil {
		tokensJsonData, ok := tm.tokensSecret.Data[api.AuthTokenSecretKeyName]
		if ok {
			// unmarshal data
			err := json.Unmarshal(tokensJsonData, &tm.tokens)
			if err != nil {
				log.Errorf("Fail to unmarshal token json data, err:%v", err)
				return
			}
		}
	}
}

// save tokensSecret to kubernetes
func (tm *tokenManager) save() {
	tokensJsonData, err := json.Marshal(tm.tokens)
	if err != nil {
		log.Errorf("Fail to marshal tokens to json, err:%v", err)
		return
	}
	tm.tokensSecret.Data = map[string][]byte{
		api.AuthTokenSecretKeyName: tokensJsonData,
	}
	err = tm.Client.Update(context.Background(), tm.tokensSecret)
	if err != nil {
		log.Errorf("Fail to save token secret, err:%v", err)
		return
	}
}
