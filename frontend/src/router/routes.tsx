import { createBrowserRouter } from 'react-router-dom';
import { Layout } from '../components/Layout';
import { Home } from '../pages/Home';
import { AllMangas } from '../pages/AllMangas';
import { TagsPage } from '../pages/TagsPage';
import { HistoryPage } from '../pages/HistoryPage';
import { FavoritesPage } from '../pages/FavoritesPage';
import { CachedPage } from '../pages/CachedPage';
import { SettingsPage } from '../pages/SettingsPage';
import { DownloadsPage } from '../pages/DownloadsPage';
import { MangaDetail } from '../pages/MangaDetail';
import { ReadPage } from '../pages/ReadPage';
import { AuthorPage } from '../pages/AuthorPage';
import { NotFound } from '../pages/NotFound';

export const router = createBrowserRouter([
  {
    path: '/',
    Component: Layout,
    children: [
      { index: true, Component: Home },
      { path: 'all', Component: AllMangas },
      { path: 'tags', Component: TagsPage },
      { path: 'history', Component: HistoryPage },
      { path: 'favorites', Component: FavoritesPage },
      { path: 'cached', Component: CachedPage },
      { path: 'settings', Component: SettingsPage },
      { path: 'downloads', Component: DownloadsPage },
      { path: 'manga/:id', Component: MangaDetail },
      { path: 'author/:name', Component: AuthorPage },
      { path: '*', Component: NotFound },
    ],
  },
  {
    path: 'read/:id',
    Component: ReadPage,
  },
]);
