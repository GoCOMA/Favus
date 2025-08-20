'use client';

import { useEffect, useState } from 'react';
import { useWebSocket } from '@/lib/contexts/MockWebSocketContext';

type Status = 'uploading' | 'completed' | 'failed';

// 각 업로드 세션의 상태를 정의하는 인터페이스
interface UploadState {
  runId: string;
  fileName: string;
  totalParts: number;
  status: Status;
  progress: number;
  completedParts: Set<number>;
  failedParts: Set<number>;
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

      setUploads((prev) => {
        const newUploads = { ...prev };
        let current = newUploads[RunID] ? { ...newUploads[RunID] } : undefined;

        switch (Type) {
          case 'session_start': {
            const startPayload = Payload as StartPayload;
            const partSize = startPayload.partMB * 1024 * 1024;
            const totalParts = Math.max(
              1,
              Math.ceil(startPayload.total / partSize),
            );
            current = {
              runId: RunID,
              fileName: startPayload.key,
              totalParts: totalParts,
              status: 'uploading',
              progress: 0,
              completedParts: new Set(),
              failedParts: new Set(),
            };
            break;
          }

          case 'part_done':
            if (current) {
              const { part } = Payload as PartDonePayload;
              const newCompleted = new Set(current.completedParts);
              newCompleted.add(part);
              const progress = (newCompleted.size / current.totalParts) * 100;
              current = {
                ...current,
                completedParts: newCompleted,
                progress,
              };
            }
            break;

          case 'session_done':
            if (current) {
              const { success } = Payload as DonePayload;
              current = {
                ...current,
                status: success ? 'completed' : 'failed',
                progress: success ? 100 : current.progress,
              };
            }
            break;

          case 'error':
            if (current) {
              const { message, partNumber } = Payload as ErrorPayload;
              const newFailed = new Set(current.failedParts);
              if (typeof partNumber === 'number') {
                newFailed.add(partNumber);
              }
              current = {
                ...current,
                status: 'failed',
                error: message,
                failedParts: newFailed,
              };
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
    return () => unsubscribe('*');
  }, [subscribe, unsubscribe]);

  const renderStatusPill = (status: UploadState['status']) => {
    const baseClasses = 'px-3 py-1 text-xs font-semibold rounded-full';
    if (status === 'uploading')
      return (
        <span className={`${baseClasses} bg-blue-100 text-blue-800`}>
          업로드 중
        </span>
      );
    if (status === 'completed')
      return (
        <span className={`${baseClasses} bg-green-100 text-green-800`}>
          완료
        </span>
      );
    return (
      <span className={`${baseClasses} bg-red-100 text-red-800`}>실패</span>
    );
  };

  const PartBars = ({
    total,
    completed,
    failed,
  }: {
    total: number;
    completed: Set<number>;
    failed: Set<number>;
  }) => {
    return (
      <div className="mt-2 px-4 rounded border border-gray-200 overflow-auto">
        <ul className="space-y-1">
          {Array.from({ length: total }, (_, i) => {
            const partNo = i + 1;
            const isFailed = failed.has(partNo);
            const isDone = completed.has(partNo);
            const cls = isFailed
              ? 'bg-red-500'
              : isDone
                ? 'bg-blue-600'
                : 'bg-gray-300';
            const title = isFailed
              ? `Part ${partNo}: 실패`
              : isDone
                ? `Part ${partNo}: 완료`
                : `Part ${partNo}: 대기`;

            return (
              <li
                key={partNo}
                className="my-3 flex items-center gap-2"
                title={title}
              >
                <span className="w-14 shrink-0 text-[14px] text-gray-500">
                  #{partNo}
                </span>
                <div className="h-3 w-full rounded-sm">
                  <div className={`h-3 w-full rounded-sm ${cls}`} />
                </div>
              </li>
            );
          })}
        </ul>

        {/* 범례 */}
        <div className="flex items-center gap-3 my-3 text-[11px] text-gray-500">
          <span className="inline-flex items-center gap-1">
            <i className="inline-block w-3 h-2 rounded-sm bg-blue-600" /> 완료
          </span>
          <span className="inline-flex items-center gap-1">
            <i className="inline-block w-3 h-2 rounded-sm bg-red-500" /> 실패
          </span>
          <span className="inline-flex items-center gap-1">
            <i className="inline-block w-3 h-2 rounded-sm bg-gray-300" /> 대기
          </span>
        </div>
      </div>
    );
  };

  return (
    <div className="bg-white shadow-lg rounded-lg p-6 border border-gray-200 mb-8">
      <h2 className="text-2xl font-bold text-gray-800 mb-2">
        실시간 업로드 현황
      </h2>
      <p className="text-gray-500 mb-4">
        CLI에서 시작된 업로드 작업이 여기에 표시됩니다. (WebSocket:{' '}
        {isConnected ? '연결됨' : '연결 끊김'})
      </p>

      {Object.keys(uploads).length === 0 ? (
        <div className="text-center py-10 bg-gray-50 rounded-lg">
          <p className="text-gray-500">진행중인 업로드가 없습니다.</p>
          <p className="text-sm text-gray-400 mt-2">
            `favus upload` 명령을 실행하면 여기에 표시됩니다.
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {Object.values(uploads).map((upload) => (
            <div
              key={upload.runId}
              className="p-4 border rounded-md bg-gray-50"
            >
              <div className="flex justify-between items-center mb-2">
                <p className="font-mono text-sm text-gray-700 truncate pr-4">
                  {upload.fileName}
                </p>
                {renderStatusPill(upload.status)}
              </div>

              <div className="my-4 flex items-center justify-between text-xs text-gray-500">
                <span>
                  {upload.completedParts.size} / {upload.totalParts} 파트
                </span>
                <span>{Math.round(upload.progress)}%</span>
              </div>

              <PartBars
                total={upload.totalParts}
                completed={upload.completedParts}
                failed={upload.failedParts}
              />

              {/* <div className="w-full bg-gray-200 rounded-full h-2.5">
                <div
                  className="bg-blue-600 h-2.5 rounded-full transition-all duration-300 ease-in-out"
                  style={{ width: `${upload.progress}%` }}
                ></div>
              </div>
              <div className="text-right text-xs text-gray-500 mt-1">
                {upload.completedParts.size} / {upload.totalParts} 파트 완료 (
                {Math.round(upload.progress)}%)
              </div> */}
              {upload.error && (
                <p className="text-red-500 text-sm mt-2">
                  에러: {upload.error}
                </p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
