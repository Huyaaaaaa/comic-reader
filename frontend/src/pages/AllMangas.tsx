import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { MangaCard } from '../components/MangaCard';
import { Loader2, X } from 'lucide-react';
import * as api from '../api';
import { Manga, comicListItemToManga } from '../types';

export function AllMangas() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [mangas, setMangas] = useState<Manga[]>([]);
  const [page, setPage] = useState(Number(searchParams.get('page')) || 1);
  const [perPage, setPerPage] = useState(20);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [settingsLoaded, setSettingsLoaded] = useState(false);

  const tagId = searchParams.get('tag_id');
  const categoryId = searchParams.get('category_id');
  const authorId = searchParams.get('author_id');
  const author = searchParams.get('author');

  // 初始化时从后端读取 page_size 设置
  useEffect(() => {
    api.getUserSettings().then((settings) => {
      if (settings.page_size) {
        setPerPage(Number(settings.page_size));
      }
      setSettingsLoaded(true);
    }).catch(() => {
      setSettingsLoaded(true);
    });
  }, []);

  useEffect(() => {
    // 等待设置加载完成后再加载漫画列表
    if (!settingsLoaded) return;

    setLoading(true);
    setError(null);

    const fetchData = async () => {
      try {
        if (tagId || categoryId || authorId || author) {
          // 使用筛选 API
          const params: any = { page };
          if (tagId) params.tag_id = Number(tagId);
          if (categoryId) params.category_id = Number(categoryId);
          if (authorId) params.author_id = Number(authorId);
          if (author) params.author = author;

          const res = await api.filterComics(params);
          const converted = (res.items ?? []).map(comicListItemToManga);
          setMangas(converted);
          setTotalPages(res.total_pages);
        } else {
          // 使用普通列表 API
          const res = await api.fetchComics(page, true, perPage);
          const converted = (res.items ?? []).map(comicListItemToManga);
          setMangas(converted);
          setTotalPages(res.total_pages);
        }
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : '加载失败';
        setError(msg);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [page, perPage, tagId, categoryId, authorId, author, settingsLoaded]);

  const clearFilter = () => {
    setSearchParams({});
    setPage(1);
  };

  const hasFilter = tagId || categoryId || authorId || author;

  const maxVisiblePages = 10;
  const startPage = Math.max(1, page - Math.floor(maxVisiblePages / 2));
  const endPage = Math.min(totalPages, startPage + maxVisiblePages - 1);

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl mb-2 text-gray-900 dark:text-white">所有漫画</h1>
          <p className="text-gray-600 dark:text-gray-400">
            第 {page} 页 / 共 {totalPages} 页
          </p>
        </div>

        <div className="flex items-center gap-2">
          <label className="text-sm text-gray-700 dark:text-gray-300">每页显示：</label>
          <select
            value={perPage}
            onChange={(e) => {
              const newValue = Number(e.target.value);
              setPerPage(newValue);
              setPage(1);
              // 保存到后端
              api.updateUserSettings({ page_size: String(newValue) }).catch((err) => {
                console.error('保存设置失败:', err);
              });
            }}
            className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
          >
            <option value={8}>8</option>
            <option value={12}>12</option>
            <option value={16}>16</option>
            <option value={20}>20</option>
          </select>
        </div>
      </div>

      {hasFilter && (
        <div className="mb-6 flex items-center gap-2 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-3">
          <span className="text-sm text-gray-700 dark:text-gray-300">当前筛选：</span>
          {author && <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded text-sm">作者: {author}</span>}
          {tagId && <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded text-sm">标签 ID: {tagId}</span>}
          {categoryId && <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded text-sm">分类 ID: {categoryId}</span>}
          {authorId && <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded text-sm">作者 ID: {authorId}</span>}
          <button
            onClick={clearFilter}
            className="ml-auto flex items-center gap-1 px-3 py-1 text-sm text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-200 transition-colors"
          >
            <X size={16} />
            清除筛选
          </button>
        </div>
      )}

      {loading && (
        <div className="flex items-center justify-center py-16">
          <Loader2 className="animate-spin text-blue-500" size={32} />
          <span className="ml-3 text-gray-500 dark:text-gray-400">加载中...</span>
        </div>
      )}

      {error && !loading && (
        <div className="text-center py-16">
          <p className="text-red-500 mb-4">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
          >
            重试
          </button>
        </div>
      )}

      {!loading && !error && (
        <>
          <div className="grid grid-cols-4 gap-4 mb-8">
            {mangas.map((manga) => (
              <MangaCard key={manga.id} manga={manga} />
            ))}
          </div>

          {totalPages > 1 && (
            <div className="flex justify-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-4 py-2 rounded-lg bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
              >
                上一页
              </button>

              <div className="flex items-center gap-1">
                {Array.from({ length: endPage - startPage + 1 }, (_, i) => startPage + i).map((p) => (
                  <button
                    key={p}
                    onClick={() => setPage(p)}
                    className={`w-10 h-10 rounded-lg transition-colors ${
                      page === p
                        ? 'bg-blue-500 text-white'
                        : 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white hover:bg-gray-300 dark:hover:bg-gray-600'
                    }`}
                  >
                    {p}
                  </button>
                ))}
              </div>

              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-4 py-2 rounded-lg bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
              >
                下一页
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
