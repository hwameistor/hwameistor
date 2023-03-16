package controller

import (
	"fmt"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type ILocalDiskNodeController interface {
	//LocalDiskNodeList
	LocalDiskNodeList(ctx *gin.Context)
}

// LocalDiskNodeController
type LocalDiskNodeController struct {
	m *manager.ServerManager
}

func NewLocalDiskNodeController(m *manager.ServerManager) ILocalDiskNodeController {
	fmt.Println("NewLocalDiskNodeController start")

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
