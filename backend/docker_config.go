package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// DockerInfo Docker 系统信息
type DockerInfo struct {
	ServerVersion  string `json:"serverVersion"`
	APIVersion     string `json:"apiVersion"`
	OS             string `json:"os"`
	Arch           string `json:"arch"`
	KernelVersion  string `json:"kernelVersion"`
	OperatingSystem string `json:"operatingSystem"`
	Containers     int    `json:"containers"`
	ContainersRunning int `json:"containersRunning"`
	ContainersPaused  int `json:"containersPaused"`
	ContainersStopped int `json:"containersStopped"`
	Images         int    `json:"images"`
	Driver         string `json:"driver"`
	MemoryLimit    bool   `json:"memoryLimit"`
	SwapLimit      bool   `json:"swapLimit"`
	CPUCfsPeriod   bool   `json:"cpuCfsPeriod"`
	CPUCfsQuota    bool   `json:"cpuCfsQuota"`
	IPv4Forwarding bool   `json:"ipv4Forwarding"`
	DockerRootDir  string `json:"dockerRootDir"`
	IndexServerAddress string `json:"indexServerAddress"`
}

// DockerConfig Docker 配置（用于前端交互，使用驼峰命名）
type DockerConfig struct {
	RegistryMirrors    []string          `json:"registryMirrors,omitempty"`
	InsecureRegistries []string          `json:"insecureRegistries,omitempty"`
	IPv6               bool              `json:"ipv6,omitempty"`
	Iptables           bool              `json:"iptables,omitempty"`
	LiveRestore        bool              `json:"liveRestore,omitempty"`
	LogDriver          string            `json:"logDriver,omitempty"`
	LogOpts            map[string]string `json:"logOpts,omitempty"`
	StorageDriver      string            `json:"storageDriver,omitempty"`
	DataRoot           string            `json:"dataRoot,omitempty"`
	CgroupDriver       string            `json:"cgroupDriver,omitempty"`
}

// DaemonConfig Docker daemon.json 配置（使用连字符命名，与文件格式一致）
type DaemonConfig struct {
	RegistryMirrors    []string          `json:"registry-mirrors,omitempty"`
	InsecureRegistries []string          `json:"insecure-registries,omitempty"`
	IPv6               bool              `json:"ipv6,omitempty"`
	Iptables           bool              `json:"iptables,omitempty"`
	LiveRestore        bool              `json:"live-restore,omitempty"`
	LogDriver          string            `json:"log-driver,omitempty"`
	LogOpts            map[string]string `json:"log-opts,omitempty"`
	StorageDriver      string            `json:"storage-driver,omitempty"`
	DataRoot           string            `json:"data-root,omitempty"`
	ExecOpts           []string          `json:"exec-opts,omitempty"`
}

