import client from './client';
import type {
  ComicListResponse,
  ComicDetail,
  ComicImage,
  ComicCacheState,
  ReaderSessionSnapshot,
  DashboardStats,
  TagWithCount,
  ComicCategory,
} from '../types';

// ===== 漫画列表 =====

export async function fetchComics(page = 1, useCache = true, pageSize = 20): Promise<ComicListResponse> {
  const { data } = await client.get<ComicListResponse>('/comics', {
    params: { page, use_cache: useCache, page_size: pageSize },
  });
  return data;
}

// ===== 漫画详情 =====

export async function fetchComicDetail(id: number): Promise<ComicDetail> {
  const { data } = await client.get<ComicDetail>(`/comics/${id}`);
  return data;
}

// ===== 阅读器图片 =====

export async function fetchComicImages(id: number): Promise<{ images: ComicImage[] }> {
  const { data } = await client.get<{ images: ComicImage[] }>(`/comics/${id}/images`);
  return data;
}

export async function fetchComicCacheState(id: number): Promise<ComicCacheState> {
  const { data } = await client.get<{ code: number; data: ComicCacheState }>(`/v2/comics/${id}/cache`);
  return data.data;
}

export async function createReaderSession(comicId: number): Promise<ReaderSessionSnapshot> {
  const { data } = await client.post<{ code: number; data: ReaderSessionSnapshot }>('/v2/reader/sessions', {
    comic_id: comicId,
  });
  return data.data;
}

export async function fetchReaderSession(sessionId: string): Promise<ReaderSessionSnapshot> {
  const { data } = await client.get<{ code: number; data: ReaderSessionSnapshot }>(`/v2/reader/sessions/${sessionId}`);
  return data.data;
}

export async function updateReaderFocus(sessionId: string, sort: number): Promise<void> {
  await client.post(`/v2/reader/sessions/${sessionId}/focus`, { sort });
}

// ===== 搜索 =====

export async function searchComics(
  keyword: string,
  page = 1,
  mode: 'local' | 'online' = 'local'
): Promise<ComicListResponse> {
  const { data } = await client.get<ComicListResponse>('/comics/search', {
    params: { keyword, page, mode },
  });
  return data;
}

// ===== 筛选 =====

export async function filterComics(params: {
  tag_id?: number;
  category_id?: number;
  author_id?: number;
  author?: string;
  page?: number;
}): Promise<ComicListResponse> {
  const { data } = await client.get<ComicListResponse>('/comics/filter', { params });
  return data;
}

// ===== 收藏 =====

export async function toggleFavorite(
  id: number,
  title: string,
  coverUrl: string
): Promise<{ is_favorited: boolean; message: string }> {
  const { data } = await client.post(`/comics/${id}/favorite`, {
    title,
    cover_url: coverUrl,
  });
  return data;
}

// ===== 阅读历史 =====

export async function addHistory(
  id: number,
  title: string,
  coverUrl: string
): Promise<void> {
  await client.post(`/comics/${id}/history`, {
    title,
    cover_url: coverUrl,
  });
}

// ===== 仪表盘 =====

export async function fetchDashboardStats(): Promise<DashboardStats> {
  const { data } = await client.get<DashboardStats>('/dashboard/stats');
  return data;
}

// ===== 健康检查 =====

export async function healthCheck(): Promise<boolean> {
  try {
    await client.get('/health');
    return true;
  } catch {
    return false;
  }
}

// ===== 标签和分类 =====

export interface Tag {
  id: number;
  name: string;
}

export async function fetchTags(): Promise<Tag[]> {
  const { data } = await client.get<{ code: number; data: Tag[] }>('/v2/tags');
  return data.data || [];
}

export async function fetchCategories(): Promise<ComicCategory[]> {
  const { data } = await client.get<{ code: number; data: ComicCategory[] }>('/v2/categories');
  return data.data || [];
}

// ===== 源站管理 =====

export interface SourceSite {
  id: number;
  url: string;
  name: string;
  image_cdn: string;
  priority: number;
  status: string;
  fail_count: number;
  latency: number;
  direct_status: string;
  direct_latency: number;
  direct_last_error: string;
  proxy_status: string;
  proxy_latency: number;
  proxy_last_error: string;
  last_check: string | null;
  created_at: string;
}

export interface SourceValidationFingerprint {
  comic_id: number;
  title: string;
  author: string;
}

export interface SourceImportSkipped {
  url: string;
  reason: string;
}

export interface ImportReleasePageSourcesResult {
  release_page_url: string;
  candidate_count: number;
  added_count: number;
  skipped_count: number;
  added: SourceSite[];
  skipped: SourceImportSkipped[];
  fingerprints: SourceValidationFingerprint[];
}

export interface SourceHealthResponse {
  id: number;
  url: string;
  status: string;
  latency_ms: number;
  direct_status: string;
  direct_latency_ms: number;
  direct_error?: string;
  proxy_status: string;
  proxy_latency_ms: number;
  proxy_error?: string;
  error?: string;
}

export interface ProxyTargetProbeResponse {
  name: string;
  url: string;
  status: string;
  latency_ms: number;
  error?: string;
}

