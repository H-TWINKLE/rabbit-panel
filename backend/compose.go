package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const composeBaseDir = "./compose_projects"

type ComposeProject struct {
	Name       string             `json:"name"`
	Status     string             `json:"status"` // "running", "partial", "stopped", "unknown"
	Containers []ComposeContainer `json:"containers,omitempty"`
}

type ComposeContainer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Service string `json:"service"`
	State   string `json:"state"`   // "running", "exited", "paused", etc.
	Status  string `json:"status"`  // 详细状态如 "Up 2 hours"
	Ports   string `json:"ports"`
}

type ComposeFileRequest struct {
	Project string `json:"project"`
	Content string `json:"content"`
}

type ComposeActionRequest struct {
	Project string `json:"project"`
	Action  string `json:"action"` // "up", "down", "restart", "pull", "logs"
}

func initCompose() {
	if err := os.MkdirAll(composeBaseDir, 0755); err != nil {
		log.Printf("无法创建 Compose 目录: %v", err)
	}
}

// 获取 Compose 项目列表
func handleComposeList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 禁止缓存
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	entries, err := os.ReadDir(composeBaseDir)
	if err != nil {
		log.Printf("读取 Compose 目录失败: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projects := make([]ComposeProject, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			projectName := entry.Name()
			projectDir := filepath.Join(composeBaseDir, projectName)
			
			// 获取项目状态
			status := getComposeProjectStatus(projectDir)
			
			projects = append(projects, ComposeProject{
				Name:   projectName,
				Status: status,
			})
		}
	}

	log.Printf("获取到 %d 个 Compose 项目", len(projects))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// 创建新项目
func handleComposeCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "项目名称不能为空", http.StatusBadRequest)
		return
	}

	projectDir := filepath.Join(composeBaseDir, req.Name)
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		http.Error(w, "项目已存在", http.StatusConflict)
		return
	}

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 创建默认 docker-compose.yml
	defaultContent := "version: '3'\nservices:\n  web:\n    image: nginx:alpine\n    ports:\n      - \"8080:80\"\n"
	if err := ioutil.WriteFile(filepath.Join(projectDir, "docker-compose.yml"), []byte(defaultContent), 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// 获取 Compose 文件内容
func handleComposeGetFile(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	if project == "" {
		http.Error(w, "Missing project parameter", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(composeBaseDir, project, "docker-compose.yml")
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 尝试 .yaml
			filePath = filepath.Join(composeBaseDir, project, "docker-compose.yaml")
			content, err = ioutil.ReadFile(filePath)
			if err != nil {
				http.Error(w, "File not found", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

// 保存 Compose 文件内容
func handleComposeSaveFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ComposeFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(composeBaseDir, req.Project, "docker-compose.yml")
	if err := ioutil.WriteFile(filePath, []byte(req.Content), 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// 获取 Compose 项目状态（包含容器详情）
func handleComposeStatus(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	if project == "" {
		http.Error(w, "Missing project parameter", http.StatusBadRequest)
		return
	}

	projectDir := filepath.Join(composeBaseDir, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// 使用 docker compose ps --format json 获取容器状态
	cmd := exec.Command("docker", "compose", "ps", "--format", "json", "-a")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		// 可能是没有运行的容器，返回空列表
		result := ComposeProject{
			Name:       project,
			Status:     "stopped",
			Containers: []ComposeContainer{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// 解析 JSON 输出（每行一个 JSON 对象）
	containers := []ComposeContainer{}
	runningCount := 0
	totalCount := 0

	// docker compose ps --format json 输出每行一个 JSON
	lines := splitLines(string(output))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var containerInfo struct {
			ID      string `json:"ID"`
			Name    string `json:"Name"`
			Service string `json:"Service"`
			State   string `json:"State"`
			Status  string `json:"Status"`
			Ports   string `json:"Ports"`
		}
		if err := json.Unmarshal([]byte(line), &containerInfo); err != nil {
			continue
		}
		totalCount++
		if containerInfo.State == "running" {
			runningCount++
		}
		containers = append(containers, ComposeContainer{
			ID:      containerInfo.ID,
			Name:    containerInfo.Name,
			Service: containerInfo.Service,
			State:   containerInfo.State,
			Status:  containerInfo.Status,
			Ports:   containerInfo.Ports,
		})
	}

	// 计算整体状态
	status := "stopped"
	if totalCount > 0 {
		if runningCount == totalCount {
			status = "running"
		} else if runningCount > 0 {
			status = "partial"
		}
	}

	result := ComposeProject{
		Name:       project,
		Status:     status,
		Containers: containers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// 辅助函数：分割行
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// 获取 Compose 项目状态（快速版本，只返回状态不返回容器详情）
func getComposeProjectStatus(projectDir string) string {
	// 检查目录是否存在
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return "unknown"
	}
	
	// 使用 docker compose ps --format json 获取容器状态
	cmd := exec.Command("docker", "compose", "ps", "--format", "json", "-a")
	cmd.Dir = projectDir
	output, err := cmd.Output()
	if err != nil {
		// 可能是没有运行的容器或者没有 compose 文件
		return "stopped"
	}
	
	// 解析输出
	lines := splitLines(string(output))
	runningCount := 0
	totalCount := 0
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		var containerInfo struct {
			State string `json:"State"`
		}
		if err := json.Unmarshal([]byte(line), &containerInfo); err != nil {
			continue
		}
		totalCount++
		if containerInfo.State == "running" {
			runningCount++
		}
	}
	
	if totalCount == 0 {
		return "stopped"
	}
	if runningCount == totalCount {
		return "running"
	}
	if runningCount > 0 {
		return "partial"
	}
	return "stopped"
}

// ComposeActionResponse 用于返回 Compose 操作结果
type ComposeActionResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// 执行 Compose 操作 (SSE 流式输出)
func handleComposeAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ComposeActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[Compose] Action: %s, project: %s", req.Action, req.Project)

	projectDir := filepath.Join(composeBaseDir, req.Project)
	
	// 检查项目目录是否存在
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		http.Error(w, "项目不存在", http.StatusNotFound)
		return
	}

	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// 发送 SSE 消息的辅助函数
	sendEvent := func(eventType, data string) {
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
		flusher.Flush()
	}

	var cmd *exec.Cmd

	switch req.Action {
	case "up":
		cmd = exec.Command("docker", "compose", "up", "-d")
	case "down":
		cmd = exec.Command("docker", "compose", "down")
	case "restart":
		cmd = exec.Command("docker", "compose", "restart")
	case "pull":
		cmd = exec.Command("docker", "compose", "pull")
	case "logs":
		cmd = exec.Command("docker", "compose", "logs", "--tail=100", "--no-color")
	default:
		sendEvent("error", "未知操作: "+req.Action)
		sendEvent("done", "failed")
		return
	}

	cmd.Dir = projectDir

	// 获取命令的 stdout 和 stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		sendEvent("error", "创建输出管道失败: "+err.Error())
		sendEvent("done", "failed")
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		sendEvent("error", "创建错误管道失败: "+err.Error())
		sendEvent("done", "failed")
		return
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		sendEvent("error", "启动命令失败: "+err.Error())
		sendEvent("done", "failed")
		return
	}

	sendEvent("log", fmt.Sprintf("执行: docker compose %s", req.Action))

	// 合并 stdout 和 stderr 读取
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				sendEvent("log", string(buf[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	// 读取 stderr
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				sendEvent("log", string(buf[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	// 等待命令完成
	err = cmd.Wait()
	
	// 稍等一下确保所有输出都发送了
	time.Sleep(100 * time.Millisecond)

	if err != nil {
		log.Printf("[Compose] Action failed, project: %s, action: %s, error: %v", req.Project, req.Action, err)
		sendEvent("error", "命令执行失败: "+err.Error())
		sendEvent("done", "failed")
	} else {
		log.Printf("[Compose] Action success, project: %s, action: %s", req.Project, req.Action)
		sendEvent("done", "success")
	}
}

// 删除 Compose 项目
func handleComposeDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Project string `json:"project"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Project == "" {
		http.Error(w, "项目名称不能为空", http.StatusBadRequest)
		return
	}

	projectDir := filepath.Join(composeBaseDir, req.Project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		http.Error(w, "项目不存在", http.StatusNotFound)
		return
	}

	// 先尝试停止容器
	cmd := exec.Command("docker", "compose", "down")
	cmd.Dir = projectDir
	cmd.Run() // 忽略错误，可能本来就没有运行

	// 删除项目目录
	if err := os.RemoveAll(projectDir); err != nil {
		http.Error(w, fmt.Sprintf("删除失败: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
