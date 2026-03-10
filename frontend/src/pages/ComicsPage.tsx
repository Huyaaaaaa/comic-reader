import { FormEvent, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { addSearchHistory, clearSearchHistory, listComics, listSearchHistory, searchComics, type ComicListItem, type SearchHistoryItem } from '../lib/api'

export default function ComicsPage() {
  const [items, setItems] = useState<ComicListItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [keyword, setKeyword] = useState('')
  const [history, setHistory] = useState<SearchHistoryItem[]>([])
  const [message, setMessage] = useState('')

  const loadDefault = async () => {
    setLoading(true)
    setError('')
    try {
      const [result, historyItems] = await Promise.all([listComics(), listSearchHistory()])
      setItems(result.comics)
      setHistory(historyItems)
      setMessage('')
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadDefault()
  }, [])

  const runSearch = async (term: string) => {
    const trimmed = term.trim()
    setLoading(true)
    setError('')
    try {
      if (!trimmed) {
        await loadDefault()
        return
      }
      const result = await searchComics(trimmed)
      setItems(result.comics)
      setMessage('本地结果已回显，后续可接远端补全联动。')
      await addSearchHistory(trimmed)
      setHistory(await listSearchHistory())
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '搜索失败')
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = (event: FormEvent) => {
    event.preventDefault()
    void runSearch(keyword)
  }

  const handleClearHistory = async () => {
    await clearSearchHistory()
    setHistory([])
  }

  return (
    <section className="page">
      <header className="page-header">
        <div>
          <h1>所有漫画</h1>
          <p>当前先走本地数据链路，搜索历史和本地搜索已经接上。</p>
        </div>
      </header>

      <section className="panel search-panel">
        <form className="search-form" onSubmit={handleSubmit}>
          <input value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder="搜索标题、作者、标签" />
          <button className="primary-button" type="submit">搜索</button>
          <button type="button" onClick={() => { setKeyword(''); void loadDefault() }}>重置</button>
        </form>
        <div className="history-strip">
          {history.map((item) => (
            <button key={`${item.keyword}-${item.searched_at}`} className="history-chip" onClick={() => { setKeyword(item.keyword); void runSearch(item.keyword) }}>
              {item.keyword}
            </button>
          ))}
          {history.length ? <button className="history-chip muted" onClick={() => void handleClearHistory()}>清空历史</button> : null}
        </div>
        {message ? <p className="hint-text">{message}</p> : null}
      </section>

      {loading ? <div className="panel">正在加载漫画列表…</div> : null}
      {error ? <div className="panel error-text">{error}</div> : null}
      {!loading && !error && items.length === 0 ? <div className="panel">当前没有匹配结果。</div> : null}

      <div className="comic-grid">
        {items.map((item) => (
          <Link key={item.id} to={`/comics/${item.id}`} className="comic-card">
            <div className="comic-cover">{item.title.slice(0, 1) || '漫'}</div>
            <div className="comic-meta">
              <strong>{item.title}</strong>
              <p>{item.subtitle || item.category_name || '暂无副标题'}</p>
              <small>缓存：L{item.cache_state.meta_level} · 图片 {item.cache_state.images_local}/{item.cache_state.images_total}</small>
            </div>
          </Link>
        ))}
      </div>
    </section>
  )
}