export interface ProxyHealthResponse {
  configured: boolean;
  available: boolean;
  status: string;
  message?: string;
  latency_ms: number;
  checked_at?: string;
  targets: ProxyTargetProbeResponse[];
}

export async function fetchSources(): Promise<SourceSite[]> {
  const { data } = await client.get<{ code: number; data: SourceSite[] }>('/v2/sources');
  return data.data || [];
}

export async function addSource(url: string, name: string, imageCdn: string): Promise<SourceSite> {
  const { data } = await client.post<{ code: number; data: SourceSite }>('/v2/sources', {
    url,
    name,
    image_cdn: imageCdn,
  });
  return data.data;
}

export async function deleteSource(id: number): Promise<void> {
  await client.delete(`/v2/sources/${id}`);
}

export async function checkSourceHealth(id: number): Promise<SourceHealthResponse> {
  const { data } = await client.post<{ code: number; data: SourceHealthResponse }>(`/v2/sources/${id}/check`);
  return data.data;
}

export async function checkProxyHealth(force = false): Promise<ProxyHealthResponse> {
  const { data } = await client.get<{ code: number; data: ProxyHealthResponse }>('/v2/sources/proxy/check', {
    params: force ? { force: 'true' } : {},
  });
  return data.data;
}

export async function importSourcesFromReleasePage(releasePageUrl: string): Promise<ImportReleasePageSourcesResult> {
  const { data } = await client.post<{ code: number; data: ImportReleasePageSourcesResult }>('/v2/sources/import-release', {
    release_page_url: releasePageUrl,
  });
  return data.data;
}

// ===== 下载管理 =====

export interface DownloadTask {
  id: number;
  comic_id: number;
  comic_title?: string;
  comic_cover_url?: string;
  task_type: string;
  status: string;
  current: number;
  progress: number;
  total: number;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export async function createDownloadTask(comicId: number, taskType: string = 'full'): Promise<DownloadTask> {
  const { data } = await client.post<{ code: number; data: DownloadTask }>('/v2/downloads', {
    comic_id: comicId,
    task_type: taskType,
  });
  return data.data;
}

export async function getDownloadTasks(group?: string): Promise<DownloadTask[]> {
  const { data } = await client.get<{ code: number; data: DownloadTask[] }>('/v2/downloads', {
    params: group ? { group } : {},
  });
  return data.data || [];
}

// ===== 用户设置 =====

export async function getUserSettings(): Promise<Record<string, string>> {
  const { data } = await client.get<{ code: number; data: Record<string, string> }>('/v2/settings');
  return data.data || {};
}

export async function updateUserSettings(settings: Record<string, string>): Promise<void> {
  await client.put('/v2/settings', { settings });
}

// ===== 导入导出 =====

export async function createExport(scope: string, comicIds?: number[]): Promise<{ job_id: number; status: string }> {
  const { data } = await client.post<{ code: number; data: { job_id: number; status: string } }>('/v2/export/create', {
    scope,
    comic_ids: comicIds,
  });
  return data.data;
}

export async function getExportStatus(id: number): Promise<{ id: number; status: string; file_size?: number }> {
  const { data } = await client.get<{ code: number; data: { id: number; status: string; file_size?: number } }>(`/v2/export/status/${id}`);
  return data.data;
}

export function getExportDownloadUrl(id: number): string {
  return `/api/v2/export/download/${id}`;
}

export async function uploadImportFile(file: File): Promise<{ file_path: string; filename: string }> {
  const formData = new FormData();
  formData.append('file', file);
  const { data } = await client.post<{ code: number; data: { file_path: string; filename: string } }>(
    '/v2/import/upload',
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }
  );
  return data.data;
}

export async function scanImport(filePath: string): Promise<{
  conflicts: { will_add: number; will_overwrite: number };
  details: Array<{ comic_id: number; title: string; conflict_type: string }>;
}> {
  const { data } = await client.post('/v2/import/scan', { file_path: filePath });
  return data.data;
}

export async function executeImport(filePath: string, strategy: string): Promise<{ job_id: number; status: string }> {
  const { data } = await client.post('/v2/import/execute', { file_path: filePath, strategy });
  return data.data;
}

export async function getImportStatus(id: number): Promise<{ id: number; status: string; summary?: string }> {
  const { data } = await client.get<{ code: number; data: { id: number; status: string; summary?: string } }>(`/v2/import/status/${id}`);
  return data.data;
}

// ===== 内容更新 =====

export async function checkContentUpdate(pages?: number): Promise<{
  has_update: boolean;
  new_comics: number;
  scanned_pages: number;
}> {
  const { data } = await client.post('/v2/updates/check-content', { pages });
  return data.data;
}

// ===== 存储管理 =====

export interface StorageStats {
  db_size_mb: number;
  download_size_mb: number;
  l1_count: number;
  l2_count: number;
  l3_count: number;
  l3_size_mb: number;
  total_size_mb: number;
}

export async function getStorageStats(): Promise<StorageStats> {
  const { data } = await client.get<{ code: number; data: StorageStats }>('/v2/storage/stats');
  return data.data;
}

export async function clearStorage(level?: string): Promise<void> {
  await client.post('/v2/storage/clear', { level });
}
