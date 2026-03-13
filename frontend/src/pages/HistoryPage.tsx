import { useManga } from '../contexts/MangaContext';
import { MangaCard } from '../components/MangaCard';

export function HistoryPage() {
  const { mangas, history } = useManga();
  const historyMangas = history
    .map((id) => mangas.find((m) => m.id === id))
    .filter(Boolean);

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
              <MangaCard key={manga!.id} manga={manga!} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
