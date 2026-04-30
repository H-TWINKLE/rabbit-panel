package model

// DockerInfo Docker 系统信息
type DockerInfo struct {
	ServerVersion      string `json:"serverVersion"`
	APIVersion         string `json:"apiVersion"`
	OS                 string `json:"os"`
	Arch               string `json:"arch"`
	KernelVersion      string `json:"kernelVersion"`
	OperatingSystem    string `json:"operatingSystem"`
	Containers         int    `json:"containers"`
	ContainersRunning  int    `json:"containersRunning"`
	ContainersPaused   int    `json:"containersPaused"`
	ContainersStopped  int    `json:"containersStopped"`
	Images             int    `json:"images"`
	Driver             string `json:"driver"`
	MemoryLimit        bool   `json:"memoryLimit"`
	SwapLimit          bool   `json:"swapLimit"`
	CPUCfsPeriod       bool   `json:"cpuCfsPeriod"`
	CPUCfsQuota        bool   `json:"cpuCfsQuota"`
	IPv4Forwarding     bool   `json:"ipv4Forwarding"`
	DockerRootDir      string `json:"dockerRootDir"`
	IndexServerAddress string `json:"indexServerAddress"`
}

// DockerConfig Docker 配置（API 响应/请求用，驼峰命名）
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

// DaemonConfig Docker daemon.json 格式配置（连字符命名，与文件格式一致）
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
