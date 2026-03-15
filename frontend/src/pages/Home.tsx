import { useCallback, useEffect, useMemo, useState } from 'react';
import { Clock, Database, PencilLine, Tags, WifiOff } from 'lucide-react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';
import { fetchComicDetail } from '../api';
import client from '../api/client';
import { MangaCard } from '../components/MangaCard';
import { useManga } from '../contexts/MangaContext';
import { useSSE } from '../hooks/useSSE';
import { Manga } from '../types';

type HomeSectionKey = 'welcome' | 'charts' | 'summary' | 'recentTags' | 'recentlyViewed';

type HomeVisibility = Record<HomeSectionKey, boolean>;

interface RecentTagStat {
  tagName: string;
  count: number;
}

const HOME_VISIBILITY_STORAGE_KEY = 'comic-home-visibility';

const DEFAULT_HOME_VISIBILITY: HomeVisibility = {
  welcome: true,
  charts: true,
  summary: true,
  recentTags: true,
  recentlyViewed: true,
};

const HOME_SECTION_OPTIONS: Array<{ key: HomeSectionKey; label: string; description: string }> = [
  { key: 'welcome', label: '欢迎信息', description: '首页顶部的欢迎文案和离线提示' },
  { key: 'charts', label: '环形图统计', description: '缓存、封面和收藏数量环形图' },
  { key: 'summary', label: '概览指标', description: '漫画总数、标签数和阅读历史汇总' },
  { key: 'recentTags', label: '最近观看标签', description: '最近 10 条阅读记录的标签 Top 10' },
  { key: 'recentlyViewed', label: '最近阅读', description: '首页底部的最近阅读卡片' },
];

function loadHomeVisibility(): HomeVisibility {
  if (typeof window === 'undefined') {
    return DEFAULT_HOME_VISIBILITY;
  }

  try {
    const raw = window.localStorage.getItem(HOME_VISIBILITY_STORAGE_KEY);
    if (!raw) {
      return DEFAULT_HOME_VISIBILITY;
    }

    const parsed = JSON.parse(raw) as Partial<HomeVisibility>;
    return {
      welcome: typeof parsed.welcome === 'boolean' ? parsed.welcome : DEFAULT_HOME_VISIBILITY.welcome,
      charts: typeof parsed.charts === 'boolean' ? parsed.charts : DEFAULT_HOME_VISIBILITY.charts,
      summary: typeof parsed.summary === 'boolean' ? parsed.summary : DEFAULT_HOME_VISIBILITY.summary,
      recentTags: typeof parsed.recentTags === 'boolean' ? parsed.recentTags : DEFAULT_HOME_VISIBILITY.recentTags,
      recentlyViewed:
        typeof parsed.recentlyViewed === 'boolean'
          ? parsed.recentlyViewed
          : DEFAULT_HOME_VISIBILITY.recentlyViewed,
    };
  } catch {
    return DEFAULT_HOME_VISIBILITY;
  }
}

