package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RegistryInfo 镜像仓库信息
type RegistryInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	Username  string `json:"username"`
	Password  string `json:"password,omitempty"` // 返回时不包含密码
	IsDefault bool   `json:"isDefault"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// CreateRegistryRequest 创建镜像仓库请求
type CreateRegistryRequest struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	IsDefault bool   `json:"isDefault"`
}

// RegistryTestResult 测试镜像仓库连接结果
type RegistryTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// 镜像仓库存储
var (
	registries     = make(map[string]*RegistryInfo)
	registriesLock sync.RWMutex
	registriesFile = "data/registries.json"
)

// initRegistries 初始化镜像仓库
func initRegistries() {
	// 确保数据目录存在
	os.MkdirAll("data", 0755)

	// 加载已保存的镜像仓库
	loadRegistries()
}

// loadRegistries 从文件加载镜像仓库
func loadRegistries() {
	registriesLock.Lock()
	defer registriesLock.Unlock()

	data, err := os.ReadFile(registriesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		return
	}

	var regs []*RegistryInfo
	if err := json.Unmarshal(data, &regs); err != nil {
		return
	}

	for _, reg := range regs {
		registries[reg.ID] = reg
	}
}

// saveRegistries 保存镜像仓库到文件
func saveRegistries() error {
	registriesLock.RLock()
	defer registriesLock.RUnlock()

	regs := make([]*RegistryInfo, 0, len(registries))
	for _, reg := range registries {
		regs = append(regs, reg)
	}

	data, err := json.MarshalIndent(regs, "", "  ")
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(registriesFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(registriesFile, data, 0644)
}

// handleRegistriesList 获取镜像仓库列表
func handleRegistriesList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	registriesLock.RLock()
	defer registriesLock.RUnlock()

	regs := make([]*RegistryInfo, 0, len(registries))
	for _, reg := range registries {
		// 复制一份，不返回密码
		regCopy := *reg
		regCopy.Password = ""
		regs = append(regs, &regCopy)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(regs)
}

// handleRegistriesCreate 创建镜像仓库
func handleRegistriesCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req CreateRegistryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求参数错误: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "仓库名称不能为空", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "仓库地址不能为空", http.StatusBadRequest)
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	reg := &RegistryInfo{
		ID:        uuid.New().String(),
		Name:      req.Name,
		URL:       req.URL,
		Username:  req.Username,
		Password:  req.Password,
		IsDefault: req.IsDefault,
		CreatedAt: now,
		UpdatedAt: now,
	}

	registriesLock.Lock()
	// 如果设置为默认，取消其他默认
	if req.IsDefault {
		for _, r := range registries {
			r.IsDefault = false
		}
	}
	registries[reg.ID] = reg
	registriesLock.Unlock()

	if err := saveRegistries(); err != nil {
		http.Error(w, "保存镜像仓库失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回时不包含密码
	regCopy := *reg
	regCopy.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(regCopy)
}

// handleRegistriesUpdate 更新镜像仓库
func handleRegistriesUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 路径中提取 ID
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/api/registries/")
	if id == "" || id == path || strings.Contains(id, "/") {
		http.Error(w, "镜像仓库 ID 不能为空", http.StatusBadRequest)
		return
	}

	var req CreateRegistryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求参数错误: "+err.Error(), http.StatusBadRequest)
		return
	}

	registriesLock.Lock()
	reg, exists := registries[id]
	if !exists {
		registriesLock.Unlock()
		http.Error(w, "镜像仓库不存在", http.StatusNotFound)
		return
	}

	// 更新字段
	if req.Name != "" {
		reg.Name = req.Name
	}
	if req.URL != "" {
		reg.URL = req.URL
	}
	if req.Username != "" {
		reg.Username = req.Username
	}
	if req.Password != "" {
		reg.Password = req.Password
	}
	if req.IsDefault {
		// 取消其他默认
		for _, r := range registries {
			r.IsDefault = false
		}
	}
	reg.IsDefault = req.IsDefault
	reg.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
	registriesLock.Unlock()

	if err := saveRegistries(); err != nil {
		http.Error(w, "保存镜像仓库失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回时不包含密码
	regCopy := *reg
	regCopy.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(regCopy)
}

// handleRegistriesRemove 删除镜像仓库
func handleRegistriesRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 路径中提取 ID
	path := r.URL.Path
	id := strings.TrimPrefix(path, "/api/registries/")
	if id == "" || id == path || strings.Contains(id, "/") {
		http.Error(w, "镜像仓库 ID 不能为空", http.StatusBadRequest)
		return
	}

	registriesLock.Lock()
	if _, exists := registries[id]; !exists {
		registriesLock.Unlock()
		http.Error(w, "镜像仓库不存在", http.StatusNotFound)
		return
	}
	delete(registries, id)
	registriesLock.Unlock()

	if err := saveRegistries(); err != nil {
		http.Error(w, "保存镜像仓库失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleRegistriesTest 测试镜像仓库连接
func handleRegistriesTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 路径中提取 ID
	path := r.URL.Path
	path = strings.TrimSuffix(path, "/test")
	id := strings.TrimPrefix(path, "/api/registries/")
	if id == "" || id == path {
		http.Error(w, "镜像仓库 ID 不能为空", http.StatusBadRequest)
		return
	}

	registriesLock.RLock()
	reg, exists := registries[id]
	registriesLock.RUnlock()

	if !exists {
		http.Error(w, "镜像仓库不存在", http.StatusNotFound)
		return
	}

	// 测试连接
	result := testRegistryConnection(reg.URL, reg.Username, reg.Password)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// testRegistryConnection 测试 Docker Registry 连接
func testRegistryConnection(registryURL, username, password string) RegistryTestResult {
	// 规范化 URL
	url := strings.TrimSuffix(registryURL, "/")
	
	// 确保有协议前缀
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	// Docker Registry API v2 端点
	apiURL := url + "/v2/"

	// 创建 HTTP 客户端，支持跳过证书验证（用于自签名证书）
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return RegistryTestResult{
			Success: false,
			Message: fmt.Sprintf("创建请求失败: %v", err),
		}
	}

	// 添加认证头
	if username != "" && password != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		req.Header.Set("Authorization", "Basic "+auth)
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(startTime)

	if err != nil {
		// 如果 HTTPS 失败，尝试 HTTP
		if strings.HasPrefix(url, "https://") {
			httpURL := strings.Replace(url, "https://", "http://", 1) + "/v2/"
			req2, _ := http.NewRequestWithContext(ctx, "GET", httpURL, nil)
			if username != "" && password != "" {
				auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
				req2.Header.Set("Authorization", "Basic "+auth)
			}
			startTime = time.Now()
			resp, err = client.Do(req2)
			latency = time.Since(startTime)
		}
		
		if err != nil {
			return RegistryTestResult{
				Success: false,
				Message: fmt.Sprintf("连接失败: %v", err),
			}
		}
	}
	defer resp.Body.Close()

	// 检查响应状态
	switch resp.StatusCode {
	case http.StatusOK:
		return RegistryTestResult{
			Success: true,
			Message: fmt.Sprintf("连接成功 (延迟: %dms)", latency.Milliseconds()),
		}
	case http.StatusUnauthorized:
		// 401 表示仓库存在但需要认证
		if username == "" {
			return RegistryTestResult{
				Success: true,
				Message: fmt.Sprintf("仓库可访问，需要认证 (延迟: %dms)", latency.Milliseconds()),
			}
		}
		return RegistryTestResult{
			Success: false,
			Message: "认证失败: 用户名或密码错误",
		}
	case http.StatusForbidden:
		return RegistryTestResult{
			Success: false,
			Message: "访问被拒绝: 权限不足",
		}
	case http.StatusNotFound:
		return RegistryTestResult{
			Success: false,
			Message: "仓库地址无效: 未找到 Registry API",
		}
	default:
		return RegistryTestResult{
			Success: false,
			Message: fmt.Sprintf("连接失败: HTTP %d", resp.StatusCode),
		}
	}
}
