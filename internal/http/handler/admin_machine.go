package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"nexus/internal/database"
	"nexus/internal/model"

	"github.com/gin-gonic/gin"
)

// ---------- Request types ----------

type createMachineRequest struct {
	Name  string `json:"name" binding:"required"`
	Notes string `json:"notes"`
}

type updateMachineRequest struct {
	Name     string `json:"name"`
	Notes    string `json:"notes"`
	IsActive *bool  `json:"is_active"`
}

// ---------- Handlers ----------

// AdminListMachines returns all machines with server count.
// GET /api/admin/machines
func AdminListMachines(c *gin.Context) {
	var machines []model.Machine
	database.DB.Order("id ASC").Find(&machines)

	type machineResp struct {
		model.Machine
		ServerCount int64 `json:"servers_count"`
	}

	result := make([]machineResp, 0, len(machines))
	for _, m := range machines {
		var count int64
		database.DB.Model(&model.Node{}).Where("machine_id = ?", m.ID).Count(&count)
		// Mask token
		m.Token = ""
		result = append(result, machineResp{
			Machine:     m,
			ServerCount: count,
		})
	}

	Success(c, result)
}

// AdminGetMachine returns a single machine with node count.
// GET /api/admin/machines/:id
func AdminGetMachine(c *gin.Context) {
	var machine model.Machine
	if err := database.DB.First(&machine, c.Param("id")).Error; err != nil {
		NotFound(c, "机器不存在")
		return
	}

	var count int64
	database.DB.Model(&model.Node{}).Where("machine_id = ?", machine.ID).Count(&count)

	machine.Token = "" // Mask token
	Success(c, gin.H{
		"machine":       machine,
		"servers_count": count,
	})
}

// AdminGetMachineLoadHistory returns machine load history.
// GET /api/admin/machines/:id/history
func AdminGetMachineLoadHistory(c *gin.Context) {
	var machine model.Machine
	if err := database.DB.First(&machine, c.Param("id")).Error; err != nil {
		NotFound(c, "机器不存在")
		return
	}

	limit := c.DefaultQuery("limit", "60")
	rangeHours := c.DefaultQuery("range_hours", "6")

	var parsedLimit int
	var parsedHours int
	parsedLimit = 60
	parsedHours = 6
	if l, err := strconv.Atoi(limit); err == nil && l >= 10 && l <= 1440 {
		parsedLimit = l
	}
	if h, err := strconv.Atoi(rangeHours); err == nil && h >= 1 && h <= 24 {
		parsedHours = h
	}

	cutoff := time.Now().Add(-time.Duration(parsedHours) * time.Hour).Unix()

	var history []model.MachineLoadHistory
	database.DB.Where("machine_id = ? AND recorded_at >= ?", machine.ID, cutoff).
		Order("recorded_at ASC").
		Limit(parsedLimit).
		Find(&history)

	Success(c, history)
}

// AdminCreateMachine creates a new machine and returns its token + install command.
// POST /api/admin/machines
func AdminCreateMachine(c *gin.Context) {
	var req createMachineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请提供机器名称")
		return
	}

	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		InternalError(c, "生成Token失败")
		return
	}
	token := hex.EncodeToString(tokenBytes)

	machine := model.Machine{
		Name:     req.Name,
		Notes:    req.Notes,
		Token:    token,
		IsActive: true,
	}

	if err := database.DB.Create(&machine).Error; err != nil {
		InternalError(c, "创建机器失败")
		return
	}

	// Build install command (Xboard-style)
	panelURL := database.GetSetting("app_url")
	if panelURL == "" {
		panelURL = fmt.Sprintf("%s://%s", scheme(c), c.Request.Host)
	}
	installCommand := fmt.Sprintf(
		`curl -fsSL https://github.com/TIUCSIB/nexus-agent/raw/main/install.sh | sudo bash -s -- --mode machine --panel %s --token %s --machine-id %d`,
		panelURL, token, machine.ID,
	)

	Success(c, gin.H{
		"id":              machine.ID,
		"name":            machine.Name,
		"token":           token,
		"notes":           machine.Notes,
		"is_active":       machine.IsActive,
		"install_command": installCommand,
	})
}

// AdminUpdateMachine updates machine name/notes/status.
// PUT /api/admin/machines/:id
func AdminUpdateMachine(c *gin.Context) {
	var machine model.Machine
	if err := database.DB.First(&machine, c.Param("id")).Error; err != nil {
		NotFound(c, "机器不存在")
		return
	}

	var req updateMachineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "无效的请求参数")
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Notes != "" {
		updates["notes"] = req.Notes
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&machine).Updates(updates).Error; err != nil {
			InternalError(c, "保存失败")
			return
		}
	}

	Success(c, true)
}

// AdminDeleteMachine deletes a machine and detaches its nodes.
// DELETE /api/admin/machines/:id
func AdminDeleteMachine(c *gin.Context) {
	var machine model.Machine
	if err := database.DB.First(&machine, c.Param("id")).Error; err != nil {
		NotFound(c, "机器不存在")
		return
	}

	// Detach nodes
	database.DB.Model(&model.Node{}).Where("machine_id = ?", machine.ID).Update("machine_id", nil)

	if err := database.DB.Delete(&machine).Error; err != nil {
		InternalError(c, "删除失败")
		return
	}

	Success(c, true)
}

