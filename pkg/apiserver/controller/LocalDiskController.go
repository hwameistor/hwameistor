package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type ILocalDiskController interface {
	// LocalDiskList
	LocalDiskList(ctx *gin.Context)
}

// LocalDiskController
type LocalDiskController struct {
	m *manager.ServerManager
}

func NewLocalDiskController(m *manager.ServerManager) ILocalDiskController {
	fmt.Println("NewLocalDiskController start")

	return &LocalDiskController{m}
}

// LocalDiskList godoc
// @Summary     摘要 获取本地磁盘列表
// @Description get LocalDiskList 状态枚举 （Active、Inactive、Unknown、Pending、Available、Bound)
// @Tags        LocalDisk
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.LocalDiskList
// @Router      /cluster/localdisks [get]
func (v *LocalDiskController) LocalDiskList(ctx *gin.Context) {

	vgs, err := v.m.LocalDiskController().ListLocalDisk()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, vgs)
}
