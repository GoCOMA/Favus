'use client';

import { createContext, useContext, useEffect, useRef, useState, ReactNode } from 'react';

interface WebSocketMessage {
  Type: string;
  RunID: string;
  Payload: string;
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

  const connect = () => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    try {
      // Use wss for production, ws for development
      const wsUrl = process.env.NODE_ENV === 'production' 
        ? `wss://${window.location.host}/ws`
        : `ws://${window.location.hostname}:8080/ws`;
      
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        reconnectAttemptsRef.current = 0;
      };

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          
          // Send message to all subscribers for this RunID
          const callback = subscriptionsRef.current.get(message.RunID);
          if (callback) {
            callback(message);
          }
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        setIsConnected(false);
        wsRef.current = null;

        // Attempt to reconnect if not a manual close
        if (event.code !== 1000 && reconnectAttemptsRef.current < maxReconnectAttempts) {
          reconnectAttemptsRef.current++;
          console.log(`Attempting to reconnect (${reconnectAttemptsRef.current}/${maxReconnectAttempts})...`);
          
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, reconnectDelay);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
    }
  };

  const disconnect = () => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current) {
      wsRef.current.close(1000, 'Manual disconnect');
      wsRef.current = null;
    }
    setIsConnected(false);
  };

  const subscribe = (id: string, callback: (message: WebSocketMessage) => void) => {
    subscriptionsRef.current.set(id, callback);
    
    // Connect if not already connected
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      connect();
    }
  };

  const unsubscribe = (id: string) => {
    subscriptionsRef.current.delete(id);
    
    // Disconnect if no more subscriptions
    if (subscriptionsRef.current.size === 0) {
      disconnect();
    }
  };

  useEffect(() => {
    return () => {
      disconnect();
    };
  }, []);

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