package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type IAuthController interface {
	Auth(ctx *gin.Context)
	Logout(ctx *gin.Context)
	Info(ctx *gin.Context)
	GetAuthMiddleWare() func(ctx *gin.Context)
}

type AuthController struct {
	m      *manager.ServerManager
	tokens map[string]struct{}
}

func NewAuthController(m *manager.ServerManager) IAuthController {
	return &AuthController{m, map[string]struct{}{}}
}

// Auth godoc
// @Summary     Authorization
// @Description Authorize user, return a token and expireAt if success
// @Tags        Auth
// @Param       body body api.AuthReqBody true "body"
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.AuthRspBody
// @Failure     500 {object} api.RspFailBody
// @Router      /cluster/auth/auth [post]
func (n *AuthController) Auth(ctx *gin.Context) {
	var req api.AuthReqBody
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, api.RspFailBody{
			ErrCode: 500,
			Desc:    "Authorization Failed, " + err.Error(),
		})
		return
	}
	if len(req.AccessId) == 0 || len(req.SecretKey) == 0 {
		ctx.JSON(http.StatusUnauthorized, api.RspFailBody{
			ErrCode: 401,
			Desc:    "Fail to authorize, AccessId or SecretKey cant be nil",
		})
		return
	}
	token, expireAt, err := n.m.AuthController().Auth(req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, api.RspFailBody{
			ErrCode: 401,
			Desc:    "Fail to authorize, " + err.Error(),
		})
		return
	}
	// auth success
	ctx.JSON(http.StatusOK, api.AuthRspBody{
		Token:    token,
		ExpireAt: expireAt,
	})
	log.Infof("User auth success, expireAt:%v", expireAt)
}

// Logout godoc
// @Summary     Logout the token
// @Description Logout the token, if verify token success, delete this token and return success
// @Tags        Auth
// @Produce     application/json
// @Success     200 {object} api.AuthLogoutRspBody
// @Failure     500 {object} api.RspFailBody
// @Router      /cluster/auth/logout [post]
func (n *AuthController) Logout(ctx *gin.Context) {
	token := ctx.Request.Header.Get(api.AuthTokenHeaderName)
	err := n.m.AuthController().Logout(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, api.RspFailBody{
			ErrCode: 401,
			Desc:    "Fail to authorize, " + err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, api.AuthLogoutRspBody{
		Success: true,
	})
	log.Infof("User logout, token:%v", token)
}

// Info godoc
// @Summary     Auth info
// @Description Get the status if enable authorization
// @Tags        Auth
// @Produce     application/json
// @Success     200 {object} api.AuthInfoRspBody
// @Router      /cluster/auth/info [get]
func (n *AuthController) Info(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, api.AuthInfoRspBody{
		Enabled: n.m.AuthController().IsEnableAuth(),
	})
}

func (n *AuthController) GetAuthMiddleWare() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		// if enable auth and verify token fail, return 401
		if n.m.AuthController().IsEnableAuth() && !n.m.AuthController().VerifyToken(ctx.Request.Header.Get(api.AuthTokenHeaderName)) {
			// abort request and return 401
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, api.RspFailBody{
				ErrCode: 401,
				Desc:    "Fail to authorize",
			})
		}
	}
}
