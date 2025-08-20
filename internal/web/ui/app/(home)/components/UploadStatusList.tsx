'use client';

import { useEffect, useState } from 'react';
import { useWebSocket } from '@/lib/contexts/WebSocketContext';

// 각 업로드 세션의 상태를 정의하는 인터페이스
interface UploadState {
  runId: string;
  fileName: string;
  totalParts: number;
  status: string; // e.g., 'uploading', 'completed', 'failed'
  progress: number; // 0-100
  completedParts: Set<number>;
  error?: string;
}

// WebSocket 메시지 페이로드 타입 정의
interface StartPayload {
  bucket: string;
  key: string;
  uploadId: string;
  total: number;
  partMB: number;
}

interface PartDonePayload {
  part: number;
  size: number;
  etag: string;
}

interface ProgressPayload {
  totalBytes: number;
}

interface DonePayload {
  success: boolean;
  uploadId: string;
}

interface ErrorPayload {
  message: string;
  partNumber?: number;
}

export default function UploadStatusList() {
  const { subscribe, unsubscribe, isConnected } = useWebSocket();
  const [uploads, setUploads] = useState<Record<string, UploadState>>({});

  useEffect(() => {
    const handleMessage = (msg: any) => {
      const { Type, RunID, Payload } = msg;

      setUploads(prev => {
        const newUploads = { ...prev };
        let current = newUploads[RunID] ? { ...newUploads[RunID] } : undefined;

        switch (Type) {
          case 'session_start': {
            const startPayload = Payload as StartPayload;
            const partSize = startPayload.partMB * 1024 * 1024;
            const totalParts = Math.ceil(startPayload.total / partSize);
            current = {
              runId: RunID,
              fileName: startPayload.key,
              totalParts: totalParts,
              status: '업로드 중',
              progress: 0,
              completedParts: new Set(),
            };
            break;
          }

          case 'part_done':
            if (current) {
              const partDonePayload = Payload as PartDonePayload;
              current.completedParts.add(partDonePayload.part);
              current.progress = (current.completedParts.size / current.totalParts) * 100;
            }
            break;

          case 'session_done':
            if (current) {
              const donePayload = Payload as DonePayload;
              current.status = donePayload.success ? '완료' : '실패';
              current.progress = donePayload.success ? 100 : current.progress;
            }
            break;

          case 'error':
            if (current) {
              const errorPayload = Payload as ErrorPayload;
              current.status = '실패';
              current.error = errorPayload.message;
            }
            break;
        }

        if (current) {
          newUploads[RunID] = current;
        }
        return newUploads;
      });
    };

    subscribe('*', handleMessage);

    return () => {
      unsubscribe('*');
    };
  }, [subscribe, unsubscribe]);

  const renderStatusPill = (status: string) => {
    const baseClasses = 'px-3 py-1 text-sm font-semibold rounded-full';
    switch (status) {
      case '업로드 중':
        return <span className={`bg-blue-100 text-blue-800 ${baseClasses}`}>업로드 중</span>;
      case '완료':
        return <span className={`bg-green-100 text-green-800 ${baseClasses}`}>완료</span>;
      case '실패':
        return <span className={`bg-red-100 text-red-800 ${baseClasses}`}>실패</span>;
      default:
        return <span className={`bg-gray-100 text-gray-800 ${baseClasses}`}>{status}</span>;
    }
  };

  return (
    <div className="bg-white shadow-lg rounded-lg p-6 border border-gray-200 mb-8">
      <h2 className="text-2xl font-bold text-gray-800 mb-2">실시간 업로드 현황</h2>
      <p className="text-gray-500 mb-4">
        CLI에서 시작된 업로드 작업이 여기에 표시됩니다. (WebSocket: {isConnected ? '연결됨' : '연결 끊김'})
      </p>
      
      {Object.keys(uploads).length === 0 ? (
        <div className="text-center py-10 bg-gray-50 rounded-lg">
          <p className="text-gray-500">진행중인 업로드가 없습니다.</p>
          <p className="text-sm text-gray-400 mt-2">`favus upload` 명령을 실행하면 여기에 표시됩니다.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {Object.values(uploads).map(upload => (
            <div key={upload.runId} className="p-4 border rounded-md bg-gray-50">
              <div className="flex justify-between items-center mb-2">
                <p className="font-mono text-sm text-gray-700 truncate pr-4">{upload.fileName}</p>
                {renderStatusPill(upload.status)}
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2.5">
                <div 
                  className="bg-blue-600 h-2.5 rounded-full transition-all duration-300 ease-in-out"
                  style={{ width: `${upload.progress}%` }}
                ></div>
              </div>
              <div className="text-right text-xs text-gray-500 mt-1">
                {upload.completedParts.size} / {upload.totalParts} 파트 완료 ({Math.round(upload.progress)}%)
              </div>
              {upload.error && (
                <p className="text-red-500 text-sm mt-2">에러: {upload.error}</p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
