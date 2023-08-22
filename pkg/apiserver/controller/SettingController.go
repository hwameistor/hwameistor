package controller

import (
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type ISettingController interface {
	EnableDRBDSetting(ctx *gin.Context)
	DRBDSettingGet(ctx *gin.Context)
}

type SettingController struct {
	m *manager.ServerManager
}

func NewSettingController(m *manager.ServerManager) ISettingController {
	return &SettingController{m}
}

// EnableDRBDSetting godoc
// @Summary 摘要 高可用设置
// @Description post EnableDRBDSetting
// @Tags        Setting
// @Param       body body api.DrbdEnableSettingReqBody true "body"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.DrbdEnableSettingRspBody
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/drbd [post]
func (n *SettingController) EnableDRBDSetting(ctx *gin.Context) {
	var failRsp api.RspFailBody

	//// 获取path中的name
	//enabledrbd := ctx.Param("enabledrbd")

	var desrb api.DrbdEnableSettingReqBody
	err := ctx.ShouldBind(&desrb)
	if err != nil {
		log.Errorf("Unmarshal err = %v", err)
		failRsp.ErrCode = 203
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}
	enabledrbd := desrb.Enable

	if enabledrbd == true {
		setting, err := n.m.SettingController().EnableHighAvailability()
		if err != nil {
			failRsp.ErrCode = 500
			failRsp.Desc = "EnableDRBDSetting Failed" + err.Error()
			ctx.JSON(http.StatusInternalServerError, failRsp)
			return
		}
		ctx.JSON(http.StatusOK, setting)
	}
}

// DRBDSettingGet godoc
// @Summary 摘要 获取高可用设置
// @Description get DRBDSettingGet
// @Tags        Setting
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.DrbdEnableSetting
// @Router      /cluster/drbd [get]
func (n *SettingController) DRBDSettingGet(ctx *gin.Context) {
	var failRsp api.RspFailBody

	setting, err := n.m.SettingController().GetDRBDSetting()
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = "GetDRBDSetting Failed" + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, setting)
}
