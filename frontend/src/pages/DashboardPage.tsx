import { useEffect, useMemo, useState } from 'react'
import StatCard from '../components/StatCard'
import { getSettings, listComics, listSources, streamEvents, syncHead } from '../lib/api'

export default function DashboardPage() {
  const [settingsCount, setSettingsCount] = useState(0)
  const [sourceCount, setSourceCount] = useState(0)
  const [comicCount, setComicCount] = useState(0)
  const [logLines, setLogLines] = useState<string[]>([])
  const [syncing, setSyncing] = useState(false)
  const [message, setMessage] = useState('')

  const refreshSummary = async () => {
    const [settings, sources, comics] = await Promise.all([getSettings(), listSources(), listComics(1, 1)])
    setSettingsCount(Object.keys(settings).length)
    setSourceCount(sources.length)
    setComicCount(comics.total)
  }

  useEffect(() => {
    void refreshSummary()

    const source = streamEvents((event) => {
      const line = `${new Date().toLocaleTimeString()} ${event.type} ${event.data}`
      setLogLines((current) => [line, ...current].slice(0, 30))
    })

    return () => source.close()
  }, [])

  const handleSync = async () => {
    setSyncing(true)
    setMessage('')
    try {
      const result = await syncHead()
      setMessage(`已同步 ${result.scanned_pages} 页，写入 ${result.updated_items} 条，当前远端总量 ${result.total_comics}。`)
      await refreshSummary()
    } catch (reason) {
      setMessage(reason instanceof Error ? reason.message : '同步失败')
    } finally {
      setSyncing(false)
    }
  }

  const cards = useMemo(
    () => [
      { label: '漫画数量', value: `${comicCount}`, hint: '当前数据库可见漫画', accent: '#355070' },
      { label: '源站数量', value: `${sourceCount}`, hint: '当前已登记源站', accent: '#b55233' },
      { label: '设置项数量', value: `${settingsCount}`, hint: '包含缓存与更新策略', accent: '#0b6e4f' },
    ],
    [comicCount, settingsCount, sourceCount],
  )

  return (
    <section className="page">
      <header className="page-header">
        <div>
          <h1>总览</h1>
          <p>现在这里可以直接触发真实源站头部同步，并通过 SSE 观察切源、重试和补全过程。</p>
        </div>
        <div className="action-row">
          <button className="primary-button" disabled={syncing} onClick={() => void handleSync()}>
            {syncing ? '同步中…' : '执行头部同步'}
          </button>
        </div>
      </header>

      {message ? <section className="panel"><p className="hint-text">{message}</p></section> : null}

      <div className="stats-grid">
        {cards.map((card) => (
          <StatCard key={card.label} {...card} />
        ))}
      </div>

      <section className="panel">
        <div className="panel-head">
          <h2>实时日志</h2>
          <span>通过 SSE 接收事件</span>
        </div>
        <div className="log-box">
          {logLines.length === 0 ? <p>暂时没有事件，后续源站检测、同步和设置变更会出现在这里。</p> : null}
          {logLines.map((line) => (
            <div key={line} className="log-line">{line}</div>
          ))}
        </div>
      </section>
    </section>
  )
}
