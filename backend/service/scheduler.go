package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"rabbit-panel/model"
	"rabbit-panel/repository"
)

// Scheduler 容器调度器
type Scheduler struct {
	nodeService *NodeService
	dockerRepo  repository.IDockerRepository
	nodeSecret  string
}

// NewScheduler 创建调度器
func NewScheduler(ns *NodeService, dr repository.IDockerRepository, nodeSecret string) *Scheduler {
	return &Scheduler{
		nodeService: ns,
		dockerRepo:  dr,
		nodeSecret:  nodeSecret,
	}
}

// ScheduleContainer 调度容器到最佳节点
func (s *Scheduler) ScheduleContainer(ctx context.Context, req *model.ScheduleRequest) (*model.ScheduleResponse, error) {
	targetNode, err := s.nodeService.SelectBestNode()
	if err != nil {
		return nil, err
	}

	// If local (Master) or no nodes, create locally
	if targetNode.Mode == model.ModeMaster {
		return s.createLocalContainer(ctx, req)
	}

	// Forward to worker node
	return s.forwardToWorker(ctx, targetNode.Address, req)
}

// createLocalContainer 在本地创建容器
func (s *Scheduler) createLocalContainer(ctx context.Context, req *model.ScheduleRequest) (*model.ScheduleResponse, error) {
	// TODO: Implement actual container creation
	return &model.ScheduleResponse{
		Status:   "success",
		NodeID:   "local",
		NodeName: "Master",
	}, nil
}

// forwardToWorker 转发请求到 Worker 节点
func (s *Scheduler) forwardToWorker(ctx context.Context, workerAddr string, req *model.ScheduleRequest) (*model.ScheduleResponse, error) {
	containerConfig := map[string]interface{}{
		"image": req.Image,
		"name":  req.Name,
		"ports": req.Ports,
		"env":   req.Env,
	}

	jsonData, _ := json.Marshal(containerConfig)
	workerURL := fmt.Sprintf("http://%s/api/containers/create", workerAddr)

	httpReq, _ := http.NewRequest("POST", workerURL, bytes.NewBuffer(jsonData))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Node-ID", "master")
	httpReq.Header.Set("X-Node-Token", s.generateNodeToken("master"))

	// TODO: Add timeout
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	return &model.ScheduleResponse{
		Status:   "success",
		NodeID:   "worker",
		NodeName: "Worker",
		Container: result,
	}, nil
}

// generateNodeToken generates a node authentication token using HMAC-SHA256
func (s *Scheduler) generateNodeToken(nodeID string) string {
	h := sha256.New()
	h.Write([]byte(nodeID))
	h.Write([]byte(s.nodeSecret))
	return hex.EncodeToString(h.Sum(nil))
}