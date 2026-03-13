import { Link } from 'react-router-dom';
import { Manga } from '../types';

interface MangaCardProps {
  manga: Manga;
}

export function MangaCard({ manga }: MangaCardProps) {
  return (
    <Link to={`/manga/${manga.id}`} className="group">
      <div className="bg-white dark:bg-gray-800 rounded-lg overflow-hidden shadow-sm hover:shadow-md transition-shadow">
        <div className="aspect-[3/4] overflow-hidden bg-gray-200 dark:bg-gray-700">
          <img
            src={manga.coverUrl}
            alt={manga.title}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
            loading="lazy"
          />
        </div>
        <div className="p-3">
          <h3 className="font-medium text-gray-900 dark:text-white truncate">
            {manga.title}
          </h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 truncate">
            {manga.author.join(', ')}
          </p>
          {manga.rating > 0 && (
            <p className="text-xs text-yellow-500 mt-1">{manga.rating}</p>
          )}
        </div>
      </div>
    </Link>
  );
}
