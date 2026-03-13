import { useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useManga } from '../contexts/MangaContext';
import { MangaCard } from '../components/MangaCard';
import { Search, Loader2 } from 'lucide-react';
import { Manga } from '../types';

export function TagsPage() {
  const { mangas, searchMangas } = useManga();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const initialSearch = searchParams.get('q') || '';
  const [searchText, setSearchText] = useState(initialSearch);
  const [results, setResults] = useState<Manga[]>([]);
  const [searched, setSearched] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleSearch = async () => {
    if (!searchText.trim()) return;
    setLoading(true);
    setSearched(true);

    const params = new URLSearchParams();
    if (searchText) params.set('q', searchText);
    navigate(`/tags?${params.toString()}`, { replace: true });

    try {
      // 先搜本地列表
      const localResults = mangas.filter(
        (m) =>
          m.title.toLowerCase().includes(searchText.toLowerCase()) ||
          m.author.some((a) => a.toLowerCase().includes(searchText.toLowerCase()))
      );
      setResults(localResults);

      // 再调后端搜索
      const { mangas: apiResults } = await searchMangas(searchText);
      if (apiResults.length > 0) {
        // 合并去重
        const ids = new Set(localResults.map((m) => m.id));
        const merged = [...localResults, ...apiResults.filter((m) => !ids.has(m.id))];
        setResults(merged);
      }
    } catch {
      // 搜索失败保持本地结果
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    setSearchText('');
    setResults([]);
    setSearched(false);
    navigate('/tags', { replace: true });
  };

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl text-gray-900 dark:text-white">搜索</h1>
        <button onClick={handleSearch} className="flex items-center gap-2 px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors">
          <Search size={20} /><span>搜索</span>
        </button>
      </div>

      <div className="mb-6">
        <input
          type="text"
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          placeholder="搜索标题、作者..."
          className="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
          onKeyDown={(e) => { if (e.key === 'Enter') handleSearch(); }}
        />
      </div>

      {searched && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <p className="text-gray-600 dark:text-gray-400">
              {loading ? '搜索中...' : `找到 ${results.length} 部漫画`}
            </p>
            <button onClick={handleReset} className="text-sm text-blue-500 hover:text-blue-600">清除</button>
          </div>

          {loading && (
            <div className="flex justify-center py-8">
              <Loader2 className="animate-spin text-blue-500" size={24} />
            </div>
          )}

          {!loading && results.length > 0 && (
            <div className="grid grid-cols-4 gap-4">
              {results.map((manga) => (
                <MangaCard key={manga.id} manga={manga} />
              ))}
            </div>
          )}

          {!loading && results.length === 0 && (
            <div className="text-center py-16">
              <p className="text-gray-500 dark:text-gray-400 text-lg">没有找到符合条件的漫画</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
