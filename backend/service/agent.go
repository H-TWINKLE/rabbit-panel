package service

import (
	"sync"
	"time"

	"rabbit-panel/repository"
)

// AgentService AI 智能体服务
type AgentService struct {
	sqliteRepo  *repository.SQLiteRepository
	dockerRepo  repository.IDockerRepository

	config      AgentConfig
	configMutex sync.RWMutex
	configPath  string
}

// AgentConfig AI 配置
type AgentConfig struct {
	APIURL  string
	APIKey  string
	Model   string
	Enabled bool
}

// NewAgentService 创建 AI 服务
func NewAgentService(sr *repository.SQLiteRepository, dr repository.IDockerRepository) *AgentService {
	return &AgentService{
		sqliteRepo: sr,
		dockerRepo:  dr,
		configPath: "./data/agent.json",
	}
}

// GetConfig 获取配置
func (s *AgentService) GetConfig() AgentConfig {
	s.configMutex.RLock()
	defer s.configMutex.RUnlock()
	return s.config
}

// SaveConfig 保存配置
func (s *AgentService) SaveConfig(cfg AgentConfig) error {
	s.configMutex.Lock()
	s.config = cfg
	s.configMutex.Unlock()
	// TODO: Save to file
	return nil
}

// GetChatHistory 获取聊天历史
func (s *AgentService) GetChatHistory(limit int) ([]repository.ChatHistoryRecord, error) {
	return s.sqliteRepo.GetChatHistory(limit)
}

// SaveChatMessage 保存聊天消息
func (s *AgentService) SaveChatMessage(role, content string) error {
	return s.sqliteRepo.SaveChatMessage(role, content)
}

// CleanupOldMessages 清理旧消息
func (s *AgentService) CleanupOldMessages(olderThan time.Duration) (int64, error) {
	return s.sqliteRepo.CleanupOldMessages(olderThan)
}