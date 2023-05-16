package api

const (
	AuthSecretName      = "hwameistor-auth"
	AuthAccessIdName    = "AccessId"
	AuthSecretKeyName   = "SecretKey"
	AuthTokenHeaderName = "Authorization"
)

type AuthReqBody struct {
	AccessId  string `json:"access_id"`
	SecretKey string `json:"secret_key"`
}

type AuthRspBody struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

type LogoutRspBody struct {
	Success bool `json:"success"`
}
