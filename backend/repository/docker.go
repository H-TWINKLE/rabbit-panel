package repository

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// IDockerRepository Docker API 操作接口
type IDockerRepository interface {
	// === Container Operations ===
	ContainerList(ctx context.Context, opts types.ContainerListOptions) ([]types.Container, error)
	ContainerCreate(ctx context.Context, config *container.Config, host *container.HostConfig, networkingConfig *network.NetworkingConfig, platform interface{}, name string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, opts container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, timeout *int) error
	ContainerRestart(ctx context.Context, containerID string, timeout *int) error
	ContainerRemove(ctx context.Context, containerID string, opts container.RemoveOptions) error
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerLogs(ctx context.Context, containerID string, opts types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerExecCreate(ctx context.Context, containerID string, config types.ExecConfig) (types.IDResponse, error)
	ContainerExecStart(ctx context.Context, execID string, config types.ExecStartCheck) error
	ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error)
	ContainerExecResize(ctx context.Context, execID string, options container.ResizeOptions) error
	ContainerRename(ctx context.Context, containerID string, name string) error
	ContainerUpdate(ctx context.Context, containerID string, update container.UpdateConfig) (container.ContainerUpdateOKBody, error)

	// === Image Operations ===
	ImageList(ctx context.Context, opts types.ImageListOptions) ([]image.Summary, error)
	ImagePull(ctx context.Context, refStr string, opts types.ImagePullOptions) (io.ReadCloser, error)
	ImageRemove(ctx context.Context, imageID string, opts types.ImageRemoveOptions) ([]types.ImageDeleteResponseItem, error)
	ImageInspectWithRaw(ctx context.Context, imageID string) (types.ImageInspect, []byte, error)
	ImageBuild(ctx context.Context, context_ io.Reader, opts types.ImageBuildOptions) (types.ImageBuildResponse, error)

	// === Network Operations ===
	NetworkList(ctx context.Context, opts types.NetworkListOptions) ([]types.NetworkResource, error)
	NetworkCreate(ctx context.Context, name string, opts types.NetworkCreate) (types.NetworkCreateResponse, error)
	NetworkRemove(ctx context.Context, networkID string) error
	NetworkInspect(ctx context.Context, networkID string, opts types.NetworkInspectOptions) (types.NetworkResource, error)
	NetworkConnect(ctx context.Context, networkID, containerID string, config *network.EndpointSettings) error
	NetworkDisconnect(ctx context.Context, networkID, containerID string, force bool) error

	// === Volume Operations ===
	VolumeList(ctx context.Context, opts volume.ListOptions) (volume.ListResponse, error)
	VolumeCreate(ctx context.Context, opts volume.CreateOptions) (volume.Volume, error)
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
	VolumeInspect(ctx context.Context, volumeID string) (volume.Volume, error)

	// === System Operations ===
	Ping(ctx context.Context) (types.Ping, error)
	Info(ctx context.Context) (types.Info, error)
	ServerVersion(ctx context.Context) (types.Version, error)
	DiskUsage(ctx context.Context, opts types.DiskUsageOptions) (types.DiskUsage, error)

	// === Container File Operations ===
	ContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error)
	CopyToContainer(ctx context.Context, containerID, dstPath string, content io.Reader, opts types.CopyToContainerOptions) error
	CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, types.ContainerPathStat, error)
}

// DockerRepository 实现 IDockerRepository
type DockerRepository struct {
	client *client.Client
}

// NewDockerRepository 创建 Docker 仓库实例
func NewDockerRepository(cli *client.Client) *DockerRepository {
	return &DockerRepository{client: cli}
}

// === Container Operations ===

func (r *DockerRepository) ContainerList(ctx context.Context, opts types.ContainerListOptions) ([]types.Container, error) {
	return r.client.ContainerList(ctx, opts)
}

func (r *DockerRepository) ContainerCreate(ctx context.Context, config *container.Config, host *container.HostConfig, networkingConfig *network.NetworkingConfig, platform interface{}, name string) (container.CreateResponse, error) {
	return r.client.ContainerCreate(ctx, config, host, networkingConfig, nil, name)
}

func (r *DockerRepository) ContainerStart(ctx context.Context, containerID string, opts container.StartOptions) error {
	return r.client.ContainerStart(ctx, containerID, opts)
}

func (r *DockerRepository) ContainerStop(ctx context.Context, containerID string, timeout *int) error {
	return r.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: timeout})
}

func (r *DockerRepository) ContainerRestart(ctx context.Context, containerID string, timeout *int) error {
	return r.client.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: timeout})
}

func (r *DockerRepository) ContainerRemove(ctx context.Context, containerID string, opts container.RemoveOptions) error {
	return r.client.ContainerRemove(ctx, containerID, opts)
}

func (r *DockerRepository) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return r.client.ContainerInspect(ctx, containerID)
}

func (r *DockerRepository) ContainerLogs(ctx context.Context, containerID string, opts types.ContainerLogsOptions) (io.ReadCloser, error) {
	return r.client.ContainerLogs(ctx, containerID, opts)
}

func (r *DockerRepository) ContainerExecCreate(ctx context.Context, containerID string, config types.ExecConfig) (types.IDResponse, error) {
	return r.client.ContainerExecCreate(ctx, containerID, config)
}

