import { type FormEvent, useEffect, useState } from 'react'
import { checkSource, createSource, deleteSource, listSources, syncHead, type SyncHeadResult } from '../lib/api'
import type { SourceSite } from '../lib/types'

const emptyForm = {
  name: '',
  base_url: '',
  navigator_url: '',
  priority: 0,
}

export default function SourcesPage() {
  const [sources, setSources] = useState<SourceSite[]>([])
  const [form, setForm] = useState(emptyForm)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [syncingSourceId, setSyncingSourceId] = useState<number | null>(null)
  const [syncMessage, setSyncMessage] = useState('')

  const refresh = async () => {
    setSources(await listSources())
  }

  useEffect(() => {
    void refresh()
  }, [])

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setLoading(true)
    setError('')
    try {
      await createSource({
        name: form.name,
        base_url: form.base_url,
        navigator_url: form.navigator_url,
        priority: Number(form.priority),
        enabled: true,
      })
      setForm(emptyForm)
      await refresh()
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '保存失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCheck = async (id: number) => {
    setError('')
    try {
      await checkSource(id)
      await refresh()
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '检测失败')
      await refresh()
    }
  }

  const handleSync = async (sourceId?: number) => {
    setError('')
    setSyncMessage('')
    setSyncingSourceId(sourceId ?? -1)
    try {
      const result: SyncHeadResult = await syncHead(sourceId ? { source_id: sourceId } : undefined)
      setSyncMessage(`已从 ${result.source_name} 扫描 ${result.scanned_pages} 页，写入 ${result.updated_items} 条，远端总量 ${result.total_comics}。`)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '同步失败')
    } finally {
      setSyncingSourceId(null)
    }
  }

  const handleDelete = async (id: number) => {
    await deleteSource(id)
    await refresh()
  }

  return (
    <section className="page two-column">
      <section className="panel">
        <div className="panel-head">
          <h1>源站管理</h1>
          <span>默认不内置任何站点地址</span>
        </div>
        <form className="stack-form" onSubmit={handleSubmit}>
          <label>
            名称
            <input value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} placeholder="例如：主站" />
          </label>
          <label>
            源站地址
            <input value={form.base_url} onChange={(event) => setForm({ ...form, base_url: event.target.value })} placeholder="请手动填写完整源站地址" />
          </label>
          <label>
            导航页地址
            <input value={form.navigator_url} onChange={(event) => setForm({ ...form, navigator_url: event.target.value })} placeholder="后续可用来自动导入站点" />
          </label>
          <label>
            优先级
            <input type="number" value={form.priority} onChange={(event) => setForm({ ...form, priority: Number(event.target.value) })} />
          </label>
          <div className="action-row">
            <button className="primary-button" disabled={loading} type="submit">
              {loading ? '保存中…' : '添加源站'}
            </button>
            <button type="button" onClick={() => void handleSync()}>
              {syncingSourceId === -1 ? '同步中…' : '按优先级头部同步'}
            </button>
          </div>
          {syncMessage ? <p className="hint-text">{syncMessage}</p> : null}
          {error ? <p className="error-text">{error}</p> : null}
        </form>
      </section>

      <section className="panel">
        <div className="panel-head">
          <h2>已配置源站</h2>
          <span>单次请求超时会自动重试，连续三次失败标记不可用</span>
        </div>
        <div className="source-list">
          {sources.length === 0 ? <p>还没有任何源站，请先添加。</p> : null}
          {sources.map((source) => (
            <article key={source.id} className="source-card">
              <div>
                <strong>{source.name}</strong>
                <p>{source.base_url}</p>
                <small>状态：{source.status} · 连续失败：{source.consecutive_failures}</small>
              </div>
              <div className="source-actions">
                <button onClick={() => void handleCheck(source.id)}>测速/检测</button>
                <button onClick={() => void handleSync(source.id)}>{syncingSourceId === source.id ? '同步中…' : '头部同步'}</button>
                <button className="danger-button" onClick={() => void handleDelete(source.id)}>删除</button>
              </div>
            </article>
          ))}
        </div>
      </section>
    </section>
  )
}
