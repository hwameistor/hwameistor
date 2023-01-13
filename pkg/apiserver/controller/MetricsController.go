package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type IMetricsController interface {
	//RestController
	ModuleStatus(ctx *gin.Context)
	OperationList(ctx *gin.Context)
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

	moduleStatus, err := v.m.MetricController().GetModuleStatus()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, moduleStatus)
}

// OperationList godoc
// @Summary 摘要 获取操作记录列表
// @Description OperationList 状态枚举 （Submitted、AddReplica、SyncReplica、PruneReplica、InProgress、Completed、ToBeAborted、Cancelled、Aborted、Failed）
// @Tags        Metric
// @Param       name query string false "name"
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.OperationMetric  "成功"
// @Router      /cluster/operations [get]
func (v *MetricsController) OperationList(ctx *gin.Context) {

	// 获取path中的page
	page := ctx.Query("page")
	// 获取path中的pageSize
	pageSize := ctx.Query("pageSize")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	operationListMetric, err := v.m.MetricController().OperationListMetric(int32(p), int32(ps))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, operationListMetric)
}
