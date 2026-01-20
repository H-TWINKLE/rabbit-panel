package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
)

// AgentConfig 智能体配置
type AgentConfig struct {
	APIURL  string `json:"api_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
	Enabled bool   `json:"enabled"`
}

// Global Agent Config
var (
	agentConfig      AgentConfig
	agentConfigMutex sync.RWMutex
	agentConfigPath  = "./data/agent.json"
)

// OpenAI API Structs

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatCompletionChunk struct {
	ID      string `json:"id"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// User Request
type AgentChatRequest struct {
	Message string `json:"message"`
	History []ChatMessage `json:"history"` // Optional history from frontend
}

// 加载配置
func loadAgentConfig() {
	os.MkdirAll("./data", 0755)
	file, err := os.Open(agentConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			agentConfig = AgentConfig{
				APIURL:  "https://api.openai.com/v1",
				Model:   "gpt-3.5-turbo",
				Enabled: true,
			}
			return
		}
		log.Printf("[Agent] 加载配置失败: %v", err)
		return
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&agentConfig)
}

// 保存配置
func saveAgentConfig() error {
	agentConfigMutex.Lock()
	defer agentConfigMutex.Unlock()
	file, err := os.Create(agentConfigPath)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(agentConfig)
}

// Get Config Handler
func handleGetAgentConfig(w http.ResponseWriter, r *http.Request) {
	agentConfigMutex.RLock()
	config := agentConfig
	agentConfigMutex.RUnlock()

	// Mask Key
	if len(config.APIKey) > 8 {
		config.APIKey = config.APIKey[:4] + "****" + config.APIKey[len(config.APIKey)-4:]
	} else if config.APIKey != "" {
		config.APIKey = "****"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// Save Config Handler
func handleSaveAgentConfig(w http.ResponseWriter, r *http.Request) {
	var req AgentConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	agentConfigMutex.Lock()
	if req.APIKey != "" && req.APIKey != "****" && !strings.Contains(req.APIKey, "****") {
		agentConfig.APIKey = req.APIKey
	}
	agentConfig.APIURL = req.APIURL
	agentConfig.Model = req.Model
	agentConfig.Enabled = req.Enabled
	agentConfigMutex.Unlock()

	saveAgentConfig()
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// System Prompt with Tools Description
const systemPrompt = `你是一个专业的 Rabbit Panel 运维助手。
你的职责是协助用户管理服务器、Docker 容器、镜像、网络、存储卷和 Compose 项目。

**回复规范:**
1. **结构化显示**: 尽量使用 Markdown 表格、列表和代码块来展示数据。
2. **工具调用**: 当你需要执行操作时，请使用 [[TOOL: function(args)]] 格式。**参数请直接写值**。
   - 正确示例: [[TOOL: start_container(abc12345)]]
   - 错误示例: [[TOOL: start_container(id="abc")]]
3. **多步执行**: 你可以连续执行多个工具调用，但不要超过 3 步。
4. **确认**: 在执行删除、停止等破坏性操作前，请先征得用户确认。
5. **专注目标**: 仅回答用户提出的问题，不要进行未经请求的分析。
6. **简洁专业**: 使用中文回复，保持精简。

**可用工具 (共 26 个):**

**容器管理:**
1. list_containers() - 列出所有容器
2. start_container(id) - 启动已存在的容器
3. stop_container(id) - 停止容器
4. restart_container(id) - 重启容器
5. get_container_logs(id) - 获取最近 50 行日志
6. inspect_container(id) - 获取容器详细信息
7. delete_container(id) - 删除容器 (需确认)
8. run_container(image, name, options...) - 创建并启动容器
   - 基础: run_container(redis:latest, my-redis)
   - 端口: run_container(redis:latest, my-redis, 6379:6379)
   - 多参数: run_container(nginx, web, 80:80, -e ENV=prod, -v /data:/app)

**镜像管理:**
9. list_images() - 列出所有镜像
10. pull_image(name) - 拉取镜像 (如: nginx:latest)
11. delete_image(id) - 删除镜像 (需确认)

**网络管理:**
12. list_networks() - 列出所有网络
13. create_network(name, driver) - 创建网络 (driver: bridge/overlay)
14. delete_network(id) - 删除网络 (需确认)
15. inspect_network(id) - 查看网络详情

**存储卷管理:**
16. list_volumes() - 列出所有存储卷
17. create_volume(name) - 创建存储卷
18. delete_volume(name) - 删除存储卷 (需确认)

**Docker Compose:**
19. list_compose_projects() - 列出所有 Compose 项目
20. compose_up(project) - 启动 Compose 项目
21. compose_down(project) - 停止并移除项目 (需确认)
22. compose_restart(project) - 重启项目
23. compose_status(project) - 查看项目状态和容器

**系统维护:**
24. system_status() - 获取 CPU 和内存状态
25. prune_containers() - 清理已停止的容器 (需确认)
26. prune_images() - 清理未使用的镜像 (需确认)`


// Chat Handler with Streaming
func handleAgentChat(w http.ResponseWriter, r *http.Request) {
	agentConfigMutex.RLock()
	config := agentConfig
	agentConfigMutex.RUnlock()

	if !config.Enabled {
		http.Error(w, "Agent is disabled", http.StatusForbidden)
		return
	}

	var req AgentChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Prepare Messages
	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
	}
	// Append history if needed (limit length in production)
	messages = append(messages, req.History...)
	messages = append(messages, ChatMessage{Role: "user", Content: req.Message})

	// Setup SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Call LLM
	streamLLM(w, flusher, config, messages)
}


// Helper to send data with proper SSE prefix
func sendSSE(w http.ResponseWriter, flusher http.Flusher, data string) {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprintf(w, "\n")
	flusher.Flush()
}

func streamLLM(w http.ResponseWriter, flusher http.Flusher, config AgentConfig, messages []ChatMessage) {
	for step := 0; step < 3; step++ { // Limit to 3 steps to prevent infinite loops
		apiURL := strings.TrimRight(config.APIURL, "/") + "/chat/completions"
		
		reqBody := ChatRequest{
			Model:    config.Model,
			Messages: messages,
			Stream:   true,
		}
		jsonData, _ := json.Marshal(reqBody)

		httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			sendSSE(w, flusher, fmt.Sprintf("Error creating request: %v", err))
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+config.APIKey)

		client := &http.Client{Timeout: 90 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			sendSSE(w, flusher, fmt.Sprintf("Error calling API: %v", err))
			return
		}

		reader := bufio.NewReader(resp.Body)
		var fullResponse strings.Builder
		
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk ChatCompletionChunk
			json.Unmarshal([]byte(data), &chunk)

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					fullResponse.WriteString(content)
					sendSSE(w, flusher, content)
				}
			}
		}
		resp.Body.Close()

		responseText := fullResponse.String()
		// Add assistant response to history for next possible step
		messages = append(messages, ChatMessage{Role: "assistant", Content: responseText})

		// Check for Tool Calls - extract ALL tool calls from response
		if strings.Contains(responseText, "[[TOOL:") {
			var toolResults strings.Builder
			hasToolCalls := false
			
			// Find all [[TOOL: xxx]] patterns
			remaining := responseText
			for {
				startIdx := strings.Index(remaining, "[[TOOL:")
				if startIdx == -1 {
					break
				}
				remaining = remaining[startIdx+7:] // Skip past "[[TOOL:"
				endIdx := strings.Index(remaining, "]]")
				if endIdx == -1 {
					break
				}
				
				toolCmd := strings.TrimSpace(remaining[:endIdx])
				remaining = remaining[endIdx+2:] // Move past "]]"
				
				// Execute Tool
				toolResult := executeTool(toolCmd)
				hasToolCalls = true
				
				// Send Tool Output to Client
				displayOutput := fmt.Sprintf("\n\n> 🛠️ **系统正在执行**: `%s`\n\n%s\n", toolCmd, toolResult)
				sendSSE(w, flusher, displayOutput)
				
				// Collect results for feedback
				toolResults.WriteString(fmt.Sprintf("工具 %s 执行结果: %s\n", toolCmd, toolResult))
			}
			
			if hasToolCalls {
				// Feed all results back to LLM in one message
				messages = append(messages, ChatMessage{Role: "user", Content: fmt.Sprintf("[SYSTEM] 所有工具执行完成:\n%s\n如无需继续执行工具，请直接总结结果。", toolResults.String())})
				continue
			}
		}
		break
	}
}

