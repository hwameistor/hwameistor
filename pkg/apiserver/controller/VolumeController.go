package controller

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type IVolumeController interface {
	//RestController
	VolumeGet(ctx *gin.Context)
	VolumeList(ctx *gin.Context)
	VolumeReplicasGet(ctx *gin.Context)
	GetVolumeMigrateOperation(ctx *gin.Context)
	VolumeMigrateOperation(ctx *gin.Context)
	GetVolumeConvertOperation(ctx *gin.Context)
	VolumeConvertOperation(ctx *gin.Context)
	GetVolumeExpandOperation(ctx *gin.Context)
	VolumeExpandOperation(ctx *gin.Context)
	VolumeOperationGet(ctx *gin.Context)
}

// VolumeController
type VolumeController struct {
	m *manager.ServerManager
}

func NewVolumeController(m *manager.ServerManager) IVolumeController {
	log.Info("NewVolumeController start")

	return &VolumeController{m}
}

// VolumeGet godoc
// @Summary     摘要 获取指定数据卷基本信息
// @Description get Volume angel1
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.Volume
// @Router      /cluster/volumes/{volumeName} [get]
func (v *VolumeController) VolumeGet(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	if volumeName == "" {
		failRsp.ErrCode = 203
		failRsp.Desc = "volumeName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}
	lv, err := v.m.VolumeController().GetLocalVolume(volumeName)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, lv)
}

// VolumeList godoc
// @Summary 摘要 获取数据卷列表信息 数据卷状态枚举 （ToBeMounted、Created、Creating、Ready、NotReady、ToBeDeleted、Deleted）
// @Description list Volume
// @Tags        Volume
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Param       volumeName query string false "volumeName"
// @Param       state query string false "state"
// @Param       fuzzy query bool false "fuzzy"
// @Param       sort query bool false "sort"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeList
// @Router      /cluster/volumes [get]
func (v *VolumeController) VolumeList(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的page
	page := ctx.Query("page")
	// 获取path中的pageSize
	pageSize := ctx.Query("pageSize")
	// 获取path中的volumeName
	volumeName := ctx.Query("volumeName")
	log.Infof("VolumeList volumeName = %v", volumeName)
	// 获取path中的state
	state := ctx.Query("state")
	// 获取path中的namespace
	namespace := ctx.Query("namespace")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	var queryPage hwameistorapi.QueryPage
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)
	queryPage.VolumeName = volumeName
	queryPage.VolumeState = hwameistorapi.VolumeStatefuzzyConvert(state)
	queryPage.NameSpace = namespace

	lvs, err := v.m.VolumeController().ListLocalVolume(queryPage)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, lvs)
}

// VolumeReplicasGet godoc
// @Summary 摘要 获取指定数据卷的副本列表信息
// @Description list volumes
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Param       volumeReplicaName query string false "volumeReplicaName"
// @Param       state query string false "state"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeReplicaList
// @Router      /cluster/volumes/{volumeName}/replicas [get]
func (v *VolumeController) VolumeReplicasGet(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	// 获取path中的volumeReplicaName
	volumeReplicaName := ctx.Query("volumeReplicaName")
	// 获取path中的state
	state := ctx.Query("state")

	if volumeName == "" {
		failRsp.ErrCode = 203
		failRsp.Desc = "volumeName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.VolumeReplicaName = volumeReplicaName
	queryPage.VolumeState = hwameistorapi.VolumeStatefuzzyConvert(state)
	queryPage.VolumeName = volumeName

	lvs, err := v.m.VolumeController().GetVolumeReplicas(queryPage)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, lvs)
}

