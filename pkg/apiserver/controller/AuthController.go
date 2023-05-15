package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
	"net/http"
)

type IAuthController interface {
	Auth(ctx *gin.Context)
	Logout(ctx *gin.Context)
}

type AuthController struct {
	m *manager.ServerManager
}

func NewAuthController(m *manager.ServerManager) IAuthController {
	return &AuthController{m}
}

func (n *AuthController) Auth(ctx *gin.Context) {
	var req api.AuthReqBody
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, api.RspFailBody{
			ErrCode: 500,
			Desc:    "Authentication Failed" + err.Error(),
		})
		return
	}
	if authResult := n.m.AuthController().Auth(req); authResult {
		// todo: token? uuid?
		ctx.JSON(http.StatusOK, api.AuthRspBody{
			Success: true,
			Token:   "??",
		})
		return
	}
	ctx.JSON(http.StatusUnauthorized, api.RspFailBody{
		ErrCode: 401,
		Desc:    "Fail to authenticate",
	})
}

func (n *AuthController) Logout(ctx *gin.Context) {
	// delete token
}
