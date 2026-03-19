import { useState, useEffect, useCallback, useMemo } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useManga } from '../contexts/MangaContext';
import { MangaCard } from '../components/MangaCard';
import { Loader2, Search } from 'lucide-react';
import * as api from '../api';

interface Tag {
  id: number;
  name: string;
}

interface Manga {
  id: string;
  title: string;
  author: string[];
  coverUrl: string;
  tags: string[];
  rating: number;
  ratingCount: number;
  favorites: number;
  categoryName: string;
  description: string;
  cached: boolean;
  favorited: boolean;
}

export function TagsPage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const tagsParam = searchParams.get('tags') || '';
  const urlSearch = searchParams.get('q') || '';

  const urlTags = useMemo(
    () => tagsParam.split(',').filter(Boolean),
    [tagsParam]
  );

  const [allTags, setAllTags] = useState<Tag[]>([]);
  const [selectedTags, setSelectedTags] = useState<string[]>(urlTags);
  const [searchText, setSearchText] = useState(urlSearch);
  const [showResults, setShowResults] = useState(urlTags.length > 0 || urlSearch !== '');
  const [loading, setLoading] = useState(true);
  const [searchResults, setSearchResults] = useState<Manga[]>([]);
  const [searching, setSearching] = useState(false);

  useEffect(() => {
    api.fetchTags().then((tagsData) => {
      setAllTags(tagsData);
      setLoading(false);
    }).catch(() => {
      setLoading(false);
    });
  }, []);

  useEffect(() => {
    setSelectedTags(urlTags);
    setSearchText(urlSearch);
    setShowResults(urlTags.length > 0 || urlSearch !== '');
  }, [urlSearch, urlTags]);

  // 如果 URL 中有参数，自动执行搜索
  const toggleTag = (tagName: string) => {
    setSelectedTags((prev) =>
      prev.includes(tagName) ? prev.filter((t) => t !== tagName) : [...prev, tagName]
    );
  };

  const performSearch = useCallback(async () => {
    setSearching(true);
    setShowResults(true);

    try {
      let results: Manga[] = [];

      // 如果选择了标签，逐个标签查询并合并去重结果
      if (selectedTags.length > 0) {
        const merged = new Map<string, Manga>();
        for (const tagName of selectedTags) {
          const selectedTag = allTags.find((t) => t.name === tagName);
          if (!selectedTag) {
            continue;
          }

          const res = await api.filterComics({ tag_id: selectedTag.id, page: 1 });
          for (const item of res.items ?? []) {
            merged.set(String(item.id), {
              id: String(item.id),
              title: item.title,
              author: item.author ? [item.author] : [],
              coverUrl: item.cover_base64
                ? `data:image/jpeg;base64,${item.cover_base64}`
                : item.cover_url,
              tags: [],
              rating: item.rating,
              ratingCount: item.rating_count,
              favorites: item.favorites,
              categoryName: '',
              description: '',
              cached: item.is_cached,
              favorited: false,
            });
          }
        }

        results = Array.from(merged.values());
      }
      // 如果有文本搜索，使用搜索 API
      else if (searchText.trim()) {
        const res = await api.searchComics(searchText.trim(), 1, 'local');
        results = (res.items ?? []).map((item: any) => ({
          id: String(item.id),
          title: item.title,
          author: item.author ? [item.author] : [],
          coverUrl: item.cover_base64
            ? `data:image/jpeg;base64,${item.cover_base64}`
            : item.cover_url,
          tags: [],
          rating: item.rating,
          ratingCount: item.rating_count,
          favorites: item.favorites,
          categoryName: '',
          description: '',
          cached: item.is_cached,
          favorited: false,
        }));
      }

      setSearchResults(results);
    } catch (error) {
      console.error('搜索失败:', error);
      setSearchResults([]);
    } finally {
      setSearching(false);
    }
  }, [allTags, searchText, selectedTags]);

  useEffect(() => {
    if (allTags.length === 0) {
      return;
    }
    if (urlTags.length > 0 || urlSearch) {
      performSearch();
    }
  }, [allTags, urlSearch, urlTags, performSearch]);

  const handleSearch = () => {
    const params = new URLSearchParams();
    if (selectedTags.length > 0) {
      params.set('tags', selectedTags.join(','));
    }
    if (searchText) {
      params.set('q', searchText);
    }
    navigate(`/tags?${params.toString()}`, { replace: true });
    performSearch();
  };

  const handleReset = () => {
    setSelectedTags([]);
    setSearchText('');
    setShowResults(false);
    setSearchResults([]);
    navigate('/tags', { replace: true });
  };

  if (loading) {
    return (
      <div className="p-8 flex items-center justify-center py-16">
        <Loader2 className="animate-spin text-blue-500" size={32} />
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl text-gray-900 dark:text-white">标签搜索</h1>

        <button
          onClick={handleSearch}
          className="flex items-center gap-2 px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
        >
          <Search size={20} />
          <span>搜索</span>
        </button>
      </div>

      {/* 文本搜索 */}
      <div className="mb-6">
        <label className="block mb-2 text-gray-700 dark:text-gray-300">
          文本搜索
        </label>
        <input
          type="text"
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          placeholder="搜索标题、作者或描述..."
          className="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              handleSearch();
            }
          }}
        />
      </div>

      {/* 标签选择 */}
      <div className="mb-8">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg text-gray-900 dark:text-white">选择标签</h2>
          {selectedTags.length > 0 && (
            <button
              onClick={handleReset}
              className="text-sm text-blue-500 hover:text-blue-600"
            >
              清除所有
            </button>
          )}
        </div>
        {allTags.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {allTags.map((tag) => (
              <button
                key={tag.id}
                onClick={() => toggleTag(tag.name)}
                className={`px-4 py-2 rounded-full transition-colors ${
                  selectedTags.includes(tag.name)
                    ? 'bg-blue-500 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white hover:bg-gray-300 dark:hover:bg-gray-600'
                }`}
              >
                {tag.name}
              </button>
            ))}
          </div>
        ) : (
          <p className="text-gray-500 dark:text-gray-400">暂无标签</p>
        )}
      </div>

      {/* 搜索结果 */}
      {showResults && (
        <div>
          {searching ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="animate-spin text-blue-500" size={32} />
              <span className="ml-3 text-gray-500 dark:text-gray-400">搜索中...</span>
            </div>
          ) : (
            <>
              <div className="mb-4">
                <p className="text-gray-600 dark:text-gray-400">
                  找到 {searchResults.length} 部漫画
                </p>
              </div>

              {searchResults.length > 0 ? (
                <div className="grid grid-cols-4 gap-4">
                  {searchResults.map((manga) => (
                    <MangaCard key={manga.id} manga={manga} />
                  ))}
                </div>
              ) : (
                <div className="text-center py-16">
                  <p className="text-gray-500 dark:text-gray-400 text-lg">
                    没有找到符合条件的漫画
                  </p>
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
}
