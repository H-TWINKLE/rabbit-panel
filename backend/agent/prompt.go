package agent

// SystemPrompt AI 系统提示词
const SystemPrompt = `你是一个 Docker 容器管理助手，可以帮助用户管理容器、镜像、网络、存储卷等资源。

你可以执行以下操作：
1. 容器管理：列出、启动、停止、重启、删除容器
2. 镜像管理：列出、删除、拉取镜像
3. 网络管理：列出、创建、删除网络
4. 存储卷管理：列出、创建、删除存储卷
5. Docker Compose 管理：启动、停止、重启 Compose 项目

当用户请求执行操作时，请解析用户意图并调用相应的工具。
`

// ToolDefinition 工具定义
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]string
}

// GetTools 返回可用工具列表
func GetTools() []ToolDefinition {
	return []ToolDefinition{
		{Name: "list_containers", Description: "列出所有容器", Parameters: map[string]string{"all": "是否显示所有容器（包括已停止）"}},
		{Name: "start_container", Description: "启动容器", Parameters: map[string]string{"id": "容器 ID"}},
		{Name: "stop_container", Description: "停止容器", Parameters: map[string]string{"id": "容器 ID"}},
		{Name: "list_images", Description: "列出所有镜像", Parameters: map[string]string{}},
		{Name: "list_volumes", Description: "列出所有存储卷", Parameters: map[string]string{}},
		{Name: "list_networks", Description: "列出所有网络", Parameters: map[string]string{}},
	}
}