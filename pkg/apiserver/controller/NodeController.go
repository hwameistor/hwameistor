package controller

import (
	"fmt"
	"net/http"
	"strconv"

	mgrpkg "sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gin-gonic/gin"
	hwameistorapi "github.com/hwameistor/hwameistor/pkg/apiserver/api"
	"github.com/hwameistor/hwameistor/pkg/apiserver/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/handler/localdisk"
)

type INodeController interface {
	//RestController
	StorageNodeGet(ctx *gin.Context)
	StorageNodeList(ctx *gin.Context)
	StorageNodeMigrateGet(ctx *gin.Context)
	StorageNodeDisksList(ctx *gin.Context)
	UpdateStorageNodeDisk(ctx *gin.Context)
	GetStorageNodeDisk(ctx *gin.Context)
	StorageNodePoolsList(ctx *gin.Context)
	StorageNodePoolGet(ctx *gin.Context)
	StorageNodePoolDisksList(ctx *gin.Context)
	StorageNodePoolDiskGet(ctx *gin.Context)
}

// NodeController
type NodeController struct {
	m           *manager.ServerManager
	diskHandler *localdisk.Handler
}

func NewNodeController(m *manager.ServerManager, mgr mgrpkg.Manager) INodeController {

	diskHandler := localdisk.NewLocalDiskHandler(mgr.GetClient(),
		mgr.GetEventRecorderFor("localdisk-controller"))

	return &NodeController{m, diskHandler}
}

// StorageNodeGet godoc
// @Summary 摘要 获取指定存储节点
// @Description get StorageNode 驱动状态 [运行中（Ready）,维护中（Maintain）, 离线（Offline)] , 节点状态 [运行中（Ready）,未就绪（NotReady）,未知（Unknown)]
// @Tags        Node
// @Param       nodeName path string false "nodeName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.StorageNode
// @Router      /cluster/nodes/{nodeName} [get]
func (n *NodeController) StorageNodeGet(ctx *gin.Context) {
	// 获取path中的name
	nodeName := ctx.Param("nodeName")

	if nodeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}
	sn, err := n.m.StorageNodeController().GetStorageNode(nodeName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, sn)
}

// StorageNodeList godoc
// @Summary     摘要 获取存储节点列表
// @Description list StorageNode  驱动状态 [运行中（Ready）,维护中（Maintain）, 离线（Offline)] , 节点状态 [运行中（Ready）,未就绪（NotReady）,未知（Unknown)]
// @Tags        Node
// @Param       name query string false "name"
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Param       nodeState query string false "nodeState"
// @Param       driverState query string false "driverState"
// @Param       fuzzy query bool false "fuzzy"
// @Param       sort query bool false "sort"
// @Accept      json
// @Produce     json
// @Success     200 {object} api.StorageNodeList
// @Router      /cluster/nodes [get]
func (n *NodeController) StorageNodeList(ctx *gin.Context) {

	// 获取path中的page
	page := ctx.Query("page")
	// 获取path中的pageSize
	pageSize := ctx.Query("pageSize")
	// 获取path中的nodeState
	nodeState := ctx.Query("nodeState")
	// 获取path中的driverState
	driverState := ctx.Query("driverState")
	// 获取path中的name
	nodeName := ctx.Query("name")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	fmt.Println("StorageNodeList driverState = %v, nodeName = %v", driverState, nodeName)

	var queryPage hwameistorapi.QueryPage
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)
	queryPage.NodeState = hwameistorapi.NodeStatefuzzyConvert(nodeState)
	queryPage.DriverState = hwameistorapi.DriverStatefuzzyConvert(driverState)
	queryPage.Name = nodeName

	sns, err := n.m.StorageNodeController().StorageNodeList(queryPage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, sns)
}

// StorageNodeMigrateGet godoc
// @Summary     摘要 获取指定节点数据卷迁移任务列表
// @Description get StorageNodeMigrate
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Accept      json
// @Produce     json
// @Success     200 {object} api.VolumeOperationListByNode
// @Router      /cluster/nodes/{nodeName}/migrates [get]
func (n *NodeController) StorageNodeMigrateGet(ctx *gin.Context) {

	// 获取path中的name
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

	sns, err := n.m.StorageNodeController().GetStorageNodeMigrate(queryPage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, sns)
}

// StorageNodeDisksList godoc
// @Summary 摘要 获取指定存储节点磁盘列表
// @Description list StorageNodeDisks
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Param       state query string false "state"
// @Param       fuzzy query bool false "fuzzy"
// @Param       sort query bool false "sort"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.LocalDiskListByNode
// @Router      /cluster/nodes/{nodeName}/disks [get]
func (n *NodeController) StorageNodeDisksList(ctx *gin.Context) {
	// 获取path中的name
	nodeName := ctx.Param("nodeName")
	if nodeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	// 获取path中的page
	page := ctx.Query("page")
	// 获取path中的pageSize
	pageSize := ctx.Query("pageSize")

	// 获取path中的state
	state := ctx.Query("state")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	var queryPage hwameistorapi.QueryPage
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)
	queryPage.DiskState = hwameistorapi.DiskStatefuzzyConvert(state)
	queryPage.NodeName = nodeName

	lds, err := n.m.StorageNodeController().LocalDiskListByNode(queryPage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, lds)
}

