package service

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"

	"rabbit-panel/model"
	"rabbit-panel/repository"
)

// NetworkService 网络服务
type NetworkService struct {
	dockerRepo repository.IDockerRepository
}

// NewNetworkService 创建网络服务
func NewNetworkService(dr repository.IDockerRepository) *NetworkService {
	return &NetworkService{dockerRepo: dr}
}

// ListNetworks 列出网络
func (s *NetworkService) ListNetworks(ctx context.Context) ([]model.NetworkInfo, error) {
	networks, err := s.dockerRepo.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]model.NetworkInfo, 0, len(networks))
	for _, n := range networks {
		info := model.NetworkInfo{
			ID:         n.ID[:12],
			Name:       n.Name,
			Driver:     n.Driver,
			Scope:      n.Scope,
			Internal:   n.Internal,
			Attachable: n.Attachable,
			Created:    n.Created.Format("2006-01-02 15:04:05"),
		}

		// Get subnet and gateway
		if n.IPAM.Config != nil && len(n.IPAM.Config) > 0 {
			info.Subnet = n.IPAM.Config[0].Subnet
			info.Gateway = n.IPAM.Config[0].Gateway
		}

		// Get connected containers
		if n.Containers != nil {
			for id := range n.Containers {
				info.Containers = append(info.Containers, id[:12])
			}
		}
		info.ContainerCount = len(info.Containers)
		info.InUse = info.ContainerCount > 0

		result = append(result, info)
	}
	return result, nil
}

// CreateNetwork 创建网络
func (s *NetworkService) CreateNetwork(ctx context.Context, req *model.CreateNetworkRequest) error {
	opts := types.NetworkCreate{
		Driver: req.Driver,
	}
	if req.Subnet != "" {
		ipamConfig := network.IPAMConfig{Subnet: req.Subnet, Gateway: req.Gateway}
		opts.IPAM = &network.IPAM{
			Config: []network.IPAMConfig{ipamConfig},
		}
	}
	opts.Internal = req.Internal
	opts.Attachable = req.Attachable
	opts.Labels = req.Labels

	_, err := s.dockerRepo.NetworkCreate(ctx, req.Name, opts)
	return err
}

// RemoveNetwork 删除网络
func (s *NetworkService) RemoveNetwork(ctx context.Context, id string) error {
	return s.dockerRepo.NetworkRemove(ctx, id)
}

// InspectNetwork 获取网络详情
func (s *NetworkService) InspectNetwork(ctx context.Context, id string) (map[string]interface{}, error) {
	n, err := s.dockerRepo.NetworkInspect(ctx, id, types.NetworkInspectOptions{Verbose: true})
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":            n.ID,
		"name":          n.Name,
		"driver":        n.Driver,
		"scope":         n.Scope,
		"internal":      n.Internal,
		"attachable":    n.Attachable,
		"ingress":       n.Ingress,
		"ipam":          n.IPAM,
		"options":       n.Options,
		"labels":        n.Labels,
		"created":       n.Created.Format("2006-01-02 15:04:05"),
		"containers":    n.Containers,
	}

	if n.IPAM.Config != nil && len(n.IPAM.Config) > 0 {
		result["subnet"] = n.IPAM.Config[0].Subnet
		result["gateway"] = n.IPAM.Config[0].Gateway
	}

	return result, nil
}

// ConnectNetwork 连接容器到网络
func (s *NetworkService) ConnectNetwork(ctx context.Context, networkID, containerID string) error {
	return s.dockerRepo.NetworkConnect(ctx, networkID, containerID, nil)
}

// DisconnectNetwork 断开容器与网络的连接
func (s *NetworkService) DisconnectNetwork(ctx context.Context, networkID, containerID string, force bool) error {
	return s.dockerRepo.NetworkDisconnect(ctx, networkID, containerID, force)
}
