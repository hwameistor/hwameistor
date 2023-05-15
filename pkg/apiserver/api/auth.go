package api

const (
	AuthSecretName    = "hwameistor-auth"
	AuthAccessIdName  = "AccessId"
	AuthSecretKeyName = "SecretKey"
)

type AuthReqBody struct {
	AccessId  string `json:"access_id"`
	SecretKey string `json:"secret_key"`
}

type AuthRspBody struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

//type LogoutReqBody struct {
//}
//
//type LogoutRspBody struct {
//	Success bool `json:"success"`
//}
