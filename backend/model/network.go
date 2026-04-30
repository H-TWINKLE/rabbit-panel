package model

// NetworkInfo 网络信息
type NetworkInfo struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Driver   string   `json:"driver"`
	Scope    string   `json:"scope"`
	Internal bool     `json:"internal"`
	Attachable bool   `json:"attachable"`
	Containers []string `json:"containers"`
	Subnet   string   `json:"subnet"`
	Gateway  string   `json:"gateway"`
	Created  string   `json:"created"`
}

// NetworkConnectRequest 网络连接请求
type NetworkConnectRequest struct {
	NetworkID   string `json:"network_id"`
	ContainerID string `json:"container_id"`
}

// NetworkDisconnectRequest 网络断开请求
type NetworkDisconnectRequest struct {
	NetworkID   string `json:"network_id"`
	ContainerID string `json:"container_id"`
	Force       bool   `json:"force"`
}

// CreateNetworkRequest 创建网络请求
type CreateNetworkRequest struct {
	Name     string `json:"name"`
	Driver   string `json:"driver"`   // bridge, overlay, host, none
	Subnet   string `json:"subnet"`
	Gateway  string `json:"gateway"`
	Internal bool   `json:"internal"`
	Attachable bool `json:"attachable"`
	Labels   map[string]string `json:"labels"`
}
