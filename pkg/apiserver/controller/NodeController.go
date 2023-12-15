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

type INodeController interface {
	StorageNodeGet(ctx *gin.Context)
	StorageNodeList(ctx *gin.Context)
	StorageNodeMigrateGet(ctx *gin.Context)
	StorageNodeDisksList(ctx *gin.Context)
	UpdateStorageNodeDisk(ctx *gin.Context)
	SetStorageNodeDiskOwner(ctx *gin.Context)
	GetStorageNodeDisk(ctx *gin.Context)
	StorageNodePoolsList(ctx *gin.Context)
	StorageNodePoolGet(ctx *gin.Context)
	StorageNodePoolDisksList(ctx *gin.Context)
	StorageNodePoolDiskGet(ctx *gin.Context)
	StorageNodeUpdate(ctx *gin.Context)
}

type NodeController struct {
	m *manager.ServerManager
}

func NewNodeController(m *manager.ServerManager) INodeController {
	return &NodeController{m}
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
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
	nodeName := ctx.Param("nodeName")

	if nodeName == "" {
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = "nodeName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}
	sn, err := n.m.StorageNodeController().GetStorageNode(nodeName)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
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
	var failRsp hwameistorapi.RspFailBody

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

	log.Infof("StorageNodeList driverState = %v, nodeName = %v", driverState, nodeName)

	var queryPage hwameistorapi.QueryPage
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)
	queryPage.NodeState = hwameistorapi.NodeStatefuzzyConvert(nodeState)
	queryPage.DriverState = hwameistorapi.DriverStatefuzzyConvert(driverState)
	queryPage.Name = nodeName

	sns, err := n.m.StorageNodeController().StorageNodeList(queryPage)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
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
// @Param       migrateOperationName query string false "migrateOperationName"
// @Param       volumeName query string false "volumeName"
// @Param       state query string false "state"
// @Accept      json
// @Produce     json
// @Success     200 {object} api.VolumeOperationListByNode
// @Router      /cluster/nodes/{nodeName}/migrates [get]
func (n *NodeController) StorageNodeMigrateGet(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
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
	// 获取path中的migrateOperationName
	migrateOperationName := ctx.Query("migrateOperationName")
	// 获取path中的volumeName
	volumeName := ctx.Query("volumeName")
	// 获取path中的state
	state := ctx.Query("state")

	p, _ := strconv.ParseInt(page, 10, 32)
	ps, _ := strconv.ParseInt(pageSize, 10, 32)

	var queryPage hwameistorapi.QueryPage
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)
	queryPage.NodeName = nodeName
	queryPage.OperationName = migrateOperationName
	queryPage.VolumeName = volumeName
	queryPage.OperationState = hwameistorapi.OperationStatefuzzyConvert(state)

	sns, err := n.m.StorageNodeController().GetStorageNodeMigrate(queryPage)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
	}

	ctx.JSON(http.StatusOK, sns)
}

// StorageNodeDisksList godoc
// @Summary 摘要 获取指定存储节点磁盘列表
// @Description list StorageNodeDisks 状态枚举 （Active、Inactive、Unknown、Pending、Available、Bound)
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
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的name
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
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, lds)
}

