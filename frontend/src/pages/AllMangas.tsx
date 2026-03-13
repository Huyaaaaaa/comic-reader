import { useState, useEffect } from 'react';
import { useManga } from '../contexts/MangaContext';
import { MangaCard } from '../components/MangaCard';
import { Loader2 } from 'lucide-react';

export function AllMangas() {
  const { mangas, currentPage, totalPages, loading, error, fetchPage } = useManga();
  const [page, setPage] = useState(currentPage || 1);

  useEffect(() => {
    fetchPage(page);
  }, [page]);

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
      </div>

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
            onClick={() => fetchPage(page)}
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
