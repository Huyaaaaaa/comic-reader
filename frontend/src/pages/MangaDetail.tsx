import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useManga } from '../contexts/MangaContext';
import { Heart, Download, BookOpen, HardDrive, Loader2 } from 'lucide-react';
import { Manga, ComicImage, comicDetailToManga } from '../types';
import * as api from '../api';

export function MangaDetail() {
  const { id } = useParams();
  const { mangas, toggleFavorite, toggleCache } = useManga();

  const [manga, setManga] = useState<Manga | null>(null);
  const [images, setImages] = useState<ComicImage[]>([]);
  const [sameAuthorMangas, setSameAuthorMangas] = useState<Manga[]>([]);
  const [loading, setLoading] = useState(true);

  // 先从列表缓存中尝试找到
  const cachedManga = mangas.find((m) => m.id === id);

  useEffect(() => {
    if (!id) return;
    setLoading(true);

    const numericId = Number(id);
    Promise.all([
      api.fetchComicDetail(numericId).catch(() => null),
      api.fetchComicImages(numericId).catch(() => ({ images: [] })),
    ]).then(([detail, imgRes]) => {
      if (detail) {
        const converted = comicDetailToManga(detail);
        setManga(converted);
        // 查找同作者作品（从当前列表中）
        const authorNames = converted.author;
        const related = mangas.filter(
          (m) => m.id !== id && m.author.some((a) => authorNames.includes(a))
        );
        setSameAuthorMangas(related);
      } else if (cachedManga) {
        setManga(cachedManga);
      }
      setImages(imgRes.images ?? []);
      setLoading(false);
    });
  }, [id]);

  if (loading) {
    return (
      <div className="p-8 flex items-center justify-center py-16">
        <Loader2 className="animate-spin text-blue-500" size={32} />
      </div>
    );
  }

  const display = manga ?? cachedManga;
  if (!display) {
    return (
      <div className="p-8 text-center py-16">
        <p className="text-gray-500 dark:text-gray-400 text-lg">漫画不存在</p>
      </div>
    );
  }

  const previewImages = images.slice(0, 10);

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
                    <Link key={author} to={`/author/${encodeURIComponent(author)}`} className="text-blue-500 hover:text-blue-600 hover:underline">
                      {author}
                    </Link>
                  ))}
                </div>
              </div>

              {display.tags.length > 0 && (
                <div className="flex items-center gap-2">
                  <span className="text-gray-500 dark:text-gray-400">标签：</span>
                  <div className="flex flex-wrap gap-2">
                    {display.tags.map((tag) => (
                      <Link key={tag} to={`/tags?tags=${encodeURIComponent(tag)}`} className="px-3 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded-full text-sm hover:bg-blue-200 dark:hover:bg-blue-800 transition-colors">
                        {tag}
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
                {display.cached ? (
                  <><HardDrive size={16} className="text-green-500" /><span className="text-green-600 dark:text-green-400">已缓存</span></>
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
                onClick={() => toggleFavorite(display.id)}
                className={`flex items-center gap-2 px-6 py-3 rounded-lg transition-colors ${display.favorited ? 'bg-red-500 text-white hover:bg-red-600' : 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white hover:bg-gray-300 dark:hover:bg-gray-600'}`}
              >
                <Heart size={20} fill={display.favorited ? 'currentColor' : 'none'} />
                <span>{display.favorited ? '已收藏' : '收藏'}</span>
              </button>
              <button
                onClick={() => toggleCache(display.id)}
                className={`flex items-center gap-2 px-6 py-3 rounded-lg transition-colors ${display.cached ? 'bg-green-500 text-white hover:bg-green-600' : 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white hover:bg-gray-300 dark:hover:bg-gray-600'}`}
              >
                <Download size={20} /><span>{display.cached ? '已下载' : '下载'}</span>
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
