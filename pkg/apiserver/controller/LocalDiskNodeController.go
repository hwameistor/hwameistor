package controller

import (
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type ILocalDiskNodeController interface {
	LocalDiskNodeList(ctx *gin.Context)
}

type LocalDiskNodeController struct {
	m *manager.ServerManager
}

func NewLocalDiskNodeController(m *manager.ServerManager) ILocalDiskNodeController {
	log.Info("NewLocalDiskNodeController start")

	return &LocalDiskNodeController{m}
}

// LocalDiskNodeList godoc
// @Summary     摘要 获取集群磁盘组列表
// @Description get LocalDiskNodeList
// @Tags        LocalDiskNode
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.LocalDiskNodeList
// @Router      /cluster/localdisknodes [get]
func (v *LocalDiskNodeController) LocalDiskNodeList(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	vgs, err := v.m.LocalDiskNodeController().ListLocalDiskNode()
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, vgs)
}
