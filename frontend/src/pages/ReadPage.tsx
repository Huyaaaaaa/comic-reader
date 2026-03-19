import axios from 'axios';
import { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useManga } from '../contexts/MangaContext';
import { ViewMode, ComicImage, ReaderSessionSnapshot } from '../types';
import { ArrowLeft, Columns, Rows, Loader2 } from 'lucide-react';
import * as api from '../api';

export function ReadPage() {
  const { id } = useParams();
  const { addToHistory } = useManga();
  const [viewMode, setViewMode] = useState<ViewMode>('waterfall');
  const [currentPage, setCurrentPage] = useState(0);
  const [images, setImages] = useState<ComicImage[]>([]);
  const [loading, setLoading] = useState(true);
  const [title, setTitle] = useState('');
  const [contentCacheStrategy, setContentCacheStrategy] = useState('passive');
  const [readerSession, setReaderSession] = useState<ReaderSessionSnapshot | null>(null);
  const waterfallRef = useRef<HTMLDivElement>(null);
  const imageRefs = useRef<(HTMLDivElement | null)[]>([]);
  const passiveCacheTriggeredRef = useRef<string | null>(null);
  const lastFocusedSortRef = useRef<number | null>(null);

  // 从后端读取默认阅读模式
  useEffect(() => {
    api.getUserSettings().then((settings) => {
      if (settings.reading_mode === 'single' || settings.reading_mode === 'waterfall') {
        setViewMode(settings.reading_mode);
      }
      if (settings.cache_l3_strategy) {
        setContentCacheStrategy(settings.cache_l3_strategy);
      }
    }).catch(() => {});
  }, []);

  useEffect(() => {
    passiveCacheTriggeredRef.current = null;
    lastFocusedSortRef.current = null;
    setReaderSession(null);
    setCurrentPage(0);
  }, [id]);

  useEffect(() => {
    if (!id) return;

    let canceled = false;
    setLoading(true);

    (async () => {
      try {
        const comicId = Number(id);
        const imageRes = await api.fetchComicImages(comicId);
        if (canceled) {
          return;
        }

        setImages(imageRes.images ?? []);

        const [session, fetchedTitle] = await Promise.all([
          api.createReaderSession(comicId),
          api.fetchComicDetail(comicId).then((d) => {
            addToHistory(id, d.title, d.cover_url || '');
            return d.title;
          }).catch(() => `漫画 #${id}`),
        ]);

        if (canceled) {
          return;
        }

        setReaderSession(session);
        setTitle(fetchedTitle);
      } catch {
        if (!canceled) {
          setReaderSession(null);
        }
      } finally {
        if (!canceled) {
          setLoading(false);
        }
      }
    })();

    return () => {
      canceled = true;
    };
  }, [id, addToHistory]);

  useEffect(() => {
    if (!readerSession?.session_id) {
      return;
    }

    let stopped = false;
    let timer: number | undefined;

    const poll = async () => {
      try {
        const next = await api.fetchReaderSession(readerSession.session_id);
        if (stopped) {
          return;
        }
        setReaderSession(next);
        const delay = next.ready_count < next.total ? 350 : 1500;
        timer = window.setTimeout(poll, delay);
      } catch {
        if (!stopped) {
          timer = window.setTimeout(poll, 1000);
        }
      }
    };

    timer = window.setTimeout(poll, 350);

    return () => {
      stopped = true;
      if (timer) {
        window.clearTimeout(timer);
      }
    };
  }, [readerSession?.session_id]);

  const reportFocus = useCallback((pageIndex: number) => {
    if (!readerSession?.session_id) {
      return;
    }

    const nextSort = pageIndex + 1;
    if (nextSort < 1 || lastFocusedSortRef.current === nextSort) {
      return;
    }

    lastFocusedSortRef.current = nextSort;
    api.updateReaderFocus(readerSession.session_id, nextSort).catch(() => {});
  }, [readerSession?.session_id]);

  useEffect(() => {
    if (!id || images.length === 0 || contentCacheStrategy !== 'passive') {
      return;
    }
    if (passiveCacheTriggeredRef.current === id) {
      return;
    }
    passiveCacheTriggeredRef.current = id;

    const comicId = Number(id);
    api.fetchComicCacheState(comicId)
      .then((state) => {
        if (state.l3_cached) {
          return;
        }
        return api.createDownloadTask(comicId, 'images');
      })
      .catch((error) => {
        if (axios.isAxiosError(error)) {
          if (error.response?.status === 404) {
            api.createDownloadTask(comicId, 'images').catch((taskError) => {
              if (!axios.isAxiosError(taskError) || taskError.response?.status !== 409) {
                console.error('触发被动缓存失败:', taskError);
              }
            });
            return;
          }
          if (error.response?.status === 409) {
            return;
          }
        }
        console.error('检查缓存状态失败:', error);
      });
  }, [id, images, contentCacheStrategy]);

  useEffect(() => {
    if (viewMode !== 'single') {
      return;
    }
    reportFocus(currentPage);
  }, [currentPage, viewMode, reportFocus]);

  useEffect(() => {
    if (viewMode !== 'waterfall') {
      return;
    }

    const container = waterfallRef.current;
    if (!container) {
      return;
    }

    let ticking = false;
    const handleScroll = () => {
      if (ticking) {
        return;
      }

      ticking = true;
      window.requestAnimationFrame(() => {
        ticking = false;
        const containerRect = container.getBoundingClientRect();
        const containerCenter = containerRect.top + containerRect.height / 2;
        let closestIndex = 0;
        let minDistance = Number.POSITIVE_INFINITY;

        imageRefs.current.forEach((node, index) => {
          if (!node) {
            return;
          }
          const rect = node.getBoundingClientRect();
          const center = rect.top + rect.height / 2;
          const distance = Math.abs(center - containerCenter);
          if (distance < minDistance) {
            minDistance = distance;
            closestIndex = index;
          }
        });

        setCurrentPage(closestIndex);
        reportFocus(closestIndex);
      });
    };

    handleScroll();
    container.addEventListener('scroll', handleScroll);
    return () => container.removeEventListener('scroll', handleScroll);
  }, [viewMode, reportFocus, readerSession?.images.length]);

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
        <Loader2 className="animate-spin text-blue-500" size={32} />
      </div>
    );
  }

  if (images.length === 0) {
    return (
      <div className="h-screen flex flex-col items-center justify-center bg-gray-50 dark:bg-gray-900">
        <p className="text-gray-500 dark:text-gray-400 text-lg mb-4">暂无图片数据</p>
        <Link to={id ? `/manga/${id}` : '/'} className="text-blue-500 hover:underline">返回</Link>
      </div>
    );
  }

  const pages = readerSession?.images ?? images.map((img) => ({
    sort: img.sort,
    filename: img.filename,
    extension: img.extension,
    status: 'ready',
    ready: true,
    source: img.local_path ? 'local' : 'remote',
    view_url: img.url,
  }));

  const handlePrevPage = () => setCurrentPage((p) => Math.max(0, p - 1));
  const handleNextPage = () => setCurrentPage((p) => Math.min(pages.length - 1, p + 1));

  const handleViewModeChange = (newMode: ViewMode) => {
    if (newMode === 'waterfall' && viewMode === 'single') {
      setViewMode(newMode);
      setTimeout(() => {
        imageRefs.current[currentPage]?.scrollIntoView({ behavior: 'smooth' });
      }, 100);
    } else if (newMode === 'single' && viewMode === 'waterfall') {
      if (waterfallRef.current) {
        const container = waterfallRef.current;
        const scrollTop = container.scrollTop;
        const containerHeight = container.clientHeight;
        const centerY = scrollTop + containerHeight / 2;
        let closestIndex = 0;
        let minDistance = Infinity;
        imageRefs.current.forEach((img, index) => {
          if (img) {
            const rect = img.getBoundingClientRect();
            const imgTop = scrollTop + rect.top;
            const distance = Math.abs(imgTop + rect.height / 2 - centerY);
            if (distance < minDistance) { minDistance = distance; closestIndex = index; }
          }
        });
        setCurrentPage(closestIndex);
      }
      setViewMode(newMode);
    } else {
      setViewMode(newMode);
    }
  };

  return (
    <div className="h-screen flex flex-col bg-gray-50 dark:bg-gray-900">
      <div className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link to={`/manga/${id}`} className="flex items-center gap-2 text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors">
              <ArrowLeft size={20} /><span>返回</span>
            </Link>
            <div className="border-l border-gray-300 dark:border-gray-600 h-6" />
            <h1 className="font-medium text-gray-900 dark:text-white">{title}</h1>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 bg-gray-100 dark:bg-gray-700 rounded-lg p-1">
              <button onClick={() => handleViewModeChange('waterfall')} className={`flex items-center gap-2 px-3 py-2 rounded transition-colors ${viewMode === 'waterfall' ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm' : 'text-gray-600 dark:text-gray-400'}`}>
                <Rows size={16} /><span className="text-sm">瀑布流</span>
              </button>
              <button onClick={() => handleViewModeChange('single')} className={`flex items-center gap-2 px-3 py-2 rounded transition-colors ${viewMode === 'single' ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm' : 'text-gray-600 dark:text-gray-400'}`}>
                <Columns size={16} /><span className="text-sm">单图流</span>
              </button>
            </div>
            {viewMode === 'single' && (
              <div className="text-sm text-gray-600 dark:text-gray-400">{currentPage + 1} / {pages.length}</div>
            )}
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-auto" ref={waterfallRef}>
        {viewMode === 'waterfall' ? (
          <div className="max-w-4xl mx-auto py-8">
            {pages.map((page, index) => (
              <div
                key={index}
                ref={(el) => (imageRefs.current[index] = el)}
                className="w-full mb-2 rounded-lg overflow-hidden bg-white dark:bg-gray-800 shadow-sm"
              >
                {page.ready ? (
                  <img src={page.view_url} alt={`Page ${index + 1}`} className="w-full block" loading={index < 2 ? 'eager' : 'lazy'} />
                ) : (
                  <div className="aspect-[3/4] flex flex-col items-center justify-center text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800">
                    <Loader2 className="animate-spin mb-3" size={24} />
                    <span className="text-sm">第 {index + 1} 张准备中</span>
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : (
          <div className="h-full flex items-center justify-center p-8">
            <div className="max-w-4xl w-full">
              {pages[currentPage]?.ready ? (
                <img src={pages[currentPage].view_url} alt={`Page ${currentPage + 1}`} className="w-full h-auto max-h-[80vh] object-contain mx-auto" />
              ) : (
                <div className="w-full max-w-4xl aspect-[3/4] max-h-[80vh] flex flex-col items-center justify-center rounded-xl bg-white dark:bg-gray-800 text-gray-500 dark:text-gray-400">
                  <Loader2 className="animate-spin mb-3" size={30} />
                  <span>第 {currentPage + 1} 张准备中</span>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {viewMode === 'single' && (
        <div className="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 px-6 py-4">
          <div className="flex items-center justify-center gap-4">
            <button onClick={handlePrevPage} disabled={currentPage === 0} className="px-6 py-2 rounded-lg bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors">上一页</button>
            <div className="text-sm text-gray-600 dark:text-gray-400">{currentPage + 1} / {pages.length}</div>
            <button onClick={handleNextPage} disabled={currentPage === pages.length - 1} className="px-6 py-2 rounded-lg bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors">下一页</button>
          </div>
        </div>
      )}
    </div>
  );
}
