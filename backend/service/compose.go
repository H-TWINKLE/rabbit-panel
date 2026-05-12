package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"rabbit-panel/model"
	"rabbit-panel/repository"
)

// ComposeService Docker Compose 服务
type ComposeService struct {
	dockerRepo repository.IDockerRepository
	fileRepo   repository.IFileRepository
}

// NewComposeService 创建 Compose 服务
func NewComposeService(dr repository.IDockerRepository, fr repository.IFileRepository) *ComposeService {
	return &ComposeService{
		dockerRepo: dr,
		fileRepo:   fr,
	}
}

// ListProjects 列出 Compose 项目
func (s *ComposeService) ListProjects() ([]model.ComposeProject, error) {
	projects, err := s.fileRepo.ListComposeProjects()
	if err != nil {
		return nil, err
	}

	result := make([]model.ComposeProject, 0, len(projects))
	for _, name := range projects {
		status := s.getProjectStatus(name)
		result = append(result, model.ComposeProject{
			Name:   name,
			Status: status,
		})
	}
	return result, nil
}

// FetchProjectStatus 获取单个 Compose 项目状态
func (s *ComposeService) FetchProjectStatus(name string) (*model.ComposeProject, error) {
	dir := s.fileRepo.GetComposeProjectDir(name)
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "ps", "--format", "json")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return &model.ComposeProject{
			Name:   name,
			Status: "unknown",
		}, nil
	}

	project := &model.ComposeProject{
		Name:   name,
		Status: "stopped",
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	running := 0
	total := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		total++
		var item struct {
			ID      string `json:"ID"`
			Name    string `json:"Name"`
			Service string `json:"Service"`
			State   string `json:"State"`
			Status  string `json:"Status"`
			Publishers []struct {
				URL           string `json:"URL"`
				TargetPort    int    `json:"TargetPort"`
				PublishedPort int    `json:"PublishedPort"`
				Protocol      string `json:"Protocol"`
			} `json:"Publishers"`
		}
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}
		if item.State == "running" {
			running++
		}
		ports := make([]string, 0, len(item.Publishers))
		for _, publisher := range item.Publishers {
			if publisher.PublishedPort > 0 && publisher.TargetPort > 0 {
				ports = append(ports, fmt.Sprintf("%d:%d", publisher.PublishedPort, publisher.TargetPort))
			}
		}
		project.Containers = append(project.Containers, model.ComposeContainer{
			ID:      item.ID,
			Name:    item.Name,
			Service: item.Service,
			State:   item.State,
			Status:  item.Status,
			Ports:   strings.Join(ports, ", "),
		})
	}

	switch {
	case total == 0:
		project.Status = "stopped"
	case running == total:
		project.Status = "running"
	case running > 0:
		project.Status = "partial"
	default:
		project.Status = "stopped"
	}

	return project, nil
}

// CreateProject 创建 Compose 项目
func (s *ComposeService) CreateProject(name, content string) error {
	return s.fileRepo.CreateComposeProject(name, content)
}

// GetProjectFile 获取项目 Compose 文件
func (s *ComposeService) GetProjectFile(name string) (string, error) {
	return s.fileRepo.GetComposeFile(name)
}

// SaveProjectFile 保存项目 Compose 文件
func (s *ComposeService) SaveProjectFile(name, content string) error {
	return s.fileRepo.SaveComposeFile(name, content)
}

// DeleteProject 删除项目
func (s *ComposeService) DeleteProject(name string) error {
	return s.fileRepo.DeleteComposeProject(name)
}

// ExecuteAction 执行 Compose 操作（up, down, restart, pull, logs）
func (s *ComposeService) ExecuteAction(name, action string, writer io.Writer) error {
	dir := s.fileRepo.GetComposeProjectDir(name)
	var cmd *exec.Cmd

	switch action {
	case "up":
		cmd = exec.Command("docker", "compose", "-f", "docker-compose.yml", "up", "-d")
	case "down":
		cmd = exec.Command("docker", "compose", "-f", "docker-compose.yml", "down")
	case "restart":
		cmd = exec.Command("docker", "compose", "-f", "docker-compose.yml", "restart")
	case "pull":
		cmd = exec.Command("docker", "compose", "-f", "docker-compose.yml", "pull")
	case "logs":
		cmd = exec.Command("docker", "compose", "-f", "docker-compose.yml", "logs", "--tail=50")
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	cmd.Dir = dir
	cmd.Stdout = writer
	cmd.Stderr = writer
	return cmd.Run()
}

// getProjectStatus 获取项目状态
func (s *ComposeService) getProjectStatus(name string) string {
	dir := s.fileRepo.GetComposeProjectDir(name)
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "ps", "--format", "json")
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Parse output to determine overall status
	scanner := bufio.NewScanner(bytes.NewReader(output))
	var running, total int
	for scanner.Scan() {
		total++
		var container struct {
			State string `json:"State"`
		}
		if json.Unmarshal(scanner.Bytes(), &container) == nil {
			if container.State == "running" {
				running++
			}
		}
	}

	if total == 0 {
		return "stopped"
	}
	if running == total {
		return "running"
	}
	if running > 0 {
		return "partial"
	}
	return "stopped"
}
