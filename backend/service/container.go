package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	"rabbit-panel/model"
	"rabbit-panel/repository"
)

// ContainerService 容器服务
type ContainerService struct {
	dockerRepo repository.IDockerRepository
	cacheRepo repository.ICacheRepository
}

// NewContainerService 创建容器服务
func NewContainerService(dr repository.IDockerRepository, cr repository.ICacheRepository) *ContainerService {
	return &ContainerService{
		dockerRepo: dr,
		cacheRepo: cr,
	}
}

// ListContainers 列出容器
func (s *ContainerService) ListContainers(ctx context.Context) ([]model.ContainerInfo, error) {
	// Check cache
	if cached, _, ok := s.cacheRepo.GetContainers(); ok {
		return cached, nil
	}

	containers, err := s.dockerRepo.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	result := s.convertContainers(containers)
	s.cacheRepo.SetContainers(result)
	return result, nil
}

// ListContainersForce 强制刷新容器列表（不使用缓存）
func (s *ContainerService) ListContainersForce(ctx context.Context) ([]model.ContainerInfo, error) {
	containers, err := s.dockerRepo.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	result := s.convertContainers(containers)
	s.cacheRepo.SetContainers(result)
	return result, nil
}

// GetContainer 获取单个容器详情
func (s *ContainerService) GetContainer(ctx context.Context, id string) (*model.ContainerInfo, error) {
	containers, err := s.dockerRepo.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		if strings.HasPrefix(c.ID, id) || c.ID == id || strings.Contains(c.ID, id) {
			result := s.convertContainers([]types.Container{c})
			return &result[0], nil
		}
	}
	return nil, fmt.Errorf("容器不存在")
}

// InspectContainer 获取容器完整配置
func (s *ContainerService) InspectContainer(ctx context.Context, id string) (types.ContainerJSON, error) {
	return s.dockerRepo.ContainerInspect(ctx, id)
}

// StartContainer 启动容器
func (s *ContainerService) StartContainer(ctx context.Context, id string) error {
	err := s.dockerRepo.ContainerStart(ctx, id, container.StartOptions{})
	s.cacheRepo.InvalidateContainers()
	return err
}

// StopContainer 停止容器
func (s *ContainerService) StopContainer(ctx context.Context, id string) error {
	err := s.dockerRepo.ContainerStop(ctx, id, nil)
	s.cacheRepo.InvalidateContainers()
	return err
}

// RestartContainer 重启容器
func (s *ContainerService) RestartContainer(ctx context.Context, id string) error {
	err := s.dockerRepo.ContainerRestart(ctx, id, nil)
	s.cacheRepo.InvalidateContainers()
	return err
}

// RemoveContainer 删除容器
func (s *ContainerService) RemoveContainer(ctx context.Context, id string, force bool) error {
	err := s.dockerRepo.ContainerRemove(ctx, id, container.RemoveOptions{Force: force})
	s.cacheRepo.InvalidateContainers()
	return err
}

// RenameContainer 重命名容器
func (s *ContainerService) RenameContainer(ctx context.Context, id, newName string) error {
	return s.dockerRepo.ContainerRename(ctx, id, newName)
}

// GetContainerLogs 获取容器日志
func (s *ContainerService) GetContainerLogs(ctx context.Context, id, tail string, follow bool) (io.ReadCloser, error) {
	tailValue := "100"
	if strings.EqualFold(strings.TrimSpace(tail), "all") {
		tailValue = "all"
	} else if strings.TrimSpace(tail) != "" {
		var tailOpt int64 = 100
		fmt.Sscanf(tail, "%d", &tailOpt)
		tailValue = fmt.Sprintf("%d", tailOpt)
	}
	return s.dockerRepo.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tailValue,
		Timestamps: true,
		Follow:     follow,
	})
}

// PullImage 拉取镜像
func (s *ContainerService) PullImage(ctx context.Context, image string) error {
	reader, err := s.dockerRepo.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	io.Copy(io.Discard, reader)
	return nil
}

// convertContainers 转换 Docker 容器类型到模型
func (s *ContainerService) convertContainers(containers []types.Container) []model.ContainerInfo {
	result := make([]model.ContainerInfo, 0, len(containers))
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		ports := ""
		var portMappings []string
		for _, p := range c.Ports {
			if p.PublicPort != 0 {
				portMappings = append(portMappings, fmt.Sprintf("%d:%d", p.PublicPort, p.PrivatePort))
			}
		}
		if len(portMappings) > 0 {
			ports = strings.Join(portMappings, ", ")
		}

		result = append(result, model.ContainerInfo{
			ID:      c.ID[:12],
			Name:    name,
			Image:   c.Image,
			Status:  c.Status,
			Ports:   ports,
			Created: time.Unix(c.Created, 0).Format("2006-01-02 15:04:05"),
			State:   c.State,
		})
	}
	return result
}
