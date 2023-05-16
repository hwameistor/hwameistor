package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
	"net/http"
)

type IAuthController interface {
	Auth(ctx *gin.Context)
	Logout(ctx *gin.Context)
	GetAuthMiddleWare(whiteList WhiteList) func(ctx *gin.Context)
}

type AuthController struct {
	m      *manager.ServerManager
	tokens map[string]struct{}
}

// WhiteList [Url][Method]
type WhiteList map[string]string

func NewAuthController(m *manager.ServerManager) IAuthController {
	return &AuthController{m, map[string]struct{}{}}
}

func (n *AuthController) Auth(ctx *gin.Context) {
	var req api.AuthReqBody
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, api.RspFailBody{
			ErrCode: 500,
			Desc:    "Authorization Failed" + err.Error(),
		})
		return
	}
	if authResult := n.m.AuthController().Auth(req); authResult {
		ctx.JSON(http.StatusOK, api.AuthRspBody{
			Success: true,
			Token:   n.generateToken(),
		})
		return
	}
	ctx.JSON(http.StatusUnauthorized, api.RspFailBody{
		ErrCode: 401,
		Desc:    "Fail to authorize",
	})
}

func (n *AuthController) Logout(ctx *gin.Context) {
	token := ctx.Request.Header.Get(api.AuthTokenHeaderName)
	if n.verifyToken(token) {
		// verify token success, continue logout
		n.deleteToken(token)
		ctx.JSON(http.StatusOK, api.LogoutRspBody{
			Success: true,
		})
		return
	}
	// token verify fail
	ctx.JSON(http.StatusUnauthorized, api.RspFailBody{
		ErrCode: 401,
		Desc:    "Fail to authorize",
	})
}

func (n *AuthController) generateToken() string {
	token := uuid.New().String()
	n.tokens[token] = struct{}{}
	return token
}

func (n *AuthController) verifyToken(token string) bool {
	_, in := n.tokens[token]
	return in
}

func (n *AuthController) deleteToken(token string) {
	delete(n.tokens, token)
}

func (n *AuthController) GetAuthMiddleWare(whiteList WhiteList) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if method, in := whiteList[ctx.Request.URL.String()]; in && method == ctx.Request.Method {
			// if this request is in whitelist, then continue
			return
		}
		// this request is not in whitelist, check token
		if !n.verifyToken(ctx.Request.Header.Get(api.AuthTokenHeaderName)) {
			// token verify fail, abort
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, api.RspFailBody{
				ErrCode: 401,
				Desc:    "Fail to authorize",
			})
		}
	}
}
