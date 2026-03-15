import { useState, useEffect, useRef, useCallback } from 'react';
import { useTheme } from '../contexts/ThemeContext';
import { CacheStrategy, ViewMode } from '../types';
import * as api from '../api';
import {
  Sun, Moon, Plus, Trash2, TestTube, Globe, Upload, Download,
  RefreshCw, ChevronDown, ChevronUp, Loader2, Save, Check, AlertCircle,
} from 'lucide-react';
interface CacheLevelSettings {
  strategy: CacheStrategy;
  count: number;
}

export function SettingsPage() {
  const { theme, toggleTheme } = useTheme();

  // === 设置加载状态 ===
  const [settingsLoaded, setSettingsLoaded] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveStatus, setSaveStatus] = useState<'idle' | 'pending' | 'saving' | 'saved' | 'error'>('idle');
  const saveTimerRef = useRef<ReturnType<typeof setTimeout>>();
  const saveStatusTimerRef = useRef<ReturnType<typeof setTimeout>>();
  const hasHydratedSettingsRef = useRef(false);

  // === 缓存策略 ===
  const [metadataCache, setMetadataCache] = useState<CacheLevelSettings>({ strategy: 'passive', count: 500 });
  const [extendedCache, setExtendedCache] = useState<CacheLevelSettings>({ strategy: 'passive', count: 500 });
  const [contentCache, setContentCache] = useState<CacheLevelSettings>({ strategy: 'passive', count: 100 });

  // === 阅读模式 ===
  const [readingMode, setReadingMode] = useState<ViewMode>('waterfall');

  // === 多源站 ===
  const [sources, setSources] = useState<api.SourceSite[]>([]);
  const [newSourceUrl, setNewSourceUrl] = useState('');
  const [newSourceName, setNewSourceName] = useState('');
  const [sourcesExpanded, setSourcesExpanded] = useState(false);
  const [sourcesLoading, setSourcesLoading] = useState(false);

  // 加载源站列表
  useEffect(() => {
    if (sourcesExpanded) {
      setSourcesLoading(true);
      api.fetchSources().then((data) => {
        setSources(data);
        setSourcesLoading(false);
      }).catch(() => {
        setSourcesLoading(false);
      });
    }
  }, [sourcesExpanded]);

  // === 更新 ===
  const [contentUpdateMode, setContentUpdateMode] = useState<'manual' | 'startup' | 'interval'>('startup');
  const [updateInterval, setUpdateInterval] = useState(30);
  const [appAutoCheck, setAppAutoCheck] = useState(true);
  const [updatesExpanded, setUpdatesExpanded] = useState(false);

  // === 导入导出 ===
  const [importExportExpanded, setImportExportExpanded] = useState(false);
  const [exportJobId, setExportJobId] = useState<number | null>(null);
  const [exportStatus, setExportStatus] = useState<string | null>(null);
  const [importStatus, setImportStatus] = useState<string | null>(null);
  const [importScanResult, setImportScanResult] = useState<Awaited<ReturnType<typeof api.scanImport>> | null>(null);
  const [importFilePath, setImportFilePath] = useState('');
  const importFileRef = useRef<HTMLInputElement>(null);

  // === 更新检查 ===
  const [checkingUpdate, setCheckingUpdate] = useState(false);
  const [updateResult, setUpdateResult] = useState<{ new_comics: number } | null>(null);

  // === 存储 ===
  const [storageStats, setStorageStats] = useState<api.StorageStats | null>(null);
  const [clearingCache, setClearingCache] = useState(false);

  // === 从后端加载设置 ===
  useEffect(() => {
    api.getUserSettings().then((settings) => {
      if (settings.cache_l1_strategy) setMetadataCache({ strategy: settings.cache_l1_strategy as CacheStrategy, count: parseInt(settings.cache_l1_count) || 500 });
      if (settings.cache_l2_strategy) setExtendedCache({ strategy: settings.cache_l2_strategy as CacheStrategy, count: parseInt(settings.cache_l2_count) || 500 });
      if (settings.cache_l3_strategy) setContentCache({ strategy: settings.cache_l3_strategy as CacheStrategy, count: parseInt(settings.cache_l3_count) || 100 });
      if (settings.reading_mode) setReadingMode(settings.reading_mode as ViewMode);
      if (settings.content_update_mode) setContentUpdateMode(settings.content_update_mode as 'manual' | 'startup' | 'interval');
      if (settings.update_interval) setUpdateInterval(parseInt(settings.update_interval) || 30);
      if (settings.app_auto_check !== undefined) setAppAutoCheck(settings.app_auto_check === 'true');
      setSettingsLoaded(true);
    }).catch(() => {
      setSettingsLoaded(true);
    });
  }, []);

  // === 自动保存设置（防抖 800ms）===
  const saveSettings = useCallback(() => {
    if (!settingsLoaded) return;
    if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    if (saveStatusTimerRef.current) clearTimeout(saveStatusTimerRef.current);
    setSaveStatus('pending');
    saveTimerRef.current = setTimeout(async () => {
      setSaving(true);
      setSaveStatus('saving');
      try {
        await api.updateUserSettings({
          cache_l1_strategy: metadataCache.strategy,
          cache_l1_count: String(metadataCache.count),
          cache_l2_strategy: extendedCache.strategy,
          cache_l2_count: String(extendedCache.count),
          cache_l3_strategy: contentCache.strategy,
          cache_l3_count: String(contentCache.count),
          reading_mode: readingMode,
          content_update_mode: contentUpdateMode,
          update_interval: String(updateInterval),
          app_auto_check: String(appAutoCheck),
        });
        setSaveStatus('saved');
        saveStatusTimerRef.current = setTimeout(() => setSaveStatus('idle'), 2000);
      } catch (e) {
        console.error('保存设置失败:', e);
        setSaveStatus('error');
        saveStatusTimerRef.current = setTimeout(() => setSaveStatus('idle'), 4000);
      } finally {
        setSaving(false);
      }
    }, 800);
  }, [settingsLoaded, metadataCache, extendedCache, contentCache, readingMode, contentUpdateMode, updateInterval, appAutoCheck]);

  useEffect(() => {
    if (!settingsLoaded) {
      return;
    }

    if (!hasHydratedSettingsRef.current) {
      hasHydratedSettingsRef.current = true;
      return;
    }

    saveSettings();
    return () => {
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
      if (saveStatusTimerRef.current) clearTimeout(saveStatusTimerRef.current);
    };
  }, [settingsLoaded, metadataCache, extendedCache, contentCache, readingMode, contentUpdateMode, updateInterval, appAutoCheck, saveSettings]);

  // === 导出处理 ===
  const handleExport = async (scope: string) => {
    try {
      setExportStatus('running');
      const { job_id } = await api.createExport(scope);
      setExportJobId(job_id);
      // 轮询状态
      const poll = setInterval(async () => {
        try {
          const status = await api.getExportStatus(job_id);
          if (status.status === 'completed') {
            clearInterval(poll);
            setExportStatus('completed');
          } else if (status.status === 'failed') {
            clearInterval(poll);
            setExportStatus('failed');
          }
        } catch { clearInterval(poll); setExportStatus('failed'); }
      }, 2000);
    } catch {
      setExportStatus('failed');
    }
  };

  // === 导入处理 ===
  const handleImportFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setImportStatus('scanning');
    try {
      const uploaded = await api.uploadImportFile(file);
      setImportFilePath(uploaded.file_path);
      setImportScanResult(null);

      const result = await api.scanImport(uploaded.file_path);
      setImportScanResult(result);
      setImportStatus('scanned');
    } catch {
      setImportStatus(null);
    }
    e.target.value = '';
  };

  const handleImportExecute = async (strategy: string) => {
    setImportStatus('importing');
    try {
      const { job_id } = await api.executeImport(importFilePath, strategy);
      const poll = setInterval(async () => {
        try {
          const status = await api.getImportStatus(job_id);
          if (status.status === 'completed') {
            clearInterval(poll);
            setImportStatus('done');
            setImportScanResult(null);
          } else if (status.status === 'failed') {
            clearInterval(poll);
            setImportStatus(null);
          }
        } catch { clearInterval(poll); setImportStatus(null); }
      }, 2000);
    } catch {
      setImportStatus(null);
    }
  };

  // === 更新检查 ===
  const handleCheckContentUpdate = async () => {
    setCheckingUpdate(true);
    setUpdateResult(null);
    try {
      const result = await api.checkContentUpdate(3);
      setUpdateResult({ new_comics: result.new_comics });
    } catch {
      // ignore
    } finally {
      setCheckingUpdate(false);
    }
  };

  // === 存储管理 ===
  const loadStorageStats = async () => {
    try {
      const stats = await api.getStorageStats();
      setStorageStats(stats);
    } catch { /* ignore */ }
  };

  const handleClearCache = async () => {
    setClearingCache(true);
    try {
      await api.clearStorage();
      await loadStorageStats();
    } catch { /* ignore */ }
    finally { setClearingCache(false); }
  };

  // 加载存储统计
  useEffect(() => { loadStorageStats(); }, []);

  const addSource = async () => {
    if (!newSourceUrl.trim()) return;
    try {
      const newSource = await api.addSource(
        newSourceUrl.trim(),
        newSourceName.trim() || '新源站',
        ''
      );
      setSources((prev) => [...prev, newSource]);
      setNewSourceUrl('');
      setNewSourceName('');
    } catch (error) {
      console.error('添加源站失败:', error);
    }
  };

  const removeSource = async (id: number) => {
    try {
      await api.deleteSource(id);
      setSources((prev) => prev.filter((s) => s.id !== id));
    } catch (error) {
      console.error('删除源站失败:', error);
    }
  };

  const testSource = async (id: number) => {
    try {
      await api.checkSourceHealth(id);
      // 重新获取源站列表以更新状态
      const data = await api.fetchSources();
      setSources(data);
    } catch (error) {
      console.error('测试源站失败:', error);
    }
  };

  const CacheSettingRow = ({ title, description, settings, onChange }: {
    title: string; description: string; settings: CacheLevelSettings;
    onChange: (s: CacheLevelSettings) => void;
  }) => (
    <div className="py-4 border-b border-gray-200 dark:border-gray-700 last:border-0">
      <div className="mb-3">
        <h3 className="font-medium text-gray-900 dark:text-white">{title}</h3>
        <p className="text-sm text-gray-500 dark:text-gray-400">{description}</p>
      </div>
      <div className="flex items-center gap-4">
        <select value={settings.strategy} onChange={(e) => onChange({ ...settings, strategy: e.target.value as CacheStrategy })}
          className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
          <option value="none">不缓存</option>
          <option value="passive">被动缓存</option>
          <option value="active">主动缓存最近的 x 部</option>
          <option value="all">全部缓存</option>
        </select>
        {settings.strategy === 'active' && (
          <div className="flex items-center gap-2">
            <input type="number" min="1" value={settings.count}
              onChange={(e) => onChange({ ...settings, count: parseInt(e.target.value) || 1 })}
              className="w-20 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
            <span className="text-sm text-gray-600 dark:text-gray-400">部</span>
          </div>
        )}
      </div>
    </div>
  );

  const SectionToggle = ({ title, expanded, onToggle, children }: {
    title: string; expanded: boolean; onToggle: () => void; children: React.ReactNode;
  }) => (
    <section className="bg-white dark:bg-gray-800 rounded-lg shadow-sm overflow-hidden">
      <button onClick={onToggle} className="w-full flex items-center justify-between p-6 hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors">
        <h2 className="text-xl text-gray-900 dark:text-white">{title}</h2>
        {expanded ? <ChevronUp size={20} className="text-gray-500" /> : <ChevronDown size={20} className="text-gray-500" />}
      </button>
      {expanded && <div className="px-6 pb-6 border-t border-gray-200 dark:border-gray-700 pt-4">{children}</div>}
    </section>
  );

  const saveFeedback = {
    pending: {
      icon: <Save size={18} className="text-blue-600 dark:text-blue-400" />,
      title: '检测到设置变更',
      description: '正在等待自动保存，缓存策略会按最新设置重新应用。',
      className: 'border-blue-200 bg-blue-50 text-blue-900 dark:border-blue-900/60 dark:bg-blue-950/30 dark:text-blue-100',
    },
    saving: {
      icon: <Loader2 size={18} className="animate-spin text-blue-600 dark:text-blue-400" />,
      title: '正在保存设置',
      description: '新的缓存策略正在提交，旧的主动缓存轮次会尽快停止。',
      className: 'border-blue-200 bg-blue-50 text-blue-900 dark:border-blue-900/60 dark:bg-blue-950/30 dark:text-blue-100',
    },
    saved: {
      icon: <Check size={18} className="text-green-600 dark:text-green-400" />,
      title: '设置已保存',
      description: '当前页面配置已经写入后端，新的缓存策略已开始按最新设置执行。',
      className: 'border-green-200 bg-green-50 text-green-900 dark:border-green-900/60 dark:bg-green-950/30 dark:text-green-100',
    },
    error: {
      icon: <AlertCircle size={18} className="text-red-600 dark:text-red-400" />,
      title: '保存失败',
      description: '这次更改还没有成功写入后端，请检查服务状态后重试。',
      className: 'border-red-200 bg-red-50 text-red-900 dark:border-red-900/60 dark:bg-red-950/30 dark:text-red-100',
    },
  } as const;

  const currentSaveFeedback = saveStatus === 'idle' ? null : saveFeedback[saveStatus];

  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-3xl text-gray-900 dark:text-white">设置</h1>
      </div>

      {currentSaveFeedback && (
        <div className={`fixed right-6 top-6 z-50 flex w-[min(420px,calc(100vw-3rem))] items-start gap-3 rounded-lg border px-4 py-3 shadow-lg ${currentSaveFeedback.className}`}>
          <div className="mt-0.5 flex-shrink-0">{currentSaveFeedback.icon}</div>
          <div>
            <p className="font-medium">{currentSaveFeedback.title}</p>
            <p className="text-sm opacity-90">{currentSaveFeedback.description}</p>
          </div>
        </div>
      )}

      <div className="max-w-2xl space-y-6">
        {saving && saveStatus === 'saving' && (
          <div className="sr-only" aria-live="polite">设置保存中</div>
        )}
        {/* 外观 */}
        <section className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-sm">
          <h2 className="text-xl mb-4 text-gray-900 dark:text-white">外观</h2>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-900 dark:text-white">主题模式</p>
              <p className="text-sm text-gray-500 dark:text-gray-400">切换黑夜或白天模式</p>
            </div>
            <button onClick={toggleTheme} className="flex items-center gap-2 px-4 py-2 rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors">
              {theme === 'light' ? <><Moon size={20} /><span>黑夜模式</span></> : <><Sun size={20} /><span>白天模式</span></>}
            </button>
          </div>
        </section>

        {/* 阅读设置 */}
        <section className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-sm">
          <h2 className="text-xl mb-4 text-gray-900 dark:text-white">阅读设置</h2>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-900 dark:text-white">默认阅读模式</p>
              <p className="text-sm text-gray-500 dark:text-gray-400">选择默认的漫画阅读模式</p>
            </div>
            <select value={readingMode} onChange={(e) => setReadingMode(e.target.value as ViewMode)}
              className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
              <option value="waterfall">瀑布长图流</option>
              <option value="single">单图流</option>
            </select>
          </div>
        </section>

        {/* 缓存策略 */}
        <section className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-sm">
          <h2 className="text-xl mb-4 text-gray-900 dark:text-white">缓存策略设置</h2>
          <p className="text-sm text-gray-600 dark:text-gray-400 mb-6">分别设置三层缓存的策略，优化存储空间和加载速度</p>
          <div className="space-y-2">
            <CacheSettingRow title="L1 元数据" description="包含漫画标题、作者、标签等基本信息" settings={metadataCache} onChange={setMetadataCache} />
            <CacheSettingRow title="L2 拓展信息" description="包含漫画封面、描述等详细信息" settings={extendedCache} onChange={setExtendedCache} />
            <CacheSettingRow title="L3 正文图片" description="包含漫画的所有章节图片内容" settings={contentCache} onChange={setContentCache} />
          </div>
          <div className="mt-6 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
            <p className="text-sm text-blue-800 dark:text-blue-200">
              提示：L1 体积最小但访问最频繁，建议全部缓存；L3 体积最大，建议按需缓存。
            </p>
          </div>
        </section>

        {/* 多源站管理 */}
        <SectionToggle title="多源站管理" expanded={sourcesExpanded} onToggle={() => setSourcesExpanded(!sourcesExpanded)}>
          <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">管理漫画数据源站地址，支持故障自动切换</p>

          {sourcesLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="animate-spin text-blue-500" size={24} />
            </div>
          ) : (
            <>
              <div className="space-y-2 mb-4">
                <input type="text" value={newSourceName} onChange={(e) => setNewSourceName(e.target.value)}
                  placeholder="源站名称（可选）"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-400" />
                <div className="flex gap-2">
                  <input type="text" value={newSourceUrl} onChange={(e) => setNewSourceUrl(e.target.value)}
                    placeholder="输入源站地址，例如 https://example.com"
                    className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-400"
                    onKeyDown={(e) => { if (e.key === 'Enter') addSource(); }} />
                  <button onClick={addSource} className="flex items-center gap-1 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors">
                    <Plus size={16} /><span>添加</span>
                  </button>
                </div>
              </div>

              {sources.length === 0 ? (
                <p className="text-gray-500 dark:text-gray-400 text-sm py-4 text-center">暂无源站，请添加至少一个源站地址</p>
              ) : (
                <div className="space-y-2">
                  {sources.map((source, index) => (
                    <div key={source.id} className="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-750 rounded-lg">
                      <span className="text-sm text-gray-500 w-6">{index + 1}</span>
                      <Globe size={16} className="text-gray-400 flex-shrink-0" />
                      <div className="flex-1 min-w-0">
                        <div className="text-sm text-gray-900 dark:text-white truncate">{source.name}</div>
                        <div className="text-xs text-gray-500 dark:text-gray-400 truncate">{source.url}</div>
                      </div>
                      <span className={`text-xs px-2 py-1 rounded ${source.status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'}`}>
                        {source.status === 'active' ? (source.latency > 0 ? `${source.latency}ms` : '在线') : '未知'}
                      </span>
                      <button onClick={() => testSource(source.id)} className="p-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors" title="测速">
                        <TestTube size={16} className="text-blue-500" />
                      </button>
                      <button onClick={() => removeSource(source.id)} className="p-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors" title="删除">
                        <Trash2 size={16} className="text-red-500" />
                      </button>
                    </div>
                  ))}
                </div>
              )}

              <div className="mt-4 p-3 bg-gray-50 dark:bg-gray-750 rounded-lg">
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  心跳检测间隔：60 分钟 | 单次请求自动重试 3 次 | 连续 3 次失败自动切换
                </p>
              </div>
            </>
          )}
        </SectionToggle>

        {/* 更新设置 */}
        <SectionToggle title="更新设置" expanded={updatesExpanded} onToggle={() => setUpdatesExpanded(!updatesExpanded)}>
          <div className="space-y-6">
            <div>
              <h3 className="font-medium text-gray-900 dark:text-white mb-3">内容更新</h3>
              <div className="space-y-3">
                {(['manual', 'startup', 'interval'] as const).map((mode) => (
                  <label key={mode} className="flex items-center gap-3 cursor-pointer">
                    <input type="radio" name="contentUpdate" value={mode} checked={contentUpdateMode === mode}
                      onChange={() => setContentUpdateMode(mode)}
                      className="w-4 h-4 text-blue-500" />
                    <span className="text-gray-900 dark:text-white">
                      {mode === 'manual' && '手动检查'}
                      {mode === 'startup' && '启动时检查'}
                      {mode === 'interval' && '定时检查'}
                    </span>
                  </label>
                ))}
                {contentUpdateMode === 'interval' && (
                  <div className="ml-7 flex items-center gap-2">
                    <span className="text-sm text-gray-600 dark:text-gray-400">间隔：</span>
                    <input type="number" min="5" value={updateInterval}
                      onChange={(e) => setUpdateInterval(parseInt(e.target.value) || 30)}
                      className="w-20 px-3 py-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
                    <span className="text-sm text-gray-600 dark:text-gray-400">分钟</span>
                  </div>
                )}
              </div>
            </div>

            <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
              <h3 className="font-medium text-gray-900 dark:text-white mb-3">应用更新</h3>
              <label className="flex items-center gap-3 cursor-pointer">
                <input type="checkbox" checked={appAutoCheck} onChange={(e) => setAppAutoCheck(e.target.checked)}
                  className="w-4 h-4 text-blue-500 rounded" />
                <span className="text-gray-900 dark:text-white">启动时自动检查更新</span>
              </label>
              <button onClick={handleCheckContentUpdate}
                disabled={checkingUpdate}
                className="mt-3 flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors disabled:opacity-50">
                {checkingUpdate ? <Loader2 size={16} className="animate-spin" /> : <RefreshCw size={16} />}
                <span>{updateResult ? `发现 ${updateResult.new_comics} 部新漫画` : '立即检查更新'}</span>
              </button>
            </div>
          </div>
        </SectionToggle>

        {/* 导入导出 */}
        <SectionToggle title="导入导出" expanded={importExportExpanded} onToggle={() => setImportExportExpanded(!importExportExpanded)}>
          <div className="space-y-4">
            <div>
              <h3 className="font-medium text-gray-900 dark:text-white mb-2">导出</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400 mb-3">导出本地缓存数据以备份或迁移</p>
              <div className="flex flex-wrap gap-2">
                {([['all_cached_comics', '全部已缓存漫画'], ['all_covers', '所有封面'], ['all_images', '所有图片']] as const).map(([scope, label]) => (
                  <button key={scope} onClick={() => handleExport(scope)}
                    disabled={exportStatus !== null && exportStatus !== 'completed' && exportStatus !== 'failed'}
                    className="flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors disabled:opacity-50">
                    {exportStatus === 'running' ? <Loader2 size={16} className="animate-spin" /> : <Upload size={16} />}
                    <span>{label}</span>
                  </button>
                ))}
              </div>
              {exportStatus === 'completed' && exportJobId && (
                <a href={api.getExportDownloadUrl(exportJobId)}
                  className="mt-3 inline-flex items-center gap-2 px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors">
                  <Download size={16} /><span>下载导出文件</span>
                </a>
              )}
              {exportStatus === 'failed' && (
                <p className="mt-2 text-sm text-red-500">导出失败，请重试</p>
              )}
            </div>

            <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
              <h3 className="font-medium text-gray-900 dark:text-white mb-2">导入</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400 mb-3">从备份文件恢复数据</p>
              <input type="file" accept=".cpack,.zip" ref={importFileRef} className="hidden"
                onChange={handleImportFileSelect} />
              <button onClick={() => importFileRef.current?.click()}
                disabled={importStatus === 'scanning' || importStatus === 'importing'}
                className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50">
                {importStatus === 'scanning' ? <Loader2 size={16} className="animate-spin" /> : <Download size={16} />}
                <span>选择文件导入</span>
              </button>
              {importScanResult && importStatus === 'scanned' && (
                <div className="mt-3 p-3 bg-gray-50 dark:bg-gray-750 rounded-lg">
                  <p className="text-sm text-gray-700 dark:text-gray-300 mb-2">
                    新增 {importScanResult.conflicts.will_add} 部，覆盖 {importScanResult.conflicts.will_overwrite} 部
                  </p>
                  <div className="flex gap-2">
                    <button onClick={() => handleImportExecute('overwrite')}
                      className="px-3 py-1 bg-red-500 text-white rounded text-sm hover:bg-red-600">覆盖导入</button>
                    <button onClick={() => handleImportExecute('skip_existing')}
                      className="px-3 py-1 bg-blue-500 text-white rounded text-sm hover:bg-blue-600">仅导入缺失</button>
                    <button onClick={() => { setImportScanResult(null); setImportStatus(null); }}
                      className="px-3 py-1 bg-gray-300 dark:bg-gray-600 text-gray-900 dark:text-white rounded text-sm">取消</button>
                  </div>
                </div>
              )}
              {importStatus === 'importing' && (
                <p className="mt-2 text-sm text-blue-500 flex items-center gap-2">
                  <Loader2 size={14} className="animate-spin" />导入中...
                </p>
              )}
              {importStatus === 'done' && (
                <p className="mt-2 text-sm text-green-500">导入完成</p>
              )}
            </div>
          </div>
        </SectionToggle>

        {/* 存储 */}
        <section className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-sm">
          <h2 className="text-xl mb-4 text-gray-900 dark:text-white">存储</h2>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-900 dark:text-white">缓存大小</p>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  当前使用 {storageStats ? `${storageStats.total_size_mb} MB` : '-- MB'}
                  {storageStats && (
                    <span className="ml-2">
                      (数据库 {storageStats.db_size_mb}MB / 下载 {storageStats.download_size_mb}MB)
                    </span>
                  )}
                </p>
                {storageStats && (
                  <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">
                    L1: {storageStats.l1_count} 条 | L2: {storageStats.l2_count} 条 | L3: {storageStats.l3_count} 条 ({storageStats.l3_size_mb}MB)
                  </p>
                )}
              </div>
              <button onClick={handleClearCache} disabled={clearingCache}
                className="px-4 py-2 rounded-lg bg-red-500 text-white hover:bg-red-600 transition-colors disabled:opacity-50">
                {clearingCache ? <Loader2 size={16} className="animate-spin" /> : '清除缓存'}
              </button>
            </div>
          </div>
        </section>

        {/* 关于 */}
        <section className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-sm">
          <h2 className="text-xl mb-4 text-gray-900 dark:text-white">关于</h2>
          <div className="space-y-2 text-sm text-gray-600 dark:text-gray-400">
            <p>版本：1.0.0</p>
            <p>Comic Go - 漫画离线浏览系统</p>
          </div>
        </section>
      </div>
    </div>
  );
}
