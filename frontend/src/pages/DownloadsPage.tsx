import { useState, useEffect, useCallback } from 'react';
import { Pause, Play, X, Loader2 } from 'lucide-react';
import * as api from '../api';
import client from '../api/client';
import { useDownloadSSE } from '../hooks/useSSE';

export function DownloadsPage() {
  const [downloads, setDownloads] = useState<api.DownloadTask[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchDownloads = useCallback(() => {
    api.getDownloadTasks().then((tasks) => {
      setDownloads(tasks);
      setLoading(false);
    }).catch(() => {
      setLoading(false);
    });
  }, []);

  useEffect(() => {
    fetchDownloads();
  }, [fetchDownloads]);

  // SSE 实时更新，替代 5 秒轮询
  useDownloadSSE(fetchDownloads);

  if (loading) {
    return (
      <div className="p-8 flex items-center justify-center py-16">
        <Loader2 className="animate-spin text-blue-500" size={32} />
      </div>
    );
  }

  return (
    <div className="p-8">
      <h1 className="text-3xl mb-6 text-gray-900 dark:text-white">下载管理</h1>
      {downloads.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-gray-500 dark:text-gray-400 text-lg">暂无下载任务</p>
        </div>
      ) : (
        <div className="space-y-4">
          {downloads.map((download) => {
            const progress = download.total > 0 ? Math.round((download.current / download.total) * 100) : 0;
            return (
              <div key={download.id} className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-sm">
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <h3 className="font-medium text-gray-900 dark:text-white">{download.comic_title || `漫画 #${download.comic_id}`}</h3>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      {download.status === 'downloading' && '下载中...'}
                      {download.status === 'completed' && '已完成'}
                      {download.status === 'paused' && '已暂停'}
                      {download.status === 'queued' && '排队中'}
                      {download.status === 'failed' && '失败'}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    {download.status === 'downloading' && (
                      <button
                        onClick={() => client.post(`/v2/downloads/${download.id}/pause`).then(fetchDownloads)}
                        className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                        title="暂停"
                      >
                        <Pause size={20} className="text-gray-700 dark:text-gray-300" />
                      </button>
                    )}
                    {(download.status === 'paused' || download.status === 'failed') && (
                      <button
                        onClick={() => {
                          const action = download.status === 'failed' ? 'retry' : 'resume';
                          client.post(`/v2/downloads/${download.id}/${action}`).then(fetchDownloads);
                        }}
                        className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                        title={download.status === 'failed' ? '重试' : '恢复'}
                      >
                        <Play size={20} className="text-gray-700 dark:text-gray-300" />
                      </button>
                    )}
                    {download.status !== 'completed' && download.status !== 'canceled' && (
                      <button
                        onClick={() => client.post(`/v2/downloads/${download.id}/cancel`).then(fetchDownloads)}
                        className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                        title="取消"
                      >
                        <X size={20} className="text-red-500" />
                      </button>
                    )}
                  </div>
                </div>
                <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div className="bg-blue-500 h-2 rounded-full transition-all duration-300" style={{ width: `${progress}%` }} />
                </div>
                <div className="mt-2 text-sm text-gray-500 dark:text-gray-400 text-right">
                  {progress}% ({download.current}/{download.total})
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
