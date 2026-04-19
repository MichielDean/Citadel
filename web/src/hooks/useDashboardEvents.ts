import { useCallback, useEffect, useRef, useState } from 'react';
import type { DashboardData } from '../api/types';

interface UseDashboardEventsOptions {
  onData?: (data: DashboardData) => void;
  onError?: (error: Error) => void;
  enabled?: boolean;
}

export function useDashboardEvents(options: UseDashboardEventsOptions = {}) {
  const { onData, onError, enabled = true } = options;
  const [data, setData] = useState<DashboardData | null>(null);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const esRef = useRef<EventSource | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout>>();

  const connect = useCallback(() => {
    if (esRef.current) {
      esRef.current.close();
    }

    const es = new EventSource('/api/dashboard/events');

    es.onopen = () => {
      setConnected(true);
      setError(null);
    };

    es.onmessage = (e) => {
      try {
        const parsed: DashboardData = JSON.parse(e.data);
        setData(parsed);
        onData?.(parsed);
      } catch {
        // ignore non-JSON messages
      }
    };

    es.onerror = () => {
      setConnected(false);
      const err = new Error('SSE connection lost');
      setError(err);
      onError?.(err);
      es.close();
      esRef.current = null;
      if (enabled) {
        clearTimeout(reconnectTimer.current);
        reconnectTimer.current = setTimeout(() => {
          connect();
        }, 3000);
      }
    };

    esRef.current = es;
  }, [onData, onError, enabled]);

  useEffect(() => {
    if (!enabled) {
      if (esRef.current) {
        esRef.current.close();
        esRef.current = null;
      }
      return;
    }
    connect();
    return () => {
      clearTimeout(reconnectTimer.current);
      if (esRef.current) {
        esRef.current.close();
        esRef.current = null;
      }
    };
  }, [connect, enabled]);

  return { data, connected, error };
}