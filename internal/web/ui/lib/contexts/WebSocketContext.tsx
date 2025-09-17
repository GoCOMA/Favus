'use client';

import { createContext, useContext, useEffect, useRef, useState, ReactNode } from 'react';

interface WebSocketMessage {
  Type: string;
  RunID: string;
  Payload: any;
}

interface WebSocketContextType {
  subscribe: (id: string, callback: (message: WebSocketMessage) => void) => void;
  unsubscribe: (id: string) => void;
  isConnected: boolean;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

export function useWebSocket() {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
}

interface WebSocketProviderProps {
  children: ReactNode;
}

export function WebSocketProvider({ children }: WebSocketProviderProps) {
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const subscriptionsRef = useRef<Map<string, (message: WebSocketMessage) => void>>(new Map());
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;
  const reconnectDelay = 3000;

  useEffect(() => {
    const connect = () => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        return;
      }

      try {
        const wsUrl = process.env.NEXT_PUBLIC_WS_URL || "ws://127.0.0.1:8765/ws";
        const ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => {
          console.log('WebSocket connected');
          setIsConnected(true);
          reconnectAttemptsRef.current = 0;
        };

        ws.onmessage = (event) => {
          try {
            const rawMessage = JSON.parse(event.data);
            const message: WebSocketMessage = {
                Type: rawMessage.type,
                RunID: rawMessage.runId,
                Payload: rawMessage.payload,
            };
            const globalCallback = subscriptionsRef.current.get('*');
            if (globalCallback) {
              globalCallback(message);
            }
            const specificCallback = subscriptionsRef.current.get(message.RunID);
            if (specificCallback) {
              specificCallback(message);
            }
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
          }
        };

        ws.onclose = (event) => {
          console.log('WebSocket disconnected:', event.code, event.reason);
          setIsConnected(false);
          wsRef.current = null;

          if (event.code !== 1000 && reconnectAttemptsRef.current < maxReconnectAttempts) {
            reconnectAttemptsRef.current++;
            console.log(`Attempting to reconnect (${reconnectAttemptsRef.current}/${maxReconnectAttempts})...`);
            reconnectTimeoutRef.current = setTimeout(connect, reconnectDelay);
          }
        };

        ws.onerror = (error) => {
          console.error('WebSocket error:', error);
        };
      } catch (error) {
        console.error('Failed to create WebSocket connection:', error);
      }
    };

    connect();

    const disconnect = () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close(1000, 'Manual disconnect');
      }
    };

    return () => {
      disconnect();
    };
  }, []);

  const subscribe = (id: string, callback: (message: WebSocketMessage) => void) => {
    subscriptionsRef.current.set(id, callback);
  };

  const unsubscribe = (id: string) => {
    subscriptionsRef.current.delete(id);
  };

  const contextValue: WebSocketContextType = {
    subscribe,
    unsubscribe,
    isConnected,
  };

  return (
    <WebSocketContext.Provider value={contextValue}>
      {children}
    </WebSocketContext.Provider>
  );
}
