package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type IVolumeGroupController interface {
	//RestController
	VolumeListByVolumeGroup(ctx *gin.Context)
}

// VolumeGroupController
type VolumeGroupController struct {
	m *manager.ServerManager
}

func NewVolumeGroupController(m *manager.ServerManager) IVolumeGroupController {
	fmt.Println("NewVolumeGroupController start")

	return &VolumeGroupController{m}
}

// VolumeListByVolumeGroup godoc
// @Summary     摘要 获取指定数据卷组中包含数据卷名称列表基本信息
// @Description get VolumeListByVolumeGroup
// @Tags        VolumeGroup
// @Param       name path string true "name"
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.VolumeGroup
// @Router      /volumegroups/volumegroups/{name} [get]
func (v *VolumeGroupController) VolumeListByVolumeGroup(ctx *gin.Context) {
	// 获取path中的name
	volumeGroupName := ctx.Param("name")

	if volumeGroupName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}
	vginfos, err := v.m.VolumeGroupController().ListVolumesByVolumeGroup(volumeGroupName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, vginfos)
}
