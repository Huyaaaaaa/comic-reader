import { useState } from 'react';
import { useTheme } from '../contexts/ThemeContext';
import { CacheStrategy } from '../types';
import {
  Sun, Moon, Plus, Trash2, TestTube, Globe, Upload, Download,
  RefreshCw, ChevronDown, ChevronUp,
} from 'lucide-react';

interface CacheLevelSettings {
  strategy: CacheStrategy;
  count: number;
}

interface SourceSite {
  id: string;
  url: string;
  status: 'unknown' | 'online' | 'offline';
  latency?: number;
}

export function SettingsPage() {
  const { theme, toggleTheme } = useTheme();

  // === 缓存策略 ===
  const [metadataCache, setMetadataCache] = useState<CacheLevelSettings>({ strategy: 'passive', count: 500 });
  const [extendedCache, setExtendedCache] = useState<CacheLevelSettings>({ strategy: 'passive', count: 500 });
  const [contentCache, setContentCache] = useState<CacheLevelSettings>({ strategy: 'passive', count: 100 });

  // === 多源站 ===
  const [sources, setSources] = useState<SourceSite[]>([]);
  const [newSourceUrl, setNewSourceUrl] = useState('');
  const [sourcesExpanded, setSourcesExpanded] = useState(false);

  // === 更新 ===
  const [contentUpdateMode, setContentUpdateMode] = useState<'manual' | 'startup' | 'interval'>('startup');
  const [updateInterval, setUpdateInterval] = useState(30);
  const [appAutoCheck, setAppAutoCheck] = useState(true);
  const [updatesExpanded, setUpdatesExpanded] = useState(false);

  // === 导入导出 ===
  const [importExportExpanded, setImportExportExpanded] = useState(false);

  const addSource = () => {
    if (!newSourceUrl.trim()) return;
    setSources((prev) => [...prev, { id: Date.now().toString(), url: newSourceUrl.trim(), status: 'unknown' }]);
    setNewSourceUrl('');
  };

  const removeSource = (id: string) => {
    setSources((prev) => prev.filter((s) => s.id !== id));
  };

  const testSource = (id: string) => {
    setSources((prev) =>
      prev.map((s) => (s.id === id ? { ...s, status: 'online' as const, latency: Math.floor(Math.random() * 500 + 50) } : s))
    );
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

  return (
    <div className="p-8">
      <h1 className="text-3xl mb-6 text-gray-900 dark:text-white">设置</h1>

      <div className="max-w-2xl space-y-6">
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
            <select className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
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
          <div className="flex gap-2 mb-4">
            <input type="text" value={newSourceUrl} onChange={(e) => setNewSourceUrl(e.target.value)}
              placeholder="输入源站地址，例如 https://example.com"
              className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-400"
              onKeyDown={(e) => { if (e.key === 'Enter') addSource(); }} />
            <button onClick={addSource} className="flex items-center gap-1 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors">
              <Plus size={16} /><span>添加</span>
            </button>
          </div>

          {sources.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400 text-sm py-4 text-center">暂无源站，请添加至少一个源站地址</p>
          ) : (
            <div className="space-y-2">
              {sources.map((source, index) => (
                <div key={source.id} className="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-750 rounded-lg">
                  <span className="text-sm text-gray-500 w-6">{index + 1}</span>
                  <Globe size={16} className="text-gray-400 flex-shrink-0" />
                  <span className="flex-1 text-sm text-gray-900 dark:text-white truncate">{source.url}</span>
                  <span className={`text-xs px-2 py-1 rounded ${source.status === 'online' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : source.status === 'offline' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'}`}>
                    {source.status === 'online' ? `${source.latency}ms` : source.status === 'offline' ? '离线' : '未测试'}
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
              <button className="mt-3 flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors">
                <RefreshCw size={16} /><span>立即检查更新</span>
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
                {['全部已缓存漫画', '所有封面', '所有图片'].map((label) => (
                  <button key={label} className="flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors">
                    <Upload size={16} /><span>{label}</span>
                  </button>
                ))}
              </div>
            </div>

            <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
              <h3 className="font-medium text-gray-900 dark:text-white mb-2">导入</h3>
              <p className="text-sm text-gray-500 dark:text-gray-400 mb-3">从备份文件恢复数据</p>
              <button className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors">
                <Download size={16} /><span>选择文件导入</span>
              </button>
              <p className="mt-2 text-xs text-gray-500 dark:text-gray-400">
                导入策略：覆盖导入 / 仅导入缺失 / 取消
              </p>
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
                <p className="text-sm text-gray-500 dark:text-gray-400">当前使用 -- MB</p>
              </div>
              <button className="px-4 py-2 rounded-lg bg-red-500 text-white hover:bg-red-600 transition-colors">清除缓存</button>
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
