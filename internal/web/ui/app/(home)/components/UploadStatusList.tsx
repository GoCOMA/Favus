'use client';

import { useEffect, useMemo, useState } from 'react';
import { useWebSocket } from '@/lib/contexts/WebSocketContext';

type Status = 'uploading' | 'completed' | 'failed';

// 각 업로드 세션의 상태를 정의하는 인터페이스
interface FileUploadState {
  uploadId: string;
  fileName: string;
  partSize: number;
  totalBytes: number;
  totalParts: number;
  completedParts: Set<number>;
  progress: number;
  status: Status;
  error?: string;
}

interface RunUploadState {
  runId: string;
  files: Record<string, FileUploadState>;
  status: Status;
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
  uploadId: string;
  part: number;
  size: number;
  etag: string;
}

interface ProgressPayload {
  uploadId: string;
  totalBytes: number;
}

interface DonePayload {
  uploadId: string;
  success: boolean;
}

interface ErrorPayload {
  uploadId: string;
  message: string;
  partNumber?: number;
}

export default function UploadStatusList() {
  const { subscribe, unsubscribe, isConnected } = useWebSocket();
  const [uploads, setUploads] = useState<Record<string, RunUploadState>>({});

  const computeRunStatus = (files: Record<string, FileUploadState>): Status => {
    const list = Object.values(files);
    if (list.some((f) => f.status === 'failed')) return 'failed';
    if (list.length > 0 && list.every((f) => f.status === 'completed'))
      return 'completed';
    return 'uploading';
  };

  const calcProgressFromParts = (
    completedParts: number,
    totalParts: number,
  ) => {
    if (totalParts <= 0) return 0;
    return Math.min(100, (completedParts / totalParts) * 100);
  };

  useEffect(() => {
    const handleMessage = (msg: any) => {
      const { Type, RunID, Payload } = msg;

      setUploads((prev) => {
        const newUploads = { ...prev };
        const run: RunUploadState = newUploads[RunID] ?? {
          runId: RunID,
          files: {},
          status: 'uploading',
        };

        switch (Type) {
          case 'session_start': {
            const startPayload = Payload as StartPayload;
            const partSize = Math.max(
              1,
              Math.floor(startPayload.partMB * 1024 * 1024),
            );
            const totalParts = Math.ceil(startPayload.total / partSize);

            // 파일 상태 초기화
            const initial: FileUploadState = {
              uploadId: startPayload.uploadId,
              fileName: startPayload.key,
              partSize,
              totalBytes: startPayload.total,
              totalParts,
              completedParts: new Set(),
              progress: 0,
              status: 'uploading',
            };

            run.files = { ...run.files, [startPayload.uploadId]: initial };
            run.status = computeRunStatus(run.files);
            newUploads[RunID] = run;
            break;
          }

          case 'part_done': {
            const p = Payload as PartDonePayload;
            const file = run.files[p.uploadId];
            if (file) {
              const newSet = new Set(file.completedParts);
              newSet.add(p.part);
              const progress = calcProgressFromParts(
                newSet.size,
                file.totalParts,
              );

              run.files = {
                ...run.files,
                [p.uploadId]: {
                  ...file,
                  completedParts: newSet,
                  progress,
                },
              };
              run.status = computeRunStatus(run.files);
              newUploads[RunID] = run;
            }
            break;
          }

          case 'progress': {
            const p = Payload as ProgressPayload;
            const file = run.files[p.uploadId];
            if (file && file.totalBytes > 0) {
              const progress = Math.min(
                100,
                (p.totalBytes / file.totalBytes) * 100,
              );
              run.files = {
                ...run.files,
                [p.uploadId]: { ...file, progress },
              };
              run.status = computeRunStatus(run.files);
              newUploads[RunID] = run;
            }
            break;
          }

          case 'session_done': {
            const p = Payload as DonePayload;
            const file = run.files[p.uploadId];
            if (file) {
              run.files = {
                ...run.files,
                [p.uploadId]: {
                  ...file,
                  status: p.success ? 'completed' : 'failed',
                  progress: p.success ? 100 : file.progress,
                },
              };
              run.status = computeRunStatus(run.files);
              newUploads[RunID] = run;
            }
            break;
          }

          case 'error': {
            const p = Payload as ErrorPayload;
            const file = run.files[p.uploadId];
            if (file) {
              run.files = {
                ...run.files,
                [p.uploadId]: {
                  ...file,
                  status: 'failed',
                  error: p.message,
                },
              };
              run.status = computeRunStatus(run.files);
              newUploads[RunID] = run;
            }
            break;
          }
        }

        return newUploads;
      });
    };

    subscribe('*', handleMessage);

    return () => {
      unsubscribe('*');
    };
  }, [subscribe, unsubscribe]);

  const renderStatusPill = (status: Status) => {
    const baseClasses = 'px-3 py-1 text-sm font-semibold rounded-full';
    if (status === 'uploading')
      return (
        <span className={`${baseClasses} bg-blue-100 text-blue-800`}>
          uploading
        </span>
      );
    if (status === 'completed')
      return (
        <span className={`${baseClasses} bg-green-100 text-green-800`}>
          completed
        </span>
      );
    return (
      <span className={`${baseClasses} bg-red-100 text-red-800`}>failed</span>
    );
  };

  const runCards = useMemo(() => {
    const list = Object.values(uploads);
    if (list.length === 0) {
      return (
        <div className="text-center py-10 bg-gray-50 rounded-lg">
          <p className="text-gray-500">진행중인 업로드가 없습니다.</p>
          <p className="text-sm text-gray-400 mt-2">
            `favus upload` 명령을 실행하면 여기에 표시됩니다.
          </p>
        </div>
      );
    }

    return list.map((run) => {
      const files = Object.values(run.files);
      const total = files.length;
      const done = files.filter((f) => f.status === 'completed').length;

      return (
        <div key={run.runId} className="p-4 border rounded-lg bg-gray-50">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <h3 className="font-semibold text-gray-800">
                RunID: <span className="font-mono">{run.runId}</span>
              </h3>
              {renderStatusPill(run.status)}
            </div>
            <div className="text-sm text-gray-500">
              {done} / {total} 파일 completed
            </div>
          </div>

          <div className="space-y-3">
            {files.map((file) => (
              <div
                key={file.uploadId}
                className="p-3 rounded-md bg-white border"
              >
                <div className="flex items-center justify-between mb-1">
                  <p className="font-mono text-sm text-gray-700 truncate pr-4">
                    {file.fileName}
                  </p>
                  {renderStatusPill(file.status)}
                </div>

                <div className="w-full bg-gray-200 rounded-full h-2.5">
                  <div
                    className="bg-blue-600 h-2.5 rounded-full transition-all duration-300 ease-in-out"
                    style={{ width: `${file.progress}%` }}
                  />
                </div>

                <div className="flex items-center justify-between text-xs text-gray-500 mt-1">
                  <span>
                    {file.completedParts.size} / {file.totalParts} 파트
                  </span>
                  <span>{Math.round(file.progress)}%</span>
                </div>

                {file.error && (
                  <p className="text-red-500 text-xs mt-2">
                    에러: {file.error}
                  </p>
                )}
              </div>
            ))}
          </div>
        </div>
      );
    });
  }, [uploads]);

  return (
    <div className="bg-white shadow-lg rounded-lg p-6 border border-gray-200 mb-8">
      <h2 className="text-2xl font-bold text-gray-800 mb-2">
        실시간 업로드 현황
      </h2>
      <p className="text-gray-500 mb-4">
        CLI에서 시작된 업로드 작업이 여기에 표시됩니다. (WebSocket:{' '}
        {isConnected ? '연결됨' : '연결 끊김'})
      </p>
      <div className="space-y-4">{runCards}</div>
    </div>
  );
}
