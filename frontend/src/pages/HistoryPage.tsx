import { useState, useEffect } from 'react';
import { MangaCard } from '../components/MangaCard';
import { Loader2 } from 'lucide-react';
import { Manga } from '../types';
import client from '../api/client';

export function HistoryPage() {
  const [historyMangas, setHistoryMangas] = useState<Manga[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    client.get('/history').then(({ data }) => {
      const items = data.items || [];
      const converted = items.map((item: any) => ({
        id: String(item.comic_id),
        title: item.title,
        author: [],
        coverUrl: item.cover_url,
        tags: [],
        rating: 0,
        ratingCount: 0,
        favorites: 0,
        categoryName: '',
        description: '',
        cached: false,
        favorited: false,
      }));
      setHistoryMangas(converted);
      setLoading(false);
    }).catch(() => {
      setLoading(false);
    });
  }, []);

  if (loading) {
    return (
      <div className="p-8 flex items-center justify-center py-16">
        <Loader2 className="animate-spin text-blue-500" size={32} />
      </div>
    );
  }

  return (
    <div className="p-8">
      <h1 className="text-3xl mb-6 text-gray-900 dark:text-white">历史观看</h1>
      {historyMangas.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-gray-500 dark:text-gray-400 text-lg">暂无观看历史</p>
        </div>
      ) : (
        <>
          <p className="text-gray-600 dark:text-gray-400 mb-6">共 {historyMangas.length} 部漫画</p>
          <div className="grid grid-cols-4 gap-4">
            {historyMangas.map((manga) => (
              <MangaCard key={manga.id} manga={manga} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
