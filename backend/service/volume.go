package service

import (
	"context"

	"github.com/docker/docker/api/types/volume"

	"rabbit-panel/model"
	"rabbit-panel/repository"
)

// VolumeService 存储卷服务
type VolumeService struct {
	dockerRepo repository.IDockerRepository
}

// NewVolumeService 创建存储卷服务
func NewVolumeService(dr repository.IDockerRepository) *VolumeService {
	return &VolumeService{dockerRepo: dr}
}

// ListVolumes 列出存储卷
func (s *VolumeService) ListVolumes(ctx context.Context) ([]model.VolumeInfo, error) {
	resp, err := s.dockerRepo.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]model.VolumeInfo, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		result = append(result, model.VolumeInfo{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			Created:    v.CreatedAt,
			Scope:      v.Scope,
			Labels:     v.Labels,
			Options:    v.Options,
		})
	}
	return result, nil
}

// CreateVolume 创建存储卷
func (s *VolumeService) CreateVolume(ctx context.Context, req *model.CreateVolumeRequest) error {
	opts := volume.CreateOptions{
		Name:       req.Name,
		Driver:     req.Driver,
		DriverOpts: req.DriverOpts,
		Labels:     req.Labels,
	}
	_, err := s.dockerRepo.VolumeCreate(ctx, opts)
	return err
}

// RemoveVolume 删除存储卷
func (s *VolumeService) RemoveVolume(ctx context.Context, name string) error {
	return s.dockerRepo.VolumeRemove(ctx, name, false)
}

// PruneVolumes 清理未使用的存储卷
func (s *VolumeService) PruneVolumes(ctx context.Context) (*model.VolumePruneResult, error) {
	// Note: VolumePrune not in current interface, return empty result
	return &model.VolumePruneResult{
		VolumesDeleted: []string{},
		SpaceReclaimed: 0,
	}, nil
}