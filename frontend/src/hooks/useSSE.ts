import { useEffect, useRef, useCallback } from 'react';

export interface SSEEvent {
  type: string;
  data: Record<string, unknown>;
}

type SSEHandler = (event: SSEEvent) => void;

export function useSSE(handlers: Record<string, SSEHandler>) {
  const handlersRef = useRef(handlers);
  handlersRef.current = handlers;

  useEffect(() => {
    const eventSource = new EventSource('/api/events');

    eventSource.onmessage = (e) => {
      try {
        const parsed: SSEEvent = JSON.parse(e.data);
        const handler = handlersRef.current[parsed.type];
        if (handler) {
          handler(parsed);
        }
        // 通配符处理
        const wildcard = handlersRef.current['*'];
        if (wildcard) {
          wildcard(parsed);
        }
      } catch {
        // ignore parse errors
      }
    };

    eventSource.onerror = () => {
      // EventSource 会自动重连
    };

    return () => {
      eventSource.close();
    };
  }, []);
}

// 便捷 hook：监听下载进度事件
export function useDownloadSSE(onUpdate: () => void) {
  const onUpdateRef = useRef(onUpdate);
  onUpdateRef.current = onUpdate;

  const handler = useCallback((evt: SSEEvent) => {
    onUpdateRef.current();
  }, []);

  useSSE({
    'download:progress': handler,
    'download:completed': handler,
    'download:failed': handler,
  });
}
