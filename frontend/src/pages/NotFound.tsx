import { Link } from 'react-router-dom';

export function NotFound() {
  return (
    <div className="flex items-center justify-center h-screen bg-gray-50 dark:bg-gray-900">
      <div className="text-center">
        <h1 className="text-6xl mb-4 text-gray-900 dark:text-white">404</h1>
        <p className="text-xl mb-6 text-gray-600 dark:text-gray-400">
          页面不存在
        </p>
        <Link
          to="/"
          className="px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors inline-block"
        >
          返回首页
        </Link>
      </div>
    </div>
  );
}
