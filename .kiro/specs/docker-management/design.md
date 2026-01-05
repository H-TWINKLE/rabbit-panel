# Docker 管理功能 - 技术设计文档

## 1. 概述

本文档描述 Docker 配置管理、仓库管理、存储卷管理三个功能模块的详细技术设计。

---

## 2. 前端架构设计

### 2.1 文件结构

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
├── api/
│   ├── dockerConfig.ts       # Docker 配置 API
│   ├── registry.ts           # 仓库 API
│   └── volumes.ts            # 存储卷 API
├── components/
│   ├── registry/
│   │   └── RegistryDialog.vue    # 仓库编辑对话框
│   └── volumes/
│       └── CreateVolumeDialog.vue # 创建卷对话框
└── types/
    └── index.ts              # 新增类型定义
```

### 2.2 路由配置

```typescript
// frontend/src/router/index.ts 新增路由
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

### 2.3 侧边栏导航

在 `MainLayout.vue` 的 `menuItems` 中新增：

```typescript
{ index: '/volumes', icon: FolderOpened, title: t('sideNav.volumes') },
{ index: '/registry', icon: OfficeBuilding, title: t('sideNav.registry') },
{ index: '/docker-config', icon: Setting, title: t('sideNav.dockerConfig') },
```

---

## 3. 类型定义

### 3.1 存储卷类型

```typescript
// types/index.ts
export interface VolumeInfo {
  name: string
  driver: string
  mountpoint: string
  scope: string
  created: string
  labels: Record<string, string>
  usageData?: {
    size: number
    refCount: number
  }
  containers: string[]  // 使用该卷的容器名称列表
}

export interface CreateVolumeRequest {
  name: string
  driver?: string
  driverOpts?: Record<string, string>
  labels?: Record<string, string>
}

export interface VolumePruneResult {
  volumesDeleted: string[]
  spaceReclaimed: number
}
```

### 3.2 仓库类型

```typescript
export interface RegistryInfo {
  id: string
  name: string
  url: string
  username?: string
  isDefault: boolean
  createdAt: string
}

export interface CreateRegistryRequest {
  name: string
  url: string
  username?: string
  password?: string
}

export interface RegistryTestResult {
  success: boolean
  message: string
  latency?: number
}
```

### 3.3 Docker 配置类型

```typescript
export interface DockerConfig {
  registryMirrors: string[]
  insecureRegistries: string[]
  ipv6: boolean
  fixedCidrV6?: string
  logDriver: string
  logOpts: Record<string, string>
  iptables: boolean
  liveRestore: boolean
  cgroupDriver: 'cgroupfs' | 'systemd'
  dataRoot?: string
  storageDriver?: string
}

export interface DockerInfo {
  serverVersion: string
  apiVersion: string
  os: string
  arch: string
  kernelVersion: string
  operatingSystem: string
  containers: number
  containersRunning: number
  containersPaused: number
  containersStopped: number
  images: number
  driver: string
  dockerRootDir: string
  memoryLimit: boolean
  swapLimit: boolean
  cpuCfsPeriod: boolean
  cpuCfsQuota: boolean
}
```

---

## 4. API 设计

### 4.1 存储卷 API

```typescript
// api/volumes.ts
import request from '@/utils/request'
import type { VolumeInfo, CreateVolumeRequest, VolumePruneResult } from '@/types'

export const volumeApi = {
  // 获取卷列表
  list(): Promise<VolumeInfo[]> {
    return request.get('/api/volumes')
  },

  // 创建卷
  create(data: CreateVolumeRequest): Promise<VolumeInfo> {
    return request.post('/api/volumes', data)
  },

  // 删除卷
  remove(name: string): Promise<void> {
    return request.delete(`/api/volumes/${encodeURIComponent(name)}`)
  },

  // 清理未使用的卷
  prune(): Promise<VolumePruneResult> {
    return request.post('/api/volumes/prune')
  },

  // 获取卷详情
  inspect(name: string): Promise<VolumeInfo> {
    return request.get(`/api/volumes/${encodeURIComponent(name)}`)
  }
}
```

### 4.2 仓库 API

