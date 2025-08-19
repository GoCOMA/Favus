'use client';

import { createContext, useContext, useEffect, useState, ReactNode } from 'react';

interface WebSocketMessage {
  Type: string;
  RunID: string;
  Payload: string;
}

interface WebSocketContextType {
  isConnected: boolean;
  subscribe: (runId: string, callback: (data: any) => void) => void;
  unsubscribe: (runId: string) => void;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

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
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [subscribers, setSubscribers] = useState<Map<string, (data: any) => void>>(new Map());

  useEffect(() => {
    const connectWebSocket = () => {
      const websocket = new WebSocket("ws://localhost:8000/ws");

      websocket.onopen = () => {
        console.log("✅ WebSocket connected globally");
        setIsConnected(true);
        setWs(websocket);
      };

      websocket.onmessage = (event) => {
        try {
          const data: WebSocketMessage = JSON.parse(event.data);
          
          if (data.Type === "progress" && data.RunID) {
            const callback = subscribers.get(data.RunID);
            if (callback) {
              const payload = JSON.parse(data.Payload);
              callback({ runId: data.RunID, payload });
            }
          }
        } catch (err) {
          console.error("WebSocket message parsing error:", err);
        }
      };

      websocket.onclose = () => {
        console.log("❌ WebSocket disconnected globally");
        setIsConnected(false);
        setWs(null);
        
        // Reconnect after 5 seconds
        setTimeout(connectWebSocket, 5000);
      };

      websocket.onerror = (error) => {
        console.error("WebSocket error:", error);
      };
    };

    connectWebSocket();

    return () => {
      ws?.close();
    };
  }, []);

  const subscribe = (runId: string, callback: (data: any) => void) => {
    setSubscribers(prev => new Map(prev.set(runId, callback)));
  };

  const unsubscribe = (runId: string) => {
    setSubscribers(prev => {
      const newMap = new Map(prev);
      newMap.delete(runId);
      return newMap;
    });
  };

  return (
    <WebSocketContext.Provider value={{ isConnected, subscribe, unsubscribe }}>
      {children}
    </WebSocketContext.Provider>
  );
}