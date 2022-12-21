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
	StorageNodeVolumeOperationYamlGet(ctx *gin.Context)
	ReserveStorageNodeDisk(ctx *gin.Context)
	RemoveReserveStorageNodeDisk(ctx *gin.Context)
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
// @Param       name path string false "name"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.StorageNode
// @Router      /nodes/storagenodes/{name} [get]
func (n *NodeController) StorageNodeGet(ctx *gin.Context) {
	// 获取path中的name
	nodeName := ctx.Param("name")

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
// @Accept      json
// @Produce     json
// @Success     200 {object} api.StorageNodeList
// @Router      /nodes/storagenodes [get]
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

	fmt.Println("StorageNodeList driverState = %v", driverState)

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
// @Router      /nodes/storagenode/{nodeName}/migrates [get]
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
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.LocalDiskListByNode
// @Router      /nodes/storagenode/{nodeName}/disks [get]
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
	queryPage.Name = nodeName

	lds, err := n.m.StorageNodeController().LocalDiskListByNode(queryPage)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, lds)
}

// StorageNodeVolumeOperationYamlGet godoc
// @Summary 摘要 获取节点数据卷操作记录yaml信息
// @Description get StorageNodeVolumeOperationYamlGet
// @Tags        Node
// @Param       migrateOperationName path string true "migrateOperationName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.YamlData  "成功"
// @Router      /nodes/storagenodeoperations/{migrateOperationName}/yaml [get]
func (n *NodeController) StorageNodeVolumeOperationYamlGet(ctx *gin.Context) {

	// 获取path中的name
	name := ctx.Param("migrateOperationName")
	fmt.Println("StorageNodeVolumeOperationYamlGet name = %v", name)

	if name == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}
	resourceYamlStr, err := n.m.StorageNodeController().GetStorageNodeVolumeMigrateYamlStr(name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, resourceYamlStr)
}

// ReserveStorageNodeDisk godoc
// @Summary 摘要 预留磁盘
// @Description post ReserveStorageNodeDisk diskname i.g sdb sdc ...
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       diskName path string true "diskName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.DiskReservedRspBody  "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /nodes/storagenode/{nodeName}/disks/{diskName}/reserve [post]
func (n *NodeController) ReserveStorageNodeDisk(ctx *gin.Context) {
	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的diskName
	diskName := ctx.Param("diskName")

	fmt.Println("ReserveStorageNodeDisk nodeName = %v, diskName = %v", nodeName, diskName)

	if nodeName == "" || diskName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, nil)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.DiskName = diskName

	diskReservedRsp, err := n.m.StorageNodeController().ReserveStorageNodeDisk(queryPage, n.diskHandler)
	if err != nil {
		var failRsp hwameistorapi.RspFailBody
		failRsp.ErrCode = 500
		failRsp.Desc = "ReserveStorageNodeDisk Failed:" + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, diskReservedRsp)
}

// RemoveReserveStorageNodeDisk godoc
// @Summary 摘要 解除磁盘预留
// @Description post RemoveReserveStorageNodeDisk
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       diskName path string true "diskName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.DiskRemoveReservedRspBody  "成功"
// @Failure     500 {object}  api.RspFailBody "失败"
// @Router      /nodes/storagenode/{nodeName}/disks/{diskName}/removereserve [post]
func (n *NodeController) RemoveReserveStorageNodeDisk(ctx *gin.Context) {
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

	removeDiskReservedRsp, err := n.m.StorageNodeController().RemoveReserveStorageNodeDisk(queryPage, n.diskHandler)
	if err != nil {
		var failRsp hwameistorapi.RspFailBody
		failRsp.ErrCode = 500
		failRsp.Desc = "ReserveStorageNodeDisk Failed:" + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, removeDiskReservedRsp)
}
