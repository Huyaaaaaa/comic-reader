import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import { Manga, comicListItemToManga, DashboardStats } from '../types';
import * as api from '../api';

export interface DownloadItem {
  id: string;
  mangaId: string;
  mangaTitle: string;
  progress: number;
  status: 'downloading' | 'completed' | 'paused';
}

interface MangaContextType {
  mangas: Manga[];
  currentPage: number;
  totalPages: number;
  fromCache: boolean;
  loading: boolean;
  error: string | null;
  fetchPage: (page: number) => Promise<void>;
  searchMangas: (keyword: string, page?: number) => Promise<{ mangas: Manga[]; totalPages: number }>;
  toggleFavorite: (id: string) => Promise<void>;
  history: string[];
  addToHistory: (id: string) => void;
  toggleCache: (id: string) => void;
  downloads: DownloadItem[];
  stats: DashboardStats | null;
  fetchStats: () => Promise<void>;
  online: boolean;
}

const MangaContext = createContext<MangaContextType | undefined>(undefined);

export function MangaProvider({ children }: { children: ReactNode }) {
  const [mangas, setMangas] = useState<Manga[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [fromCache, setFromCache] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [history, setHistory] = useState<string[]>(() => {
    const saved = localStorage.getItem('comic-history');
    return saved ? JSON.parse(saved) : [];
  });
  const [downloads] = useState<DownloadItem[]>([]);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [online, setOnline] = useState(true);

  useEffect(() => {
    localStorage.setItem('comic-history', JSON.stringify(history));
  }, [history]);

  useEffect(() => {
    api.healthCheck().then(setOnline);
  }, []);

  const fetchPage = useCallback(async (page: number) => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.fetchComics(page);
      const converted = (res.items ?? []).map(comicListItemToManga);
      setMangas(converted);
      setCurrentPage(res.current_page);
      setTotalPages(res.total_pages);
      setFromCache(res.from_cache);
      setOnline(true);
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : '加载失败';
      setError(msg);
      setOnline(false);
    } finally {
      setLoading(false);
    }
  }, []);

  const searchMangas = useCallback(async (keyword: string, page = 1) => {
    const res = await api.searchComics(keyword, page);
    const converted = (res.items ?? []).map(comicListItemToManga);
    return { mangas: converted, totalPages: res.total_pages };
  }, []);

  const toggleFavorite = useCallback(async (id: string) => {
    const manga = mangas.find((m) => m.id === id);
    if (!manga) return;
    try {
      const res = await api.toggleFavorite(Number(id), manga.title, manga.coverUrl);
      setMangas((prev) =>
        prev.map((m) => (m.id === id ? { ...m, favorited: res.is_favorited } : m))
      );
    } catch {
      setMangas((prev) =>
        prev.map((m) => (m.id === id ? { ...m, favorited: !m.favorited } : m))
      );
    }
  }, [mangas]);

  const toggleCache = useCallback((id: string) => {
    setMangas((prev) =>
      prev.map((m) => (m.id === id ? { ...m, cached: !m.cached } : m))
    );
  }, []);

  const addToHistory = useCallback((id: string) => {
    setHistory((prev) => {
      const filtered = prev.filter((item) => item !== id);
      return [id, ...filtered].slice(0, 50);
    });
    const manga = mangas.find((m) => m.id === id);
    if (manga) {
      api.addHistory(Number(id), manga.title, manga.coverUrl).catch(() => {});
    }
  }, [mangas]);

  const fetchStats = useCallback(async () => {
    try {
      const data = await api.fetchDashboardStats();
      setStats(data);
    } catch {
      // ignore
    }
  }, []);

  return (
    <MangaContext.Provider
      value={{
        mangas, currentPage, totalPages, fromCache, loading, error,
        fetchPage, searchMangas, toggleFavorite, toggleCache,
        history, addToHistory, downloads, stats, fetchStats, online,
      }}
    >
      {children}
    </MangaContext.Provider>
  );
}

export function useManga() {
  const context = useContext(MangaContext);
  if (!context) {
    throw new Error('useManga must be used within MangaProvider');
  }
  return context;
}