// Tool Execution Logic
func executeTool(command string) string {
	// Parse func(args)
	idx := strings.Index(command, "(")
	if idx == -1 {
		return "Invalid command format"
	}
	name := strings.TrimSpace(command[:idx])
	argsStr := command[idx+1 : len(command)-1] // Remove parens
	args := strings.Split(argsStr, ",")
	
	// Clean arguments: handle "id='abc'" and quotes
	cleanArgs := make([]string, 0)
	for _, a := range args {
		trimmed := strings.TrimSpace(a)
		if trimmed == "" {
			continue
		}
		// Handle key=value
		if strings.Contains(trimmed, "=") {
			parts := strings.SplitN(trimmed, "=", 2)
			trimmed = strings.TrimSpace(parts[1])
		}
		// Strip quotes
		trimmed = strings.Trim(trimmed, "\"'")
		cleanArgs = append(cleanArgs, trimmed)
	}

	arg1 := ""
	if len(cleanArgs) > 0 {
		arg1 = cleanArgs[0]
	}

	ctx := context.Background()

	switch name {
	// ========== Container Management ==========
	case "list_containers":
		containers, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{All: true})
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		var sb strings.Builder
		sb.WriteString("| ID | 名称 | 状态 |\n")
		sb.WriteString("| --- | --- | --- |\n")
		for _, c := range containers {
			cName := c.ID[:12]
			if len(c.Names) > 0 {
				cName = strings.TrimPrefix(c.Names[0], "/")
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", c.ID[:12], cName, c.Status))
		}
		return sb.String()

	case "start_container":
		if arg1 == "" { return "Missing container ID" }
		err := dockerClient.ContainerStart(ctx, arg1, types.ContainerStartOptions{})
		if err != nil { return fmt.Sprintf("Error starting: %v", err) }
		return fmt.Sprintf("✅ Container %s started.", arg1)

	case "stop_container":
		if arg1 == "" { return "Missing container ID" }
		err := dockerClient.ContainerStop(ctx, arg1, container.StopOptions{})
		if err != nil { return fmt.Sprintf("Error stopping: %v", err) }
		return fmt.Sprintf("✅ Container %s stopped.", arg1)

	case "restart_container":
		if arg1 == "" { return "Missing container ID" }
		err := dockerClient.ContainerRestart(ctx, arg1, container.StopOptions{})
		if err != nil { return fmt.Sprintf("Error restarting: %v", err) }
		return fmt.Sprintf("✅ Container %s restarted.", arg1)

	case "get_container_logs":
		if arg1 == "" { return "Missing container ID" }
		opts := container.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: "50"}
		out, err := dockerClient.ContainerLogs(ctx, arg1, opts)
		if err != nil { return fmt.Sprintf("Error reading logs: %v", err) }
		buf := new(bytes.Buffer)
		io.Copy(buf, out)
		return "```\n" + buf.String() + "\n```"
	
	case "inspect_container":
		if arg1 == "" { return "Missing container ID" }
		info, err := dockerClient.ContainerInspect(ctx, arg1)
		if err != nil { return fmt.Sprintf("Error inspecting: %v", err) }
		return fmt.Sprintf("| 属性 | 值 |\n| --- | --- |\n| 名称 | %s |\n| 状态 | %s |\n| IP地址 | %s |\n| 镜像 | %s |", 
			strings.TrimPrefix(info.Name, "/"), info.State.Status, info.NetworkSettings.IPAddress, info.Config.Image)

	case "delete_container":
		if arg1 == "" { return "Missing container ID" }
		err := dockerClient.ContainerRemove(ctx, arg1, container.RemoveOptions{Force: true})
		if err != nil { return fmt.Sprintf("Error deleting: %v", err) }
		return fmt.Sprintf("✅ Container %s deleted.", arg1)

	case "run_container":
		// Flexible container creation using docker CLI
		// Format: run_container(image, name, [options...])
		// Options: -p port:port, -e VAR=val, -v /host:/container, --restart always
		if arg1 == "" { return "Missing image name" }
		
		// Build docker run command
		cmdArgs := []string{"run", "-d"}
		
		// Add container name if provided
		if len(cleanArgs) > 1 && cleanArgs[1] != "" {
			cmdArgs = append(cmdArgs, "--name", cleanArgs[1])
		}
		
		// Process remaining args as docker options
		for i := 2; i < len(cleanArgs); i++ {
			opt := cleanArgs[i]
			if strings.HasPrefix(opt, "-p") || strings.HasPrefix(opt, "-e") || 
			   strings.HasPrefix(opt, "-v") || strings.HasPrefix(opt, "--restart") {
				// Split option with value: "-p 8080:80" or "-p=8080:80"
				if strings.Contains(opt, "=") {
					parts := strings.SplitN(opt, "=", 2)
					cmdArgs = append(cmdArgs, parts[0], parts[1])
				} else if strings.Contains(opt, " ") {
					parts := strings.SplitN(opt, " ", 2)
					cmdArgs = append(cmdArgs, parts[0], parts[1])
				} else {
					// Direct value like "8080:80" for port
					cmdArgs = append(cmdArgs, "-p", opt)
				}
			} else if strings.Contains(opt, ":") {
				// Assume it's a port mapping like "6379:6379"
				cmdArgs = append(cmdArgs, "-p", opt)
			}
		}
		
		// Add image name
		cmdArgs = append(cmdArgs, arg1)
		
		// Execute docker run
		cmd := exec.Command("docker", cmdArgs...)
		output, err := cmd.CombinedOutput()
		outputStr := strings.TrimSpace(string(output))
		
		if err != nil {
			return fmt.Sprintf("❌ Error: %v\n%s", err, outputStr)
		}
		
		// Get container ID (first 12 chars)
		containerID := outputStr
		if len(containerID) > 12 {
			containerID = containerID[:12]
		}
		
		displayName := containerID
		if len(cleanArgs) > 1 && cleanArgs[1] != "" {
			displayName = cleanArgs[1]
		}
		
		return fmt.Sprintf("✅ Container %s created and started (ID: %s)", displayName, containerID)

	// ========== Image Management ==========
	case "list_images":
		images, err := dockerClient.ImageList(ctx, types.ImageListOptions{})
		if err != nil { return fmt.Sprintf("Error: %v", err) }
		var sb strings.Builder
		sb.WriteString("| ID | 名称 | 大小 |\n")
		sb.WriteString("| --- | --- | --- |\n")
		for _, img := range images {
			name := "<none>"
			if len(img.RepoTags) > 0 {
				name = img.RepoTags[0]
			}
			size := fmt.Sprintf("%.1f MB", float64(img.Size)/1024/1024)
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", img.ID[7:19], name, size))
		}
		return sb.String()

	case "pull_image":
		if arg1 == "" { return "Missing image name" }
		out, err := dockerClient.ImagePull(ctx, arg1, types.ImagePullOptions{})
		if err != nil { return fmt.Sprintf("❌ Error pulling: %v", err) }
		defer out.Close()
		// Must consume the entire stream to ensure pull completes
		_, copyErr := io.Copy(io.Discard, out)
		if copyErr != nil { return fmt.Sprintf("❌ Error during pull: %v", copyErr) }
		// Verify image actually exists
		images, _ := dockerClient.ImageList(ctx, types.ImageListOptions{})
		for _, img := range images {
			for _, tag := range img.RepoTags {
				if tag == arg1 || strings.HasPrefix(tag, arg1) {
					return fmt.Sprintf("✅ Image %s pulled successfully (ID: %s)", arg1, img.ID[7:19])
				}
			}
		}
		return fmt.Sprintf("⚠️ Pull completed but image %s not found. Try: list_images() to check.", arg1)

	case "delete_image":
		if arg1 == "" { return "Missing image ID" }
		_, err := dockerClient.ImageRemove(ctx, arg1, types.ImageRemoveOptions{Force: true})
		if err != nil { return fmt.Sprintf("Error deleting: %v", err) }
		return fmt.Sprintf("✅ Image %s deleted.", arg1)

	// ========== Network Management ==========
	case "list_networks":
		networks, err := dockerClient.NetworkList(ctx, types.NetworkListOptions{})
		if err != nil { return fmt.Sprintf("Error: %v", err) }
		var sb strings.Builder
		sb.WriteString("| ID | 名称 | 驱动 | 范围 |\n")
		sb.WriteString("| --- | --- | --- | --- |\n")
		for _, n := range networks {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", n.ID[:12], n.Name, n.Driver, n.Scope))
		}
		return sb.String()

	case "create_network":
		if arg1 == "" { return "Missing network name" }
		driver := "bridge"
		if len(cleanArgs) > 1 {
			driver = cleanArgs[1]
		}
		resp, err := dockerClient.NetworkCreate(ctx, arg1, types.NetworkCreate{Driver: driver})
		if err != nil { return fmt.Sprintf("Error creating: %v", err) }
		return fmt.Sprintf("✅ Network %s created (ID: %s)", arg1, resp.ID[:12])

	case "delete_network":
		if arg1 == "" { return "Missing network ID" }
		err := dockerClient.NetworkRemove(ctx, arg1)
		if err != nil { return fmt.Sprintf("Error deleting: %v", err) }
		return fmt.Sprintf("✅ Network %s deleted.", arg1)

	case "inspect_network":
		if arg1 == "" { return "Missing network ID" }
		info, err := dockerClient.NetworkInspect(ctx, arg1, types.NetworkInspectOptions{})
		if err != nil { return fmt.Sprintf("Error inspecting: %v", err) }
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("**网络**: %s\n\n", info.Name))
		sb.WriteString("| 属性 | 值 |\n| --- | --- |\n")
		sb.WriteString(fmt.Sprintf("| ID | %s |\n", info.ID[:12]))
		sb.WriteString(fmt.Sprintf("| 驱动 | %s |\n", info.Driver))
		sb.WriteString(fmt.Sprintf("| 范围 | %s |\n", info.Scope))
		if len(info.Containers) > 0 {
			sb.WriteString(fmt.Sprintf("| 容器数 | %d |\n", len(info.Containers)))
		}
		return sb.String()

	// ========== Volume Management ==========
	case "list_volumes":
		vols, err := dockerClient.VolumeList(ctx, volume.ListOptions{})
		if err != nil { return fmt.Sprintf("Error: %v", err) }
		var sb strings.Builder
		sb.WriteString("| 名称 | 驱动 |\n")
		sb.WriteString("| --- | --- |\n")
		for _, v := range vols.Volumes {
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", v.Name, v.Driver))
		}
		return sb.String()

	case "create_volume":
		if arg1 == "" { return "Missing volume name" }
		vol, err := dockerClient.VolumeCreate(ctx, volume.CreateOptions{Name: arg1})
		if err != nil { return fmt.Sprintf("Error creating: %v", err) }
		return fmt.Sprintf("✅ Volume %s created.", vol.Name)

	case "delete_volume":
		if arg1 == "" { return "Missing volume name" }
		err := dockerClient.VolumeRemove(ctx, arg1, true)
		if err != nil { return fmt.Sprintf("Error deleting: %v", err) }
		return fmt.Sprintf("✅ Volume %s deleted.", arg1)

	// ========== Docker Compose ==========
	case "list_compose_projects":
		files, err := os.ReadDir(composeBaseDir)
		if err != nil { return fmt.Sprintf("Error reading compose dir: %v", err) }
		var sb strings.Builder
		sb.WriteString("| 项目名 | 状态 |\n")
		sb.WriteString("| --- | --- |\n")
		for _, f := range files {
			if f.IsDir() {
				status := getComposeProjectStatus(filepath.Join(composeBaseDir, f.Name()))
				sb.WriteString(fmt.Sprintf("| %s | %s |\n", f.Name(), status))
			}
		}
		return sb.String()

	case "compose_up":
		if arg1 == "" { return "Missing project name" }
		projectDir := filepath.Join(composeBaseDir, arg1)
		cmd := exec.Command("docker", "compose", "-f", filepath.Join(projectDir, "docker-compose.yml"), "up", "-d")
		output, err := cmd.CombinedOutput()
		if err != nil { return fmt.Sprintf("Error: %v\n%s", err, string(output)) }
		return fmt.Sprintf("✅ Compose project %s started.\n```\n%s\n```", arg1, string(output))

	case "compose_down":
		if arg1 == "" { return "Missing project name" }
		projectDir := filepath.Join(composeBaseDir, arg1)
		cmd := exec.Command("docker", "compose", "-f", filepath.Join(projectDir, "docker-compose.yml"), "down")
		output, err := cmd.CombinedOutput()
		if err != nil { return fmt.Sprintf("Error: %v\n%s", err, string(output)) }
		return fmt.Sprintf("✅ Compose project %s stopped and removed.\n```\n%s\n```", arg1, string(output))

	case "compose_restart":
		if arg1 == "" { return "Missing project name" }
		projectDir := filepath.Join(composeBaseDir, arg1)
		cmd := exec.Command("docker", "compose", "-f", filepath.Join(projectDir, "docker-compose.yml"), "restart")
		output, err := cmd.CombinedOutput()
		if err != nil { return fmt.Sprintf("Error: %v\n%s", err, string(output)) }
		return fmt.Sprintf("✅ Compose project %s restarted.\n```\n%s\n```", arg1, string(output))

	case "compose_status":
		if arg1 == "" { return "Missing project name" }
		projectDir := filepath.Join(composeBaseDir, arg1)
		cmd := exec.Command("docker", "compose", "-f", filepath.Join(projectDir, "docker-compose.yml"), "ps", "--format", "table {{.Name}}\t{{.Status}}")
		output, err := cmd.CombinedOutput()
		if err != nil { return fmt.Sprintf("Error: %v\n%s", err, string(output)) }
		return fmt.Sprintf("**项目 %s 状态:**\n```\n%s\n```", arg1, string(output))

	// ========== System Maintenance ==========
	case "system_status":
		cpu, _ := getCPUUsage()
		mem, _ := getMemoryUsage()
		disk, _ := getDiskUsage()
		return fmt.Sprintf("| 指标 | 使用率 |\n| --- | --- |\n| CPU | %.1f%% |\n| 内存 | %.1f%% |\n| 磁盘 | %.1f%% |", cpu, mem.Usage, disk.Usage)

	case "prune_containers":
		report, err := dockerClient.ContainersPrune(ctx, filters.Args{})
		if err != nil { return fmt.Sprintf("Error: %v", err) }
		return fmt.Sprintf("✅ 已清理 %d 个容器，释放 %d bytes", len(report.ContainersDeleted), report.SpaceReclaimed)

	case "prune_images":
		report, err := dockerClient.ImagesPrune(ctx, filters.Args{})
		if err != nil { return fmt.Sprintf("Error: %v", err) }
		return fmt.Sprintf("✅ 已清理 %d 个镜像，释放 %.1f MB", len(report.ImagesDeleted), float64(report.SpaceReclaimed)/1024/1024)

	default:
		return "Unknown tool: " + name
	}
}

