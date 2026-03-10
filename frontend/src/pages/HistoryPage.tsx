import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import ComicCover from '../components/ComicCover'
import { listHistory, type HistoryItem } from '../lib/api'

export default function HistoryPage() {
  const [items, setItems] = useState<HistoryItem[]>([])

  useEffect(() => {
    void listHistory().then((result) => setItems(result.history))
  }, [])

  return (
    <section className="page">
      <header className="page-header">
        <div>
          <h1>阅读历史</h1>
          <p>长图流阅读位置会记录为图序号和图内滚动比例。</p>
        </div>
      </header>
      <div className="stack-list">
        {items.length === 0 ? <div className="panel">还没有阅读历史。</div> : null}
        {items.map((item) => (
          <article key={item.comic_id} className="panel list-row">
            <ComicCover
              comicId={item.comic_id}
              title={item.comic.title}
              coverURL={item.comic.cover_url}
              coverLocalRelPath={item.comic.cover_local_rel_path}
              className="list-cover media-frame"
            />
            <div className="list-copy">
              <strong>{item.comic.title}</strong>
              <p>{item.comic.subtitle || item.comic.category_name}</p>
              <small>上次位置：第 {item.locator.sort + 1} 张 · 比例 {Math.round((item.locator.offset_ratio ?? 0) * 100)}%</small>
            </div>
            <div className="action-row">
              <Link className="primary-link" to={`/reader/${item.comic_id}`}>继续阅读</Link>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}
