export type SourceSite = {
  id: number
  name: string
  base_url: string
  navigator_url: string
  priority: number
  enabled: boolean
  status: string
  last_latency_ms: number | null
  last_checked_at: string | null
  consecutive_failures: number
  last_error: string
}

export type SettingsMap = Record<string, string>

export type SSEEvent = {
  type: string
  timestamp: string
  payload: unknown
}