// VolumeMigrateOperation godoc
// @Summary 摘要 指定数据卷迁移操作
// @Description post VolumeMigrateOperation body i.g. body { SrcNode string ,SelectedNode string}
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Param       body body api.VolumeMigrateReqBody true "reqBody"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeMigrateRspBody
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/volumes/{volumeName}/migrate [post]
func (v *VolumeController) VolumeMigrateOperation(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody
	// 获取path中的name
	name := ctx.Param("volumeName")
	log.Infof("VolumeMigrateOperation name = %v", name)

	if name == "" {
		failRsp.ErrCode = 203
		failRsp.Desc = "volumeName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	var vmrb hwameistorapi.VolumeMigrateReqBody
	err := ctx.ShouldBind(&vmrb)
	if err != nil {
		log.Errorf("Unmarshal err = %v", err)
		failRsp.ErrCode = 203
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	log.Infof("VolumeMigrateOperation vmrb = %v", vmrb)

	if vmrb.SrcNode == "" || vmrb.SrcNode == "string" {
		failRsp.ErrCode = 203
		failRsp.Desc = "SrcNode cannot be empty or cannot be string"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	if vmrb.SelectedNode == "string" {
		vmrb.SelectedNode = ""
	}

	sourceNodeName := vmrb.SrcNode
	targetNodeName := vmrb.SelectedNode
	abort := vmrb.Abort

	volumeMigrate, err := v.m.VolumeController().CreateVolumeMigrate(name, sourceNodeName, targetNodeName, abort)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeMigrateOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}
	ctx.JSON(http.StatusOK, volumeMigrate)
}

// VolumeConvertOperation godoc
// @Summary 摘要 指定数据卷转换操作
// @Description post VolumeConvertOperation
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Param       body body api.VolumeConvertReqBody true "reqBody"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeConvertRspBody
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/volumes/{volumeName}/convert [post]
func (v *VolumeController) VolumeConvertOperation(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	var vcrb hwameistorapi.VolumeConvertReqBody
	err := ctx.ShouldBind(&vcrb)
	if err != nil {
		log.Errorf("Unmarshal err = %v", err)
		failRsp.ErrCode = 203
		failRsp.Desc = "Unmarshal Failed: " + err.Error()
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}
	abort := vcrb.Abort

	log.Infof("VolumeConvertOperation volumeName = %v, abort = %v", volumeName, abort)
	if volumeName == "" {
		failRsp.ErrCode = 203
		failRsp.Desc = "volumeName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	volumeConvert, err := v.m.VolumeController().CreateVolumeConvert(volumeName, abort)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeConvertOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, volumeConvert)
}

// GetVolumeMigrateOperation godoc
// @Summary 摘要 获取指定数据卷迁移操作
// @Description get GetVolumeMigrateOperation 状态枚举 （Submitted、AddReplica、SyncReplica、PruneReplica、InProgress、Completed、ToBeAborted、Cancelled、Aborted、Failed）
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeMigrateOperation
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/volumes/{volumeName}/migrate [get]
func (v *VolumeController) GetVolumeMigrateOperation(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody
	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	log.Infof("VolumeConvertOperation volumeName = %v", volumeName)
	if volumeName == "" {
		failRsp.ErrCode = 203
		failRsp.Desc = "volumeName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	volumeConvert, err := v.m.VolumeController().GetVolumeMigrate(volumeName)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeConvertOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, volumeConvert)
}

// GetVolumeConvertOperation godoc
// @Summary 摘要 获取指定数据卷转换操作
// @Description get GetVolumeConvertOperation 状态枚举 （Submitted、InProgress、Completed、ToBeAborted、Aborted）
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeConvertOperation
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/volumes/{volumeName}/convert [get]
func (v *VolumeController) GetVolumeConvertOperation(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	log.Infof("VolumeConvertOperation volumeName = %v", volumeName)
	if volumeName == "" {
		failRsp.ErrCode = 203
		failRsp.Desc = "volumeName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	volumeConvert, err := v.m.VolumeController().GetVolumeConvert(volumeName)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeConvertOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, volumeConvert)
}

// GetVolumeExpandOperation godoc
// @Summary 摘要 获取指定数据卷扩容操作
// @Description get GetVolumeExpandOperation
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeExpandOperation
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/volumes/{volumeName}/expand [get]
func (v *VolumeController) GetVolumeExpandOperation(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	log.Infof("VolumeConvertOperation volumeName = %v", volumeName)
	if volumeName == "" {
		failRsp.ErrCode = 203
		failRsp.Desc = "volumeName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	volumeConvert, err := v.m.VolumeController().GetVolumeConvert(volumeName)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeConvertOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, volumeConvert)
}

// VolumeExpandOperation godoc
// @Summary 摘要 指定数据卷扩容
// @Description post VolumeExpandOperation
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeConvertOperation
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/volumes/{volumeName}/expand [post]
func (v *VolumeController) VolumeExpandOperation(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	log.Infof("VolumeConvertOperation volumeName = %v", volumeName)
	if volumeName == "" {
		failRsp.ErrCode = 203
		failRsp.Desc = "volumeName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	volumeConvert, err := v.m.VolumeController().GetVolumeConvert(volumeName)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeConvertOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, volumeConvert)
}

// VolumeOperationGet godoc
// @Summary 摘要 获取指定数据卷操作记录信息 状态枚举 (Submitted、AddReplica、SyncReplica、PruneReplica、InProgress、Completed、ToBeAborted、Cancelled、Aborted、Failed)
// @Description get VolumeOperation
// @Tags        Volume
// @Param       volumeName path string true "volumeName"
// @Param       volumeEventName query string false "volumeEventName"
// @Param       state query string false "state"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.VolumeOperationByVolume
// @Router      /cluster/volumes/{volumeName}/operations [get]
func (v *VolumeController) VolumeOperationGet(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
	volumeName := ctx.Param("volumeName")

	if volumeName == "" {
		failRsp.ErrCode = 203
		failRsp.Desc = "volumeName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	// 获取path中的volumeEventName
	volumeEventName := ctx.Query("volumeEventName")
	// 获取path中的state
	state := ctx.Query("state")

	var queryPage hwameistorapi.QueryPage
	queryPage.VolumeEventName = volumeEventName
	queryPage.VolumeState = hwameistorapi.VolumeStatefuzzyConvert(state)
	queryPage.VolumeName = volumeName

	volumeOperation, err := v.m.VolumeController().GetVolumeOperation(queryPage)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = "VolumeConvertOperation Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, volumeOperation)
}
