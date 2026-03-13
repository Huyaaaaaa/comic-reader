// ===== 后端 API 响应类型 =====

export interface ComicListItem {
  id: number;
  title: string;
  cover_url: string;
  cover_base64: string;
  rating: number;
  rating_count: number;
  favorites: number;
  author: string;
  author_id: number;
  is_cached: boolean;
  local_saved: number;
  local_total: number;
}

export interface ComicListResponse {
  items: ComicListItem[];
  current_page: number;
  total_pages: number;
  from_cache: boolean;
}

export interface ComicAuthor {
  author_id: number;
  author_name: string;
}

export interface ComicTag {
  tag_id: number;
  tag_name: string;
}

export interface ComicDetail {
  id: number;
  title: string;
  subtitle: string;
  author: string;
  author_id: number;
  authors: ComicAuthor[];
  cover_url: string;
  rating: number;
  rating_count: number;
  favorites: number;
  category_id: number;
  category_name: string;
  tags: ComicTag[];
  created_at: string;
  updated_at: string;
  reader_url: string;
  is_favorited: boolean;
}

export interface ComicImage {
  sort: number;
  comic_id: number;
  filename: string;
  extension: string;
  url: string;
  local_path: string;
}

export interface DashboardStats {
  total_comics: number;
  cached_comics: number;
  cover_cached: number;
  total_tags: number;
  favorites_count: number;
  history_count: number;
  downloading_count: number;
  pending_downloads: number;
}

// ===== 前端内部类型 =====

export interface Manga {
  id: string;
  title: string;
  author: string[];
  coverUrl: string;
  tags: string[];
  rating: number;
  ratingCount: number;
  favorites: number;
  categoryName: string;
  description: string;
  cached: boolean;
  favorited: boolean;
  chapters: Chapter[];
  cacheLevel: {
    metadata: boolean;
    cover: boolean;
    content: boolean;
  };
}

export interface Chapter {
  id: string;
  title: string;
  pages: string[];
}

export type ViewMode = 'waterfall' | 'single';

export type CacheStrategy = 'none' | 'passive' | 'active' | 'all';

export interface CacheSettings {
  metadata: { strategy: CacheStrategy; count?: number };
  extended: { strategy: CacheStrategy; count?: number };
  content: { strategy: CacheStrategy; count?: number };
}

// ===== 转换函数 =====

export function comicListItemToManga(item: ComicListItem): Manga {
  return {
    id: String(item.id),
    title: item.title,
    author: item.author ? [item.author] : [],
    coverUrl: item.cover_base64
      ? `data:image/jpeg;base64,${item.cover_base64}`
      : item.cover_url,
    tags: [],
    rating: item.rating,
    ratingCount: item.rating_count,
    favorites: item.favorites,
    categoryName: '',
    description: '',
    cached: item.is_cached,
    favorited: false,
    chapters: [],
    cacheLevel: {
      metadata: item.is_cached,
      cover: !!item.cover_base64,
      content: item.local_saved > 0 && item.local_saved === item.local_total,
    },
  };
}

export function comicDetailToManga(detail: ComicDetail): Manga {
  return {
    id: String(detail.id),
    title: detail.title,
    author: detail.authors?.map((a) => a.author_name) ?? (detail.author ? [detail.author] : []),
    coverUrl: detail.cover_url,
    tags: detail.tags?.map((t) => t.tag_name) ?? [],
    rating: detail.rating,
    ratingCount: detail.rating_count,
    favorites: detail.favorites,
    categoryName: detail.category_name ?? '',
    description: detail.subtitle ?? '',
    cached: false,
    favorited: detail.is_favorited,
    chapters: [],
    cacheLevel: { metadata: true, cover: false, content: false },
  };
}
