#!/bin/bash

# Rabbit Panel 管理脚本
# 用法: ./rabbit.sh {start|stop|restart|status|log|build}

set -e

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# 配置
PID_FILE="rabbit-panel.pid"
LOG_FILE="rabbit-panel.log"
FRONTEND_DIR="frontend"
BACKEND_DIR="backend"

# 颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 支持的构建目标与默认编译参数
SUPPORTED_TARGETS=("amd64" "arm64" "armv7")
BUILD_LDFLAGS="${BUILD_LDFLAGS:--s -w}"

# 校验 Go 环境
ensure_go() {
    if ! command -v go > /dev/null 2>&1; then
        echo -e "${RED}错误: 未检测到 go，请先安装 Go 1.22+${NC}"
        exit 1
    fi
}

# 校验 Node.js 环境
ensure_node() {
    if ! command -v node > /dev/null 2>&1; then
        echo -e "${RED}错误: 未检测到 node，请先安装 Node.js 18+${NC}"
        exit 1
    fi
    if ! command -v npm > /dev/null 2>&1; then
        echo -e "${RED}错误: 未检测到 npm${NC}"
        exit 1
    fi
}

# 根据主机架构推断默认构建目标
detect_host_target() {
    case "$(uname -m)" in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        armv7l|armv7)
            echo "armv7"
            ;;
        *)
            echo ""
            ;;
    esac
}

# 判断是否为受支持的构建目标
is_supported_target() {
    local needle="$1"
    for target in "${SUPPORTED_TARGETS[@]}"; do
        if [ "$target" = "$needle" ]; then
            return 0
        fi
    done
    return 1
}

# 构建前端
build_frontend() {
    echo -e "${BLUE}→ 构建前端...${NC}"
    
    ensure_node
    
    cd "$FRONTEND_DIR"
    
    # 检查是否需要安装依赖
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}  安装前端依赖...${NC}"
        npm install
    fi
    
    # 构建前端
    npm run build
    
    cd "$SCRIPT_DIR"
    
    echo -e "${GREEN}  前端构建完成${NC}"
}

# 针对指定目标执行编译
build_target() {
    local target="$1"
    local goarch=""
    local goarm=""
    local suffix=""

    case "$target" in
        amd64)
            goarch="amd64"
            suffix="linux-amd64"
            ;;
        arm64)
            goarch="arm64"
            suffix="linux-arm64"
            ;;
        armv7)
            goarch="arm"
            goarm="7"
            suffix="linux-armv7"
            ;;
        *)
            echo -e "${RED}不支持的构建目标: $target${NC}"
            exit 1
            ;;
    esac

    local output_dir="dist/rabbit-panel-${suffix}-release"
    local output_file="${output_dir}/rabbit-panel-${suffix}"

    mkdir -p "$output_dir"

    # 进入后端目录编译
    cd "$BACKEND_DIR"

    # 使用国内代理加速构建
    local -a envs=("GOOS=linux" "GOARCH=$goarch" "CGO_ENABLED=0" "GOPROXY=https://goproxy.cn,direct")
    if [ -n "$goarm" ]; then
        envs+=("GOARM=$goarm")
    fi

    local upper_target
    upper_target=$(echo "$target" | tr '[:lower:]' '[:upper:]')
    local cc_var="CC_${upper_target}"
    local cc_value="${!cc_var}"
    if [ -n "$cc_value" ]; then
        envs+=("CC=$cc_value")
    fi

    echo -e "${GREEN}→ 正在构建 rabbit-panel (${suffix})...${NC}"
    env "${envs[@]}" go build -trimpath -ldflags "${BUILD_LDFLAGS}" -o "../$output_file" .
    chmod +x "../$output_file"

    cd "$SCRIPT_DIR"

    if [ "$(detect_host_target)" = "$target" ]; then
        cp "$output_file" ./rabbit-panel
        chmod +x ./rabbit-panel
    fi

    echo -e "${GREEN}   输出文件: $output_file${NC}"
}

# 检查二进制文件
check_binary() {
    BINARY=""
    if [ -f "dist/rabbit-panel-linux-amd64-release/rabbit-panel-linux-amd64" ]; then
        BINARY="dist/rabbit-panel-linux-amd64-release/rabbit-panel-linux-amd64"
    elif [ -f "dist/rabbit-panel-linux-arm64-release/rabbit-panel-linux-arm64" ]; then
        BINARY="dist/rabbit-panel-linux-arm64-release/rabbit-panel-linux-arm64"
    elif [ -f "dist/rabbit-panel-linux-armv7-release/rabbit-panel-linux-armv7" ]; then
        BINARY="dist/rabbit-panel-linux-armv7-release/rabbit-panel-linux-armv7"
    elif [ -f "./rabbit-panel" ]; then
        BINARY="./rabbit-panel"
    else
        echo -e "${RED}错误: 找不到编译好的二进制文件${NC}"
        echo "请先运行: ./rabbit.sh build"
        exit 1
    fi
}

