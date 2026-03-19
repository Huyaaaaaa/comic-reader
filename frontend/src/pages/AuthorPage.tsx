import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { MangaCard } from '../components/MangaCard';
import { User, Loader2 } from 'lucide-react';
import { Manga, comicListItemToManga } from '../types';
import { filterComics } from '../api';

export function AuthorPage() {
  const { name } = useParams();
  const [authorMangas, setAuthorMangas] = useState<Manga[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!name) return;
    setLoading(true);

    filterComics({ author: name })
      .then((res) => {
        const converted = (res.items ?? []).map(comicListItemToManga);
        setAuthorMangas(converted);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [name]);

  return (
    <div className="p-8">
      <div className="flex items-center gap-3 mb-6">
        <User className="text-blue-500" size={32} />
        <div>
          <h1 className="text-3xl text-gray-900 dark:text-white">{name}</h1>
          <p className="text-gray-600 dark:text-gray-400">的作品</p>
        </div>
      </div>

      {loading && (
        <div className="flex justify-center py-8">
          <Loader2 className="animate-spin text-blue-500" size={24} />
        </div>
      )}

      {authorMangas.length > 0 ? (
        <>
          <p className="text-gray-600 dark:text-gray-400 mb-6">共 {authorMangas.length} 部作品</p>
          <div className="grid grid-cols-4 gap-4">
            {authorMangas.map((manga) => (
              <MangaCard key={manga.id} manga={manga} />
            ))}
          </div>
        </>
      ) : !loading ? (
        <div className="text-center py-16">
          <p className="text-gray-500 dark:text-gray-400 text-lg">暂无作品</p>
        </div>
      ) : null}
    </div>
  );
}
