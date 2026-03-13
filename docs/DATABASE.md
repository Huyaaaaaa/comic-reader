# 数据库设计文档

## 概述

本项目使用 SQLite 作为数据库，通过 GORM 进行 ORM 映射。

## 表结构

### 1. comics - 漫画主表

存储漫画的基本信息。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，漫画ID |
| title | TEXT | 标题 |
| subtitle | TEXT | 副标题 |
| author | TEXT | 作者 |
| author_id | INTEGER | 作者ID |
| cover_url | TEXT | 封面URL |
| cover_base64 | TEXT | 封面Base64编码 |
| rating | REAL | 评分 |
| rating_count | INTEGER | 评分人数 |
| favorites | INTEGER | 收藏数 |
| category_id | INTEGER | 分类ID |
| category_name | TEXT | 分类名称 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |
| has_cover_cached | BOOLEAN | 是否已缓存封面 |
| cover_cached_at | TIMESTAMP | 封面缓存时间 |
| cover_size | INTEGER | 封面大小(字节) |

**索引**:
- `idx_comics_author_id` ON `author_id`

### 2. comic_images - 漫画图片

存储漫画的图片列表。

| 字段 | 类型 | 说明 |
|------|------|------|
| comic_id | INTEGER | 漫画ID (主键) |
| sort | INTEGER | 排序号 (主键) |
| filename | TEXT | 文件名 |
| extension | TEXT | 扩展名 |
| url | TEXT | 图片URL |
| local_path | TEXT | 本地路径 |

**主键**: (comic_id, sort)

### 3. tags - 标签表

存储所有标签。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，标签ID |
| name | TEXT | 标签名称 |

### 4. comic_tags - 漫画标签关联

多对多关联表。

| 字段 | 类型 | 说明 |
|------|------|------|
| comic_id | INTEGER | 漫画ID (主键) |
| tag_id | INTEGER | 标签ID (主键) |
**主键**: (comic_id, tag_id)

**索引**:
- `idx_comic_tags_comic_id` ON `comic_id`

### 5. comic_authors - 漫画作者关联

支持多作者。

| 字段 | 类型 | 说明 |
|------|------|------|
| comic_id | INTEGER | 漫画ID (主键) |
| author_id | INTEGER | 作者ID (主键) |
| author_name | TEXT | 作者名称 (主键) |
| position | INTEGER | 位置顺序 |

**主键**: (comic_id, author_id, author_name)

**索引**:
- `idx_comic_authors_comic_id` ON `comic_id`
- `idx_comic_authors_author_id` ON `author_id`

### 6. comic_metadata_state - 漫画元数据状态

跟踪漫画的缓存状态。

| 字段 | 类型 | 说明 |
|------|------|------|
| comic_id | INTEGER | 主键，漫画ID |
| list_cached | BOOLEAN | 列表是否已缓存 |
| detail_cached | BOOLEAN | 详情是否已缓存 |
| author_ready | BOOLEAN | 作者信息是否就绪 |
| tags_ready | BOOLEAN | 标签信息是否就绪 |
| last_list_cached_at | TIMESTAMP | 列表缓存时间 |
| last_detail_cached_at | TIMESTAMP | 详情缓存时间 |
| updated_at | TIMESTAMP | 更新时间 |

**索引**:
- `idx_metadata_state_author_ready` ON `author_ready`
- `idx_metadata_state_tags_ready` ON `tags_ready`

### 7. reading_history - 阅读历史

| 字段 | 类型 | 说明 |
|------|------|
| comic_id | INTEGER | 主键，漫画ID |
| title | TEXT | 标题 |
| cover_url | TEXT | 封面URL |
| last_read_at | TIMESTAMP | 最后阅读时间 |

### 8. comic_favorites - 收藏记录

| 字段 | 类型 | 说明 |
|------|------|------|
| comic_id | INTEGER | 主键，漫画ID |
| title | TEXT | 标题 |
| cover_url | TEXT | 封面URL |
| favorited_at | TIMESTAMP | 收藏时间 |
| updated_at | TIMESTAMP | 更新时间 |

**索引**:
- `idx_comic_favorites_favorited_at` ON `favorited_at`

### 9. comic_list_cache - 列表缓存

缓存列表页数据。

| 字段 | 类型 | 说明 |
|------|------|------|
| comic_id | INTEGER | 漫画ID (主键) |
| page | INTEGER | 页码 (主键) |
| position | INTEGER | 位置 |
| title | TEXT | 标题 |
| author | TEXT | 作者 |
| author_id | INTEGER | 作者ID |
| cover_url | TEXT | 封面URL |
| rating | REAL | 评分 |
| rating_count | INTEGER | 评分人数 |
| favorites | INTEGER | 收藏数 |
| cached_at | TIMESTAMP | 缓存时间 |

**主键**: (comic_id, page)

**索引**:
- `idx_comic_list_cache_page_position` ON `(page, position)`

### 10. download_tasks - 下载任务

| 字段 | 类型 | 说明 |
|------|------|------|
| comic_id | INTEGER | 主键，漫画ID |
| title | TEXT | 标题 |
| total | INTEGER | 总数 |
| current | INTEGER | 当前进度 |
| status | TEXT | 状态 (pending/downloading/completed/failed) |
| error | TEXT | 错误信息 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### 11. cover_cache_queue - 封面缓存队列

| 字段 | 类型 | 说明 |
|------|------|------|
| comic_id | INTEGER | 主键，漫画ID |
| title | TEXT | 标题 |
| cover_url | TEXT | 封面URL |
| priority | INTEGER | 优先级 |
| status | TEXT | 状态 (pending/processing/completed/failed) |
| retry_count | INTEGER | 重试次数 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### 12. system_metadata - 系统元数据

存储系统级别的配置和状态。

| 字段 | 类型 | 说明 |
|------|------|------|
| key | TEXT | 主键，键名 |
| value | TEXT | 值 |
| updated_at | TIMESTAMP | 更新时间 |

## 数据迁移

使用 GORM 的 AutoMigrate 功能自动创建和更新表结构：

```go
db.AutoMigrate(
    &model.Comic{},
    &model.ComicImage{},
    &model.Tag{},
    &model.ComicTag{},
    &model.ComicAuthor{},
    &model.ComicMetadataState{},
    &model.ReadingHistory{},
    &model.ComicFavorite{},
    &model.ComicListCache{},
    &model.DownloadTask{},
    &model.CoverCacheQueue{},
    &model.SystemMetadata{},
)
```

## 查询优化建议

1. **使用索引**: 为常用查询字段添加索引
2. **分页查询**: 使用 LIMIT 和 OFFSET
3. **预加载关联**: 使用 GORM 的 Preload
4. **批量操作**: 使用批量插入和更新
5. **连接池**: 配置合适的连接池大小

## 备份和恢复

### 备份

```bash
# 备份数据库
cp data/comics.db data/comics_backup_$(date +%Y%m%d_%H%M%S).db
```

### 恢复

```bash
# 恢复数据库
cp data/comics_backup_20240101_120000.db data/comics.db
```

## 性能监控

使用 GORM 的日志功能监控慢查询：

```go
db.Logger = logger.Default.LogMode(logger.Info)
```

## 数据清理

定期清理过期数据：

1. 清理旧的列表缓存
2. 清理失败的下载任务
3. 清理过期的封面缓存队列
