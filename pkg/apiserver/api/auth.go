package api

import "time"

const (
	AuthSecretName      = "hwameistor-auth"
	AuthTokenSecretName = "hwameistor-auth-tokens"

	AuthAccessIdName    = "AccessId"
	AuthSecretKeyName   = "SecretKey"
	AuthTokenHeaderName = "Authorization"

	AuthTokenExpireTime = 7 * 24 * time.Hour
	AuthEnableEnvName   = "EnableAuth"

	CheckTokenExpireTime = 2 * time.Hour
)

type AuthReqBody struct {
	AccessId  string `json:"access_id,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
}

type AuthRspBody struct {
	Token    string `json:"token"`
	ExpireAt int64  `json:"expire_at"`
}

type AuthLogoutRspBody struct {
	Success bool `json:"success"`
}

type AuthInfoRspBody struct {
	Enabled bool `json:"enabled"`
}