```typescript
// api/registry.ts
import request from '@/utils/request'
import type { RegistryInfo, CreateRegistryRequest, RegistryTestResult } from '@/types'

export const registryApi = {
  // 获取仓库列表
  list(): Promise<RegistryInfo[]> {
    return request.get('/api/registries')
  },

  // 添加仓库
  create(data: CreateRegistryRequest): Promise<RegistryInfo> {
    return request.post('/api/registries', data)
  },

  // 更新仓库
  update(id: string, data: Partial<CreateRegistryRequest>): Promise<RegistryInfo> {
    return request.put(`/api/registries/${id}`, data)
  },

  // 删除仓库
  remove(id: string): Promise<void> {
    return request.delete(`/api/registries/${id}`)
  },

  // 测试仓库连接
  test(id: string): Promise<RegistryTestResult> {
    return request.post(`/api/registries/${id}/test`)
  }
}
```

### 4.3 Docker 配置 API

```typescript
// api/dockerConfig.ts
import request from '@/utils/request'
import type { DockerConfig, DockerInfo } from '@/types'

export const dockerConfigApi = {
  // 获取 Docker 信息
  getInfo(): Promise<DockerInfo> {
    return request.get('/api/docker/info')
  },

  // 获取当前配置
  getConfig(): Promise<DockerConfig> {
    return request.get('/api/docker/config')
  },

  // 更新配置
  updateConfig(config: Partial<DockerConfig>): Promise<void> {
    return request.put('/api/docker/config', config)
  },

  // 重启 Docker 服务
  restart(): Promise<void> {
    return request.post('/api/docker/restart')
  }
}
```

---

## 5. Store 设计

### 5.1 存储卷 Store

```typescript
// stores/volumes.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { volumeApi } from '@/api/volumes'
import type { VolumeInfo } from '@/types'

export type SortField = 'name' | 'driver' | 'created'
export type SortOrder = 'asc' | 'desc'

export const useVolumeStore = defineStore('volumes', () => {
  const volumes = ref<VolumeInfo[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  
  const searchQuery = ref('')
  const sortField = ref<SortField>('name')
  const sortOrder = ref<SortOrder>('asc')
  const currentPage = ref(1)
  const pageSize = ref(10)

  const filteredVolumes = computed(() => {
    let result = [...volumes.value]
    
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      result = result.filter(v => 
        v.name.toLowerCase().includes(query) ||
        v.driver.toLowerCase().includes(query)
      )
    }
    
    result.sort((a, b) => {
      const aVal = a[sortField.value]?.toLowerCase() ?? ''
      const bVal = b[sortField.value]?.toLowerCase() ?? ''
      if (aVal < bVal) return sortOrder.value === 'asc' ? -1 : 1
      if (aVal > bVal) return sortOrder.value === 'asc' ? 1 : -1
      return 0
    })
    
    return result
  })

  const paginatedVolumes = computed(() => {
    const start = (currentPage.value - 1) * pageSize.value
    return filteredVolumes.value.slice(start, start + pageSize.value)
  })

  const totalVolumes = computed(() => filteredVolumes.value.length)

  // 未使用的卷数量
  const unusedCount = computed(() => 
    volumes.value.filter(v => v.containers.length === 0).length
  )

  async function fetchVolumes() {
    loading.value = true
    try {
      volumes.value = await volumeApi.list()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch volumes'
    } finally {
      loading.value = false
    }
  }

  async function createVolume(data: CreateVolumeRequest) {
    await volumeApi.create(data)
    await fetchVolumes()
  }

  async function removeVolume(name: string) {
    await volumeApi.remove(name)
    await fetchVolumes()
  }

  async function pruneVolumes() {
    const result = await volumeApi.prune()
    await fetchVolumes()
    return result
  }

  function setSearch(query: string) {
    searchQuery.value = query
    currentPage.value = 1
  }

  function setPage(page: number) {
    currentPage.value = page
  }

  function setPageSize(size: number) {
    pageSize.value = size
    currentPage.value = 1
  }

  return {
    volumes,
    loading,
    error,
    searchQuery,
    sortField,
    sortOrder,
    currentPage,
    pageSize,
    filteredVolumes,
    paginatedVolumes,
    totalVolumes,
    unusedCount,
    fetchVolumes,
    createVolume,
    removeVolume,
    pruneVolumes,
    setSearch,
    setPage,
    setPageSize,
  }
})
```

### 5.2 仓库 Store

