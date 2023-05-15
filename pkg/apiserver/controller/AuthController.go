package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
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

}

func (n *AuthController) Logout(ctx *gin.Context) {

}
