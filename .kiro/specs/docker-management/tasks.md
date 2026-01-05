# Implementation Plan: Docker 管理功能

## Overview

本实现计划将 Docker 配置管理、仓库管理、存储卷管理三个功能模块分阶段实现。从基础设施开始，然后按复杂度递增的顺序实现各个功能页面。

## Tasks

- [x] 1. 基础设施搭建
  - [x] 1.1 添加类型定义到 `frontend/src/types/index.ts`
    - 添加 VolumeInfo, CreateVolumeRequest, VolumePruneResult 类型
    - 添加 RegistryInfo, CreateRegistryRequest, RegistryTestResult 类型
    - 添加 DockerConfig, DockerInfo 类型
    - _Requirements: US-1, US-2, US-3_

  - [x] 1.2 添加路由配置到 `frontend/src/router/index.ts`
    - 添加 /volumes 路由
    - 添加 /registry 路由
    - 添加 /docker-config 路由
    - _Requirements: US-1, US-2, US-3_

  - [x] 1.3 添加国际化翻译
    - 更新 `frontend/src/locales/zh-CN.ts` 添加 sideNav、volumes、registry、dockerConfig 模块
    - 更新 `frontend/src/locales/en-US.ts` 添加对应英文翻译
    - _Requirements: US-1, US-2, US-3_

  - [x] 1.4 更新侧边栏导航 `frontend/src/layouts/MainLayout.vue`
    - 添加存储卷管理、仓库管理、Docker 配置三个导航项
    - 导入对应图标 (FolderOpened, OfficeBuilding, Setting)
    - _Requirements: US-1, US-2, US-3_

- [x] 2. Checkpoint - 基础设施验证
  - 确保路由配置正确
  - 确保侧边栏导航显示正常
  - 确保国际化翻译生效

- [x] 3. 存储卷管理功能实现
  - [x] 3.1 创建存储卷 API 模块 `frontend/src/api/volumes.ts`
    - 实现 list, create, remove, prune, inspect 方法
    - _Requirements: US-3_

  - [x] 3.2 创建存储卷 Store `frontend/src/stores/volumes.ts`
    - 实现状态管理：volumes, loading, error
    - 实现搜索、排序、分页功能
    - 实现 fetchVolumes, createVolume, removeVolume, pruneVolumes 方法
    - _Requirements: US-3_

  - [x] 3.3 创建创建卷对话框组件 `frontend/src/components/volumes/CreateVolumeDialog.vue`
    - 表单：卷名称、驱动选择
    - 表单验证
    - _Requirements: US-3_

  - [x] 3.4 创建存储卷管理页面 `frontend/src/views/Volumes.vue`
    - 页面标题和操作按钮（创建、清理未使用）
    - 搜索栏
    - 表格：名称、驱动、挂载点、使用者、创建时间、操作
    - 分页
    - 删除确认对话框
    - 清理确认对话框
    - _Requirements: US-3_

- [x] 4. Checkpoint - 存储卷功能验证
  - 确保存储卷列表正常显示
  - 确保创建、删除、清理功能正常
  - 如有问题请告知

- [x] 5. 仓库管理功能实现
  - [x] 5.1 创建仓库 API 模块 `frontend/src/api/registry.ts`
    - 实现 list, create, update, remove, test 方法
    - _Requirements: US-2_

  - [x] 5.2 创建仓库 Store `frontend/src/stores/registry.ts`
    - 实现状态管理：registries, loading
    - 实现 fetchRegistries, createRegistry, updateRegistry, removeRegistry, testRegistry 方法
    - _Requirements: US-2_

  - [x] 5.3 创建仓库编辑对话框组件 `frontend/src/components/registry/RegistryDialog.vue`
    - 表单：名称、地址、用户名、密码
    - 支持添加和编辑模式
    - 测试连接按钮
    - 表单验证
    - _Requirements: US-2_

  - [x] 5.4 创建仓库管理页面 `frontend/src/views/Registry.vue`
    - 页面标题和添加按钮
    - 卡片列表展示仓库
    - 每个卡片显示：名称、地址、用户名、操作按钮
    - 编辑、删除、测试连接功能
    - _Requirements: US-2_

- [x] 6. Checkpoint - 仓库管理功能验证
  - 确保仓库列表正常显示
  - 确保添加、编辑、删除、测试连接功能正常
  - 如有问题请告知

- [x] 7. Docker 配置管理功能实现
  - [x] 7.1 创建 Docker 配置 API 模块 `frontend/src/api/dockerConfig.ts`
    - 实现 getInfo, getConfig, updateConfig, restart 方法
    - _Requirements: US-1_

  - [x] 7.2 创建 Docker 配置 Store `frontend/src/stores/dockerConfig.ts`
    - 实现状态管理：config, info, loading, saving
    - 实现 fetchConfig, updateConfig, restartDocker 方法
    - _Requirements: US-1_

  - [x] 7.3 创建 Docker 配置管理页面 `frontend/src/views/DockerConfig.vue`
    - Docker 信息卡片（版本、系统信息等）
    - 配置表单分组：
      - 镜像加速器（可添加多个）
      - 私有仓库（可添加多个）
      - 网络配置（IPv6、iptables）
      - 日志配置（驱动、选项）
      - 运行时配置（Live Restore、Cgroup Driver）
    - 保存按钮
    - 重启 Docker 按钮（带确认提示）
    - _Requirements: US-1_

- [x] 8. Final Checkpoint - 全部功能验证
  - 确保所有三个管理页面正常工作
  - 确保响应式布局在移动端正常
  - 确保国际化切换正常
  - 如有问题请告知

## Notes

- 每个阶段完成后都有 Checkpoint，用于验证功能
- 后端 API 需要同步实现，前端会调用对应接口
- 所有页面遵循现有项目的代码风格和组件模式
- 危险操作（删除、重启）都需要确认提示
