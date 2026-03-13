# 漫画阅读器修复计划

## 测试发现的问题

### 1. 设置等页面暂时无法访问
- **问题**: 设置页面显示"暂时无法访问"
- **原因**: 页面未开发
- **修复**: 实现基础设置页面，包含反爬虫模式、每页数量等配置

### 2. 所有漫画页面分页问题
- **问题**: 首次访问显示几千页，点进详情页返回后只有1页
- **原因**: 状态管理问题，totalPages 没有正确保存
- **修复**: 修复 `frontend/src/store/comicStore.js` 的状态持久化

### 3. 导航名称错误
- **问题**: "所有漫画"的导航名不对
- **原因**: 可能是翻译或命名问题
- **修复**: 检查并修正 `frontend/src/App.jsx` 中的导航配置

### 4. 排版问题
- **问题**: 漫画列表排版有问题，应该一行4-5个
- **原因**: Grid 布局配置不当
- **修复**: 调整 `frontend/src/pages/ComicList.jsx` 的 Grid 列数
  - 使用响应式布局：`{ xs: 2, sm: 3, md: 4, lg: 5, xl: 6 }`

### 5. 每页数量设置缺失
- **问题**: 看不到调整每页数量的设置
- **原因**: UI 未实现
- **修复**: 在 `ComicList.jsx` 添加 PageSize 选择器
  - 选项：20, 50, 100, 200

### 6. 详情页标签和作者不可点击
- **问题**: 标签和作者无法点击进行筛选
- **原因**: 未绑定点击事件
- **修复**: 在 `frontend/src/pages/ComicDetail.jsx` 添加点击跳转
  - 标签点击 → 跳转到筛选页面 `/comics?tag=xxx`
  - 作者点击 → 跳转到筛选页面 `/comics?author=xxx`

### 7. 阅读页默认模式错误
- **问题**: 默认是单页流，应该是瀑布长图流
- **原因**: 阅读模式配置错误
- **修复**: 修改 `frontend/src/pages/Reader.jsx`
  - 默认模式改为瀑布流（waterfall）
  - 添加模式切换按钮：瀑布流 / 单页流

### 8. 单页流图片比例被压缩
- **问题**: 图片比例不正确
- **原因**: CSS 使用了固定高度或 `object-fit: cover`
- **修复**: 改为 `object-fit: contain` 保持原始比例

## 修复优先级

1. **高优先级**（影响核心功能）:
   - 问题2: 分页状态丢失
   - 问题4: 排版问题
   - 问题7: 阅读模式错误
   - 问题8: 图片比例

2. **中优先级**（影响用户体验）:
   - 问题5: 每页数量设置
   - 问题6: 标签作者点击

3. **低优先级**（功能完善）:
   - 问题1: 设置页面
   - 问题3: 导航名称

## 文件修改清单

### 前端文件
1. `frontend/src/store/comicStore.js` - 修复状态管理
2. `frontend/src/pages/ComicList.jsx` - 修复排版和添加每页数量选择
3. `frontend/src/pages/ComicDetail.jsx` - 添加标签作者点击
4. `frontend/src/pages/Reader.jsx` - 修改默认模式和图片比例
5. `frontend/src/pages/Settings.jsx` - 实现设置页面
6. `frontend/src/App.jsx` - 修正导航名称

### 后端文件
- 无需修改（后端功能正常）

## 详细修复方案

### 修复2: 分页状态丢失

**文件**: `frontend/src/store/comicStore.js`

```javascript
// 添加 localStorage 持久化
const useComicStore = create(
  persist(
    (set, get) => ({
      comics: [],
      currentPage: 1,
      totalPages: 1,
      pageSize: 100,
      loading: false,

      setComics: (comics, currentPage, totalPages) => {
        set({ comics, currentPage, totalPages })
    },

      setPageSize: (pageSize) => {
        set({ pageSize, currentPage: 1 })
      }
    }),
    {
      name: 'comic-store',
      partialPersist: (state) => ({
      pageSize: state.pageSize,
        currentPage: state.currentPage,
        totalPages: state.totalPages
      })
    }
  )
)
```

