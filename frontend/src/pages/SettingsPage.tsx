import { useEffect, useState } from 'react'
import { getSettings, updateSetting } from '../lib/api'
import type { SettingsMap } from '../lib/types'

export default function SettingsPage() {
  const [settings, setSettings] = useState<SettingsMap>({})
  const [status, setStatus] = useState('')

  useEffect(() => {
    void getSettings().then(setSettings)
  }, [])

  const handleSave = async (key: string, value: string) => {
    setStatus('保存中…')
    try {
      await updateSetting(key, value)
      setSettings((current) => ({ ...current, [key]: JSON.stringify(value) }))
      setStatus('已保存')
    } catch (reason) {
      setStatus(reason instanceof Error ? reason.message : '保存失败')
    }
  }

  return (
    <section className="page">
      <header className="page-header">
        <div>
          <h1>设置</h1>
          <p>这里先承接缓存预设、更新模式和源站调度的核心参数。</p>
        </div>
        <span>{status}</span>
      </header>

      <div className="settings-grid">
        <SettingField label="L1 缓存模式" value={settings.l1_cache_mode} onSave={(value) => void handleSave('l1_cache_mode', value)} />
        <SettingField label="L2 缓存模式" value={settings.l2_cache_mode} onSave={(value) => void handleSave('l2_cache_mode', value)} />
        <SettingField label="L3 缓存模式" value={settings.l3_cache_mode} onSave={(value) => void handleSave('l3_cache_mode', value)} />
        <SettingField label="内容更新模式" value={settings.content_update_mode} onSave={(value) => void handleSave('content_update_mode', value)} />
        <SettingField label="应用更新模式" value={settings.app_update_mode} onSave={(value) => void handleSave('app_update_mode', value)} />
        <SettingField label="源站心跳间隔(分钟)" value={settings.source_heartbeat_interval_minutes} onSave={(value) => void handleSave('source_heartbeat_interval_minutes', value)} />
      </div>
    </section>
  )
}

type SettingFieldProps = {
  label: string
  value?: string
  onSave: (value: string) => void
}

function normalizeValue(value: string) {
  return value.split('"').join('')
}

function SettingField({ label, value = '', onSave }: SettingFieldProps) {
  const [draft, setDraft] = useState(normalizeValue(value))

  useEffect(() => {
    setDraft(normalizeValue(value))
  }, [value])

  return (
    <article className="setting-card">
      <label>
        <span>{label}</span>
        <input value={draft} onChange={(event) => setDraft(event.target.value)} />
      </label>
      <button onClick={() => onSave(draft)}>保存</button>
    </article>
  )
}
