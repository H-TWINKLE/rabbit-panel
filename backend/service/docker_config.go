package service

import (
	"encoding/json"
	"os"
	"runtime"

	"rabbit-panel/model"
)

// DockerConfigService Docker 配置服务
type DockerConfigService struct{}

// NewDockerConfigService 创建 Docker 配置服务
func NewDockerConfigService() *DockerConfigService {
	return &DockerConfigService{}
}

// GetConfigPath 获取 Docker 配置文件路径
func GetDockerConfigPath() string {
	if runtime.GOOS == "windows" {
		return "C:\\ProgramData\\docker\\config\\daemon.json"
	}
	return "/etc/docker/daemon.json"
}

// DaemonConfig Docker daemon.json 配置
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

// ToPublicConfig 转换为 API 格式
func (dc *DaemonConfig) ToPublicConfig() *model.DockerConfig {
	cfg := &model.DockerConfig{
		RegistryMirrors:    dc.RegistryMirrors,
		InsecureRegistries: dc.InsecureRegistries,
		IPv6:               dc.IPv6,
		Iptables:           dc.Iptables,
		LiveRestore:        dc.LiveRestore,
		LogDriver:          dc.LogDriver,
		LogOpts:            dc.LogOpts,
		StorageDriver:      dc.StorageDriver,
		DataRoot:           dc.DataRoot,
	}

	// Extract cgroup driver from exec-opts
	for _, opt := range dc.ExecOpts {
		if len(opt) > 20 && opt[:20] == "native.cgroupdriver=" {
			cfg.CgroupDriver = opt[20:]
		}
	}

	if cfg.LogDriver == "" {
		cfg.LogDriver = "json-file"
	}
	if cfg.LogOpts == nil {
		cfg.LogOpts = map[string]string{}
	}

	return cfg
}

// FromPublicConfig 从 API 格式转换
func FromPublicConfig(cfg *model.DockerConfig) *DaemonConfig {
	dc := &DaemonConfig{
		RegistryMirrors:    cfg.RegistryMirrors,
		InsecureRegistries: cfg.InsecureRegistries,
		IPv6:               cfg.IPv6,
		Iptables:           cfg.Iptables,
		LiveRestore:        cfg.LiveRestore,
		LogDriver:          cfg.LogDriver,
		LogOpts:            cfg.LogOpts,
		StorageDriver:      cfg.StorageDriver,
		DataRoot:           cfg.DataRoot,
	}

	if cfg.CgroupDriver != "" && cfg.CgroupDriver != "cgroupfs" {
		dc.ExecOpts = []string{"native.cgroupdriver=" + cfg.CgroupDriver}
	}

	return dc
}

// LoadDaemonConfig 加载 daemon.json
func LoadDaemonConfig() (*DaemonConfig, error) {
	data, err := os.ReadFile(GetDockerConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &DaemonConfig{Iptables: true}, nil
		}
		return nil, err
	}
	var dc DaemonConfig
	if err := json.Unmarshal(data, &dc); err != nil {
		return nil, err
	}
	return &dc, nil
}

// SaveDaemonConfig 保存 daemon.json
func SaveDaemonConfig(dc *DaemonConfig) error {
	data, err := json.MarshalIndent(dc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(GetDockerConfigPath(), data, 0644)
}
