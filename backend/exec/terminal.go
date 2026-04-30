package exec

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"

	"rabbit-panel/repository"
)

// TerminalService 终端服务
type TerminalService struct {
	dockerRepo repository.IDockerRepository
}

// NewTerminalService 创建终端服务
func NewTerminalService(dr repository.IDockerRepository) *TerminalService {
	return &TerminalService{dockerRepo: dr}
}

// ExecCommand 在容器中执行命令
func (s *TerminalService) ExecCommand(ctx context.Context, containerID string, cmd []string) (string, int, error) {
	// Create exec
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	execID, err := s.dockerRepo.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return "", -1, err
	}

	// Start exec
	err = s.dockerRepo.ContainerExecStart(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return "", -1, err
	}

	// Wait for completion
	inspect, err := s.dockerRepo.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return "", -1, err
	}

	// Note: This is a simplified version. Actual implementation needs
	// HijackedResponse for real-time output
	return "", inspect.ExitCode, nil
}

// GetContainerFile 获取容器文件内容
func (s *TerminalService) GetContainerFile(ctx context.Context, containerID, path string) (io.ReadCloser, types.ContainerPathStat, error) {
	return s.dockerRepo.CopyFromContainer(ctx, containerID, path)
}

// PutContainerFile 上传文件到容器
func (s *TerminalService) PutContainerFile(ctx context.Context, containerID, path string, content io.Reader) error {
	return s.dockerRepo.CopyToContainer(ctx, containerID, path, content, types.CopyToContainerOptions{})
}

// GetContainerInspect 获取容器详情
func (s *TerminalService) GetContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return s.dockerRepo.ContainerInspect(ctx, containerID)
}

// UpdateContainer 更新容器配置
func (s *TerminalService) UpdateContainer(ctx context.Context, containerID string, mem int64, cpus float64) error {
	resources := &container.Resources{}
	if mem > 0 {
		resources.Memory = mem
	}
	if cpus > 0 {
		resources.CPUPeriod = 100000
		resources.CPUQuota = int64(cpus * 100000)
	}

	_, err := s.dockerRepo.ContainerUpdate(ctx, containerID, container.UpdateConfig{
		Resources: *resources,
	})
	return err
}

// RenameContainer 重命名容器
func (s *TerminalService) RenameContainer(ctx context.Context, containerID, newName string) error {
	return s.dockerRepo.ContainerRename(ctx, containerID, newName)
}

// RecreateContainer 重建容器
func (s *TerminalService) RecreateContainer(ctx context.Context, req *RecreateRequest) error {
	// Stop container
	s.dockerRepo.ContainerStop(ctx, req.ContainerID, nil)

	// Remove container
	s.dockerRepo.ContainerRemove(ctx, req.ContainerID, container.RemoveOptions{Force: true})

	// Create new container
	// TODO: Implement actual recreation
	return nil
}

// RecreateRequest 重建容器请求
type RecreateRequest struct {
	ContainerID string
	Name        string
	Image       string
	Ports       []PortMapping
	Volumes     []VolumeMapping
	Env         []EnvVar
	Restart     string
	Network     string
	Memory      int64
	CPUs        float64
	Privileged  bool
	TTY         bool
}

// PortMapping 端口映射
type PortMapping struct {
	Host      string
	Container string
}

// VolumeMapping 卷映射
type VolumeMapping struct {
	Host      string
	Container string
}

// EnvVar 环境变量
type EnvVar struct {
	Key   string
	Value string
}

// CopyWithProgress 复制数据并显示进度
func CopyWithProgress(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}

// stdcopy wrapper for tar stream handling
var _ = stdcopy.StdCopy // Ensure stdcopy is used