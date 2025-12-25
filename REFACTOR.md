# Fusion 重构计划

## 概述

将 Fusion RSS 阅读器从现有技术栈重构到新技术栈：

- **前端**: SvelteKit → React 19 + TanStack Router + shadcn/ui + Zustand
- **后端**: Echo + GORM → Gin + 纯 SQL (modernc.org/sqlite)
- **架构**: 扁平化分层 (handler → store → model)

## 设计文档

| 文档                                 | 描述                       |
| ------------------------------------ | -------------------------- |
| [前端设计](/docs/frontend-design.md) | 布局、组件、路由、交互设计 |
| [后端设计](/docs/backend-design.md)  | 数据库、API、项目结构      |

## 技术决策

| 决策项   | 选择                        |
| -------- | --------------------------- |
| 前端框架 | React 19                    |
| 路由     | TanStack Router             |
| UI 组件  | shadcn/ui                   |
| 状态管理 | Zustand                     |
| 后端框架 | Gin                         |
| 数据库   | SQLite (modernc.org/sqlite) |
| ORM      | 纯 SQL                      |
| 后端分层 | handler → store → model     |
| 多语言   | 暂时移除，仅英文            |

## 实施步骤

### 阶段 1: 后端重构

详细设计见 `/docs/backend-design.md`

#### 1.1 创建新目录结构

- [ ] 创建 `backend/` 目录
- [ ] 初始化 go.mod
- [ ] 设置基础配置

#### 1.2 实现 store 层 (纯 SQL)

- [ ] `store/store.go` - 数据库初始化和迁移
- [ ] `store/group.go` - Group CRUD
- [ ] `store/feed.go` - Feed CRUD
- [ ] `store/item.go` - Item CRUD
- [ ] `store/bookmark.go` - Bookmark CRUD

#### 1.3 实现 handler 层 (Gin)

- [ ] `handler/handler.go` - 路由注册、中间件
- [ ] `handler/session.go` - 认证
- [ ] `handler/group.go` - Group API
- [ ] `handler/feed.go` - Feed API
- [ ] `handler/item.go` - Item API
- [ ] `handler/bookmark.go` - Bookmark API

#### 1.4 迁移 pull 服务

- [ ] `pull/puller.go` - 定时拉取
- [ ] `pull/client.go` - HTTP 客户端
- [ ] `pull/backoff.go` - 退避算法

#### 1.5 迁移辅助模块

- [ ] `config/config.go` - 配置加载
- [ ] `auth/password.go` - 密码哈希
- [ ] `pkg/httpx/` - HTTP 工具

#### 1.6 核心测试

- [ ] `auth/password_test.go`
- [ ] `pull/backoff_test.go`
- [ ] `store/store_test.go` (内存 SQLite)

### 阶段 2: 前端重构

详细设计见 `/docs/frontend-design.md`

#### 2.1 初始化项目

- [ ] 创建 React 19 + TypeScript + Vite 项目
- [ ] 配置 TanStack Router
- [ ] 配置 Tailwind CSS + shadcn/ui
- [ ] 配置 Zustand

#### 2.2 实现基础设施

- [ ] `lib/api/` - API 客户端
- [ ] `lib/utils.ts` - 工具函数
- [ ] `store/app.ts` - Zustand 状态管理
- [ ] `routes/__root.tsx` - 根布局

#### 2.3 实现路由

- [ ] `routes/index.tsx` - 主视图 (`/`)
- [ ] `routes/login.tsx` - 登录页

#### 2.4 实现布局组件

- [ ] `components/layout/Sidebar.tsx` - 侧边栏
- [ ] `components/layout/ItemList.tsx` - 文章列表
- [ ] `components/layout/ItemContent.tsx` - 文章内容

#### 2.5 实现功能组件

- [ ] `components/feed/FeedItem.tsx` - Feed 项
- [ ] `components/feed/AddFeedDialog.tsx` - 添加 Feed 对话框
- [ ] `components/item/ItemCard.tsx` - 文章卡片
- [ ] `components/item/ItemActions.tsx` - 文章操作按钮
- [ ] `components/settings/SettingsDialog.tsx` - 设置对话框

#### 2.6 实现交互功能

- [ ] 全局搜索（Cmd/Ctrl + K）
- [ ] 键盘快捷键（j/k/o/u/b/v）
- [ ] 响应式布局（三栏/两栏/单栏）
- [ ] 主题切换（亮色/暗色/系统）
- [ ] Feed/Group 右键菜单和操作按钮

### 阶段 3: 集成和部署

#### 3.1 前后端集成

- [ ] 前端嵌入配置
- [ ] Dockerfile 更新
- [ ] 构建脚本更新

#### 3.2 数据迁移

- [ ] 编写迁移脚本（基于 `docs/old-database-schema.md` 定义的数据结构）
- [ ] 测试迁移流程

### 阶段 4: 清理

- [ ] 更新 Dockerfile
- [ ] 更新 GitHub Action 等
- [ ] 更新 README
