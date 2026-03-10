import type { SettingsMap, SourceSite } from './types'

export type ComicListItem = {
  id: number
  title: string
  subtitle: string
  cover_url: string
  cover_local_rel_path: string
  rating: number
  rating_count: number
  favorites: number
  category_id: number | null
  category_name: string
  cache_state: {
    meta_level: number
    cover_ready: boolean
    images_total: number
    images_local: number
    offline_ready: boolean
  }
}

export type ComicDetail = {
  id: number
  title: string
  subtitle: string
  cover_url: string
  cover_local_rel_path: string
  rating: number
  rating_count: number
  favorites: number
  category_id: number | null
  category_name: string
  authors: Array<{ id: number; name: string; position?: number }>
  tags: Array<{ id: number; name: string }>
  images_total: number
  created_at: string
  updated_at: string
  cache_state: {
    meta_level: number
    cover_ready: boolean
    images_total: number
    images_local: number
    offline_ready: boolean
  }
}

export type ComicImage = {
  comic_id: number
  sort: number
  image_url: string
  extension: string
  local_rel_path: string
  file_size: number
  cached: boolean
}

export type ComicListResponse = {
  comics: ComicListItem[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export type FavoriteItem = {
  comic_id: number
  comic: ComicListItem
  ensure_offline: boolean
  offline_ready: boolean
  created_at: string
  updated_at: string
}

export type ReadingLocator = {
  mode: string
  sort: number
  offset_ratio: number
}

export type HistoryItem = {
  comic_id: number
  comic: ComicListItem
  locator: ReadingLocator
  last_read_at: string
}

export type SearchHistoryItem = {
  keyword: string
  searched_at: string
}

export type SyncHeadResult = {
  source_id: number
  source_name: string
  scanned_pages: number
  total_pages: number
  scanned_items: number
  updated_items: number
  total_comics: number
  last_page_count: number
  captured_at: string
}

export type SyncComicResult = {
  source_id: number
  source_name: string
  comic_id: number
  title: string
  images_total: number
  authors_total: number
  tags_total: number
  captured_at: string
}

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? '').trim()

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
    ...init,
  })

  const payload = await response.json()
  if (!response.ok) {
    throw new Error(payload.error ?? 'request failed')
  }
  return payload.data as T
}

export async function getSettings(): Promise<SettingsMap> {
  return request<SettingsMap>('/api/settings')
}

export async function updateSetting(key: string, value: unknown): Promise<void> {
  await request(`/api/settings/${key}`, {
    method: 'PUT',
    body: JSON.stringify({ value }),
  })
}

export async function listSources(): Promise<SourceSite[]> {
  const data = await request<{ sources: SourceSite[] }>('/api/sources')
  return data.sources
}

export async function createSource(input: Partial<SourceSite>): Promise<void> {
  await request('/api/sources', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export async function updateSource(id: number, input: Partial<SourceSite>): Promise<void> {
  await request(`/api/sources/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
}

export async function deleteSource(id: number): Promise<void> {
  await request(`/api/sources/${id}`, { method: 'DELETE' })
}

export async function checkSource(id: number): Promise<void> {
  await request(`/api/sources/${id}/check`, { method: 'POST' })
}

export async function syncHead(input?: { source_id?: number; pages?: number }): Promise<SyncHeadResult> {
  return request<SyncHeadResult>('/api/sync/head', {
    method: 'POST',
    body: JSON.stringify(input ?? {}),
  })
}

export async function syncComicDetail(comicId: number, input?: { source_id?: number }): Promise<SyncComicResult> {
  return request<SyncComicResult>(`/api/sync/comics/${comicId}/detail`, {
    method: 'POST',
    body: JSON.stringify(input ?? {}),
  })
}

export function streamEvents(onEvent: (event: MessageEvent<string>) => void): EventSource {
  const source = new EventSource(`${API_BASE_URL}/api/events/stream`)
  const eventNames = [
    'source.check.ok',
    'source.check.retry',
    'source.check.failed',
    'source.request.start',
    'source.request.ok',
    'source.request.retry',
    'source.request.timeout',
    'source.request.failed',
    'source.request.switch',
    'sync.head.started',
    'sync.head.page',
    'sync.head.completed',
    'sync.head.failed',
    'sync.comic.started',
    'sync.comic.completed',
    'sync.comic.failed',
    'cover.cache.completed',
    'cover.cache.failed',
    'image.cache.completed',
    'image.cache.failed',
    'settings.updated',
  ]
  eventNames.forEach((name) => source.addEventListener(name, onEvent as EventListener))
  return source
}

export async function listComics(page = 1, pageSize = 100, search = ''): Promise<ComicListResponse> {
  const keyword = search ? `&search=${encodeURIComponent(search)}` : ''
  return request<ComicListResponse>(`/api/comics?page=${page}&page_size=${pageSize}${keyword}`)
}

export async function searchComics(keyword: string, page = 1, pageSize = 20): Promise<ComicListResponse> {
  return request<ComicListResponse>(`/api/search?keyword=${encodeURIComponent(keyword)}&page=${page}&page_size=${pageSize}`)
}

export async function getComicDetail(id: number): Promise<ComicDetail> {
  return request<ComicDetail>(`/api/comics/${id}`)
}

export async function getComicImages(id: number): Promise<{ images: ComicImage[]; total: number }> {
  return request<{ images: ComicImage[]; total: number }>(`/api/comics/${id}/images`)
}

export function buildCoverProxyURL(comicId: number): string {
  const query = new URLSearchParams({ comic_id: String(comicId) })
  return `${API_BASE_URL}/api/covers/proxy?${query.toString()}`
}

export function buildImageProxyURL(image: ComicImage): string {
  const query = new URLSearchParams({
    comic_id: String(image.comic_id),
    sort: String(image.sort),
  })
  return `${API_BASE_URL}/api/images/proxy?${query.toString()}`
}

export async function listFavorites(comicId?: number): Promise<{ favorites: FavoriteItem[]; total: number; page: number; page_size: number }> {
  const suffix = comicId ? `?comic_id=${comicId}` : ''
  return request(`/api/favorites${suffix}`)
}

export async function addFavorite(comicId: number, ensureOffline = false): Promise<void> {
  await request('/api/favorites', {
    method: 'POST',
    body: JSON.stringify({ comic_id: comicId, ensure_offline: ensureOffline }),
  })
}

export async function removeFavorite(comicId: number): Promise<void> {
  await request(`/api/favorites/${comicId}`, { method: 'DELETE' })
}

export async function listHistory(): Promise<{ history: HistoryItem[]; total: number; page: number; page_size: number }> {
  return request('/api/history')
}

export async function postHistory(comicId: number, locator: ReadingLocator): Promise<void> {
  await request('/api/history', {
    method: 'POST',
    body: JSON.stringify({ comic_id: comicId, locator }),
  })
}

export async function listSearchHistory(): Promise<SearchHistoryItem[]> {
  const data = await request<{ history: SearchHistoryItem[] }>('/api/search/history')
  return data.history
}

export async function addSearchHistory(keyword: string): Promise<void> {
  await request('/api/search/history', {
    method: 'POST',
    body: JSON.stringify({ keyword }),
  })
}

export async function clearSearchHistory(): Promise<void> {
  await request('/api/search/history', { method: 'DELETE' })
}
