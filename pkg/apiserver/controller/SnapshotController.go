package controller

import (
	"github.com/gin-gonic/gin"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type ISnapshotController interface {
	SnapshotList(ctx *gin.Context)
	SnapshotGet(ctx *gin.Context)
}

type SnapshotController struct {
	m *manager.ServerManager
}

func NewSnapshotController(m *manager.ServerManager) ISnapshotController {
	log.Info("NewVolumeController start")

	return &SnapshotController{m}
}

// SnapshotList godoc
// @Summary 摘要 全局快照列表展式
// @Description list Snapshot
// @Tags        Snapshot
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Param       snapshotName query string false "snapshotName"
// @Param       state query string false "state"
// @Param       volumeName query string false "volumeName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.SnapshotList
// @Router      /cluster/snapshots [get]
func (v *SnapshotController) SnapshotList(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// get page form path
	page := ctx.Query("page")
	// get pagesize form path
	pageSize := ctx.Query("pageSize")
	// get snapshotName form path
	snapshotName := ctx.Query("snapshotName")
	log.Infof("SnapshotList snapshot = %v", snapshotName)

	// get state form path
	state := ctx.Query("state")
	// get volumeName from path
	volumeName := ctx.Query("volumeName")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	var queryPage hwameistorapi.QueryPage
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)
	queryPage.SnapshotName = snapshotName
	queryPage.SnapshotState = hwameistorapi.VolumeStatefuzzyConvert(state)
	queryPage.VolumeName = volumeName

	lvs, err := v.m.SnapshotController().ListLocalSnapshot(queryPage)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, lvs)
}

// SnapshotGet godoc
// @Summary 摘要 获取指定快照
// @Description get SnapshotGet 状态枚举 （Creating, Ready, NotReady, ToBeDeleted, Deleted）
// @Tags        Volume
// @Param       snapshotName path string true "snapshotName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.Snapshot
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/volumes/{volumeName}/snapshot [get]
func (v *SnapshotController) SnapshotGet(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	snapshotName := ctx.Param("snapshotName")

	if snapshotName == "" {
		failRsp.ErrCode = 203
		failRsp.Desc = "snapshotName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	lvs, err := v.m.SnapshotController().GetLocalSnapshot(snapshotName)

	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = "GetVolumeExpand Failed: " + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, lvs)
}