```typescript
// stores/registry.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { registryApi } from '@/api/registry'
import type { RegistryInfo, CreateRegistryRequest } from '@/types'

export const useRegistryStore = defineStore('registry', () => {
  const registries = ref<RegistryInfo[]>([])
  const loading = ref(false)

  async function fetchRegistries() {
    loading.value = true
    try {
      registries.value = await registryApi.list()
    } finally {
      loading.value = false
    }
  }

  async function createRegistry(data: CreateRegistryRequest) {
    await registryApi.create(data)
    await fetchRegistries()
  }

  async function updateRegistry(id: string, data: Partial<CreateRegistryRequest>) {
    await registryApi.update(id, data)
    await fetchRegistries()
  }

  async function removeRegistry(id: string) {
    await registryApi.remove(id)
    await fetchRegistries()
  }

  async function testRegistry(id: string) {
    return await registryApi.test(id)
  }

  return {
    registries,
    loading,
    fetchRegistries,
    createRegistry,
    updateRegistry,
    removeRegistry,
    testRegistry,
  }
})
```

### 5.3 Docker 配置 Store

```typescript
// stores/dockerConfig.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { dockerConfigApi } from '@/api/dockerConfig'
import type { DockerConfig, DockerInfo } from '@/types'

export const useDockerConfigStore = defineStore('dockerConfig', () => {
  const config = ref<DockerConfig | null>(null)
  const info = ref<DockerInfo | null>(null)
  const loading = ref(false)
  const saving = ref(false)

  async function fetchConfig() {
    loading.value = true
    try {
      const [configData, infoData] = await Promise.all([
        dockerConfigApi.getConfig(),
        dockerConfigApi.getInfo()
      ])
      config.value = configData
      info.value = infoData
    } finally {
      loading.value = false
    }
  }

  async function updateConfig(newConfig: Partial<DockerConfig>) {
    saving.value = true
    try {
      await dockerConfigApi.updateConfig(newConfig)
      await fetchConfig()
    } finally {
      saving.value = false
    }
  }

  async function restartDocker() {
    await dockerConfigApi.restart()
  }

  return {
    config,
    info,
    loading,
    saving,
    fetchConfig,
    updateConfig,
    restartDocker,
  }
})
```

---

## 6. 页面组件设计

### 6.1 存储卷管理页面 (Volumes.vue)

**布局结构：**
- 页面标题 + 操作按钮（创建、清理未使用）
- 搜索栏
- 表格：名称、驱动、挂载点、使用者、创建时间、操作
- 分页

**功能：**
- 列表展示所有存储卷
- 搜索过滤
- 创建新卷（对话框）
- 删除单个卷（确认提示）
- 批量清理未使用的卷（确认提示，显示将删除的数量）

### 6.2 仓库管理页面 (Registry.vue)

**布局结构：**
- 页面标题 + 添加按钮
- 卡片列表展示仓库
- 每个卡片显示：名称、地址、用户名、状态、操作按钮

**功能：**
- 列表展示已配置的仓库
- 添加新仓库（对话框）
- 编辑仓库配置
- 删除仓库
- 测试仓库连接

### 6.3 Docker 配置管理页面 (DockerConfig.vue)

**布局结构：**
- Docker 信息卡片（版本、系统信息等）
- 配置表单（分组展示）
  - 镜像加速器（可添加多个）
  - 私有仓库（可添加多个）
  - 网络配置（IPv6、iptables）
  - 日志配置（驱动、选项）
  - 运行时配置（Live Restore、Cgroup Driver）
- 保存按钮 + 重启 Docker 按钮

**功能：**
- 显示当前 Docker 信息
- 编辑各项配置
- 保存配置（写入 daemon.json）
- 重启 Docker 服务（需确认）

---

## 7. 国际化

### 7.1 zh-CN.ts 新增

