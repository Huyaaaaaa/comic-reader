import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { addFavorite, buildImageProxyURL, getComicImages, listFavorites, postHistory, removeFavorite, syncComicDetail, type ComicImage } from '../lib/api'

export default function ReaderPage() {
  const params = useParams()
  const [images, setImages] = useState<ComicImage[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [syncMessage, setSyncMessage] = useState('')
  const [currentSort, setCurrentSort] = useState(0)
  const [isFavorite, setIsFavorite] = useState(false)
  const itemRefs = useRef<Array<HTMLElement | null>>([])
  const attemptedSyncRef = useRef(false)

  const comicId = useMemo(() => Number(params.id), [params.id])

  const loadReader = async (id: number) => {
    const [result, favorites] = await Promise.all([getComicImages(id), listFavorites(id)])
    setImages(result.images)
    setIsFavorite(favorites.total > 0)
    return result.images
  }

  useEffect(() => {
    attemptedSyncRef.current = false
    if (!comicId) {
      setError('无效漫画 ID')
      setLoading(false)
      return
    }
    setLoading(true)
    setError('')
    setSyncMessage('')
    void loadReader(comicId)
      .then(async (loadedImages) => {
        if (loadedImages.length === 0 && !attemptedSyncRef.current) {
          attemptedSyncRef.current = true
          setSyncMessage('本地还没有正文索引，正在从源站补全…')
          await syncComicDetail(comicId)
          const refreshed = await loadReader(comicId)
          if (refreshed.length > 0) {
            setSyncMessage('正文索引已补全，图片会通过本地代理逐张缓存。')
          }
        }
      })
      .catch((reason) => setError(reason instanceof Error ? reason.message : '加载失败'))
      .finally(() => setLoading(false))
  }, [comicId])

  useEffect(() => {
    if (!images.length || !comicId) return
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries.filter((entry) => entry.isIntersecting)
        if (!visible.length) return
        const sort = Number((visible[0].target as HTMLElement).dataset.sort ?? '0')
        setCurrentSort(sort)
      },
      { threshold: 0.6 },
    )
    itemRefs.current.forEach((node) => node && observer.observe(node))
    return () => observer.disconnect()
  }, [images, comicId])

  useEffect(() => {
    if (!comicId || !images.length) return
    const timer = window.setTimeout(() => {
      void postHistory(comicId, { mode: 'long_scroll', sort: currentSort, offset_ratio: 0.5 })
    }, 300)
    return () => window.clearTimeout(timer)
  }, [comicId, currentSort, images.length])

  const toggleFavorite = async () => {
    if (!comicId) return
    if (isFavorite) {
      await removeFavorite(comicId)
      setIsFavorite(false)
      return
    }
    await addFavorite(comicId, false)
    setIsFavorite(true)
  }

  return (
    <section className="page reader-page">
      <header className="page-header sticky-head">
        <div>
          <h1>阅读页</h1>
          <p>当前位置：第 {currentSort + 1} 张 / 共 {images.length || 0} 张</p>
        </div>
        <div className="action-row">
          <button className="primary-button" onClick={() => void toggleFavorite()}>{isFavorite ? '取消收藏' : '加入收藏'}</button>
          <Link className="primary-link" to={comicId ? `/comics/${comicId}` : '/comics'}>返回详情</Link>
        </div>
      </header>

      {loading ? <div className="panel">正在准备阅读内容…</div> : null}
      {syncMessage ? <div className="panel"><p className="hint-text">{syncMessage}</p></div> : null}
      {error ? <div className="panel error-text">{error}</div> : null}
      {!loading && !error && images.length === 0 ? <div className="panel">当前还没有正文图片索引，后续同步或导入后会在这里展示。</div> : null}

      <div className="reader-stack">
        {images.map((image, index) => (
          <article
            key={`${image.comic_id}-${image.sort}`}
            ref={(node) => {
              itemRefs.current[index] = node
            }}
            data-sort={image.sort}
            className="reader-image-card panel"
          >
            <div className="reader-image-frame media-frame">
              <img loading={index < 3 ? 'eager' : 'lazy'} src={buildImageProxyURL(image)} alt={`第 ${image.sort + 1} 张`} />
            </div>
            <div>
              <strong>{image.cached ? '已缓存' : '通过代理缓存中'}</strong>
              <p>第 {image.sort + 1} 张 · {image.extension.toUpperCase()}</p>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}
