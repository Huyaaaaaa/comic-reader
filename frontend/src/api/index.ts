import client from './client';
import type {
  ComicListResponse,
  ComicDetail,
  ComicImage,
  DashboardStats,
} from '../types';

// ===== 漫画列表 =====

export async function fetchComics(page = 1, useCache = true): Promise<ComicListResponse> {
  const { data } = await client.get<ComicListResponse>('/comics', {
    params: { page, use_cache: useCache },
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
