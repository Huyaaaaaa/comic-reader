import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listFavorites, type FavoriteItem } from '../lib/api'

export default function FavoritesPage() {
  const [items, setItems] = useState<FavoriteItem[]>([])

  useEffect(() => {
    void listFavorites().then((result) => setItems(result.favorites))
  }, [])

  return (
    <section className="page">
      <header className="page-header">
        <div>
          <h1>收藏</h1>
          <p>本地收藏与离线状态会在这里集中展示。</p>
        </div>
      </header>
      <div className="stack-list">
        {items.length === 0 ? <div className="panel">还没有收藏任何漫画。</div> : null}
        {items.map((item) => (
          <article key={item.comic_id} className="panel list-row">
            <div>
              <strong>{item.comic.title}</strong>
              <p>{item.comic.subtitle || item.comic.category_name}</p>
              <small>{item.offline_ready ? '已离线可读' : '尚未完全离线'}</small>
            </div>
            <div className="action-row">
              <Link className="primary-link" to={`/comics/${item.comic_id}`}>详情</Link>
              <Link className="primary-link" to={`/reader/${item.comic_id}`}>继续阅读</Link>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}