// UpdateStorageNodeDisk godoc
// @Summary 摘要 更新磁盘
// @Description post UpdateStorageNodeDisk diskname i.g sdb sdc ...
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       diskName path string true "diskName"
// @Param       body body api.DiskReqBody false "reqBody"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.DiskReqBody  "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/nodes/{nodeName}/disks/{diskName} [post]
func (n *NodeController) UpdateStorageNodeDisk(ctx *gin.Context) {
	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的diskName
	diskName := ctx.Param("diskName")

	var drb hwameistorapi.DiskReqBody
	err := ctx.ShouldBind(&drb)
	if err != nil {
		fmt.Errorf("Unmarshal err = %v", err)
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}
	reserve := drb.Reserve

	if nodeName == "" || diskName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.DiskName = diskName

	if reserve == true {
		diskReservedRsp, err := n.m.StorageNodeController().ReserveStorageNodeDisk(queryPage, n.diskHandler)
		if err != nil {
			var failRsp hwameistorapi.RspFailBody
			failRsp.ErrCode = 500
			failRsp.Desc = "ReserveStorageNodeDisk Failed:" + err.Error()
			ctx.JSON(http.StatusInternalServerError, failRsp)
			return
		}
		ctx.JSON(http.StatusOK, diskReservedRsp)
	} else {
		removeDiskReservedRsp, err := n.m.StorageNodeController().ReleaseReserveStorageNodeDisk(queryPage, n.diskHandler)
		if err != nil {
			var failRsp hwameistorapi.RspFailBody
			failRsp.ErrCode = 500
			failRsp.Desc = "ReserveStorageNodeDisk Failed:" + err.Error()
			ctx.JSON(http.StatusInternalServerError, failRsp)
			return
		}

		ctx.JSON(http.StatusOK, removeDiskReservedRsp)
	}
}

// GetStorageNodeDisk godoc
// @Summary 摘要 获取指定磁盘信息
// @Description get GetStorageNodeDisk diskname i.g sdb sdc ...
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       diskName path string true "diskName"
// @Param       fuzzy query bool false "fuzzy"
// @Param       sort query bool false "sort"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.LocalDiskInfo  "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/nodes/{nodeName}/disks/{diskName} [get]
func (n *NodeController) GetStorageNodeDisk(ctx *gin.Context) {
	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的diskName
	diskName := ctx.Param("diskName")

	if nodeName == "" || diskName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.DiskName = diskName

	return
}

// StorageNodePoolsList godoc
// @Summary 摘要 获取指定节点存储池列表信息
// @Description get StorageNodePoolsList
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       fuzzy query bool false "fuzzy"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.StoragePoolList  "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/nodes/{nodeName}/pools [get]
func (n *NodeController) StorageNodePoolsList(ctx *gin.Context) {
	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的diskName
	diskName := ctx.Param("diskName")

	if nodeName == "" || diskName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.DiskName = diskName

	return
}

// StorageNodePoolGet godoc
// @Summary 摘要 获取指定节点指定存储池信息
// @Description get StorageNodePoolGet
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       poolName path string true "poolName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.StoragePool  "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/nodes/{nodeName}/pools/{poolName} [get]
func (n *NodeController) StorageNodePoolGet(ctx *gin.Context) {
	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的diskName
	diskName := ctx.Param("diskName")

	if nodeName == "" || diskName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.DiskName = diskName

	return
}

// StorageNodePoolDisksList godoc
// @Summary 摘要 获取指定节点指定存储池磁盘列表信息
// @Description get StorageNodePoolDisksList
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       poolName path string true "poolName"
// @Param       fuzzy query bool false "fuzzy"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.LocalDisksItemsList  "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/nodes/{nodeName}/pools/{poolName}/disks [get]
func (n *NodeController) StorageNodePoolDisksList(ctx *gin.Context) {
	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的diskName
	diskName := ctx.Param("diskName")

	if nodeName == "" || diskName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.DiskName = diskName

	return
}

// StorageNodePoolDiskGet godoc
// @Summary 摘要 获取指定节点指定存储池指定磁盘信息
// @Description get StorageNodePoolDiskGet
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       poolName path string true "poolName"
// @Param       diskName path string true "diskName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.LocalDiskInfo  "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /cluster/nodes/{nodeName}/pools/{poolName}/disks/{diskName} [get]
func (n *NodeController) StorageNodePoolDiskGet(ctx *gin.Context) {
	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的diskName
	diskName := ctx.Param("diskName")

	if nodeName == "" || diskName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.DiskName = diskName

	return
}
