package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
)

type IPoolController interface {
	//RestController
	StoragePoolGet(ctx *gin.Context)
	StoragePoolList(ctx *gin.Context)
	StorageNodesGetByPoolName(ctx *gin.Context)
	StorageNodeDisksGetByPoolName(ctx *gin.Context)
}

// PoolController
type PoolController struct {
	m *manager.ServerManager
}

func NewPoolController(m *manager.ServerManager) IPoolController {
	return &PoolController{m}
}

// StoragePoolGet godoc
// @Summary 摘要 获取指定存储池基本信息
// @Description get Pool
// @Tags        Pool
// @Param       name path string true "name"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.StoragePool
// @Router      /pools/storagepools/{name} [get]
func (n *PoolController) StoragePoolGet(ctx *gin.Context) {
	// 获取path中的name
	poolName := ctx.Param("name")

	if poolName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}
	sp, err := n.m.StoragePoolController().GetStoragePool(poolName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
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
// @Accept      json
// @Produce     json
// @Success     200 {object} api.StoragePoolList
// @Router      /pools/storagepools [get]
func (n *PoolController) StoragePoolList(ctx *gin.Context) {

	// 获取path中的page
	page := ctx.Query("page")
	// 获取path中的pageSize
	pageSize := ctx.Query("pageSize")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	lds, err := n.m.StoragePoolController().StoragePoolList(int32(p), int32(ps))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, lds)
}

// StorageNodesGetByPoolName godoc
// @Summary 摘要 获取指定存储池存储节点列表信息
// @Description get StorageNodesGetByPoolName
// @Tags        Pool
// @Param       storagePoolName path string true "storagePoolName"
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.StorageNodeListByPool
// @Router      /pools/storagepool/{storagePoolName}/nodes [get]
func (n *PoolController) StorageNodesGetByPoolName(ctx *gin.Context) {
	// 获取path中的name
	storagePoolName := ctx.Param("storagePoolName")

	if storagePoolName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	// 获取path中的page
	page := ctx.Query("page")
	// 获取path中的pageSize
	pageSize := ctx.Query("pageSize")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	sn, err := n.m.StoragePoolController().GetStorageNodeByPoolName(storagePoolName, int32(p), int32(ps))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, sn)
}

// StorageNodeDisksGetByPoolName godoc
// @Summary 摘要 获取指定存储池指定存储节点磁盘列表信息
// @Description get StorageNodeDisksGetByPoolName
// @Tags        Pool
// @Param       nodeName path string true "nodeName"
// @Param       storagePoolName path string true "storagePoolName"
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.NodeDiskListByPool
// @Router      /pools/storagepool/{storagePoolName}/nodes/{nodeName}/disks [get]
func (n *PoolController) StorageNodeDisksGetByPoolName(ctx *gin.Context) {
	// 获取path中的StoragePoolName
	storagePoolName := ctx.Param("storagePoolName")

	if storagePoolName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	// 获取path中的NodeName
	nodeName := ctx.Param("nodeName")

	if nodeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
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
	queryPage.PoolName = storagePoolName

	sndisksByPoolName, err := n.m.StoragePoolController().StorageNodeDisksGetByPoolName(queryPage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, sndisksByPoolName)
}