// handleDockerInfo 获取 Docker 系统信息
func handleDockerInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	info, err := dockerClient.Info(ctx)
	if err != nil {
		http.Error(w, "获取 Docker 信息失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	version, err := dockerClient.ServerVersion(ctx)
	if err != nil {
		http.Error(w, "获取 Docker 版本失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result := DockerInfo{
		ServerVersion:     version.Version,
		APIVersion:        version.APIVersion,
		OS:                info.OSType,
		Arch:              info.Architecture,
		KernelVersion:     info.KernelVersion,
		OperatingSystem:   info.OperatingSystem,
		Containers:        info.Containers,
		ContainersRunning: info.ContainersRunning,
		ContainersPaused:  info.ContainersPaused,
		ContainersStopped: info.ContainersStopped,
		Images:            info.Images,
		Driver:            info.Driver,
		MemoryLimit:       info.MemoryLimit,
		SwapLimit:         info.SwapLimit,
		CPUCfsPeriod:      info.CPUCfsPeriod,
		CPUCfsQuota:       info.CPUCfsQuota,
		IPv4Forwarding:    info.IPv4Forwarding,
		DockerRootDir:     info.DockerRootDir,
		IndexServerAddress: info.IndexServerAddress,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleDockerConfigGet 获取 Docker 配置
func handleDockerConfigGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	configPath := getDockerConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 返回空配置
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(DockerConfig{
				Iptables: true, // 默认值
			})
			return
		}
		http.Error(w, "读取 Docker 配置失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 解析 daemon.json（连字符格式）
	var daemonConfig DaemonConfig
	if err := json.Unmarshal(data, &daemonConfig); err != nil {
		http.Error(w, "解析 Docker 配置失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 解析 cgroup driver（从 exec-opts 中提取）
	cgroupDriver := "cgroupfs"
	for _, opt := range daemonConfig.ExecOpts {
		if strings.HasPrefix(opt, "native.cgroupdriver=") {
			cgroupDriver = strings.TrimPrefix(opt, "native.cgroupdriver=")
			break
		}
	}

	// 转换为前端格式（驼峰命名）
	config := DockerConfig{
		RegistryMirrors:    daemonConfig.RegistryMirrors,
		InsecureRegistries: daemonConfig.InsecureRegistries,
		IPv6:               daemonConfig.IPv6,
		Iptables:           daemonConfig.Iptables,
		LiveRestore:        daemonConfig.LiveRestore,
		LogDriver:          daemonConfig.LogDriver,
		LogOpts:            daemonConfig.LogOpts,
		StorageDriver:      daemonConfig.StorageDriver,
		DataRoot:           daemonConfig.DataRoot,
		CgroupDriver:       cgroupDriver,
	}

	// 设置默认值
	if config.LogDriver == "" {
		config.LogDriver = "json-file"
	}
	if config.LogOpts == nil {
		config.LogOpts = map[string]string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// handleDockerConfigUpdate 更新 Docker 配置
func handleDockerConfigUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var newConfig DockerConfig
	if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
		http.Error(w, "请求参数错误: "+err.Error(), http.StatusBadRequest)
		return
	}

	configPath := getDockerConfigPath()

	// 读取现有配置
	var existingConfig map[string]interface{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			http.Error(w, "读取 Docker 配置失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
		existingConfig = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(data, &existingConfig); err != nil {
			existingConfig = make(map[string]interface{})
		}
	}

	// 更新配置（转换为 daemon.json 格式）
	if len(newConfig.RegistryMirrors) > 0 {
		existingConfig["registry-mirrors"] = newConfig.RegistryMirrors
	} else {
		delete(existingConfig, "registry-mirrors")
	}

	if len(newConfig.InsecureRegistries) > 0 {
		existingConfig["insecure-registries"] = newConfig.InsecureRegistries
	} else {
		delete(existingConfig, "insecure-registries")
	}

	if newConfig.IPv6 {
		existingConfig["ipv6"] = true
	} else {
		delete(existingConfig, "ipv6")
	}

	// iptables 默认为 true，只有显式设置为 false 时才写入
	if !newConfig.Iptables {
		existingConfig["iptables"] = false
	} else {
		delete(existingConfig, "iptables")
	}

	if newConfig.LiveRestore {
		existingConfig["live-restore"] = true
	} else {
		delete(existingConfig, "live-restore")
	}

	if newConfig.LogDriver != "" && newConfig.LogDriver != "json-file" {
		existingConfig["log-driver"] = newConfig.LogDriver
	} else {
		delete(existingConfig, "log-driver")
	}

	if len(newConfig.LogOpts) > 0 {
		// 过滤空值
		logOpts := make(map[string]string)
		for k, v := range newConfig.LogOpts {
			if v != "" {
				logOpts[k] = v
			}
		}
		if len(logOpts) > 0 {
			existingConfig["log-opts"] = logOpts
		} else {
			delete(existingConfig, "log-opts")
		}
	} else {
		delete(existingConfig, "log-opts")
	}

	if newConfig.StorageDriver != "" {
		existingConfig["storage-driver"] = newConfig.StorageDriver
	}

	if newConfig.DataRoot != "" {
		existingConfig["data-root"] = newConfig.DataRoot
	}

	// 处理 cgroup driver
	if newConfig.CgroupDriver != "" && newConfig.CgroupDriver != "cgroupfs" {
		existingConfig["exec-opts"] = []string{"native.cgroupdriver=" + newConfig.CgroupDriver}
	} else {
		delete(existingConfig, "exec-opts")
	}

	// 保存配置
	newData, err := json.MarshalIndent(existingConfig, "", "  ")
	if err != nil {
		http.Error(w, "序列化配置失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		http.Error(w, "保存 Docker 配置失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "配置已保存，需要重启 Docker 服务生效"})
}

// handleDockerRestart 重启 Docker 服务
func handleDockerRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "linux" {
		cmd = exec.Command("systemctl", "restart", "docker")
	} else if runtime.GOOS == "darwin" {
		// macOS 使用 launchctl
		cmd = exec.Command("osascript", "-e", `do shell script "killall Docker && open -a Docker" with administrator privileges`)
	} else {
		http.Error(w, "不支持的操作系统", http.StatusInternalServerError)
		return
	}

	if err := cmd.Run(); err != nil {
		http.Error(w, "重启 Docker 服务失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Docker 服务正在重启"})
}

// getDockerConfigPath 获取 Docker 配置文件路径
func getDockerConfigPath() string {
	if runtime.GOOS == "windows" {
		return "C:\\ProgramData\\docker\\config\\daemon.json"
	}
	return "/etc/docker/daemon.json"
}
