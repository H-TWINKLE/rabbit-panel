package model

// RegistryInfo 镜像仓库信息
type RegistryInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	Username  string `json:"username"`
	Password  string `json:"password,omitempty"`
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
