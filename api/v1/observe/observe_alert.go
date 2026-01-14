package observe

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"main.go/global"
	"main.go/model/common/request"
	"main.go/model/common/response"
	observe "main.go/model/observe"
)

type ObserveAlertApi struct {
}

// CreateAlert 创建告警
func (m *ObserveAlertApi) CreateAlert(c *gin.Context) {
	var req observe.AlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败!", zap.Error(err))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	if err, alert := observeService.CreateAlert(req); err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败: "+err.Error(), c)
	} else {
		response.OkWithData(alert, c)
	}
}

// DeleteAlert 删除告警
func (m *ObserveAlertApi) DeleteAlert(c *gin.Context) {
	idStr := c.Param("alertId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	if err := observeService.DeleteAlert(id); err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
	} else {
		response.OkWithMessage("删除成功", c)
	}
}

// DeleteAlertBatch 批量删除告警
func (m *ObserveAlertApi) DeleteAlertBatch(c *gin.Context) {
	var ids request.IdsReq
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	if err := observeService.DeleteAlertBatch(ids); err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败", c)
	} else {
		response.OkWithMessage("批量删除成功", c)
	}
}

// UpdateAlert 更新告警
func (m *ObserveAlertApi) UpdateAlert(c *gin.Context) {
	idStr := c.Param("alertId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	var req observe.AlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败!", zap.Error(err))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	if err := observeService.UpdateAlert(id, req); err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败: "+err.Error(), c)
	} else {
		response.OkWithMessage("更新成功", c)
	}
}

// GetAlert 根据ID获取告警
func (m *ObserveAlertApi) GetAlert(c *gin.Context) {
	idStr := c.Param("alertId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	if err, alert := observeService.GetAlert(id); err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败", c)
	} else {
		response.OkWithData(alert, c)
	}
}

// GetAlertList 分页获取告警列表
func (m *ObserveAlertApi) GetAlertList(c *gin.Context) {
	var pageInfo request.PageInfo
	_ = c.ShouldBindQuery(&pageInfo)
	status := c.Query("status")
	severity := c.Query("severity")

	if err, list, total := observeService.GetAlertList(pageInfo, status, severity); err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
	} else {
		response.OkWithDetailed(response.PageResult{
			List:       list,
			TotalCount: total,
			CurrPage:   pageInfo.PageNumber,
			PageSize:   pageInfo.PageSize,
		}, "获取成功", c)
	}
}
