package controller

import (
	"github.com/gin-gonic/gin"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type IMetricsController interface {
	ModuleStatus(ctx *gin.Context)
	OperationList(ctx *gin.Context)
	OperationGet(ctx *gin.Context)
	EventList(ctx *gin.Context)
	EventGet(ctx *gin.Context)
}

type MetricsController struct {
	m *manager.ServerManager
}

func NewMetricsController(m *manager.ServerManager) IMetricsController {
	log.Info("NewMetricsController start")

	return &MetricsController{m}
}

// ModuleStatus godoc
// @Summary     摘要 获取基础监控指标
// @Description get ModuleStatus
// @Tags        Metric
// @Param       name query string false "name"
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.ModuleStatus
// @Router      /cluster/status [get]
func (v *MetricsController) ModuleStatus(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	moduleStatus, err := v.m.MetricController().GetModuleStatus()
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
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
// @Success     200 {object}  api.OperationMetric
// @Router      /cluster/operations [get]
func (v *MetricsController) OperationList(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	page := ctx.Query("page")
	pageSize := ctx.Query("pageSize")
	name := ctx.Query("name")
	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	operationListMetric, err := v.m.MetricController().OperationListMetric(int32(p), int32(ps), name)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, operationListMetric)
}

// OperationGet godoc
// @Summary 摘要 获取指定操作记录
// @Description OperationGet eventType枚举 (Migrate、Expand、Convert)
// @Tags        Metric
// @Param       operationName path string true "operationName"
// @Param       eventType query string true "eventType"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.Operation
// @Router      /cluster/operations/:operationName [get]
func (v *MetricsController) OperationGet(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody
	eventName := ctx.Param("operationName")
	eventType := ctx.Param("eventType")
	operation, err := v.m.MetricController().GetOperation(eventName, eventType)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, operation)
}

// EventList godoc
// @Summary 摘要 获取审计日志
// @Description EventList 排序  resourceType枚举（Cluster;StorageNode;DiskNode;Pool;Volume;DiskVolume;Disk）  sort枚举 （time、name、type）
// @Tags        Metric
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Param       resourceName query string false "resourceName"
// @Param       resourceType query string false "resourceType"
// @Param       sort query string false "sort"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.EventActionList
// @Router      /cluster/events [get]
func (v *MetricsController) EventList(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	page := ctx.Query("page")
	pageSize := ctx.Query("pageSize")
	resourceName := ctx.Query("resourceName")
	resourceType := ctx.Query("resourceType")
	sort := ctx.Query("sort")
	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)
	var queryPage hwameistorapi.QueryPage
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)
	queryPage.ResourceName = resourceName
	queryPage.ResourceType = resourceType
	queryPage.Sort = sort

	events, err := v.m.MetricController().EventList(queryPage)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, events)
}

// EventGet godoc
// @Summary 摘要 获取指定事件
// @Description EventGet
// @Tags        Metric
// @Param       eventName path string true "eventName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.Event
// @Router      /cluster/events/:eventName [get]
func (v *MetricsController) EventGet(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody
	eventName := ctx.Param("eventName")
	event, err := v.m.MetricController().GetEvent(eventName)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, event)
}
