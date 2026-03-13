# API 接口文档

## 基础信息

- **Base URL**: `http://localhost:8080/api`
- **Content-Type**: `application/json`
- **字符编码**: UTF-8

## 通用响应格式

### 成功响应

```json
{
  "data": {},
  "message": "success"
}
```

### 错误响应

```json
{
  "error": "错误信息"
}
```

## 接口列表

### 1. 健康检查

**GET** `/health`

检查服务器状态。

**响应示例**:
```json
{
  "status": "ok"
}
```

---

### 2. 获取漫画列表

**GET** `/comics`

获取漫画列表，支持分页和缓存。

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|----|------|------|------|
| page | int | 否 | 页码，默认1 |
| use_cache | bool | 否 | 是否使用缓存，默认true |

**响应示例**:
```json
{
  "items": [
    {
      "id": 12345,
      "title": "漫画标题",
      "cover_url": "https://example.com/cover.jpg",
      "cover_base64": "base64编码的封面图片",
      "rating": 9.2,
      "rating_count": 100,
      "favorites": 500,
      "author": "作者名",
      "author_id": 123,
      "is_cached": true,
      "local_saved": 0,
      "local_total": 0
    }
  ],
  "current_page": 1,
  "total_pages": 100,
  "from_cache": true
}
```

---

### 3. 获取漫画详情

**GET** `/comics/:id`

获取指定漫画的详细信息。

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 漫画ID |

**响应示例**:
```json
{
  "id": 12345,
  "title": "漫画标题",
  "subtitle": "副标题",
  "author": "作者名",
  "author_id": 123,
  "authors": [
    {
      "author_id": 123,
      "author_name": "作者名"
    }
  ],
  "cover_url": "https://example.com/cover.jpg",
  "rating": 9.2,
  "rating_count": 100,
  "favorites": 500,
  "category_id": 1,
  "category_name": "同人志",
  "tags": [
    {
      "tag_id": 1,
      "tag_name": "标签名"
    }
  ],
  "created_at": "2024-01-01 12:00:00",
  "updated_at": "2024-01-02 12:00:00",
  "reader_url": "/readOnline2.php?ID=12345",
  "is_favorited": false
}
```

---

### 4. 获取阅读器图片列表

**GET** `/comics/:id/images`

获取漫画的所有图片URL。

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 漫画ID |

**响应示例**:
```json
{
  "images": [
    {
      "sort": 1,
      "comic_id": 12345,
      "filename": "001",
      "extension": "jpg",
      "url": "https://example.com/images/001_w900.jpg",
      "local_path": ""
    }
  ]
}
```

---

### 5. 搜索漫画

**GET** `/comics/search`

搜索漫画，支持本地和在线搜索。

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 是 | 搜索关键词 |
| page | int | 否 | 页码，默认1 |
| mode | string | 否 | 搜索模式 (local/online)，默认local |

**响应示例**:
```json
{
  "items": [...],
  "current_page": 1,
  "total_pages": 10,
  "from_cache": true
}
```

---

### 6. 筛选漫画

**GET** `/comics/filter`

根据标签、分类、作者筛选漫画。

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| tag_id | int | 否 | 标签ID |
| category_id | int | 否 | 分类ID |
| author_id | int | 否 | 作者ID |
| author | string | 否 | 作者名 |
| page | int | 否 | 页码，默认1 |

**响应示例**:
```json
{
  "items": [...],
  "current_page": 1,
  "total_pages": 10,
  "from_cache": false
}
```

---

### 7. 添加阅读历史

**POST** `/comics/:id/history`

添加或更新阅读历史记录。

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|----|------|------|
| id | int | 是 | 漫画ID |

**请求体**:
```json
{
  "title": "漫画标题",
  "cover_url": "https://example.com/cover.jpg"
}
```

**响应示例**:
```json
{
  "message": "添加成功"
}
```

---

### 8. 切换收藏状态

**POST** `/comics/:id/favorite`

添加或取消收藏。

**路径参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 漫画ID |

**请求体**:
```json
{
  "title": "漫画标题",
  "cover_url": "https://example.com/cover.jpg"
}
```

**响应示例**:
```json
{
  "is_favorited": true,
  "message": "已收藏"
}
```

---

### 9. 获取仪表盘统计

**GET** `/dashboard/stats`

获取系统统计数据。

**响应示例**:
```json
{
  "total_comics": 10000,
  "cached_comics": 5000,
  "cover_cached": 5000,
  "total_tags": 500,
  "favorites_count": 100,
  "history_count": 200,
  "downloading_count": 2,
  "pending_downloads": 5
}
```

---

## 错误码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
| 501 | 功能未实现 |

## 调用示例

### cURL

```bash
# 获取漫画列表
curl "http://localhost:8080/api/comics?page=1&use_cache=true"

# 获取漫画详情
curl "http://localhost:8080/api/comics/12345"

# 搜索漫画
curl "http://localhost:8080/api/comics/search?keyword=test&page=1&mode=local"

# 添加收藏
curl -X POST "http://localhost:8080/api/comics/12345/favorite" \
  -H "Content-Type: application/json" \
  -d '{"title":"漫画标题","cover_url":"https://example.com/cover.jpg"}'
```

### JavaScript (Axios)

```javascript
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api',
});

// 获取漫画列表
const getComics = async (page = 1) => {
  const response = await api.get('/comics', {
    params: { page, use_cache: true }
  });
  return response.data;
};

// 获取漫画详情
const getComicDetail = async (id) => {
  const response = await api.get(`/comics/${id}`);
  return response.data;
};

// 切换收藏
const toggleFavorite = async (id, title, coverUrl) => {
  const response = await api.post(`/comics/${id}/favorite`, {
    title,
    cover_url: coverUrl
  });
  return response.data;
};
```

## 注意事项

1. 所有时间戳使用 ISO 8601 格式
2. 图片URL可能需要特定的 Referer 头
3. 部分接口可能触发爬虫限流，请注意请求频率
4. Base64 编码的封面图片可能较大，建议按需加载
