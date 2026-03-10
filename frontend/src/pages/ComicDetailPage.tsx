import { useEffect, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { addFavorite, getComicDetail, listFavorites, removeFavorite, syncComicDetail, type ComicDetail } from '../lib/api'

export default function ComicDetailPage() {
  const params = useParams()
  const [comic, setComic] = useState<ComicDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [isFavorite, setIsFavorite] = useState(false)
  const [syncing, setSyncing] = useState(false)
  const [syncMessage, setSyncMessage] = useState('')
  const attemptedSyncRef = useRef(false)

  const loadComic = async (id: number) => {
    const [detail, favorites] = await Promise.all([getComicDetail(id), listFavorites(id)])
    setComic(detail)
    setIsFavorite(favorites.total > 0)
  }

  useEffect(() => {
    const id = Number(params.id)
    attemptedSyncRef.current = false
    if (!id) {
      setError('无效漫画 ID')
      setLoading(false)
      return
    }
    setLoading(true)
    setError('')
    setSyncMessage('')
    void loadComic(id)
      .catch((reason) => setError(reason instanceof Error ? reason.message : '加载失败'))
      .finally(() => setLoading(false))
  }, [params.id])

  useEffect(() => {
    if (!comic || syncing || attemptedSyncRef.current) {
      return
    }
    const needsSync = comic.cache_state.meta_level < 2 || comic.images_total === 0 || comic.authors.length === 0 || comic.tags.length === 0
    if (!needsSync) {
      return
    }
    attemptedSyncRef.current = true
    setSyncing(true)
    setSyncMessage('本地详情已显示，正在从源站补全作者、标签和正文索引…')
    void syncComicDetail(comic.id)
      .then(async () => {
        await loadComic(comic.id)
        setSyncMessage('详情补全完成。')
      })
      .catch((reason) => {
        setSyncMessage(reason instanceof Error ? reason.message : '详情补全失败')
      })
      .finally(() => setSyncing(false))
  }, [comic, syncing])

  const toggleFavorite = async () => {
    if (!comic) return
    if (isFavorite) {
      await removeFavorite(comic.id)
      setIsFavorite(false)
      return
    }
    await addFavorite(comic.id, false)
    setIsFavorite(true)
  }

  if (loading) {
    return <section className="page"><div className="panel">正在进入详情页框架…</div></section>
  }

  if (error || !comic) {
    return <section className="page"><div className="panel error-text">{error || '未找到漫画'}</div></section>
  }

  return (
    <section className="page">
      <header className="page-header">
        <div>
          <h1>{comic.title}</h1>
          <p>{comic.subtitle || comic.category_name || '暂无副标题'}</p>
        </div>
        <div className="action-row">
          <button className="primary-button" onClick={() => void toggleFavorite()}>{isFavorite ? '取消收藏' : '加入收藏'}</button>
          <Link className="primary-link" to={`/reader/${comic.id}`}>开始阅读</Link>
        </div>
      </header>

      {syncMessage ? <section className="panel"><p className="hint-text">{syncing ? `${syncMessage}` : syncMessage}</p></section> : null}

      <section className="detail-hero panel">
        <div className="detail-cover media-frame">
          {comic.cover_url ? <img src={comic.cover_url} alt={comic.title} /> : <span>{comic.title.slice(0, 1) || '漫'}</span>}
        </div>
        <div className="detail-info">
          <div className="detail-badges">
            <span>评分 {comic.rating || 0}</span>
            <span>收藏 {comic.favorites}</span>
            <span>{comic.category_name || '未知分类'}</span>
            <span>{isFavorite ? '已收藏' : '未收藏'}</span>
          </div>
          <div>
            <h2>作者</h2>
            <p>{comic.authors.length ? comic.authors.map((item) => item.name).join(' / ') : '暂无作者信息'}</p>
          </div>
          <div>
            <h2>标签</h2>
            <p>{comic.tags.length ? comic.tags.map((item) => item.name).join(' · ') : '暂无标签信息'}</p>
          </div>
          <div>
            <h2>缓存状态</h2>
            <p>L{comic.cache_state.meta_level} · 封面 {comic.cache_state.cover_ready ? '已缓存' : '未缓存'} · 图片 {comic.cache_state.images_local}/{comic.cache_state.images_total}</p>
          </div>
        </div>
      </section>
    </section>
  )
}
