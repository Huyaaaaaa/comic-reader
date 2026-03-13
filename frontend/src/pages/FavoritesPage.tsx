import { useManga } from '../contexts/MangaContext';
import { MangaCard } from '../components/MangaCard';

export function FavoritesPage() {
  const { mangas } = useManga();
  const favoriteMangas = mangas.filter((manga) => manga.favorited);

  return (
    <div className="p-8">
      <h1 className="text-3xl mb-6 text-gray-900 dark:text-white">收藏</h1>
      {favoriteMangas.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-gray-500 dark:text-gray-400 text-lg">暂无收藏的漫画</p>
        </div>
      ) : (
        <>
          <p className="text-gray-600 dark:text-gray-400 mb-6">共 {favoriteMangas.length} 部漫画</p>
          <div className="grid grid-cols-4 gap-4">
            {favoriteMangas.map((manga) => (
              <MangaCard key={manga.id} manga={manga} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
