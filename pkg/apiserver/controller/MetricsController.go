package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
	"net/http"
)

type IMetricsController interface {
	//RestController
	ModuleStatus(ctx *gin.Context)
}

// MetricsController
type MetricsController struct {
	m *manager.ServerManager
}

func NewMetricsController(m *manager.ServerManager) IMetricsController {
	fmt.Println("NewMetricsController start")

	return &MetricsController{m}
}

// Get godoc
// @Summary     摘要 获取基础监控指标
// @Description get ModuleStatus
// @Tags        Metric
// @Param       name query string false "name"
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.ModuleStatus  "成功"
// @Router      /cluster/status [get]
func (v *MetricsController) ModuleStatus(ctx *gin.Context) {

	moduleStatus, err := v.m.MetricController().GetModuleStatusMetric()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, moduleStatus)
}
