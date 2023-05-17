package api

import "time"

const (
	AuthSecretName      = "hwameistor-auth"
	AuthAccessIdName    = "AccessId"
	AuthSecretKeyName   = "SecretKey"
	AuthTokenHeaderName = "Authorization"
	AuthTokenExpireTime = 12 * 60 * 60 * time.Second
)

type AuthReqBody struct {
	AccessId  string `json:"access_id,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
}

type AuthRspBody struct {
	Token string `json:"token,omitempty"`
}

type AuthLogoutRspBody struct {
	Success bool `json:"success,omitempty"`
}

type AuthInfoRspBody struct {
	Enabled bool `json:"enabled,omitempty"`
}
