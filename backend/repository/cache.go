package repository

import (
	"sync"
	"time"

	"rabbit-panel/model"
)

// ICacheRepository 内存缓存接口
type ICacheRepository interface {
	// Container cache
	GetContainers() ([]model.ContainerInfo, time.Time, bool)
	SetContainers([]model.ContainerInfo)
	InvalidateContainers()

	// Image cache
	GetImages() ([]model.ImageInfo, time.Time, bool)
	SetImages([]model.ImageInfo)
	InvalidateImages()

	// CPU stats cache
	GetCPUUsage() (float64, time.Time, bool)
	SetCPUUsage(float64, []uint64)
}

// CacheRepository 内存缓存实现
type CacheRepository struct {
	containersCache struct {
		sync.RWMutex
		data      []model.ContainerInfo
		lastFetch time.Time
	}
	imagesCache struct {
		sync.RWMutex
		data      []model.ImageInfo
		lastFetch time.Time
	}
	cpuStatsCache struct {
		sync.RWMutex
		lastCPU  []uint64
		lastTime time.Time
		cpuUsage float64
	}
	cacheTTL time.Duration
}

// NewCacheRepository 创建缓存仓库
func NewCacheRepository() *CacheRepository {
	return &CacheRepository{cacheTTL: 2 * time.Second}
}

// === Container Cache ===

// GetContainers 获取容器缓存
func (r *CacheRepository) GetContainers() ([]model.ContainerInfo, time.Time, bool) {
	r.containersCache.RLock()
	defer r.containersCache.RUnlock()
	if time.Since(r.containersCache.lastFetch) < r.cacheTTL && len(r.containersCache.data) > 0 {
		return r.containersCache.data, r.containersCache.lastFetch, true
	}
	return nil, time.Time{}, false
}

// SetContainers 设置容器缓存
func (r *CacheRepository) SetContainers(containers []model.ContainerInfo) {
	r.containersCache.Lock()
	r.containersCache.data = containers
	r.containersCache.lastFetch = time.Now()
	r.containersCache.Unlock()
}

// InvalidateContainers 使容器缓存失效
func (r *CacheRepository) InvalidateContainers() {
	r.containersCache.Lock()
	r.containersCache.lastFetch = time.Time{}
	r.containersCache.Unlock()
}

// === Image Cache ===

// GetImages 获取镜像缓存
func (r *CacheRepository) GetImages() ([]model.ImageInfo, time.Time, bool) {
	r.imagesCache.RLock()
	defer r.imagesCache.RUnlock()
	// Images cache TTL is longer (4s)
	if time.Since(r.imagesCache.lastFetch) < 4*time.Second && len(r.imagesCache.data) > 0 {
		return r.imagesCache.data, r.imagesCache.lastFetch, true
	}
	return nil, time.Time{}, false
}

// SetImages 设置镜像缓存
func (r *CacheRepository) SetImages(images []model.ImageInfo) {
	r.imagesCache.Lock()
	r.imagesCache.data = images
	r.imagesCache.lastFetch = time.Now()
	r.imagesCache.Unlock()
}

// InvalidateImages 使镜像缓存失效
func (r *CacheRepository) InvalidateImages() {
	r.imagesCache.Lock()
	r.imagesCache.lastFetch = time.Time{}
	r.imagesCache.Unlock()
}

// === CPU Stats Cache ===

// GetCPUUsage 获取 CPU 使用率缓存
func (r *CacheRepository) GetCPUUsage() (float64, time.Time, bool) {
	r.cpuStatsCache.RLock()
	defer r.cpuStatsCache.RUnlock()
	if time.Since(r.cpuStatsCache.lastTime) < 2*time.Second && len(r.cpuStatsCache.lastCPU) > 0 {
		return r.cpuStatsCache.cpuUsage, r.cpuStatsCache.lastTime, true
	}
	return 0, time.Time{}, false
}

// SetCPUUsage 设置 CPU 使用率缓存
func (r *CacheRepository) SetCPUUsage(usage float64, lastCPU []uint64) {
	r.cpuStatsCache.Lock()
	r.cpuStatsCache.cpuUsage = usage
	r.cpuStatsCache.lastCPU = lastCPU
	r.cpuStatsCache.lastTime = time.Now()
	r.cpuStatsCache.Unlock()
}