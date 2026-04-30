package model

import "time"

// 节点模式
const (
	ModeMaster = "master"
	ModeWorker = "worker"
)

// 节点状态
const (
	NodeStatusOnline  = "online"
	NodeStatusOffline = "offline"
	NodeStatusError   = "error"
)

// NodeInfo 节点信息
type NodeInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Address     string            `json:"address"`    // IP:Port
	Mode        string            `json:"mode"`       // master / worker
	Status      string            `json:"status"`     // online, offline, error
	CPU         float64           `json:"cpu"`        // CPU 使用率 %
	Memory      float64           `json:"memory"`     // 内存使用率 %
	Disk        float64           `json:"disk"`       // 磁盘使用率 %
	Containers  int               `json:"containers"` // 容器数量
	LastSeen    time.Time         `json:"-"`          // 内部使用
	LastSeenStr string            `json:"last_seen"`  // API 响应用
	Labels      map[string]string `json:"labels"`     // 节点标签
}

// ScheduleRequest 容器调度请求
type ScheduleRequest struct {
	Image       string            `json:"image"`
	Name        string            `json:"name"`
	Ports       map[string]string `json:"ports"`       // "8080:80"
	Env         map[string]string `json:"env"`
	Labels      map[string]string `json:"labels"`
	NodeID      string            `json:"node_id"`     // 指定节点（可选）
	Constraints map[string]string `json:"constraints"` // 调度约束（可选）
}

// ScheduleResponse 容器调度响应
type ScheduleResponse struct {
	Status    string                 `json:"status"`
	NodeID    string                 `json:"node_id"`
	NodeName  string                 `json:"node_name"`
	Container map[string]interface{} `json:"container"`
}
