import { NavLink } from 'react-router-dom';
import {
  Home,
  Library,
  Tags,
  History,
  Heart,
  HardDrive,
  Settings,
  Sun,
  Moon,
  Download,
} from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import { useManga } from '../contexts/MangaContext';

const navItems = [
  { path: '/', label: '首页', icon: Home },
  { path: '/all', label: '所有漫画', icon: Library },
  { path: '/tags', label: '标签搜索', icon: Tags },
  { path: '/history', label: '历史观看', icon: History },
  { path: '/favorites', label: '收藏', icon: Heart },
  { path: '/cached', label: '本地已保存', icon: HardDrive },
  { path: '/settings', label: '设置', icon: Settings },
];

export function Sidebar() {
  const { theme, toggleTheme } = useTheme();
  const { downloads } = useManga();

  const activeDownloads = downloads.filter((d) => d.status === 'downloading').length;

  return (
    <div className="w-64 h-screen bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 flex flex-col">
      <div className="p-6">
        <h1 className="font-bold text-xl text-gray-900 dark:text-white">漫画阅读器</h1>
      </div>

      <nav className="flex-1 px-3">
        {navItems.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            className={({ isActive }) =>
              `flex items-center gap-3 px-4 py-3 rounded-lg mb-1 transition-colors ${
                isActive
                  ? 'bg-blue-500 text-white'
                  : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
              }`
            }
          >
            <item.icon size={20} />
            <span>{item.label}</span>
          </NavLink>
        ))}
      </nav>

      <div className="p-3 border-t border-gray-200 dark:border-gray-800">
        <button
          onClick={toggleTheme}
          className="flex items-center gap-3 px-4 py-3 rounded-lg mb-2 w-full text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
        >
          {theme === 'light' ? <Moon size={20} /> : <Sun size={20} />}
          <span>{theme === 'light' ? '黑夜模式' : '白天模式'}</span>
        </button>

        <NavLink
          to="/downloads"
          className={({ isActive }) =>
            `flex items-center gap-3 px-4 py-3 rounded-lg relative transition-colors ${
              isActive
                ? 'bg-blue-500 text-white'
                : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
            }`
          }
        >
          <Download size={20} />
          <span>下载管理</span>
          {activeDownloads > 0 && (
            <span className="absolute right-3 top-1/2 -translate-y-1/2 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
              {activeDownloads}
            </span>
          )}
        </NavLink>
      </div>
    </div>
  );
}
