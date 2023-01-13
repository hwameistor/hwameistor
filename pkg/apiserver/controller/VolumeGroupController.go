package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type IVolumeGroupController interface {
	//VolumeGroupList
	VolumeGroupList(ctx *gin.Context)
	// VolumeGroupGet
	VolumeGroupGet(ctx *gin.Context)
}

// VolumeGroupController
type VolumeGroupController struct {
	m *manager.ServerManager
}

func NewVolumeGroupController(m *manager.ServerManager) IVolumeGroupController {
	fmt.Println("NewVolumeGroupController start")

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

	vgs, err := v.m.VolumeGroupController().ListVolumeGroup()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
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

	// 获取path中的vgName
	vgName := ctx.Param("vgName")
	if vgName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	vg, err := v.m.VolumeGroupController().GetVolumeGroupByVolumeGroupName(vgName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, vg)
}
