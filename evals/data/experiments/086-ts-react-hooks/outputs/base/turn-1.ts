// data hooks
import { useState, useEffect, useRef, useCallback } from 'react';

// existing hooks...

export function useWebSocket<T>(url: string, protocols?: string | string[], onMessage: (message: T) => void) {
  const [ WebSocketClass ] = useState(() => globalThis.WebSocket || globalThis.MozWebSocket);
  const [connection, setConnection] = useState<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [reconnectTimeout, setReconnectTimeout] = useState<number | null>(null);
  const [heartbeatInterval, setHeartbeatInterval] = useState<number | null>(null);
  const reconnectDelay = useRef(1000);
  const messageQueue = useRef<T[]>([]);

  const establishConnection = useCallback(() => {
    const ws = new WebSocketClass(url, protocols);
    ws.onopen = () => {
      setConnected(true);
      setConnection(ws);
      if (heartbeatInterval) {
        clearInterval(heartbeatInterval);
      }
      setHeartbeatInterval(setInterval(() => ws.send('heartbeat'), 30000));
    };
    ws.onclose = () => {
      setConnected(false);
      setConnection(null);
      clearInterval(heartbeatInterval);
      setHeartbeatInterval(null);
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
      }
      reconnectTimeoutId();
    };
    ws.onerror = () => {
      setConnected(false);
      setConnection(null);
    };
    ws.onmessage = (event) => {
      try {
        const message: T = JSON.parse(event.data);
        onMessage(message);
        if (messageQueue.current.length) {
          messageQueue.current.forEach((message) => onMessage(message));
          messageQueue.current = [];
        }
      } catch (error) {
        console.error('Error parsing message:', error);
      }
    };
  }, [url, protocols, onMessage]);

  const reconnectTimeoutId = useCallback(() => {
    setReconnectTimeout(setTimeout(() => {
      reconnectDelay.current *= 2;
      if (reconnectDelay.current > 30000) {
        reconnectDelay.current = 30000;
      }
      establishConnection();
    }, reconnectDelay.current));
  }, [establishConnection]);

  useEffect(() => {
    establishConnection();
    return () => {
      if (connection) {
        connection.close();
      }
    };
  }, [establishConnection]);

  const send = useCallback((message: T) => {
    if (connected && connection) {
      connection.send(JSON.stringify(message));
    } else {
      messageQueue.current.push(message);
    }
  }, [connected, connection]);

  return { connected, send };
}

// existing hooks...