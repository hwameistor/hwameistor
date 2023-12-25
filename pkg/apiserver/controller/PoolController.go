package controller

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type IPoolController interface {
	StoragePoolGet(ctx *gin.Context)
	StoragePoolList(ctx *gin.Context)
	StorageNodesGetByPoolName(ctx *gin.Context)
	StorageNodeDisksGetByPoolName(ctx *gin.Context)
	StoragePoolExpand(ctx *gin.Context)
}

type PoolController struct {
	m *manager.ServerManager
}

func NewPoolController(m *manager.ServerManager) IPoolController {
	return &PoolController{m}
}

// StoragePoolGet godoc
// @Summary 摘要 获取指定存储池基本信息
// @Description get Pool angel
// @Tags        Pool
// @Param       poolName path string true "poolName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.StoragePool
// @Router      /cluster/pools/{poolName} [get]
func (n *PoolController) StoragePoolGet(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
	poolName := ctx.Param("poolName")

	if poolName == "" {
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = "poolName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}
	sp, err := n.m.StoragePoolController().GetStoragePool(poolName)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, sp)
}

// StoragePoolList godoc
// @Summary     摘要 获取存储池列表信息
// @Description list StoragePools
// @Tags        Pool
// @Param       name query string false "name"
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Param       fuzzy query bool false "fuzzy"
// @Param       sort query bool false "sort"
// @Accept      json
// @Produce     json
// @Success     200 {object} api.StoragePoolList
// @Router      /cluster/pools [get]
func (n *PoolController) StoragePoolList(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
	name := ctx.Query("name")
	// 获取path中的page
	page := ctx.Query("page")
	// 获取path中的pageSize
	pageSize := ctx.Query("pageSize")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	var queryPage hwameistorapi.QueryPage
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)
	queryPage.PoolName = name

	lds, err := n.m.StoragePoolController().StoragePoolList(queryPage)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, lds)
}

// StorageNodesGetByPoolName godoc
// @Summary 摘要 获取指定存储池存储节点列表信息
// @Description get StorageNodesGetByPoolName
// @Tags        Pool
// @Param       poolName path string true "poolName"
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Param       nodeName query string false "nodeName"
// @Param       state query string false "state"
// @Param       fuzzy query bool false "fuzzy"
// @Param       sort query bool false "sort"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.StorageNodeListByPool
// @Router      /cluster/pools/{poolName}/nodes [get]
func (n *PoolController) StorageNodesGetByPoolName(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
	storagePoolName := ctx.Param("poolName")

	if storagePoolName == "" {
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = "storagePoolName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	// 获取path中的page
	page := ctx.Query("page")
	// 获取path中的pageSize
	pageSize := ctx.Query("pageSize")

	// 获取path中的nodeName
	nodeName := ctx.Query("nodeName")
	// 获取path中的pageSize
	state := ctx.Query("state")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.NodeState = hwameistorapi.NodeStatefuzzyConvert(state)
	queryPage.PoolName = storagePoolName
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)

	sn, err := n.m.StoragePoolController().GetStorageNodeByPoolName(queryPage)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, sn)
}

// StorageNodeDisksGetByPoolName godoc
// @Summary 摘要 获取指定存储池指定存储节点磁盘列表信息
// @Description get StorageNodeDisksGetByPoolName
// @Tags        Pool
// @Param       nodeName path string true "nodeName"
// @Param       poolName path string true "poolName"
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Param       fuzzy query bool false "fuzzy"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.NodeDiskListByPool
// @Router      /cluster/pools/{poolName}/nodes/{nodeName}/disks [get]
func (n *PoolController) StorageNodeDisksGetByPoolName(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的StoragePoolName
	poolName := ctx.Param("poolName")

	if poolName == "" {
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = "poolName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	// 获取path中的NodeName
	nodeName := ctx.Param("nodeName")

	if nodeName == "" {
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = "nodeName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	// 获取path中的page
	page := ctx.Query("page")
	// 获取path中的pageSize
	pageSize := ctx.Query("pageSize")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	var queryPage hwameistorapi.QueryPage
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)
	queryPage.NodeName = nodeName
	queryPage.PoolName = poolName

	sndisksByPoolName, err := n.m.StoragePoolController().StorageNodeDisksGetByPoolName(queryPage)
	if err != nil {
		failRsp.ErrCode = 500
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, sndisksByPoolName)
}

// StoragePoolExpand godoc
// @Summary     Storage pool expand
// @Description expand new disk for storage pool
// @Tags        Pool
// @Param       body body api.StoragePoolExpansionReqBody true "body"
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.StoragePoolExpansionRspBody
// @Failure     500 {object} api.RspFailBody
// @Router      /cluster/pools/expand [post]
func (n *PoolController) StoragePoolExpand(ctx *gin.Context) {
	var req hwameistorapi.StoragePoolExpansionReqBody
	err := ctx.ShouldBind(&req)
	if err != nil {
		log.Errorf("Unmarshal err = %v", err)
		ctx.JSON(http.StatusInternalServerError, hwameistorapi.RspFailBody{
			ErrCode: http.StatusInternalServerError,
			Desc:    err.Error(),
		})
		return
	}

	if req.NodeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, hwameistorapi.RspFailBody{
			ErrCode: http.StatusNonAuthoritativeInfo,
			Desc:    "nodeName cant be empty",
		})
		return
	}

	if req.DiskType != v1alpha1.DiskClassNameHDD && req.DiskType != v1alpha1.DiskClassNameSSD && req.DiskType != v1alpha1.DiskClassNameNVMe {
		ctx.JSON(http.StatusNonAuthoritativeInfo, hwameistorapi.RspFailBody{
			ErrCode: http.StatusNonAuthoritativeInfo,
			Desc:    "DiskType must be HDD/SSD/NVMe",
		})
		return
	}

	if req.Owner != v1alpha1.LocalStorage && req.Owner != v1alpha1.LocalDiskManager {
		ctx.JSON(http.StatusNonAuthoritativeInfo, hwameistorapi.RspFailBody{
			ErrCode: http.StatusNonAuthoritativeInfo,
			Desc:    fmt.Sprintf("owner must be %s or %s", v1alpha1.LocalStorage, v1alpha1.LocalDiskManager),
		})
		return
	}

	if err = n.m.StoragePoolController().ExpandStoragePool(req.NodeName, req.DiskType, req.Owner); err != nil {
		ctx.JSON(http.StatusInternalServerError, hwameistorapi.RspFailBody{
			ErrCode: http.StatusInternalServerError,
			Desc:    err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, hwameistorapi.StoragePoolExpansionRspBody{
		Success: true,
	})
}