# 设置环境变量
set_env() {
    export MODE="${MODE:-master}"
    export PORT="${PORT:-9999}"
    export HOST="${HOST:-0.0.0.0}"

    if [ -z "$JWT_SECRET" ]; then
        export JWT_SECRET="rabbit-panel-secret-key-change-in-production"
    fi

    if [ -z "$NODE_SECRET" ]; then
        export NODE_SECRET="rabbit-panel-node-secret-change-in-production"
    fi
}

# 启动
start() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            echo -e "${YELLOW}Rabbit Panel 已经在运行中 (PID: $PID)${NC}"
            return
        else
            echo "清理过期的 PID 文件..."
            rm "$PID_FILE"
        fi
    fi

    check_binary
    set_env

    echo -e "${GREEN}正在启动 Rabbit Panel...${NC}"
    echo "模式: $MODE | 端口: $PORT | 架构: $(uname -m)"
    
    nohup "$BINARY" > "$LOG_FILE" 2>&1 &
    PID=$!
    echo "$PID" > "$PID_FILE"
    
    sleep 1
    if ps -p "$PID" > /dev/null 2>&1; then
        echo -e "${GREEN}启动成功! (PID: $PID)${NC}"
        echo "日志文件: $LOG_FILE"
    else
        echo -e "${RED}启动失败，请查看日志:${NC}"
        cat "$LOG_FILE"
    fi
}

# 停止
stop() {
    if [ ! -f "$PID_FILE" ]; then
        echo -e "${YELLOW}未找到 PID 文件，Rabbit Panel 可能未运行${NC}"
        return
    fi

    PID=$(cat "$PID_FILE")
    if ! ps -p "$PID" > /dev/null 2>&1; then
        echo "进程 $PID 不存在，清理 PID 文件..."
        rm "$PID_FILE"
        return
    fi

    echo "正在停止 Rabbit Panel (PID: $PID)..."
    kill "$PID"

    for i in {1..10}; do
        if ! ps -p "$PID" > /dev/null 2>&1; then
            echo -e "${GREEN}已停止${NC}"
            rm "$PID_FILE"
            return
        fi
        sleep 0.5
    done

    echo "强制停止..."
    kill -9 "$PID"
    rm "$PID_FILE"
    echo -e "${GREEN}已强制停止${NC}"
}

# 状态
status() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            echo -e "${GREEN}Rabbit Panel 正在运行 (PID: $PID)${NC}"
            return
        fi
    fi
    echo -e "${RED}Rabbit Panel 未运行${NC}"
}

# 日志
log() {
    if [ -f "$LOG_FILE" ]; then
        tail -f "$LOG_FILE"
    else
        echo "日志文件不存在"
    fi
}

# 编译
build() {
    ensure_go

    local target="${1:-auto}"
    local skip_frontend="${2:-}"

    # 构建前端（除非指定跳过）
    if [ "$skip_frontend" != "--skip-frontend" ]; then
        build_frontend
    fi

    if [ "$target" = "auto" ]; then
        target="$(detect_host_target)"
        if [ -z "$target" ]; then
            echo -e "${RED}无法识别当前系统架构: $(uname -m)${NC}"
            exit 1
        fi
    fi

    if [ "$target" = "all" ]; then
        for t in "${SUPPORTED_TARGETS[@]}"; do
            build_target "$t"
        done
        return
    fi

    if ! is_supported_target "$target"; then
        echo -e "${RED}未知的构建目标: $target${NC}"
        echo "可用目标: auto | all | ${SUPPORTED_TARGETS[*]}"
        exit 1
    fi

    build_target "$target"
}

# 帮助信息
show_help() {
    echo "Rabbit Panel 管理脚本"
    echo ""
    echo "用法: $0 <命令> [参数]"
    echo ""
    echo "命令:"
    echo "  start              启动 Rabbit Panel"
    echo "  stop               停止 Rabbit Panel"
    echo "  restart            重启 Rabbit Panel"
    echo "  status             查看运行状态"
    echo "  log                查看日志"
    echo "  build [target]     构建项目"
    echo ""
    echo "构建参数:"
    echo "  auto               自动检测当前架构 (默认)"
    echo "  all                构建所有架构"
    echo "  amd64              构建 Linux AMD64"
    echo "  arm64              构建 Linux ARM64"
    echo "  armv7              构建 Linux ARMv7"
    echo "  --skip-frontend    跳过前端构建"
    echo ""
    echo "示例:"
    echo "  $0 build                    # 构建当前架构"
    echo "  $0 build all                # 构建所有架构"
    echo "  $0 build amd64              # 只构建 AMD64"
    echo "  $0 build auto --skip-frontend  # 跳过前端构建"
}

# 主逻辑
case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop
        sleep 1
        start
        ;;
    status)
        status
        ;;
    log)
        log
        ;;
    build)
        shift
        build "$@"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status|build|log|help}"
        echo "build 参数: auto(默认)|all|amd64|arm64|armv7 [--skip-frontend]"
        exit 1
        ;;
esac
