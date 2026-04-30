/*
 * Rabbit Panel - 容器运维面板
 * Author: reisen7
 * Email: reisen7@foxmail.com
 */
package main

import (
	"embed"
	"log"
	"runtime"

	"rabbit-panel/config"
	"rabbit-panel/router"
)

//go:embed dist
var _ embed.FS

func main() {
	app, err := config.NewApp()
	if err != nil {
		log.Fatalf("初始化应用失败: %v", err)
	}

	r := router.NewRouter(app)
	r.Register()

	logStartup(app)

	addr := app.Host + ":" + app.Port
	if err := app.Engine.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 记录启动信息
func logStartup(app *config.App) {
	log.Printf("容器运维面板启动成功！")
	log.Printf("监听地址: %s:%s", app.Host, app.Port)
	log.Printf("本地访问: http://localhost:%s", app.Port)

	if app.Mode == "master" {
		log.Printf("Master 节点: 管理所有 Worker 节点")
	} else {
		log.Printf("Worker 节点")
	}

	log.Printf("系统架构: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Printf("按 Ctrl+C 停止服务")
}