// AdminResetMachineToken resets the machine's token.
// POST /api/admin/machines/:id/reset-token
func AdminResetMachineToken(c *gin.Context) {
	var machine model.Machine
	if err := database.DB.First(&machine, c.Param("id")).Error; err != nil {
		NotFound(c, "机器不存在")
		return
	}

	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		InternalError(c, "生成Token失败")
		return
	}
	token := hex.EncodeToString(tokenBytes)

	if err := database.DB.Model(&machine).Update("token", token).Error; err != nil {
		InternalError(c, "重置失败")
		return
	}

	Success(c, gin.H{"token": token})
}

// AdminGetMachineInstallCommand returns the one-click install command for a machine.
// POST /api/admin/machines/:id/install-command
func AdminGetMachineInstallCommand(c *gin.Context) {
	var machine model.Machine
	if err := database.DB.First(&machine, c.Param("id")).Error; err != nil {
		NotFound(c, "机器不存在")
		return
	}

	panelURL := database.GetSetting("app_url")
	if panelURL == "" {
		panelURL = fmt.Sprintf("%s://%s", scheme(c), c.Request.Host)
	}
	installCommand := fmt.Sprintf(
		`curl -fsSL https://github.com/TIUCSIB/nexus-agent/raw/main/install.sh | sudo bash -s -- --mode machine --panel %s --token %s --machine-id %d`,
		panelURL, machine.Token, machine.ID,
	)

	Success(c, gin.H{"command": installCommand})
}

// AdminListMachineNodes returns nodes under a machine.
// GET /api/admin/machines/:id/nodes
func AdminListMachineNodes(c *gin.Context) {
	var machine model.Machine
	if err := database.DB.First(&machine, c.Param("id")).Error; err != nil {
		NotFound(c, "机器不存在")
		return
	}

	var nodes []model.Node
	database.DB.Where("machine_id = ?", machine.ID).Order("sort ASC, id ASC").Find(&nodes)

	Success(c, nodes)
}

// scheme returns the request scheme (http/https).
func scheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	s := c.Request.Header.Get("X-Forwarded-Proto")
	if s == "https" {
		return "https"
	}
	return "http"
}

// MachineHeartbeat records machine alive status (called by agent).
// POST /api/internal/machine/:id/heartbeat
func MachineHeartbeat(c *gin.Context) {
	var machine model.Machine
	if err := database.DB.First(&machine, c.Param("id")).Error; err != nil {
		NotFound(c, "机器不存在")
		return
	}

	now := time.Now()
	database.DB.Model(&machine).Update("last_seen_at", &now)

	Success(c, gin.H{"data": true})
}

// MachineGetNodes returns all active nodes under a machine (for agent discovery).
// GET /api/internal/machine/:id/nodes
func MachineGetNodes(c *gin.Context) {
	var machine model.Machine
	if err := database.DB.First(&machine, c.Param("id")).Error; err != nil {
		NotFound(c, "机器不存在")
		return
	}

	var nodes []model.Node
	database.DB.Where("machine_id = ? AND status = 1", machine.ID).Order("sort ASC, id ASC").Find(&nodes)

	type nodeInfo struct {
		ID       uint   `json:"id"`
		Name     string `json:"name"`
		Protocol string `json:"type"`
		Address  string `json:"host"`
		Port     int    `json:"port"`
		Sort     int    `json:"sort"`
		Status   int    `json:"status"`
	}

	result := make([]nodeInfo, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, nodeInfo{
			ID:       n.ID,
			Name:     n.Name,
			Protocol: n.Protocol,
			Address:  n.Address,
			Port:     n.Port,
			Sort:     n.Sort,
			Status:   n.Status,
		})
	}

	Success(c, gin.H{
		"nodes":        result,
		"pull_interval": database.GetSettingInt("node_pull_interval", 60),
		"push_interval": database.GetSettingInt("node_push_interval", 60),
	})
}

// MachineReportLoad records machine load status (called by agent).
// POST /api/internal/machine/:id/load
func MachineReportLoad(c *gin.Context) {
	var machine model.Machine
	if err := database.DB.First(&machine, c.Param("id")).Error; err != nil {
		NotFound(c, "机器不存在")
		return
	}

	var req struct {
		CPU         float64 `json:"cpu"`
		MemTotal    int64   `json:"mem_total"`
		MemUsed     int64   `json:"mem_used"`
		DiskTotal   int64   `json:"disk_total"`
		DiskUsed    int64   `json:"disk_used"`
		NetInSpeed  float64 `json:"net_in_speed"`
		NetOutSpeed float64 `json:"net_out_speed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "无效的负载数据")
		return
	}

	now := time.Now()

	// Update current load status on machine
	database.DB.Model(&machine).Updates(map[string]interface{}{
		"load_status": model.LoadStatus{
			CPU:         req.CPU,
			MemTotal:    req.MemTotal,
			MemUsed:     req.MemUsed,
			DiskTotal:   req.DiskTotal,
			DiskUsed:    req.DiskUsed,
			NetInSpeed:  req.NetInSpeed,
			NetOutSpeed: req.NetOutSpeed,
		},
		"last_seen_at": &now,
	})

	// Record load history
	history := model.MachineLoadHistory{
		MachineID:   machine.ID,
		CPU:         req.CPU,
		MemTotal:    req.MemTotal,
		MemUsed:     req.MemUsed,
		DiskTotal:   req.DiskTotal,
		DiskUsed:    req.DiskUsed,
		NetInSpeed:  req.NetInSpeed,
		NetOutSpeed: req.NetOutSpeed,
		RecordedAt:  now.Unix(),
	}
	database.DB.Create(&history)

	Success(c, gin.H{"data": true})
}