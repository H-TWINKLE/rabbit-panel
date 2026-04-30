package model

// ContainerInfo 容器列表项（API 响应用）
type ContainerInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Status  string `json:"status"`
	Ports   string `json:"ports"`
	Memory  string `json:"memory"`
	Created string `json:"created"`
	State   string `json:"state"`
}

// ImageInfo 镜像列表项
type ImageInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Tag     string `json:"tag"`
	Size    string `json:"size"`
	Created string `json:"created"`
}

// SystemStats 系统监控数据
type SystemStats struct {
	CPU         float64 `json:"cpu"`
	Memory      float64 `json:"memory"`
	MemoryUsed  uint64  `json:"memory_used"`  // KB
	MemoryTotal uint64  `json:"memory_total"` // KB
	Disk        float64 `json:"disk"`
	DiskUsed    uint64  `json:"disk_used"`    // KB
	DiskTotal   uint64  `json:"disk_total"`   // KB
	Time        string  `json:"time"`
}

// MemoryInfo 内存使用信息
type MemoryInfo struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Usage     float64 `json:"usage"`
}

// DiskInfo 磁盘使用信息
type DiskInfo struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Usage     float64 `json:"usage"`
}

// ContainerStats 容器资源统计
type ContainerStats struct {
	CPUPercent     float64 `json:"cpu_percent"`
	CPUCores       int     `json:"cpu_cores"`
	MemoryUsage    int64   `json:"memory_usage"`
	MemoryLimit    int64   `json:"memory_limit"`
	MemoryPercent  float64 `json:"memory_percent"`
	HasMemoryLimit bool    `json:"has_memory_limit"`
	NetworkRx      int64   `json:"network_rx"`
	NetworkTx      int64   `json:"network_tx"`
	BlockRead      int64   `json:"block_read"`
	BlockWrite     int64   `json:"block_write"`
	PIDs           uint64  `json:"pids"`
}

// ExecRequest 执行命令请求
type ExecRequest struct {
	ContainerID string   `json:"container_id"`
	Command     []string `json:"command"`
}

// ExecResponse 执行命令响应
type ExecResponse struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}

// FileInfo 文件信息
type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

// PortMapping 端口映射
type PortMapping struct {
	Host      string `json:"host"`
	Container string `json:"container"`
}

// VolumeMapping 卷映射
type VolumeMapping struct {
	Host      string `json:"host"`
	Container string `json:"container"`
}

// EnvVar 环境变量
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RecreateContainerRequest 重建容器请求
type RecreateContainerRequest struct {
	ContainerID string          `json:"container_id"`
	Name       string          `json:"name"`
	Image      string          `json:"image"`
	Ports      []PortMapping   `json:"ports"`
	Volumes    []VolumeMapping `json:"volumes"`
	Env        []EnvVar        `json:"env"`
	Restart    string          `json:"restart"`
	Network    string          `json:"network"`
	Memory     int64           `json:"memory"`
	CPUs       float64         `json:"cpus"`
	Privileged bool             `json:"privileged"`
	TTY        bool             `json:"tty"`
}

// CreateContainerRequest 创建容器请求
type CreateContainerRequest struct {
	Image     string          `json:"image"`
	Name      string          `json:"name"`
	Ports     []PortMapping   `json:"ports"`
	Volumes   []VolumeMapping `json:"volumes"`
	Env       []EnvVar        `json:"env"`
	Restart   string          `json:"restart"`
	Network   string          `json:"network"`
	Memory    int64           `json:"memory"`
	CPUs      float64         `json:"cpus"`
	Privileged bool            `json:"privileged"`
	TTY       bool             `json:"tty"`
}

// ContainerActionRequest 容器操作请求
type ContainerActionRequest struct {
	ContainerID string `json:"container_id"`
	Action      string `json:"action"` // start, stop, restart, remove
}