```typescript
// sideNav 新增
sideNav: {
  // ... 现有
  volumes: '存储卷管理',
  registry: '仓库管理',
  dockerConfig: 'Docker 配置',
},

// 新增 volumes 模块
volumes: {
  title: '存储卷管理',
  create: '创建存储卷',
  prune: '清理未使用',
  pruneConfirm: '确定要清理所有未使用的存储卷吗？',
  pruneSuccess: '已清理 {count} 个存储卷，释放 {size}',
  name: '卷名称',
  driver: '驱动',
  mountpoint: '挂载点',
  usedBy: '使用者',
  unused: '未使用',
  confirmRemove: '确定要删除此存储卷吗？',
  removeSuccess: '存储卷删除成功',
  createSuccess: '存储卷创建成功',
  inUse: '存储卷正在使用中，无法删除',
  namePlaceholder: '请输入存储卷名称',
  nameRequired: '请输入存储卷名称',
  driverPlaceholder: '选择驱动（默认 local）',
},

// 新增 registry 模块
registry: {
  title: '仓库管理',
  add: '添加仓库',
  edit: '编辑仓库',
  name: '仓库名称',
  url: '仓库地址',
  username: '用户名',
  password: '密码',
  testConnection: '测试连接',
  testing: '测试中...',
  testSuccess: '连接成功',
  testFailed: '连接失败',
  confirmRemove: '确定要删除此仓库配置吗？',
  removeSuccess: '仓库删除成功',
  createSuccess: '仓库添加成功',
  updateSuccess: '仓库更新成功',
  namePlaceholder: '请输入仓库名称',
  urlPlaceholder: '例如：https://registry.example.com',
  nameRequired: '请输入仓库名称',
  urlRequired: '请输入仓库地址',
  default: '默认仓库',
},

// 新增 dockerConfig 模块
dockerConfig: {
  title: 'Docker 配置管理',
  info: 'Docker 信息',
  version: '版本',
  apiVersion: 'API 版本',
  os: '操作系统',
  arch: '架构',
  kernel: '内核版本',
  rootDir: '数据目录',
  storageDriver: '存储驱动',
  
  // 配置项
  registryMirrors: '镜像加速器',
  registryMirrorsHelp: '配置 Docker 镜像加速器地址',
  addMirror: '添加加速器',
  mirrorPlaceholder: '例如：https://mirror.example.com',
  
  insecureRegistries: '私有仓库',
  insecureRegistriesHelp: '允许使用 HTTP 协议的私有仓库',
  addRegistry: '添加私有仓库',
  registryPlaceholder: '例如：192.168.1.100:5000',
  
  ipv6: 'IPv6 支持',
  ipv6Help: '启用 Docker 的 IPv6 网络支持',
  
  logDriver: '日志驱动',
  logDriverHelp: '容器默认日志驱动',
  logOpts: '日志选项',
  maxSize: '单文件大小',
  maxFile: '文件数量',
  
  iptables: 'iptables',
  iptablesHelp: '允许 Docker 管理 iptables 规则',
  
  liveRestore: 'Live Restore',
  liveRestoreHelp: 'Docker 重启时保持容器运行',
  
  cgroupDriver: 'Cgroup 驱动',
  cgroupDriverHelp: '容器 cgroup 管理驱动',
  
  save: '保存配置',
  saveSuccess: '配置保存成功',
  saveWarning: '配置已保存，需要重启 Docker 服务才能生效',
  
  restart: '重启 Docker',
  restartConfirm: '重启 Docker 服务将影响所有运行中的容器，确定继续吗？',
  restartSuccess: 'Docker 服务重启成功',
  restarting: '正在重启...',
},
```

### 7.2 en-US.ts 新增

