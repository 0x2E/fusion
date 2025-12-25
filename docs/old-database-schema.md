# 数据库结构文档

本文档记录端老版本的数据库结构.

## 数据表

### groups

分组表，用于组织 Feed。

| 字段       | 类型     | 约束                                        | 说明       |
| ---------- | -------- | ------------------------------------------- | ---------- |
| id         | uint     | PRIMARY KEY                                 | 主键       |
| created_at | datetime |                                             | 创建时间   |
| updated_at | datetime |                                             | 更新时间   |
| deleted_at | integer  | UNIQUE INDEX (`idx_name`: deleted_at, name) | 软删除标记 |
| name       | string   | NOT NULL, UNIQUE INDEX (`idx_name`)         | 分组名称   |

**初始数据**: 系统启动时自动创建 ID=1 的 "Default" 分组。

---

### feeds

RSS/Atom 订阅源表。

| 字段                 | 类型     | 约束                                        | 说明                   |
| -------------------- | -------- | ------------------------------------------- | ---------------------- |
| id                   | uint     | PRIMARY KEY                                 | 主键                   |
| created_at           | datetime |                                             | 创建时间               |
| updated_at           | datetime |                                             | 更新时间               |
| deleted_at           | integer  | UNIQUE INDEX (`idx_link`: deleted_at, link) | 软删除标记             |
| name                 | string   | NOT NULL                                    | 订阅源名称             |
| link                 | string   | NOT NULL, UNIQUE INDEX (`idx_link`)         | 订阅源 URL             |
| last_build           | datetime | NULLABLE                                    | 内容最后更新时间       |
| failure              | string   | DEFAULT ''                                  | 最近一次拉取的错误信息 |
| consecutive_failures | uint     | DEFAULT 0                                   | 连续失败次数           |
| suspended            | bool     | DEFAULT false                               | 是否暂停拉取           |
| req_proxy            | string   | NULLABLE                                    | 请求代理设置           |
| group_id             | uint     | FOREIGN KEY → groups.id                     | 所属分组 ID            |

**虚拟字段** (不存储在数据库):

- `unread_count`: 未读条目数量

---

### items

订阅条目表，存储 Feed 中的文章/条目。

| 字段       | 类型     | 约束                                                 | 说明          |
| ---------- | -------- | ---------------------------------------------------- | ------------- |
| id         | uint     | PRIMARY KEY                                          | 主键          |
| created_at | datetime |                                                      | 创建时间      |
| updated_at | datetime |                                                      | 更新时间      |
| deleted_at | integer  | UNIQUE INDEX (`idx_guid`: deleted_at, guid, feed_id) | 软删除标记    |
| title      | string   | NULLABLE                                             | 条目标题      |
| guid       | string   | UNIQUE INDEX (`idx_guid`)                            | 条目唯一标识  |
| link       | string   | NULLABLE                                             | 条目链接      |
| content    | string   | NULLABLE                                             | 条目内容      |
| pub_date   | datetime | NULLABLE                                             | 发布时间      |
| unread     | bool     | DEFAULT true, INDEX                                  | 是否未读      |
| bookmark   | bool     | DEFAULT false, INDEX                                 | 是否收藏      |
| feed_id    | uint     | FOREIGN KEY → feeds.id, UNIQUE INDEX (`idx_guid`)    | 所属订阅源 ID |

---

## 表关系

```
┌─────────┐       ┌─────────┐       ┌─────────┐
│ groups  │ 1───N │  feeds  │ 1───N │  items  │
└─────────┘       └─────────┘       └─────────┘
```

- **groups → feeds**: 一对多关系，一个分组可包含多个订阅源
- **feeds → items**: 一对多关系，一个订阅源包含多个条目

---

## 索引说明

### groups 表

- `idx_name`: (deleted_at, name) - 复合唯一索引，支持软删除下的名称唯一性

### feeds 表

- `idx_link`: (deleted_at, link) - 复合唯一索引，支持软删除下的 URL 唯一性

### items 表

- `idx_guid`: (deleted_at, guid, feed_id) - 复合唯一索引，确保同一 Feed 下 GUID 唯一
- `unread`: 单列索引，加速未读条目查询
- `bookmark`: 单列索引，加速收藏条目查询

---

## 软删除机制

所有表都使用软删除：

- `deleted_at` 字段类型为 integer（Unix 时间戳）
- 删除时设置为当前时间戳，而非物理删除
- 唯一索引包含 `deleted_at` 字段，允许"已删除"的记录与新记录共存同名/同值

---

## 迁移历史

### v0.8.7 之后

- 为 `feeds.link` 添加唯一索引
- 迁移时自动删除重复的 Feed（保留 ID 最小的记录）
