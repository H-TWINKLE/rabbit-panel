package model

// VolumeInfo 存储卷信息
type VolumeInfo struct {
	Name          string            `json:"name"`
	Driver        string            `json:"driver"`
	Mountpoint    string            `json:"mountpoint"`
	Created       string            `json:"created"`
	Scope         string            `json:"scope"`
	Labels        map[string]string `json:"labels"`
	Options       map[string]string `json:"options"`
	UsageData     *VolumeUsageData  `json:"usageData,omitempty"`
	Containers    []string          `json:"containers"`
	ContainerCount int              `json:"container_count"`
	InUse         bool              `json:"in_use"`
}

// VolumeUsageData 存储卷使用数据
type VolumeUsageData struct {
	Size     int64 `json:"size"`
	RefCount int64 `json:"refCount"`
}

// CreateVolumeRequest 创建存储卷请求
type CreateVolumeRequest struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	DriverOpts map[string]string `json:"driverOpts"`
	Labels     map[string]string `json:"labels"`
}

// VolumePruneResult 清理未使用卷结果
type VolumePruneResult struct {
	VolumesDeleted []string `json:"volumesDeleted"`
	SpaceReclaimed uint64   `json:"spaceReclaimed"`
}
