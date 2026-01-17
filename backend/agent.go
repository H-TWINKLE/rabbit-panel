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
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
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
你的职责是协助用户管理服务器和 Docker 容器。

**回复规范:**
1. **结构化显示**: 尽量使用 Markdown 表格、列表和代码块来展示数据。
2. **工具调用**: 当你需要执行操作时，请使用 [[TOOL: function(args)]] 格式。**参数请直接写值**。
   - 正确示例: [[TOOL: start_container(abc12345)]]
   - 错误示例: [[TOOL: start_container(id="abc")]]
3. **多步执行**: 你可以连续执行多个工具调用，但不要超过 3 步。
4. **确认**: 在执行停止、重启或删除容器等关键操作前，请先征得用户确认。
5. **专注目标**: 仅回答用户提出的问题，**不要进行未经请求的日志分析或故障诊断**。如果用户只让你列出容器，就只列出容器，不要输出额外的分析。
6. **简洁专业**: 使用中文回复，保持精简。

**可用工具:**
1. list_containers(): 列出所有容器
2. start_container(id): 启动容器
3. stop_container(id): 停止容器
4. restart_container(id): 重启容器
5. get_container_logs(id): 获取最近 50 行日志
6. inspect_container(id): 获取容器详细信息
7. system_status(): 获取 CPU 和内存状态`

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

		// Check for Tool Calls
		if strings.Contains(responseText, "[[TOOL:") {
			parts := strings.Split(responseText, "[[TOOL:")
			if len(parts) > 1 {
				toolPart := parts[1]
				endIdx := strings.Index(toolPart, "]]")
				if endIdx != -1 {
					toolCmd := strings.TrimSpace(toolPart[:endIdx])
					
					// Execute Tool
					toolResult := executeTool(toolCmd)
					
					// Send Tool Output to Client with clear separation
					displayOutput := fmt.Sprintf("\n\n---\n> 🛠️ **系统正在执行**: `[[TOOL: %s]]`\n\n%s\n\n---\n", toolCmd, toolResult)
					sendSSE(w, flusher, displayOutput)
					
					// Feed back to LLM
					messages = append(messages, ChatMessage{Role: "user", Content: fmt.Sprintf("[SYSTEM] 工具执行结果 (%s):\n%s\n请根据结果继续回答。", toolCmd, toolResult)})
					continue
				}
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
		return fmt.Sprintf("Container %s started.", arg1)

	case "stop_container":
		if arg1 == "" { return "Missing container ID" }
		err := dockerClient.ContainerStop(ctx, arg1, container.StopOptions{})
		if err != nil { return fmt.Sprintf("Error stopping: %v", err) }
		return fmt.Sprintf("Container %s stopped.", arg1)

	case "restart_container":
		if arg1 == "" { return "Missing container ID" }
		err := dockerClient.ContainerRestart(ctx, arg1, container.StopOptions{})
		if err != nil { return fmt.Sprintf("Error restarting: %v", err) }
		return fmt.Sprintf("Container %s restarted.", arg1)

	case "get_container_logs":
		if arg1 == "" { return "Missing container ID" }
		opts := container.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: "50"}
		out, err := dockerClient.ContainerLogs(ctx, arg1, opts)
		if err != nil { return fmt.Sprintf("Error reading logs: %v", err) }
		buf := new(bytes.Buffer)
		io.Copy(buf, out)
		return buf.String()
	
	case "inspect_container":
		if arg1 == "" { return "Missing container ID" }
		info, err := dockerClient.ContainerInspect(ctx, arg1)
		if err != nil { return fmt.Sprintf("Error inspecting: %v", err) }
		return fmt.Sprintf("| 属性 | 值 |\n| --- | --- |\n| 名称 | %s |\n| 状态 | %s |\n| IP地址 | %s |\n| 镜像 | %s |", 
			strings.TrimPrefix(info.Name, "/"), info.State.Status, info.NetworkSettings.IPAddress, info.Config.Image)

	case "system_status":
		cpu, _ := getCPUUsage()
		mem, _ := getMemoryUsage()
		return fmt.Sprintf("| 指标 | 使用率 |\n| --- | --- |\n| CPU | %.1f%% |\n| 内存 | %.1f%% (%d MB) |", cpu, mem.Usage, mem.Used/1024)

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