```typescript
// sideNav additions
sideNav: {
  // ... existing
  volumes: 'Volumes',
  registry: 'Registry',
  dockerConfig: 'Docker Config',
},

// volumes module
volumes: {
  title: 'Volume Management',
  create: 'Create Volume',
  prune: 'Prune Unused',
  pruneConfirm: 'Are you sure you want to remove all unused volumes?',
  pruneSuccess: 'Removed {count} volumes, reclaimed {size}',
  name: 'Name',
  driver: 'Driver',
  mountpoint: 'Mount Point',
  usedBy: 'Used By',
  unused: 'Unused',
  confirmRemove: 'Are you sure you want to remove this volume?',
  removeSuccess: 'Volume removed successfully',
  createSuccess: 'Volume created successfully',
  inUse: 'Volume is in use and cannot be removed',
  namePlaceholder: 'Enter volume name',
  nameRequired: 'Volume name is required',
  driverPlaceholder: 'Select driver (default: local)',
},

// registry module
registry: {
  title: 'Registry Management',
  add: 'Add Registry',
  edit: 'Edit Registry',
  name: 'Name',
  url: 'URL',
  username: 'Username',
  password: 'Password',
  testConnection: 'Test Connection',
  testing: 'Testing...',
  testSuccess: 'Connection successful',
  testFailed: 'Connection failed',
  confirmRemove: 'Are you sure you want to remove this registry?',
  removeSuccess: 'Registry removed successfully',
  createSuccess: 'Registry added successfully',
  updateSuccess: 'Registry updated successfully',
  namePlaceholder: 'Enter registry name',
  urlPlaceholder: 'e.g., https://registry.example.com',
  nameRequired: 'Registry name is required',
  urlRequired: 'Registry URL is required',
  default: 'Default',
},

// dockerConfig module
dockerConfig: {
  title: 'Docker Configuration',
  info: 'Docker Info',
  version: 'Version',
  apiVersion: 'API Version',
  os: 'OS',
  arch: 'Architecture',
  kernel: 'Kernel Version',
  rootDir: 'Root Directory',
  storageDriver: 'Storage Driver',
  
  registryMirrors: 'Registry Mirrors',
  registryMirrorsHelp: 'Configure Docker registry mirror URLs',
  addMirror: 'Add Mirror',
  mirrorPlaceholder: 'e.g., https://mirror.example.com',
  
  insecureRegistries: 'Insecure Registries',
  insecureRegistriesHelp: 'Allow HTTP registries',
  addRegistry: 'Add Registry',
  registryPlaceholder: 'e.g., 192.168.1.100:5000',
  
  ipv6: 'IPv6 Support',
  ipv6Help: 'Enable IPv6 networking for Docker',
  
  logDriver: 'Log Driver',
  logDriverHelp: 'Default logging driver for containers',
  logOpts: 'Log Options',
  maxSize: 'Max Size',
  maxFile: 'Max Files',
  
  iptables: 'iptables',
  iptablesHelp: 'Allow Docker to manage iptables rules',
  
  liveRestore: 'Live Restore',
  liveRestoreHelp: 'Keep containers running during Docker restart',
  
  cgroupDriver: 'Cgroup Driver',
  cgroupDriverHelp: 'Cgroup management driver for containers',
  
  save: 'Save Configuration',
  saveSuccess: 'Configuration saved successfully',
  saveWarning: 'Configuration saved. Docker restart required to apply changes.',
  
  restart: 'Restart Docker',
  restartConfirm: 'Restarting Docker will affect all running containers. Continue?',
  restartSuccess: 'Docker service restarted successfully',
  restarting: 'Restarting...',
},
```

---

## 8. 后端 API 接口（参考）

后端需要实现以下接口：

### 8.1 存储卷接口
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/volumes | 获取卷列表 |
| POST | /api/volumes | 创建卷 |
| DELETE | /api/volumes/:name | 删除卷 |
| POST | /api/volumes/prune | 清理未使用的卷 |

### 8.2 仓库接口
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/registries | 获取仓库列表 |
| POST | /api/registries | 添加仓库 |
| PUT | /api/registries/:id | 更新仓库 |
| DELETE | /api/registries/:id | 删除仓库 |
| POST | /api/registries/:id/test | 测试连接 |

### 8.3 Docker 配置接口
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/docker/info | 获取 Docker 信息 |
| GET | /api/docker/config | 获取当前配置 |
| PUT | /api/docker/config | 更新配置 |
| POST | /api/docker/restart | 重启 Docker |

---

## 9. 实现顺序

1. **阶段 1：基础设施**
   - 添加类型定义
   - 添加路由配置
   - 添加侧边栏导航
   - 添加国际化翻译

2. **阶段 2：存储卷管理**（最简单，先实现）
   - 创建 API 模块
   - 创建 Store
   - 创建页面组件
   - 创建对话框组件

3. **阶段 3：仓库管理**
   - 创建 API 模块
   - 创建 Store
   - 创建页面组件
   - 创建对话框组件

4. **阶段 4：Docker 配置管理**
   - 创建 API 模块
   - 创建 Store
   - 创建页面组件

---

## 10. 注意事项

1. **安全性**：仓库密码需要安全存储，前端不显示明文
2. **用户体验**：危险操作（删除、重启）需要确认提示
3. **错误处理**：所有 API 调用需要适当的错误处理和用户提示
4. **响应式**：页面需要适配移动端
5. **国际化**：所有文本使用 i18n
