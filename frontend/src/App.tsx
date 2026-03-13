import { RouterProvider } from 'react-router-dom';
import { ThemeProvider } from './contexts/ThemeContext';
import { MangaProvider } from './contexts/MangaContext';
import { router } from './router/routes';

export default function App() {
  return (
    <ThemeProvider>
      <MangaProvider>
        <RouterProvider router={router} />
      </MangaProvider>
    </ThemeProvider>
  );
}
