import { useState, useEffect, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useManga } from '../contexts/MangaContext';
import { ViewMode, ComicImage } from '../types';
import { ArrowLeft, Columns, Rows, Loader2 } from 'lucide-react';
import * as api from '../api';

export function ReadPage() {
  const { id } = useParams();
  const { mangas, addToHistory } = useManga();
  const [viewMode, setViewMode] = useState<ViewMode>('waterfall');
  const [currentPage, setCurrentPage] = useState(0);
  const [images, setImages] = useState<ComicImage[]>([]);
  const [loading, setLoading] = useState(true);
  const waterfallRef = useRef<HTMLDivElement>(null);
  const imageRefs = useRef<(HTMLImageElement | null)[]>([]);

  const manga = mangas.find((m) => m.id === id);

  useEffect(() => {
    if (!id) return;
    if (manga) addToHistory(manga.id);

    setLoading(true);
    api.fetchComicImages(Number(id))
      .then((res) => setImages(res.images ?? []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [id]);

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

  const pages = images.map((img) => img.url);
  const title = manga?.title ?? `漫画 #${id}`;

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
              <img key={index} ref={(el) => (imageRefs.current[index] = el)} src={page} alt={`Page ${index + 1}`} className="w-full mb-2" loading="lazy" />
            ))}
          </div>
        ) : (
          <div className="h-full flex items-center justify-center p-8">
            <div className="max-w-4xl w-full">
              <img src={pages[currentPage]} alt={`Page ${currentPage + 1}`} className="w-full h-auto max-h-[80vh] object-contain mx-auto" />
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