export function Home() {
  const { stats, fetchStats, fetchPage, online } = useManga();
  const [recentlyViewed, setRecentlyViewed] = useState<Manga[]>([]);
  const [recentTagStats, setRecentTagStats] = useState<RecentTagStat[]>([]);
  const [editingLayout, setEditingLayout] = useState(false);
  const [visibility, setVisibility] = useState<HomeVisibility>(() => loadHomeVisibility());

  useEffect(() => {
    let cancelled = false;

    const loadHomeData = async () => {
      await Promise.allSettled([fetchStats(), fetchPage(1)]);

      try {
        const { data } = await client.get('/history');
        const items = Array.isArray(data?.items) ? data.items : [];

        if (cancelled) {
          return;
        }

        const converted = items.slice(0, 4).map((item: any) => ({
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
        setRecentlyViewed(converted);

        const recentIds: number[] = Array.from(
          new Set(
            items
              .slice(0, 10)
              .map((item: any) => Number(item.comic_id))
              .filter((id: number) => Number.isFinite(id))
          )
        );

        if (recentIds.length === 0) {
          setRecentTagStats([]);
          return;
        }

        const detailResults = await Promise.allSettled(
          recentIds.map((comicId) => fetchComicDetail(comicId))
        );

        if (cancelled) {
          return;
        }

        const tagCounter = new Map<string, number>();

        detailResults.forEach((result) => {
          if (result.status !== 'fulfilled') {
            return;
          }

          const uniqueTags = new Set<string>();
          result.value.tags?.forEach((tag) => {
            const tagName = tag.tag_name?.trim();
            if (!tagName || uniqueTags.has(tagName)) {
              return;
            }

            uniqueTags.add(tagName);
            tagCounter.set(tagName, (tagCounter.get(tagName) ?? 0) + 1);
          });
        });

        const sortedTags = Array.from(tagCounter.entries())
          .sort((a, b) => {
            if (b[1] !== a[1]) {
              return b[1] - a[1];
            }

            return a[0].localeCompare(b[0], 'zh-Hans-CN');
          })
          .slice(0, 10)
          .map(([tagName, count]) => ({ tagName, count }));

        setRecentTagStats(sortedTags);
      } catch {
        if (!cancelled) {
          setRecentlyViewed([]);
          setRecentTagStats([]);
        }
      }
    };

    loadHomeData();

    return () => {
      cancelled = true;
    };
  }, [fetchPage, fetchStats]);

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }

    window.localStorage.setItem(HOME_VISIBILITY_STORAGE_KEY, JSON.stringify(visibility));
  }, [visibility]);

  const refreshStats = useCallback(() => {
    fetchStats();
  }, [fetchStats]);

  useSSE({
    'download:completed': refreshStats,
    'cache:l1_done': refreshStats,
    'cache:l2_done': refreshStats,
    'cache:l3_done': refreshStats,
  });

  const total = stats?.total_comics ?? 0;
  const l1Cached = stats?.l1_cached_count ?? stats?.cached_comics ?? 0;
  const l2Cached = stats?.l2_cached_count ?? stats?.cover_cached ?? 0;
  const l3Cached = stats?.l3_cached_count ?? 0;

  const charts = useMemo(
    () => [
      { title: 'L1 元数据', value: Math.min(l1Cached, total), color: '#3b82f6' },
      { title: 'L2 拓展信息', value: Math.min(l2Cached, total), color: '#10b981' },
      { title: 'L3 正文图片', value: Math.min(l3Cached, total), color: '#f59e0b' },
    ],
    [l1Cached, l2Cached, l3Cached, total]
  );

  const shouldShowStatsCard = visibility.charts || visibility.summary || visibility.recentTags;

  const toggleSection = (key: HomeSectionKey) => {
    setVisibility((prev) => ({
      ...prev,
      [key]: !prev[key],
    }));
  };

  return (
    <div className="p-8">
      <div className="mb-6 flex items-start justify-between gap-4">
        <div>
          {visibility.welcome ? (
            <>
              <h1 className="mb-2 text-3xl text-gray-900 dark:text-white">欢迎回来</h1>
              <p className="text-gray-600 dark:text-gray-400">继续你的漫画之旅</p>
            </>
          ) : (
            <>
              <h1 className="mb-2 text-3xl text-gray-900 dark:text-white">首页概览</h1>
              <p className="text-gray-600 dark:text-gray-400">按你的阅读习惯安排首页信息</p>
            </>
          )}

          {!online && (
            <div className="mt-3 flex items-center gap-2 text-amber-600 dark:text-amber-400">
              <WifiOff size={16} />
              <span className="text-sm">当前为离线模式，显示本地缓存数据</span>
            </div>
          )}
        </div>

        <button
          type="button"
          onClick={() => setEditingLayout((prev) => !prev)}
          className="inline-flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-4 py-2 text-sm text-gray-700 shadow-sm transition hover:border-blue-400 hover:text-blue-600 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-200 dark:hover:border-blue-500 dark:hover:text-blue-400"
        >
          <PencilLine size={16} />
          {editingLayout ? '完成编辑' : '编辑首页'}
        </button>
      </div>

      {editingLayout && (
        <div className="mb-8 rounded-xl border border-blue-100 bg-blue-50/80 p-4 dark:border-blue-900/60 dark:bg-blue-950/30">
          <div className="mb-3">
            <h2 className="text-base text-gray-900 dark:text-white">首页显示设置</h2>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              仅保存在当前浏览器，本次修改刷新后依然有效
            </p>
          </div>

          <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
            {HOME_SECTION_OPTIONS.map((option) => {
              const enabled = visibility[option.key];

              return (
                <button
                  key={option.key}
                  type="button"
                  onClick={() => toggleSection(option.key)}
                  className={`rounded-xl border px-4 py-3 text-left transition ${
                    enabled
                      ? 'border-blue-200 bg-white shadow-sm dark:border-blue-700 dark:bg-gray-900'
                      : 'border-gray-200 bg-gray-50 text-gray-500 dark:border-gray-700 dark:bg-gray-900/40 dark:text-gray-400'
                  }`}
                >
                  <div className="mb-2 flex items-center justify-between gap-3">
                    <span className="text-sm font-medium text-gray-900 dark:text-white">{option.label}</span>
                    <span
                      className={`rounded-full px-2.5 py-1 text-xs ${
                        enabled
                          ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/60 dark:text-blue-300'
                          : 'bg-gray-200 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
                      }`}
                    >
                      {enabled ? '显示中' : '已隐藏'}
                    </span>
                  </div>
                  <p className="text-sm text-gray-600 dark:text-gray-400">{option.description}</p>
                </button>
              );
            })}
          </div>
        </div>
      )}

      {shouldShowStatsCard && (
        <div className="mb-8 rounded-lg bg-white p-6 shadow-sm dark:bg-gray-800">
          <div className="mb-4 flex items-center gap-2">
            <Database className="text-blue-500" size={24} />
            <h2 className="text-xl text-gray-900 dark:text-white">本地缓存统计</h2>
          </div>

          {visibility.charts && (
            <div className="mb-6 grid grid-cols-3 gap-6">
              {charts.map(({ title, value, color }) => {
                const data = [
                  { name: title, value, color },
                  { name: '其他', value: Math.max(0, total - value), color: '#e5e7eb' },
                ];

                return (
                  <div key={title}>
                    <h3 className="mb-3 text-center text-gray-700 dark:text-gray-300">{title}</h3>
                    <ResponsiveContainer width="100%" height={200}>
                      <PieChart>
                        <Pie data={data} cx="50%" cy="50%" innerRadius={60} outerRadius={80} dataKey="value">
                          {data.map((entry, index) => (
                            <Cell key={index} fill={entry.color} />
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
          )}

          {visibility.summary && (
            <div className={`pt-4 ${visibility.charts ? 'border-t border-gray-200 dark:border-gray-700' : ''}`}>
              <div className="flex flex-col items-center gap-3 text-gray-900 dark:text-white md:flex-row md:justify-center md:gap-8">
                <span>
                  本地漫画总数：<span className="font-medium text-blue-500">{total}</span>
                </span>
                <span>
                  标签数：<span className="font-medium text-green-500">{stats?.total_tags ?? 0}</span>
                </span>
                <span>
                  阅读历史：<span className="font-medium text-amber-500">{stats?.history_count ?? 0}</span>
                </span>
              </div>
            </div>
          )}

          {visibility.recentTags && (
            <div
              className={`pt-5 ${
                visibility.charts || visibility.summary ? 'mt-5 border-t border-gray-200 dark:border-gray-700' : ''
              }`}
            >
              <div className="mb-4 flex items-center gap-2">
                <Tags className="text-blue-500" size={20} />
                <div>
                  <h3 className="text-lg text-gray-900 dark:text-white">最近观看标签 Top 10</h3>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    基于最近 10 条阅读记录统计每个标签对应的已浏览漫画数
                  </p>
                </div>
              </div>

              {recentTagStats.length > 0 ? (
                <div className="flex flex-wrap gap-3">
                  {recentTagStats.map((tag) => (
                    <div
                      key={tag.tagName}
                      className="rounded-full bg-[#0d6efd] px-4 py-2 text-sm text-white transition-colors hover:bg-[#0b5ed7]"
                    >
                      <span className="max-w-56 truncate">
                        {tag.tagName}（{tag.count}）
                      </span>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="rounded-xl border border-dashed border-gray-200 px-4 py-6 text-sm text-gray-500 dark:border-gray-700 dark:text-gray-400">
                  最近阅读里还没有足够的标签信息，等你再看几部漫画后这里会自动出现。
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {visibility.recentlyViewed && recentlyViewed.length > 0 && (
        <section className="mb-8">
          <div className="mb-4 flex items-center gap-2">
            <Clock className="text-blue-500" size={24} />
            <h2 className="text-xl text-gray-900 dark:text-white">最近阅读</h2>
          </div>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
            {recentlyViewed.map((manga) => (
              <MangaCard key={manga.id} manga={manga} />
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