func (r *DockerRepository) ContainerExecStart(ctx context.Context, execID string, config types.ExecStartCheck) error {
	return r.client.ContainerExecStart(ctx, execID, config)
}

func (r *DockerRepository) ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error) {
	return r.client.ContainerExecAttach(ctx, execID, config)
}

func (r *DockerRepository) ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
	return r.client.ContainerExecInspect(ctx, execID)
}

func (r *DockerRepository) ContainerExecResize(ctx context.Context, execID string, options container.ResizeOptions) error {
	return r.client.ContainerExecResize(ctx, execID, options)
}

func (r *DockerRepository) ContainerRename(ctx context.Context, containerID string, name string) error {
	return r.client.ContainerRename(ctx, containerID, name)
}

func (r *DockerRepository) ContainerUpdate(ctx context.Context, containerID string, update container.UpdateConfig) (container.ContainerUpdateOKBody, error) {
	return r.client.ContainerUpdate(ctx, containerID, update)
}

// === Image Operations ===

func (r *DockerRepository) ImageList(ctx context.Context, opts types.ImageListOptions) ([]image.Summary, error) {
	return r.client.ImageList(ctx, opts)
}

func (r *DockerRepository) ImagePull(ctx context.Context, refStr string, opts types.ImagePullOptions) (io.ReadCloser, error) {
	return r.client.ImagePull(ctx, refStr, opts)
}

func (r *DockerRepository) ImageRemove(ctx context.Context, imageID string, opts types.ImageRemoveOptions) ([]types.ImageDeleteResponseItem, error) {
	return r.client.ImageRemove(ctx, imageID, opts)
}

func (r *DockerRepository) ImageInspectWithRaw(ctx context.Context, imageID string) (types.ImageInspect, []byte, error) {
	return r.client.ImageInspectWithRaw(ctx, imageID)
}

func (r *DockerRepository) ImageBuild(ctx context.Context, context_ io.Reader, opts types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	return r.client.ImageBuild(ctx, context_, opts)
}

// === Network Operations ===

func (r *DockerRepository) NetworkList(ctx context.Context, opts types.NetworkListOptions) ([]types.NetworkResource, error) {
	return r.client.NetworkList(ctx, opts)
}

func (r *DockerRepository) NetworkCreate(ctx context.Context, name string, opts types.NetworkCreate) (types.NetworkCreateResponse, error) {
	return r.client.NetworkCreate(ctx, name, opts)
}

func (r *DockerRepository) NetworkRemove(ctx context.Context, networkID string) error {
	return r.client.NetworkRemove(ctx, networkID)
}

func (r *DockerRepository) NetworkInspect(ctx context.Context, networkID string, opts types.NetworkInspectOptions) (types.NetworkResource, error) {
	return r.client.NetworkInspect(ctx, networkID, opts)
}

func (r *DockerRepository) NetworkConnect(ctx context.Context, networkID, containerID string, config *network.EndpointSettings) error {
	return r.client.NetworkConnect(ctx, networkID, containerID, config)
}

func (r *DockerRepository) NetworkDisconnect(ctx context.Context, networkID, containerID string, force bool) error {
	return r.client.NetworkDisconnect(ctx, networkID, containerID, force)
}

// === Volume Operations ===

func (r *DockerRepository) VolumeList(ctx context.Context, opts volume.ListOptions) (volume.ListResponse, error) {
	return r.client.VolumeList(ctx, opts)
}

func (r *DockerRepository) VolumeCreate(ctx context.Context, opts volume.CreateOptions) (volume.Volume, error) {
	return r.client.VolumeCreate(ctx, opts)
}

func (r *DockerRepository) VolumeRemove(ctx context.Context, volumeID string, force bool) error {
	return r.client.VolumeRemove(ctx, volumeID, force)
}

func (r *DockerRepository) VolumeInspect(ctx context.Context, volumeID string) (volume.Volume, error) {
	return r.client.VolumeInspect(ctx, volumeID)
}

// === System Operations ===

func (r *DockerRepository) Ping(ctx context.Context) (types.Ping, error) {
	return r.client.Ping(ctx)
}

func (r *DockerRepository) Info(ctx context.Context) (types.Info, error) {
	return r.client.Info(ctx)
}

func (r *DockerRepository) ServerVersion(ctx context.Context) (types.Version, error) {
	return r.client.ServerVersion(ctx)
}

func (r *DockerRepository) DiskUsage(ctx context.Context, opts types.DiskUsageOptions) (types.DiskUsage, error) {
	return r.client.DiskUsage(ctx, opts)
}

// === Container File Operations ===

func (r *DockerRepository) ContainerStats(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error) {
	return r.client.ContainerStats(ctx, containerID, stream)
}

func (r *DockerRepository) CopyToContainer(ctx context.Context, containerID, dstPath string, content io.Reader, opts types.CopyToContainerOptions) error {
	return r.client.CopyToContainer(ctx, containerID, dstPath, content, opts)
}

func (r *DockerRepository) CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
	return r.client.CopyFromContainer(ctx, containerID, srcPath)
}
