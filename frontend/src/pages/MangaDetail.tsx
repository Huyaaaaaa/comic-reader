import axios from 'axios';
import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useManga } from '../contexts/MangaContext';
import { Heart, Download, BookOpen, HardDrive, Loader2 } from 'lucide-react';
import { Manga, ComicImage, ComicDetail, ComicCacheState, comicDetailToManga } from '../types';
import * as api from '../api';

const ACTIVE_DOWNLOAD_STATUSES = new Set(['queued', 'downloading', 'verifying', 'paused']);

export function MangaDetail() {
  const { id } = useParams();
  const { toggleFavorite, toggleCache, addToHistory } = useManga();

  const [manga, setManga] = useState<Manga | null>(null);
  const [detail, setDetail] = useState<ComicDetail | null>(null);
  const [images, setImages] = useState<ComicImage[]>([]);
  const [sameAuthorMangas, setSameAuthorMangas] = useState<Manga[]>([]);
  const [loading, setLoading] = useState(true);
  const [cacheState, setCacheState] = useState<ComicCacheState | null>(null);
  const [downloadStatus, setDownloadStatus] = useState<'none' | 'downloading' | 'completed'>('none');

  const refreshCacheAndDownloadState = useCallback(async (comicId: number) => {
    let nextCacheState: ComicCacheState | null = null;
    try {
      nextCacheState = await api.fetchComicCacheState(comicId);
      setCacheState(nextCacheState);
      if (nextCacheState.l3_cached) {
        setDownloadStatus('completed');
        setManga((current) => (current ? { ...current, cached: true } : current));
        return;
      }
    } catch (error) {
      if (!axios.isAxiosError(error) || error.response?.status !== 404) {
        console.error('获取缓存状态失败:', error);
      }
      setCacheState(null);
    }

    try {
      const tasks = await api.getDownloadTasks('active');
      const activeTask = tasks.find(
        (task) => task.comic_id === comicId && ACTIVE_DOWNLOAD_STATUSES.has(task.status)
      );
      setDownloadStatus(activeTask ? 'downloading' : 'none');
      setManga((current) =>
        current ? { ...current, cached: Boolean(nextCacheState?.l3_cached) } : current
      );
    } catch (error) {
      console.error('获取下载任务失败:', error);
    }
  }, []);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    setCacheState(null);
    setDownloadStatus('none');

    const numericId = Number(id);
    Promise.all([
      api.fetchComicDetail(numericId).catch(() => null),
      api.fetchComicImages(numericId).catch(() => ({ images: [] })),
    ]).then(async ([detailData, imgRes]) => {
      if (detailData) {
        setDetail(detailData);
        const converted = comicDetailToManga(detailData);
        setManga(converted);
        addToHistory(id, converted.title, converted.coverUrl);

        // 从后端获取同作者作品
        if (converted.author.length > 0) {
          try {
            const res = await api.filterComics({ author: converted.author[0] });
            const related = (res.items ?? [])
              .map((item: any) => ({
                id: String(item.id),
                title: item.title,
                author: item.author ? [item.author] : [],
                coverUrl: item.cover_url,
                tags: [],
                rating: item.rating || 0,
                ratingCount: item.rating_count || 0,
                favorites: item.favorites || 0,
                categoryName: '',
                description: '',
                cached: item.is_cached || false,
                favorited: false,
              }))
              .filter((m: Manga) => m.id !== id);
            setSameAuthorMangas(related);
          } catch {}
        }
      }
      setImages(imgRes.images ?? []);
      setLoading(false);
      refreshCacheAndDownloadState(numericId).catch(() => {});
    });
  }, [id, addToHistory, refreshCacheAndDownloadState]);

  useEffect(() => {
    if (!id || downloadStatus !== 'downloading') {
      return;
    }

    const numericId = Number(id);
    const timer = window.setInterval(() => {
      refreshCacheAndDownloadState(numericId).catch(() => {});
    }, 2000);

    return () => window.clearInterval(timer);
  }, [id, downloadStatus, refreshCacheAndDownloadState]);

  if (loading) {
    return (
      <div className="p-8 flex items-center justify-center py-16">
        <Loader2 className="animate-spin text-blue-500" size={32} />
      </div>
    );
  }

  const display = manga;
  if (!display) {
    return (
      <div className="p-8 text-center py-16">
        <p className="text-gray-500 dark:text-gray-400 text-lg">漫画不存在</p>
      </div>
    );
  }

  const previewImages = images.slice(0, 5);
  const isCached = Boolean(cacheState?.l3_cached || manga?.cached || downloadStatus === 'completed');

  return (
    <div className="p-8">
      <div className="max-w-7xl mx-auto">
        <div className="flex gap-8 mb-8">
          <div className="w-64 flex-shrink-0">
            <img src={display.coverUrl} alt={display.title} className="w-full rounded-lg shadow-lg" />
          </div>

          <div className="flex-1">
            <h1 className="text-4xl mb-4 text-gray-900 dark:text-white">{display.title}</h1>

            <div className="space-y-3 mb-6">
              <div className="flex items-center gap-2">
                <span className="text-gray-500 dark:text-gray-400">作者：</span>
                <div className="flex flex-wrap gap-2">
                  {display.author.map((author) => (
                    <Link key={author} to={`/all?author=${encodeURIComponent(author)}`} className="text-blue-500 hover:text-blue-600 hover:underline">
                      {author}
                    </Link>
                  ))}
                </div>
              </div>

              {detail && detail.tags && detail.tags.length > 0 && (
                <div className="flex items-center gap-2">
                  <span className="text-gray-500 dark:text-gray-400">标签：</span>
                  <div className="flex flex-wrap gap-2">
                    {detail.tags.map((tag) => (
                      <Link key={tag.tag_id} to={`/all?tag_id=${tag.tag_id}`} className="px-3 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded-full text-sm hover:bg-blue-200 dark:hover:bg-blue-800 transition-colors">
                        {tag.tag_name}
                      </Link>
                    ))}
                  </div>
                </div>
              )}

              {display.rating > 0 && (
                <div className="flex items-center gap-2">
                  <span className="text-gray-500 dark:text-gray-400">评分：</span>
                  <span className="text-yellow-500 font-medium">{display.rating}</span>
                  <span className="text-gray-400 text-sm">({display.ratingCount} 人)</span>
                </div>
              )}

              <div className="flex items-center gap-2">
                <span className="text-gray-500 dark:text-gray-400">状态：</span>
                {isCached ? (
                  <><HardDrive size={16} className="text-green-500" /><span className="text-green-600 dark:text-green-400">已缓存</span></>
                ) : downloadStatus === 'downloading' ? (
                  <><Loader2 size={16} className="animate-spin text-yellow-500" /><span className="text-yellow-600 dark:text-yellow-400">正在缓存</span></>
                ) : (
                  <span className="text-gray-600 dark:text-gray-400">未缓存</span>
                )}
              </div>
            </div>

            {display.description && (
              <p className="text-gray-700 dark:text-gray-300 mb-6 leading-relaxed">{display.description}</p>
            )}

            <div className="flex gap-4">
              <Link to={`/read/${display.id}`} className="flex items-center gap-2 px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors">
                <BookOpen size={20} /><span>开始阅读</span>
              </Link>
              <button
                onClick={async () => {
                  const result = await toggleFavorite(display.id, display.title, display.coverUrl);
                  if (result !== undefined && manga) {
                    setManga({ ...manga, favorited: result });
                  }
                }}
                className={`flex items-center gap-2 px-6 py-3 rounded-lg transition-colors ${manga?.favorited ? 'bg-red-500 text-white hover:bg-red-600' : 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white hover:bg-gray-300 dark:hover:bg-gray-600'}`}
              >
                <Heart size={20} fill={manga?.favorited ? 'currentColor' : 'none'} />
                <span>{manga?.favorited ? '已收藏' : '收藏'}</span>
              </button>
              <button
                onClick={async () => {
                  if (downloadStatus === 'downloading' || isCached) return;
                  setDownloadStatus('downloading');
                  try {
                    await toggleCache(display.id, display.title, display.coverUrl);
                    await refreshCacheAndDownloadState(Number(display.id));
                  } catch (error) {
                    if (axios.isAxiosError(error) && error.response?.status === 409) {
                      await refreshCacheAndDownloadState(Number(display.id));
                      setDownloadStatus('downloading');
                      return;
                    }
                    setDownloadStatus('none');
                  }
                }}
                disabled={downloadStatus === 'downloading'}
                className={`flex items-center gap-2 px-6 py-3 rounded-lg transition-colors ${
                  isCached
                    ? 'bg-green-500 text-white hover:bg-green-600'
                    : downloadStatus === 'downloading'
                    ? 'bg-yellow-500 text-white cursor-wait'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white hover:bg-gray-300 dark:hover:bg-gray-600'
                }`}
              >
                {downloadStatus === 'downloading' ? (
                  <><Loader2 size={20} className="animate-spin" /><span>缓存中...</span></>
                ) : (
                  <><Download size={20} /><span>{isCached ? '已缓存' : '下载'}</span></>
                )}
              </button>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-3 gap-8">
          <div className="col-span-2">
            <h2 className="text-2xl mb-4 text-gray-900 dark:text-white">预览</h2>
            {previewImages.length > 0 ? (
              <div className="grid grid-cols-5 gap-2">
                {previewImages.map((img, index) => (
                  <Link key={index} to={`/read/${display.id}`} className="aspect-[3/4] overflow-hidden rounded-lg bg-gray-200 dark:bg-gray-700 hover:opacity-80 transition-opacity">
                    <img src={img.url} alt={`Preview ${index + 1}`} className="w-full h-full object-cover" />
                  </Link>
                ))}
              </div>
            ) : (
              <p className="text-gray-500 dark:text-gray-400">暂无预览图片</p>
            )}
          </div>

          <div>
            <h2 className="text-2xl mb-4 text-gray-900 dark:text-white">同作者作品</h2>
            {sameAuthorMangas.length > 0 ? (
              <div className="space-y-4">
                {sameAuthorMangas.slice(0, 5).map((m) => (
                  <Link key={m.id} to={`/manga/${m.id}`} className="flex gap-3 bg-white dark:bg-gray-800 rounded-lg p-3 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                    <img src={m.coverUrl} alt={m.title} className="w-16 h-20 object-cover rounded" />
                    <div className="flex-1 min-w-0">
                      <h3 className="font-medium text-gray-900 dark:text-white truncate">{m.title}</h3>
                      <p className="text-sm text-gray-500 dark:text-gray-400 truncate">{m.author.join(', ')}</p>
                    </div>
                  </Link>
                ))}
              </div>
            ) : (
              <p className="text-gray-500 dark:text-gray-400 text-sm">暂无其他作品</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