### 修复4: 排版问题

**文件**: `frontend/src/pages/ComicList.jsx`

```jsx
// 修改 Grid 配置
<Row gutter={[16, 16]}>
  {comics.map(comic => (
    <Col xs={12} sm={8} md={6} lg={4} xl={4} key={comic.id}>
      <Card
        hoverable
        cover={<img alt={comic.title} src={comic.cover} />}
        onClick={() => navigate(`/comic/${comic.id}`)}
      >
        <Card.Meta title={comic.title} />
      </Card>
    </Col>
  ))}
</Row>
```

### 修复5: 每页数量设置

**文件**: `frontend/src/pages/ComicList.jsx`

```jsx
// 添加 PageSize 选择器
<Space style={{ marginBottom: 16 }}>
  <span>每页显示：</span>
  <Select
    value={pageSize}
    onChange={(value) => {
      setPageSize(value)
      fetchComics(1, value)
    }}
    options={[
      { label: '20条', value: 20 },
   { label: '50条', value: 50 },
      { label: '100条', value: 100 },
      { label: '200条', value: 200 }
    ]}
  />
</Space>
```

### 修复6: 标签作者点击

**文件**: `frontend/src/pages/ComicDetail.jsx`

```jsx
// 标签改为可点击
<Space wrap>
  {comic.tags.map(tag => (
    <Tag
      key={tag.id}
      color="blue"
      style={{ cursor: 'pointer' }}
      onClick={() => navigate(`/comics?tag=${tag.name}`)}
    >
      {tag.name}
    </Tag>
  ))}
</Space>

// 作者改为可点击
<Space wrap>
  {comic.authors.map(author => (
    <Tag
      key={author.id}
      color="green"
      style={{ cursor: 'pointer' }}
      onClick={() => navigate(`/comics?author=${author.name}`)}
    >
      {author.name}
    </Tag>
  ))}
</Space>
```

### 修复7: 阅读模式

**文件**: `frontend/src/pages/Reader.jsx`

```jsx
const [viewMode, setViewMode] = useState('waterfall') // 默认瀑布流

// 瀑布流模式
{viewMode === 'waterfall' && (
  <div style={{ padding: '20px' }}>
    {images.map((img, index) => (
      <img
        key={index}
        src={img.url}
        alt={`Page ${index + 1}`}
     style={{
          width: '100%',
          display: 'block',
          marginBottom: '10px'
        }}
      />
    ))}
  </div>
)}

// 单页流模式
{viewMode === 'single' && (
  <div style={{
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    height: '100vh'
  }}>
    <img
      src={images[currentPage]?.url}
   alt={`Page ${currentPage + 1}`}
      style={{
        maxWidth: '100%',
     maxHeight: '100%',
        objectFit: 'contain' // 保持比例
      }}
    />
  </div>
)}

// 模式切换按钮
<Button onClick={() => setViewMode(viewMode === 'waterfall' ? 'single' : 'waterfall')}>
  {viewMode === 'waterfall' ? '单页模式' : '瀑布流模式'}
</Button>
```

### 修复8: 图片比例

**文件**: `frontend/src/pages/Reader.jsx`

```css
/* 确保使用 contain 而不是 cover */
img {
  object-fit: contain;
  max-width: 100%;
  max-height: 100%;
}
```

## 测试检查清单

修复完成后需要测试：

- [ ] 分页在详情页返回后是否保持
- [ ] 列表页一行显示4-5个漫画
- [ ] 可以调整每页显示数量（20/50/100/200）
- [ ] 标签点击跳转到筛选页面
- [ ] 作者点击跳转到筛选页面
- [ ] 阅读页默认是瀑布流模式
- [ ] 可以切换单页/瀑布流模式
- [ ] 单页模式图片比例正确
- [ ] 设置页面可以访问
- [ ] 导航名称正确

## 预计修改时间

- 总计约 8 个文件修改
- 预计 30-45 分钟完成所有修复
