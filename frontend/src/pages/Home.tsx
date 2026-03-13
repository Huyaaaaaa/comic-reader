import { useEffect } from 'react';
import { useManga } from '../contexts/MangaContext';
import { MangaCard } from '../components/MangaCard';
import { Clock, Database, WifiOff } from 'lucide-react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';

export function Home() {
  const { mangas, history, stats, fetchStats, fetchPage, online } = useManga();

  useEffect(() => {
    fetchStats();
    if (mangas.length === 0) fetchPage(1);
  }, []);

  const recentlyViewed = history
    .slice(0, 4)
    .map((id) => mangas.find((m) => m.id === id))
    .filter(Boolean);

  const total = stats?.total_comics ?? mangas.length;
  const cached = stats?.cached_comics ?? 0;
  const covers = stats?.cover_cached ?? 0;
  const favs = stats?.favorites_count ?? 0;

  const charts = [
    { title: '漫画缓存', value: cached, color: '#3b82f6' },
    { title: '封面缓存', value: covers, color: '#10b981' },
    { title: '收藏数量', value: favs, color: '#f59e0b' },
  ];

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-3xl mb-2 text-gray-900 dark:text-white">欢迎回来</h1>
        <p className="text-gray-600 dark:text-gray-400">继续你的漫画之旅</p>
        {!online && (
          <div className="mt-3 flex items-center gap-2 text-amber-600 dark:text-amber-400">
            <WifiOff size={16} />
            <span className="text-sm">当前为离线模式，显示本地缓存数据</span>
          </div>
        )}
      </div>

      <div className="mb-8 bg-white dark:bg-gray-800 rounded-lg p-6 shadow-sm">
        <div className="flex items-center gap-2 mb-4">
          <Database className="text-blue-500" size={24} />
          <h2 className="text-xl text-gray-900 dark:text-white">本地缓存统计</h2>
        </div>

        <div className="grid grid-cols-3 gap-6 mb-6">
          {charts.map(({ title, value, color }) => {
            const data = [
              { name: title, value, color },
              { name: '其他', value: Math.max(0, total - value), color: '#e5e7eb' },
            ];
            return (
              <div key={title}>
                <h3 className="text-center mb-3 text-gray-700 dark:text-gray-300">{title}</h3>
                <ResponsiveContainer width="100%" height={200}>
                  <PieChart>
                    <Pie data={data} cx="50%" cy="50%" innerRadius={60} outerRadius={80} dataKey="value">
                      {data.map((entry, i) => (
                        <Cell key={i} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
                <p className="text-center text-sm text-gray-600 dark:text-gray-400">
                  {value} / {total}
                </p>
              </div>
            );
          })}
        </div>

        <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
          <div className="flex justify-center gap-8 text-gray-900 dark:text-white">
            <span>本地漫画总数：<span className="font-medium text-blue-500">{total}</span></span>
            <span>标签数：<span className="font-medium text-green-500">{stats?.total_tags ?? 0}</span></span>
            <span>阅读历史：<span className="font-medium text-amber-500">{stats?.history_count ?? history.length}</span></span>
          </div>
        </div>
      </div>

      {recentlyViewed.length > 0 && (
        <section className="mb-8">
          <div className="flex items-center gap-2 mb-4">
            <Clock className="text-blue-500" size={24} />
            <h2 className="text-xl text-gray-900 dark:text-white">最近阅读</h2>
          </div>
          <div className="grid grid-cols-4 gap-4">
            {recentlyViewed.map((manga) => (
              <MangaCard key={manga!.id} manga={manga!} />
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