// UpdateStorageNodeDisk godoc
// @Summary 摘要 更新磁盘
// @Description post UpdateStorageNodeDisk devicePath i.g /dev/sdb /dev/sdc ...
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       devicePath path string true "devicePath"
// @Param       body body api.DiskReqBody false "reqBody"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.DiskReqBody
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/nodes/{nodeName}/disks/{devicePath} [post]
func (n *NodeController) UpdateStorageNodeDisk(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的devicePath
	devicePath := ctx.Param("devicePath")
	log.Infof("devicePath = %v", devicePath)

	var drb hwameistorapi.DiskReqBody
	err := ctx.ShouldBind(&drb)
	if err != nil {
		log.Errorf("Unmarshal err = %v", err)
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}
	reserve := drb.Reserve

	if nodeName == "" || devicePath == "" {
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = "nodeName or devicePath cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.DeviceShortPath = devicePath

	if reserve == true {
		diskReservedRsp, err := n.m.StorageNodeController().ReserveStorageNodeDisk(queryPage)
		if err != nil {
			failRsp.ErrCode = http.StatusInternalServerError
			failRsp.Desc = "ReserveStorageNodeDisk Failed:" + err.Error()
			ctx.JSON(http.StatusInternalServerError, failRsp)
			return
		}
		ctx.JSON(http.StatusOK, diskReservedRsp)
	} else {
		removeDiskReservedRsp, err := n.m.StorageNodeController().RemoveReserveStorageNodeDisk(queryPage)
		if err != nil {
			failRsp.ErrCode = http.StatusInternalServerError
			failRsp.Desc = "ReserveStorageNodeDisk Failed:" + err.Error()
			ctx.JSON(http.StatusInternalServerError, failRsp)
			return
		}

		ctx.JSON(http.StatusOK, removeDiskReservedRsp)
	}
}

// SetStorageNodeDiskOwner godoc
// @Summary 摘要 更新磁盘owner
// @Description post SetStorageNodeDiskOwner devicePath i.g /dev/sdb /dev/sdc ...
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       devicePath path string true "devicePath"
// @Param       body body api.DiskReqBody false "reqBody"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.DiskReqBody
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/nodes/{nodeName}/disks/{devicePath}/owner [post]
func (n *NodeController) SetStorageNodeDiskOwner(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody
	nodeName := ctx.Param("nodeName")
	devicePath := ctx.Param("devicePath")
	log.Infof("devicePath = %v", devicePath)

	var drb hwameistorapi.DiskOwnerReqBody
	err := ctx.ShouldBind(&drb)
	if err != nil {
		log.Errorf("Unmarshal err = %v", err)
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = err.Error()
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}
	owner := drb.Owner

	if nodeName == "" || devicePath == "" {
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = "nodeName or devicePath cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}
	if owner != v1alpha1.LocalStorage && owner != v1alpha1.LocalDiskManager {
		// check owner reasonable
		ctx.JSON(http.StatusNonAuthoritativeInfo, hwameistorapi.RspFailBody{
			ErrCode: http.StatusNonAuthoritativeInfo,
			Desc:    fmt.Sprintf("owner must be %s or %s", v1alpha1.LocalStorage, v1alpha1.LocalDiskManager),
		})
		return
	}
	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.DeviceShortPath = devicePath
	queryPage.Owner = owner

	//if owner is system, can`t change owner
	ownerRspBody, err := n.m.StorageNodeController().SetStorageNodeDiskOwner(queryPage)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = "ReserveStorageNodeDisk Failed:" + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}
	ctx.JSON(http.StatusOK, ownerRspBody)

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
// @Success     200 {object}  api.LocalDiskInfo
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/nodes/{nodeName}/disks/{diskName} [get]
func (n *NodeController) GetStorageNodeDisk(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的diskName
	diskName := ctx.Param("diskName")

	if nodeName == "" || diskName == "" {
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = "nodeName or diskName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.DiskName = diskName

	localDiskInfo, err := n.m.StorageNodeController().GetStorageNodeDisk(queryPage)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = "GetStorageNodeDisk Failed:" + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, localDiskInfo)
}

// StorageNodePoolsList godoc
// @Summary 摘要 获取指定节点存储池列表信息
// @Description get StorageNodePoolsList
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       fuzzy query bool false "fuzzy"
// @Param       page query int32 true "page"
// @Param       pageSize query int32 true "pageSize"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.StoragePoolList
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/nodes/{nodeName}/pools [get]
func (n *NodeController) StorageNodePoolsList(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的nodeName
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
	queryPage.NodeName = nodeName
	queryPage.Page = int32(p)
	queryPage.PageSize = int32(ps)

	storagePoolList, err := n.m.StorageNodeController().StorageNodePoolsList(queryPage)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = "StorageNodePoolsList Failed:" + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, storagePoolList)
}

// StorageNodePoolGet godoc
// @Summary 摘要 获取指定节点指定存储池信息
// @Description get StorageNodePoolGet
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       poolName path string true "poolName"
// @Accept      json
// @Produce     json
// @Success     200 {object}  api.StoragePool
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/nodes/{nodeName}/pools/{poolName} [get]
func (n *NodeController) StorageNodePoolGet(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的diskName
	poolName := ctx.Param("poolName")

	if nodeName == "" || poolName == "" {
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = "nodeName or poolName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.PoolName = poolName

	storagePool, err := n.m.StorageNodeController().StorageNodePoolGet(queryPage)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = "StorageNodePoolGet Failed:" + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, storagePool)
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
// @Success     200 {object}  api.LocalDisksItemsList
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/nodes/{nodeName}/pools/{poolName}/disks [get]
func (n *NodeController) StorageNodePoolDisksList(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的poolName
	poolName := ctx.Param("poolName")

	if nodeName == "" || poolName == "" {
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = "nodeName or poolName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.PoolName = poolName

	localDisksItemsList, err := n.m.StorageNodeController().StorageNodePoolDisksList(queryPage)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = "StorageNodePoolDisksList Failed:" + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, localDisksItemsList)
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
// @Success     200 {object}  api.LocalDiskInfo
// @Failure     500 {object}  api.RspFailBody
// @Router      /cluster/nodes/{nodeName}/pools/{poolName}/disks/{diskName} [get]
func (n *NodeController) StorageNodePoolDiskGet(ctx *gin.Context) {
	var failRsp hwameistorapi.RspFailBody

	// 获取path中的nodeName
	nodeName := ctx.Param("nodeName")

	// 获取path中的poolName
	poolName := ctx.Param("poolName")

	// 获取path中的diskName
	diskName := ctx.Param("diskName")

	if nodeName == "" || diskName == "" || poolName == "" {
		failRsp.ErrCode = http.StatusNonAuthoritativeInfo
		failRsp.Desc = "nodeName or diskName or poolName cannot be empty"
		ctx.JSON(http.StatusNonAuthoritativeInfo, failRsp)
		return
	}

	var queryPage hwameistorapi.QueryPage
	queryPage.NodeName = nodeName
	queryPage.DiskName = diskName
	queryPage.PoolName = poolName

	localDiskInfo, err := n.m.StorageNodeController().StorageNodePoolDiskGet(queryPage)
	if err != nil {
		failRsp.ErrCode = http.StatusInternalServerError
		failRsp.Desc = "StorageNodePoolDiskGet Failed:" + err.Error()
		ctx.JSON(http.StatusInternalServerError, failRsp)
		return
	}

	ctx.JSON(http.StatusOK, localDiskInfo)
}

// StorageNodeUpdate godoc
// @Summary     Set node enable/disable
// @Tags        Node
// @Param       nodeName path string true "nodeName"
// @Param       body body api.NodeUpdateReqBody true "body"
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} api.NodeUpdateRspBody
// @Failure     500 {object} api.RspFailBody
// @Router      /cluster/nodes/{nodeName} [post]
func (n *NodeController) StorageNodeUpdate(ctx *gin.Context) {
	nodeName := ctx.Param("nodeName")

	if nodeName == "" {
		ctx.JSON(http.StatusNonAuthoritativeInfo, hwameistorapi.RspFailBody{
			ErrCode: http.StatusNonAuthoritativeInfo,
			Desc:    "nodeName cant be empty",
		})
		return
	}

	var req hwameistorapi.NodeUpdateReqBody
	err := ctx.ShouldBind(&req)
	if err != nil {
		log.Errorf("Unmarshal err = %v", err)
		ctx.JSON(http.StatusInternalServerError, hwameistorapi.RspFailBody{
			ErrCode: http.StatusInternalServerError,
			Desc:    err.Error(),
		})
		return
	}

	if req.Enable == nil {
		ctx.JSON(http.StatusNonAuthoritativeInfo, hwameistorapi.RspFailBody{
			ErrCode: http.StatusNonAuthoritativeInfo,
			Desc:    "enable parameter cant be empty",
		})
		return
	}

	err = n.m.StorageNodeController().UpdateLocalStorageNode(nodeName, *req.Enable)
	if err != nil {
		log.Errorf("Fail to update local disk node, req = %v, err = %v", req, err)
		ctx.JSON(http.StatusInternalServerError, hwameistorapi.RspFailBody{
			ErrCode: http.StatusInternalServerError,
			Desc:    err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, hwameistorapi.NodeUpdateRspBody{Success: true})
}
