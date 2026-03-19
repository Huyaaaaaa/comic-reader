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

type ProxyMode = 'off' | 'fallback' | 'always';
type ProxyType = 'http' | 'socks5';

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

  // === 网络代理 ===
  const [proxyMode, setProxyMode] = useState<ProxyMode>('off');
  const [proxyType, setProxyType] = useState<ProxyType>('http');
  const [proxyHost, setProxyHost] = useState('');
  const [proxyPort, setProxyPort] = useState('');
  const [proxyUsername, setProxyUsername] = useState('');
  const [proxyPassword, setProxyPassword] = useState('');
  const [proxyTimeoutMs, setProxyTimeoutMs] = useState('6000');
  const [proxyHealth, setProxyHealth] = useState<api.ProxyHealthResponse | null>(null);
  const [checkingProxyHealth, setCheckingProxyHealth] = useState(false);

  // === 多源站 ===
  const [sources, setSources] = useState<api.SourceSite[]>([]);
  const [newSourceUrl, setNewSourceUrl] = useState('');
  const [newSourceName, setNewSourceName] = useState('');
  const [releasePageUrl, setReleasePageUrl] = useState('');
  const [releaseImporting, setReleaseImporting] = useState(false);
  const [releaseImportResult, setReleaseImportResult] = useState<api.ImportReleasePageSourcesResult | null>(null);
  const [releaseImportError, setReleaseImportError] = useState<string | null>(null);
  const [sourcesExpanded, setSourcesExpanded] = useState(false);
  const [sourcesLoading, setSourcesLoading] = useState(false);
  const [testingSourceId, setTestingSourceId] = useState<number | null>(null);

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
      if (settings.network_proxy_mode) setProxyMode(settings.network_proxy_mode as ProxyMode);
      if (settings.network_proxy_type) setProxyType(settings.network_proxy_type as ProxyType);
      if (settings.network_proxy_host) setProxyHost(settings.network_proxy_host);
      if (settings.network_proxy_port) setProxyPort(settings.network_proxy_port);
      if (settings.network_proxy_username) setProxyUsername(settings.network_proxy_username);
      if (settings.network_proxy_password) setProxyPassword(settings.network_proxy_password);
      if (settings.network_proxy_timeout_ms) setProxyTimeoutMs(settings.network_proxy_timeout_ms);
      if (settings.content_update_mode) setContentUpdateMode(settings.content_update_mode as 'manual' | 'startup' | 'interval');
      if (settings.update_interval) setUpdateInterval(parseInt(settings.update_interval) || 30);
      if (settings.app_auto_check !== undefined) setAppAutoCheck(settings.app_auto_check === 'true');
      setSettingsLoaded(true);
    }).catch(() => {
      setSettingsLoaded(true);
    });
  }, []);

  const refreshProxyHealth = useCallback(async (force = false) => {
    const configured = proxyHost.trim() !== '' && proxyPort.trim() !== '';
    if (!configured) {
      setProxyHealth({
        configured: false,
        available: false,
        status: 'unconfigured',
        message: '代理未配置完整',
        latency_ms: 0,
        targets: [],
      });
      return;
    }

    setCheckingProxyHealth(true);
    try {
      const result = await api.checkProxyHealth(force);
      setProxyHealth(result);
    } catch (error) {
      console.error('检测代理失败:', error);
      setProxyHealth({
        configured: true,
        available: false,
        status: 'unavailable',
        message: error instanceof Error ? error.message : '代理检测失败',
        latency_ms: 0,
        targets: [],
      });
    } finally {
      setCheckingProxyHealth(false);
    }
  }, [proxyHost, proxyPort]);

  useEffect(() => {
    if (!settingsLoaded) {
      return;
    }
    refreshProxyHealth(false).catch(() => {});
  }, [settingsLoaded, refreshProxyHealth]);

  const buildSettingsPayload = useCallback(() => ({
    cache_l1_strategy: metadataCache.strategy,
    cache_l1_count: String(metadataCache.count),
    cache_l2_strategy: extendedCache.strategy,
    cache_l2_count: String(extendedCache.count),
    cache_l3_strategy: contentCache.strategy,
    cache_l3_count: String(contentCache.count),
    reading_mode: readingMode,
    network_proxy_mode: proxyMode,
    network_proxy_type: proxyType,
    network_proxy_host: proxyHost.trim(),
    network_proxy_port: proxyPort.trim(),
    network_proxy_username: proxyUsername.trim(),
    network_proxy_password: proxyPassword,
    network_proxy_timeout_ms: proxyTimeoutMs.trim() || '6000',
    content_update_mode: contentUpdateMode,
    update_interval: String(updateInterval),
    app_auto_check: String(appAutoCheck),
  }), [metadataCache, extendedCache, contentCache, readingMode, proxyMode, proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword, proxyTimeoutMs, contentUpdateMode, updateInterval, appAutoCheck]);

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
        await api.updateUserSettings(buildSettingsPayload());
        setSaveStatus('saved');
        refreshProxyHealth(true).catch(() => {});
        saveStatusTimerRef.current = setTimeout(() => setSaveStatus('idle'), 2000);
      } catch (e) {
        console.error('保存设置失败:', e);
        setSaveStatus('error');
        saveStatusTimerRef.current = setTimeout(() => setSaveStatus('idle'), 4000);
      } finally {
        setSaving(false);
      }
    }, 800);
  }, [settingsLoaded, buildSettingsPayload, refreshProxyHealth]);

  const flushSettingsNow = useCallback(async () => {
    if (!settingsLoaded) return;
    if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    if (saveStatusTimerRef.current) clearTimeout(saveStatusTimerRef.current);
    setSaving(true);
    setSaveStatus('saving');
    try {
      await api.updateUserSettings(buildSettingsPayload());
      setSaveStatus('saved');
      await refreshProxyHealth(true);
      saveStatusTimerRef.current = setTimeout(() => setSaveStatus('idle'), 2000);
    } catch (error) {
      console.error('立即保存设置失败:', error);
      setSaveStatus('error');
      saveStatusTimerRef.current = setTimeout(() => setSaveStatus('idle'), 4000);
      throw error;
    } finally {
      setSaving(false);
    }
  }, [settingsLoaded, buildSettingsPayload, refreshProxyHealth]);

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
  }, [settingsLoaded, metadataCache, extendedCache, contentCache, readingMode, proxyMode, proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword, proxyTimeoutMs, contentUpdateMode, updateInterval, appAutoCheck, saveSettings]);

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

  const importFromReleasePage = async () => {
    if (!releasePageUrl.trim()) return;

    setReleaseImporting(true);
    setReleaseImportError(null);
    setReleaseImportResult(null);
    try {
      const result = await api.importSourcesFromReleasePage(releasePageUrl.trim());
      setReleaseImportResult(result);
      setReleasePageUrl('');
      const data = await api.fetchSources();
      setSources(data);
    } catch (error) {
      const message = error instanceof Error ? error.message : '导入失败';
      setReleaseImportError(message);
      setReleaseImportResult(null);
    } finally {
      setReleaseImporting(false);
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
    setTestingSourceId(id);
    try {
      await flushSettingsNow();
      const result = await api.checkSourceHealth(id);
      setSources((prev) =>
        prev.map((source) =>
          source.id === id
            ? {
                ...source,
                status: result.status,
                latency: result.latency_ms,
                direct_status: result.direct_status,
                direct_latency: result.direct_latency_ms,
                direct_last_error: result.direct_error || '',
                proxy_status: result.proxy_status,
                proxy_latency: result.proxy_latency_ms,
                proxy_last_error: result.proxy_error || '',
              }
            : source
        )
      );
    } catch (error) {
      console.error('测试源站失败:', error);
    } finally {
      setTestingSourceId(null);
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
  const proxyConfigured = proxyHost.trim() !== '' && proxyPort.trim() !== '';
  const proxyUsable = proxyHealth?.available === true;

  const renderAccessBadge = (
    label: string,
    status: string | undefined,
    latency: number | undefined,
    accent: 'direct' | 'proxy',
    enabled: boolean = true,
    error?: string
  ) => {
    if (!enabled) {
      return (
        <span title={error || undefined} className="rounded-full bg-gray-100 px-2.5 py-1 text-xs text-gray-500 dark:bg-gray-700 dark:text-gray-400">
          {label} {accent === 'proxy' && proxyConfigured ? '不可用' : '未配置'}
        </span>
      );
    }

    const normalizedStatus = status || 'unknown';
    if (normalizedStatus === 'available') {
      const classes = accent === 'proxy'
        ? 'bg-amber-100 text-orange-700 dark:bg-amber-900/30 dark:text-orange-300'
        : 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400';
      const value = latency && latency > 0 ? `${latency}ms` : '在线';
      return (
        <span title={error || undefined} className={`rounded-full px-2.5 py-1 text-xs ${classes}`}>
          {label} {value}
        </span>
      );
    }

    if (normalizedStatus === 'unavailable') {
      return (
        <span title={error || undefined} className="rounded-full bg-gray-200 px-2.5 py-1 text-xs text-gray-600 dark:bg-gray-700 dark:text-gray-300">
          {label} 暂不可用
        </span>
      );
    }

    return (
      <span title={error || undefined} className="rounded-full bg-gray-100 px-2.5 py-1 text-xs text-gray-500 dark:bg-gray-700 dark:text-gray-400">
        {label} 未检测
      </span>
    );
  };

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

        <section className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-sm">
          <h2 className="text-xl mb-4 text-gray-900 dark:text-white">网络代理</h2>
          <p className="mb-6 text-sm text-gray-600 dark:text-gray-400">
            代理会先用 Google / YouTube 连通性探针检测自身是否可用。只有代理可用时，源站测速、心跳和在线抓取才会尝试使用 Proxy。
          </p>

          <div className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <label className="space-y-2">
                <span className="text-sm text-gray-700 dark:text-gray-300">代理模式</span>
                <select
                  value={proxyMode}
                  onChange={(e) => setProxyMode(e.target.value as ProxyMode)}
                  className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                >
                  <option value="off">不使用代理</option>
                  <option value="fallback">直连源站全部不可用后使用 Proxy</option>
                  <option value="always">全局通过 Proxy 访问</option>
                </select>
              </label>

              <label className="space-y-2">
                <span className="text-sm text-gray-700 dark:text-gray-300">代理类型</span>
                <select
                  value={proxyType}
                  onChange={(e) => setProxyType(e.target.value as ProxyType)}
                  className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                >
                  <option value="http">HTTP / Clash Mixed Port</option>
                  <option value="socks5">SOCKS5</option>
                </select>
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-[minmax(0,1fr),180px]">
              <label className="space-y-2">
                <span className="text-sm text-gray-700 dark:text-gray-300">代理主机</span>
                <input
                  type="text"
                  value={proxyHost}
                  onChange={(e) => setProxyHost(e.target.value)}
                  placeholder="例如 127.0.0.1"
                  className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-900 placeholder-gray-400 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                />
              </label>

              <label className="space-y-2">
                <span className="text-sm text-gray-700 dark:text-gray-300">端口</span>
                <input
                  type="number"
                  min="1"
                  value={proxyPort}
                  onChange={(e) => setProxyPort(e.target.value)}
                  placeholder="7890"
                  className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-900 placeholder-gray-400 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <label className="space-y-2">
                <span className="text-sm text-gray-700 dark:text-gray-300">用户名（可选）</span>
                <input
                  type="text"
                  value={proxyUsername}
                  onChange={(e) => setProxyUsername(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                />
              </label>

              <label className="space-y-2">
                <span className="text-sm text-gray-700 dark:text-gray-300">密码（可选）</span>
                <input
                  type="password"
                  value={proxyPassword}
                  onChange={(e) => setProxyPassword(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                />
              </label>
            </div>

            <div className="grid gap-4 md:grid-cols-[220px,minmax(0,1fr)]">
              <label className="space-y-2">
                <span className="text-sm text-gray-700 dark:text-gray-300">请求超时（毫秒）</span>
                <input
                  type="number"
                  min="1000"
                  step="500"
                  value={proxyTimeoutMs}
                  onChange={(e) => setProxyTimeoutMs(e.target.value)}
                  className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
                />
              </label>

              <div className="rounded-lg border border-blue-200 bg-blue-50/70 px-4 py-3 text-sm text-blue-900 dark:border-blue-900/60 dark:bg-blue-950/20 dark:text-blue-100">
                建议：Clash 混合端口直接选 HTTP。`fallback` 模式会先尝试所有直连可用源站，全部失败后才切到 Proxy 可用源站。
              </div>
            </div>

            <div className={`rounded-lg border px-4 py-4 ${
              !proxyConfigured
                ? 'border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-900/40'
                : proxyUsable
                ? 'border-green-200 bg-green-50 dark:border-green-900/50 dark:bg-green-950/20'
                : 'border-amber-200 bg-amber-50 dark:border-amber-900/50 dark:bg-amber-950/20'
            }`}>
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-white">代理连通状态</p>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    {proxyHealth?.message || (proxyConfigured ? '等待检测代理可用性' : '填写代理地址后可检测')}
                  </p>
                  {proxyHealth?.checked_at && (
                    <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">最近检测：{proxyHealth.checked_at}</p>
                  )}
                </div>
                <button
                  onClick={() => refreshProxyHealth(true)}
                  disabled={checkingProxyHealth || !proxyConfigured}
                  className="flex items-center gap-2 rounded-lg bg-blue-500 px-4 py-2 text-white transition-colors hover:bg-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {checkingProxyHealth ? <Loader2 size={16} className="animate-spin" /> : <RefreshCw size={16} />}
                  <span>{checkingProxyHealth ? '检测中' : '检测代理'}</span>
                </button>
              </div>

              {proxyHealth?.targets && proxyHealth.targets.length > 0 && (
                <div className="mt-3 flex flex-wrap gap-2">
                  {proxyHealth.targets.map((target) => (
                    <span
                      key={target.name}
                      title={target.error || target.url}
                      className={`rounded-full px-2.5 py-1 text-xs ${
                        target.status === 'available'
                          ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
                          : 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300'
                      }`}
                    >
                      {target.name} {target.status === 'available' ? `${target.latency_ms}ms` : '失败'}
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>
        </section>

        {/* 多源站管理 */}
        <SectionToggle title="多源站管理" expanded={sourcesExpanded} onToggle={() => setSourcesExpanded(!sourcesExpanded)}>
          <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">管理漫画数据源站地址。在线抓取会按优先级选择直连可用源站；当代理模式为 fallback 时，直连源站全部不可用后才会转为使用 Proxy 可用源站。</p>

          {sourcesLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="animate-spin text-blue-500" size={24} />
            </div>
          ) : (
            <>
              <div className="mb-6 rounded-lg border border-blue-200 bg-blue-50/70 p-4 dark:border-blue-900/60 dark:bg-blue-950/20">
                <h3 className="mb-2 font-medium text-gray-900 dark:text-white">从发布页自动导入</h3>
                <p className="mb-3 text-sm text-gray-600 dark:text-gray-400">
                  这里的发布页网址只用于本次导入，不会保存到设置中。系统会自动提取候选网址，做可达性和漫画内容指纹校验，再把通过校验且未录入的源站加入列表。
                </p>
                <div className="flex gap-2">
                  <input
                    type="url"
                    value={releasePageUrl}
                    onChange={(e) => {
                      setReleasePageUrl(e.target.value);
                      setReleaseImportError(null);
                      setReleaseImportResult(null);
                    }}
                    placeholder="输入地址发布页，例如 https://2026-01-31-akalist.top/"
                    className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-400"
                    onKeyDown={(e) => { if (e.key === 'Enter') importFromReleasePage(); }}
                  />
                  <button
                    onClick={importFromReleasePage}
                    disabled={releaseImporting || !releasePageUrl.trim()}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50"
                  >
                    {releaseImporting ? <Loader2 size={16} className="animate-spin" /> : <RefreshCw size={16} />}
                    <span>{releaseImporting ? '导入中' : '导入可用源站'}</span>
                  </button>
                </div>

                {releaseImportError && (
                  <p className="mt-3 text-sm text-red-600 dark:text-red-400">{releaseImportError}</p>
                )}

                {releaseImportResult && (
                  <div className="mt-4 rounded-lg border border-gray-200 bg-white/80 p-3 text-sm dark:border-gray-700 dark:bg-gray-900/60">
                    <p className="text-gray-900 dark:text-white">
                      本次从发布页提取了 <span className="font-medium text-blue-600 dark:text-blue-400">{releaseImportResult.candidate_count}</span> 个候选网址，
                      新增 <span className="font-medium text-green-600 dark:text-green-400">{releaseImportResult.added_count}</span> 个，
                      跳过 <span className="font-medium text-amber-600 dark:text-amber-400">{releaseImportResult.skipped_count}</span> 个。
                    </p>
                    {releaseImportResult.added.length > 0 && (
                      <div className="mt-3">
                        <p className="mb-1 text-gray-700 dark:text-gray-300">已新增：</p>
                        <div className="space-y-1">
                          {releaseImportResult.added.map((source) => (
                            <div key={source.id} className="truncate text-green-700 dark:text-green-400">
                              {source.url}
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                    {releaseImportResult.skipped.length > 0 && (
                      <div className="mt-3">
                        <p className="mb-1 text-gray-700 dark:text-gray-300">已跳过：</p>
                        <div className="max-h-40 space-y-1 overflow-auto">
                          {releaseImportResult.skipped.map((item) => (
                            <div key={`${item.url}-${item.reason}`} className="text-gray-600 dark:text-gray-400">
                              <span className="break-all">{item.url}</span>
                              <span className="ml-2 text-xs">({item.reason})</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                )}
              </div>

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
                    <div key={source.id} className="flex flex-wrap items-center gap-3 rounded-lg bg-gray-50 p-3 dark:bg-gray-750">
                      <span className="text-sm text-gray-500 w-6">{index + 1}</span>
                      <Globe size={16} className="text-gray-400 flex-shrink-0" />
                      <div className="flex-1 min-w-0">
                        <div className="text-sm text-gray-900 dark:text-white truncate">{source.name}</div>
                        <div className="text-xs text-gray-500 dark:text-gray-400 truncate">{source.url}</div>
                      </div>
                      <div className="flex flex-wrap items-center gap-2">
                        {renderAccessBadge('直连', source.direct_status, source.direct_latency, 'direct', true, source.direct_last_error)}
                        {renderAccessBadge('Proxy', source.proxy_status, source.proxy_latency, 'proxy', proxyConfigured && proxyUsable, proxyUsable ? source.proxy_last_error : '代理当前不可用，已跳过代理检测')}
                      </div>
                      <button
                        onClick={() => testSource(source.id)}
                        disabled={testingSourceId === source.id}
                        className="p-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors disabled:cursor-wait disabled:opacity-70"
                        title="测速"
                      >
                        {testingSourceId === source.id ? (
                          <Loader2 size={16} className="animate-spin text-blue-500" />
                        ) : (
                          <TestTube size={16} className="text-blue-500" />
                        )}
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
                  心跳检测间隔：60 分钟 | 手动测速与心跳都会刷新直连 / Proxy 两列状态 | 新增源站后可先手动测速确认可达性
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
