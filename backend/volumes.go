package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
)

// VolumeInfo 存储卷信息
type VolumeInfo struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	Created    string            `json:"created"`
	Scope      string            `json:"scope"`
	Labels     map[string]string `json:"labels"`
	Options    map[string]string `json:"options"`
	UsageData  *VolumeUsageData  `json:"usageData,omitempty"`
	Containers []string          `json:"containers"`
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

// VolumePruneResult 清理存储卷结果
type VolumePruneResult struct {
	VolumesDeleted []string `json:"volumesDeleted"`
	SpaceReclaimed uint64   `json:"spaceReclaimed"`
}

// handleVolumesList 获取存储卷列表
func handleVolumesList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	volumeList, err := dockerClient.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		http.Error(w, "获取存储卷列表失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取所有容器以检查卷的使用情况
	containers, _ := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	volumeContainers := make(map[string][]string)
	for _, c := range containers {
		for _, m := range c.Mounts {
			if m.Type == "volume" {
				containerName := ""
				if len(c.Names) > 0 {
					containerName = strings.TrimPrefix(c.Names[0], "/")
				} else {
					containerName = c.ID[:12]
				}
				volumeContainers[m.Name] = append(volumeContainers[m.Name], containerName)
			}
		}
	}

	volumes := make([]VolumeInfo, 0, len(volumeList.Volumes))
	for _, v := range volumeList.Volumes {
		vol := VolumeInfo{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			Created:    v.CreatedAt,
			Scope:      v.Scope,
			Labels:     v.Labels,
			Options:    v.Options,
			Containers: volumeContainers[v.Name],
		}
		if vol.Containers == nil {
			vol.Containers = []string{}
		}
		if v.UsageData != nil {
			vol.UsageData = &VolumeUsageData{
				Size:     v.UsageData.Size,
				RefCount: v.UsageData.RefCount,
			}
		}
		volumes = append(volumes, vol)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(volumes)
}

// handleVolumesCreate 创建存储卷
func handleVolumesCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req CreateVolumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求参数错误: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "存储卷名称不能为空", http.StatusBadRequest)
		return
	}

	if req.Driver == "" {
		req.Driver = "local"
	}

	ctx := context.Background()
	vol, err := dockerClient.VolumeCreate(ctx, volume.CreateOptions{
		Name:       req.Name,
		Driver:     req.Driver,
		DriverOpts: req.DriverOpts,
		Labels:     req.Labels,
	})
	if err != nil {
		http.Error(w, "创建存储卷失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result := VolumeInfo{
		Name:       vol.Name,
		Driver:     vol.Driver,
		Mountpoint: vol.Mountpoint,
		Created:    vol.CreatedAt,
		Scope:      vol.Scope,
		Labels:     vol.Labels,
		Options:    vol.Options,
		Containers: []string{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleVolumesInspect 获取存储卷详情
func handleVolumesInspect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 路径中提取存储卷名称
	path := r.URL.Path
	name := strings.TrimPrefix(path, "/api/volumes/")
	if name == "" || name == path {
		http.Error(w, "存储卷名称不能为空", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	vol, err := dockerClient.VolumeInspect(ctx, name)
	if err != nil {
		http.Error(w, "获取存储卷详情失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取使用此卷的容器
	containers, _ := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	var volumeContainers []string
	for _, c := range containers {
		for _, m := range c.Mounts {
			if m.Type == "volume" && m.Name == name {
				containerName := ""
				if len(c.Names) > 0 {
					containerName = strings.TrimPrefix(c.Names[0], "/")
				} else {
					containerName = c.ID[:12]
				}
				volumeContainers = append(volumeContainers, containerName)
				break
			}
		}
	}
	if volumeContainers == nil {
		volumeContainers = []string{}
	}

	result := VolumeInfo{
		Name:       vol.Name,
		Driver:     vol.Driver,
		Mountpoint: vol.Mountpoint,
		Created:    vol.CreatedAt,
		Scope:      vol.Scope,
		Labels:     vol.Labels,
		Options:    vol.Options,
		Containers: volumeContainers,
	}
	if vol.UsageData != nil {
		result.UsageData = &VolumeUsageData{
			Size:     vol.UsageData.Size,
			RefCount: vol.UsageData.RefCount,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleVolumesRemove 删除存储卷
func handleVolumesRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 路径中提取存储卷名称
	path := r.URL.Path
	name := strings.TrimPrefix(path, "/api/volumes/")
	if name == "" || name == path {
		http.Error(w, "存储卷名称不能为空", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := dockerClient.VolumeRemove(ctx, name, false); err != nil {
		http.Error(w, "删除存储卷失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleVolumesPrune 清理未使用的存储卷
func handleVolumesPrune(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	log.Printf("[Volumes] 开始清理未使用的存储卷...")
	
	// 使用 all=true 过滤器，清理所有未使用的卷（包括命名卷）
	pruneFilters := filters.NewArgs()
	pruneFilters.Add("all", "true")
	
	report, err := dockerClient.VolumesPrune(ctx, pruneFilters)
	if err != nil {
		log.Printf("[Volumes] 清理存储卷失败: %v", err)
		http.Error(w, "清理存储卷失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[Volumes] 清理完成，删除了 %d 个卷，释放 %d 字节", len(report.VolumesDeleted), report.SpaceReclaimed)

	result := VolumePruneResult{
		VolumesDeleted: report.VolumesDeleted,
		SpaceReclaimed: report.SpaceReclaimed,
	}
	
	// 确保返回非 nil 数组
	if result.VolumesDeleted == nil {
		result.VolumesDeleted = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
