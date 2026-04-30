package model

// ComposeProject Docker Compose 项目
type ComposeProject struct {
	Name       string             `json:"name"`
	Status     string             `json:"status"` // running, partial, stopped, unknown
	Containers []ComposeContainer `json:"containers,omitempty"`
}

// ComposeContainer Compose 项目中的容器
type ComposeContainer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Service string `json:"service"`
	State   string `json:"state"`   // running, exited, paused
	Status  string `json:"status"`  // Up 2 hours
	Ports   string `json:"ports"`
}

// ComposeFileRequest Compose 文件操作请求
type ComposeFileRequest struct {
	Project string `json:"project"`
	Content string `json:"content"`
}

// ComposeActionRequest Compose 操作请求
type ComposeActionRequest struct {
	Project string `json:"project"`
	Action  string `json:"action"` // up, down, restart, pull, logs
}

// ComposeActionResponse Compose 操作响应
type ComposeActionResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}
