package service

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types"

	"rabbit-panel/model"
	"rabbit-panel/repository"
	"rabbit-panel/tool"
)

// NodeService 节点服务
type NodeService struct {
	dockerRepo  repository.IDockerRepository
	cacheRepo  repository.ICacheRepository
	mode       string
	nodeSecret string
	nodeID     string
	host       string
	port       string

	mu    sync.RWMutex
	nodes map[string]*model.NodeInfo
}

// NewNodeService 创建节点服务
func NewNodeService(dr repository.IDockerRepository, cr repository.ICacheRepository, mode, nodeSecret, nodeID, host, port string) *NodeService {
	service := &NodeService{
		dockerRepo:  dr,
		cacheRepo:   cr,
		mode:        mode,
		nodeSecret:  nodeSecret,
		nodeID:      nodeID,
		host:        host,
		port:        port,
		nodes:       make(map[string]*model.NodeInfo),
	}
	service.ensureLocalNode()
	return service
}

// RegisterNode 注册节点
func (s *NodeService) RegisterNode(node *model.NodeInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	node.LastSeen = time.Now()
	node.Status = model.NodeStatusOnline
	s.nodes[node.ID] = node
}

// UpdateHeartbeat 更新节点心跳
func (s *NodeService) UpdateHeartbeat(nodeID string, resources model.SystemStats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if node, ok := s.nodes[nodeID]; ok {
		node.CPU = resources.CPU
		node.Memory = resources.Memory
		node.Disk = resources.Disk
		node.LastSeen = time.Now()
		node.Status = model.NodeStatusOnline
	}
}

// GetAllNodes 获取所有节点
func (s *NodeService) GetAllNodes() []*model.NodeInfo {
	s.ensureLocalNode()
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.NodeInfo, 0, len(s.nodes))
	for _, n := range s.nodes {
		n.LastSeenStr = n.LastSeen.Format("2006-01-02 15:04:05")
		result = append(result, n)
	}
	return result
}

// GetNode 获取单个节点
func (s *NodeService) GetNode(nodeID string) (*model.NodeInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	node, ok := s.nodes[nodeID]
	if ok {
		node.LastSeenStr = node.LastSeen.Format("2006-01-02 15:04:05")
	}
	return node, ok
}

// RemoveNode 移除节点
func (s *NodeService) RemoveNode(nodeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.nodes, nodeID)
}

// SelectBestNode 选择最佳节点（CPU 和内存负载最低）
func (s *NodeService) SelectBestNode() (*model.NodeInfo, error) {
	s.ensureLocalNode()
	s.mu.RLock()
	defer s.mu.RUnlock()
	var best *model.NodeInfo
	minLoad := 100.0
	for _, n := range s.nodes {
		if n.Status != model.NodeStatusOnline {
			continue
		}
		load := (n.CPU + n.Memory) / 2
		if load < minLoad {
			minLoad = load
			best = n
		}
	}
	if best == nil {
		return nil, ErrNoAvailableNodes
	}
	return best, nil
}

// GetMode 获取运行模式
func (s *NodeService) GetMode() string {
	return s.mode
}

func (s *NodeService) ensureLocalNode() {
	if s.mode != "master" {
		return
	}

	hostname, _ := os.Hostname()
	address := s.host
	if address == "" || address == "0.0.0.0" {
		address = "127.0.0.1"
	}
	if s.port != "" {
		address = address + ":" + s.port
	}

	cpu, _ := tool.GetCPUUsage()
	mem, _ := tool.GetMemoryUsage()
	disk, _ := tool.GetDiskUsage()
	containers, _ := s.dockerRepo.ContainerList(context.Background(), types.ContainerListOptions{All: true})

	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.nodes[s.nodeID]
	if !ok {
		node = &model.NodeInfo{
			ID:      s.nodeID,
			Name:    hostname,
			Address: address,
			Mode:    model.ModeMaster,
			Labels:  map[string]string{},
		}
		s.nodes[s.nodeID] = node
	}

	node.CPU = cpu
	node.Memory = mem.Usage
	node.Disk = disk.Usage
	node.Containers = len(containers)
	node.LastSeen = time.Now()
	node.Status = model.NodeStatusOnline
}

var ErrNoAvailableNodes = &NodeError{"no available nodes"}

type NodeError struct {
	msg string
}

func (e *NodeError) Error() string {
	return e.msg
}
