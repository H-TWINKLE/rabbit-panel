# Docker 管理功能需求规格

## 概述
为 Rabbit Panel 添加三个新的管理页面：Docker 配置管理、仓库管理、存储卷管理。这些页面将帮助用户更方便地管理 Docker 守护进程的配置和资源。

## 用户故事

### US-1: Docker 配置管理
**作为** 系统管理员  
**我希望** 能够通过 Web 界面管理 Docker 守护进程的配置  
**以便** 无需手动编辑 daemon.json 文件即可调整 Docker 设置

#### 验收标准
- [ ] 可以查看当前 Docker 配置
- [ ] 可以配置镜像加速器（registry-mirrors）
- [ ] 可以配置私有仓库（insecure-registries）
- [ ] 可以启用/禁用 IPv6
- [ ] 可以配置日志切割（log-driver, log-opts）
- [ ] 可以启用/禁用 iptables
- [ ] 可以启用/禁用 Live Restore
- [ ] 可以配置 cgroup driver（cgroupfs/systemd）
- [ ] 可以查看/修改 Docker Socket 路径
- [ ] 修改配置后需要重启 Docker 服务才能生效
- [ ] 显示配置修改的警告提示

### US-2: 仓库管理
**作为** 开发人员  
**我希望** 能够管理 Docker 镜像仓库  
**以便** 方便地从私有仓库拉取和推送镜像

#### 验收标准
- [ ] 可以查看已配置的仓库列表
- [ ] 可以添加新的镜像仓库（地址、用户名、密码）
- [ ] 可以编辑已有仓库配置
- [ ] 可以删除仓库配置
- [ ] 可以测试仓库连接
- [ ] 支持 Docker Hub、Harbor、阿里云等常见仓库
- [ ] 密码以安全方式存储和显示

### US-3: 存储卷管理
**作为** 系统管理员  
**我希望** 能够管理 Docker 存储卷  
**以便** 更好地管理容器数据持久化

#### 验收标准
- [ ] 可以查看所有存储卷列表
- [ ] 显示卷名称、驱动、挂载点、大小、创建时间
- [ ] 可以创建新的存储卷
- [ ] 可以删除未使用的存储卷
- [ ] 可以批量清理未使用的存储卷（prune）
- [ ] 显示卷的使用状态（被哪些容器使用）
- [ ] 删除前显示确认提示

---

## 技术设计

### 前端组件结构

```
frontend/src/
├── views/
│   ├── DockerConfig.vue      # Docker 配置管理页面
│   ├── Registry.vue          # 仓库管理页面
│   └── Volumes.vue           # 存储卷管理页面
├── stores/
│   ├── dockerConfig.ts       # Docker 配置状态管理
│   ├── registry.ts           # 仓库状态管理
│   └── volumes.ts            # 存储卷状态管理
└── components/
    ├── docker-config/
    │   └── ConfigForm.vue    # 配置表单组件
    ├── registry/
    │   └── RegistryDialog.vue # 仓库编辑对话框
    └── volumes/
        └── CreateVolumeDialog.vue # 创建卷对话框
```

### 后端 API 设计

#### Docker 配置 API
```
GET    /api/docker/config          # 获取当前配置
PUT    /api/docker/config          # 更新配置
POST   /api/docker/config/restart  # 重启 Docker 服务
```

#### 仓库管理 API
```
GET    /api/registries             # 获取仓库列表
POST   /api/registries             # 添加仓库
PUT    /api/registries/:id         # 更新仓库
DELETE /api/registries/:id         # 删除仓库
POST   /api/registries/:id/test    # 测试仓库连接
```

#### 存储卷 API
```
GET    /api/volumes                # 获取卷列表
POST   /api/volumes                # 创建卷
DELETE /api/volumes/:name          # 删除卷
POST   /api/volumes/prune          # 清理未使用的卷
```

### 路由配置

```typescript
// 新增路由
{
  path: 'docker-config',
  name: 'dockerConfig',
  component: () => import('@/views/DockerConfig.vue'),
  meta: { title: 'Docker 配置' }
},
{
  path: 'registry',
  name: 'registry',
  component: () => import('@/views/Registry.vue'),
  meta: { title: '仓库管理' }
},
{
  path: 'volumes',
  name: 'volumes',
  component: () => import('@/views/Volumes.vue'),
  meta: { title: '存储卷管理' }
}
```

### 国际化

需要在 `zh-CN.ts` 和 `en-US.ts` 中添加以下翻译键：

```typescript
// sideNav 新增
sideNav: {
  dockerConfig: 'Docker 配置',
  registry: '仓库管理',
  volumes: '存储卷管理',
}

// 新增 dockerConfig 模块
dockerConfig: {
  title: 'Docker 配置管理',
  registryMirrors: '镜像加速器',
  insecureRegistries: '私有仓库',
  ipv6: 'IPv6',
  logDriver: '日志驱动',
  logOpts: '日志选项',
  iptables: 'iptables',
  liveRestore: 'Live Restore',
  cgroupDriver: 'Cgroup 驱动',
  socketPath: 'Socket 路径',
  restartRequired: '修改配置后需要重启 Docker 服务',
  restartDocker: '重启 Docker',
  // ...
}

// 新增 registry 模块
registry: {
  title: '仓库管理',
  add: '添加仓库',
  address: '仓库地址',
  username: '用户名',
  password: '密码',
  testConnection: '测试连接',
  // ...
}

// 新增 volumes 模块
volumes: {
  title: '存储卷管理',
  create: '创建存储卷',
  prune: '清理未使用',
  driver: '驱动',
  mountpoint: '挂载点',
  usedBy: '使用者',
  // ...
}
```

---

## 实现任务

### 阶段 1: 基础设施
- [ ] 1.1 添加路由配置
- [ ] 1.2 添加侧边栏导航项
- [ ] 1.3 添加国际化翻译

### 阶段 2: 存储卷管理（最简单，先实现）
- [ ] 2.1 创建 volumes store
- [ ] 2.2 创建 Volumes.vue 页面
- [ ] 2.3 实现卷列表展示
- [ ] 2.4 实现创建卷功能
- [ ] 2.5 实现删除卷功能
- [ ] 2.6 实现批量清理功能

### 阶段 3: 仓库管理
- [ ] 3.1 创建 registry store
- [ ] 3.2 创建 Registry.vue 页面
- [ ] 3.3 实现仓库列表展示
- [ ] 3.4 实现添加/编辑仓库对话框
- [ ] 3.5 实现测试连接功能

### 阶段 4: Docker 配置管理
- [ ] 4.1 创建 dockerConfig store
- [ ] 4.2 创建 DockerConfig.vue 页面
- [ ] 4.3 实现配置表单
- [ ] 4.4 实现配置保存
- [ ] 4.5 实现重启 Docker 功能

---

## 依赖关系
- 后端需要实现相应的 API 接口
- 需要 Docker API 权限来读取/修改配置
- 修改 daemon.json 需要 root 权限

## 风险与注意事项
1. 修改 Docker 配置可能导致 Docker 服务不可用
2. 重启 Docker 会影响所有运行中的容器
3. 需要在界面上明确提示用户操作的风险
4. 建议在修改配置前备份原配置
