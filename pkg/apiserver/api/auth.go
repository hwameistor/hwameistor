package api

const (
	AuthSecretName      = "hwameistor-auth"
	AuthAccessIdName    = "AccessId"
	AuthSecretKeyName   = "SecretKey"
	AuthTokenHeaderName = "Authorization"
	AuthTokenExpireTime = 12 * 60 * 60
)

type AuthReqBody struct {
	AccessId  string `json:"access_id,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
}

type AuthRspBody struct {
	Success bool   `json:"success,omitempty"`
	Token   string `json:"token,omitempty"`
}

type LogoutRspBody struct {
	Success bool `json:"success,omitempty"`
}
