import { Navigate, Route, Routes } from 'react-router-dom'
import Layout from './components/Layout'
import ComicDetailPage from './pages/ComicDetailPage'
import ComicsPage from './pages/ComicsPage'
import DashboardPage from './pages/DashboardPage'
import DownloadsPage from './pages/DownloadsPage'
import FavoritesPage from './pages/FavoritesPage'
import HistoryPage from './pages/HistoryPage'
import ReaderPage from './pages/ReaderPage'
import SettingsPage from './pages/SettingsPage'
import SourcesPage from './pages/SourcesPage'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<DashboardPage />} />
        <Route path="comics" element={<ComicsPage />} />
        <Route path="comics/:id" element={<ComicDetailPage />} />
        <Route path="reader/:id" element={<ReaderPage />} />
        <Route path="favorites" element={<FavoritesPage />} />
        <Route path="history" element={<HistoryPage />} />
        <Route path="sources" element={<SourcesPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="downloads" element={<DownloadsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