// ============== Chat History Persistence ==============

// ChatHistoryMessage represents a stored chat message
type ChatHistoryMessage struct {
	ID        int64     `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Initialize chat history table (call from main.go after authDB is initialized)
func initChatHistoryTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS chat_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("创建聊天历史表失败: %v", err)
	}
	
	// Run initial cleanup
	cleanupOldMessages(db)
	
	return nil
}

// Cleanup messages older than 7 days
func cleanupOldMessages(db *sql.DB) {
	result, err := db.Exec(`DELETE FROM chat_messages WHERE created_at < datetime('now', '-7 days')`)
	if err != nil {
		log.Printf("[Agent] 清理旧消息失败: %v", err)
		return
	}
	if affected, _ := result.RowsAffected(); affected > 0 {
		log.Printf("[Agent] 已清理 %d 条过期消息", affected)
	}
}

// Start cleanup scheduler (runs every hour)
func startChatHistoryCleanupScheduler(db *sql.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			cleanupOldMessages(db)
		}
	}()
}

// Handler: Get chat history
func handleGetChatHistory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT id, role, content, created_at FROM chat_messages ORDER BY id ASC LIMIT 100`)
		if err != nil {
			http.Error(w, "Failed to load history", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var messages []ChatHistoryMessage
		for rows.Next() {
			var msg ChatHistoryMessage
			if err := rows.Scan(&msg.ID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
				continue
			}
			messages = append(messages, msg)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}

// Handler: Save chat message
func handleSaveChatMessage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg ChatMessage
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		_, err := db.Exec(`INSERT INTO chat_messages (role, content) VALUES (?, ?)`, msg.Role, msg.Content)
		if err != nil {
			http.Error(w, "Failed to save message", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}

// Handler: Clear chat history
func handleClearChatHistory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := db.Exec(`DELETE FROM chat_messages`)
		if err != nil {
			http.Error(w, "Failed to clear history", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}
