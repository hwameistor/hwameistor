package controller

import (
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type IVolumeGroupController interface {
	VolumeGroupList(ctx *gin.Context)
	VolumeGroupGet(ctx *gin.Context)
}

type VolumeGroupController struct {
	m *manager.ServerManager
}

func NewVolumeGroupController(m *manager.ServerManager) IVolumeGroupController {
	log.Info("NewVolumeGroupController start")

	return &VolumeGroupController{m}
}

// VolumeGroupList godoc
// @Summary     摘要 获取数据卷组列表
// @Description get VolumeGroupList
// @Tags        VolumeGroup
// @Param       name query string false "name"
// @Param       fuzzy query bool false "fuzzy"
// @Param       sort query bool false "sort"
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.VolumeGroupList
// @Router      /cluster/volumegroups [get]
func (v *VolumeGroupController) VolumeGroupList(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	vgs, err := v.m.VolumeGroupController().ListVolumeGroup()
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, vgs)
}

// VolumeGroupGet godoc
// @Summary     摘要 查看数据卷组
// @Description get VolumeGroupGet
// @Tags        VolumeGroup
// @Param       vgName path string false "vgName"
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.VolumeGroup
// @Router      /cluster/volumegroups/{vgName} [get]
func (v *VolumeGroupController) VolumeGroupGet(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的vgName
	vgName := ctx.Param("vgName")
	if vgName == "" {
		failRsp.ErrCode = 203
		failRsp.Desc = "vgName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	vg, err := v.m.VolumeGroupController().GetVolumeGroupByVolumeGroupName(vgName)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, vg)
}
