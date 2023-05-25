package api

import "time"

const (
	AuthTokenSecretName = "hwameistor-auth-tokens"

	AuthEnableEnvName    = "EnableAuth"
	AuthAccessIdEnvName  = "AuthAccessId"
	AuthSecretKeyEnvName = "AuthSecretKey"

	AuthTokenHeaderName = "Authorization"

	AuthTokenExpireTime  = 7 * 24 * time.Hour
	CheckTokenExpireTime = time.Hour
)

type AuthReqBody struct {
	AccessId  string `json:"accessId,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
}

type AuthRspBody struct {
	Token    string `json:"token"`
	ExpireAt int64  `json:"expireAt"`
}

type AuthLogoutRspBody struct {
	Success bool `json:"success"`
}

type AuthInfoRspBody struct {
	Enabled bool `json:"enabled"`
}
