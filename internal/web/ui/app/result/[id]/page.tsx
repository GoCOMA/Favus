'use client';

import { use, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { BatchResult, BatchFileItem } from '@/lib/types';
import { useWebSocket } from '@/lib/contexts/WebSocketContext';
import { HeaderSection } from './components/HeaderSection';
import { SummarySection } from './components/SummarySection';
import { FileList } from './components/FileList';
import { FileDetail } from './components/FileDetail';
import { TimeInfo } from './components/TimeInfo';
import { ErrorFallback } from './components/ErrorFallback';
import { LoadingFallback } from './components/LoadingFallback';
import { BatchErrorFallback } from './components/BatchErrorFallback';

interface Props {
  params: Promise<{ id: string }>;
}

export default function ResultPage({ params }: Props) {
  const router = useRouter();
  const { id } = use(params);
  const { subscribe, unsubscribe } = useWebSocket();
  const [batchResult, setBatchResult] = useState<BatchResult | null>(null);
  const [selectedFile, setSelectedFile] = useState<BatchFileItem | null>(null);
  const [loading, setLoading] = useState(true);

  // 초기 빈 상태 세팅
  useEffect(() => {
    const initial: BatchResult = {
      batchId: id,
      totalFiles: 0,
      completedFiles: 0,
      failedFiles: 0,
      pendingFiles: 0,
      processingFiles: 0,
      overallStatus: 'pending',
      overallProgress: 0,
      files: [],
      createdAt: new Date().toISOString(),
      startedAt: new Date().toISOString(),
      metadata: {
        batchName: `배치 ${id}`,
        description: `실시간 처리`,
        tags: ['batch', 'processing'],
      },
    };
    setBatchResult(initial);
    setLoading(false);
  }, [id]);

  // WebSocket progress 반영
  useEffect(() => {
    const handleProgressMessage = (message: any) => {
      if (message.Type === 'progress' && message.RunID === id) {
        try {
          const payload = JSON.parse(message.Payload);
          const { filename, bytes, totalBytes } = payload;

          setBatchResult((prev) => {
            if (!prev) return null;

            const updatedFiles = prev.files.some((f) => f.fileName === filename)
              ? prev.files.map((file) =>
                  file.fileName === filename
                    ? {
                        ...file,
                        progress: Math.min((bytes / totalBytes) * 100, 100),
                        status: bytes >= totalBytes ? 'completed' : 'processing',
                        fileSize: totalBytes,
                      }
                    : file
                )
              : [
                  ...prev.files,
                  {
                    id: filename,
                    fileName: filename,
                    fileSize: totalBytes,
                    status: bytes >= totalBytes ? 'completed' : 'processing',
                    progress: Math.min((bytes / totalBytes) * 100, 100),
                  },
                ];

            const completedFiles = updatedFiles.filter(
              (f) => f.status === 'completed'
            ).length;
            const processingFiles = updatedFiles.filter(
              (f) => f.status === 'processing'
            ).length;
            const pendingFiles = updatedFiles.filter(
              (f) => f.status === 'pending'
            ).length;

            return {
              ...prev,
              files: updatedFiles,
              totalFiles: updatedFiles.length,
              completedFiles,
              processingFiles,
              pendingFiles,
              overallProgress: Math.round(
                (completedFiles / updatedFiles.length) * 100
              ),
              overallStatus:
                completedFiles === updatedFiles.length
                  ? 'completed'
                  : 'processing',
            };
          });
        } catch (err) {
          console.error('Failed to parse progress payload:', err);
        }
      }
    };

    subscribe(id, handleProgressMessage);
    return () => unsubscribe(id);
  }, [id, subscribe, unsubscribe]);

  if (loading) return <LoadingFallback />;
  if (!batchResult) return <BatchErrorFallback id={id} router={router} />;

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
      <div className="max-w-7xl mx-auto px-4 py-12">
        <HeaderSection batchResult={batchResult} />
        <SummarySection batchResult={batchResult} />
        <FileList
          batchResult={batchResult}
          selectedFile={selectedFile}
          setSelectedFile={setSelectedFile}
        />
        {selectedFile && <FileDetail file={selectedFile} />}
        <TimeInfo batchResult={batchResult} />
        <div className="flex gap-4 mt-8">
          <button
            onClick={() => router.push('/')}
            className="flex-1 px-8 py-4 bg-gradient-to-r from-gray-600 to-slate-700 text-white rounded-xl hover:from-gray-700 hover:to-slate-800 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium"
          >
            홈으로 돌아가기
          </button>
        </div>
      </div>
    </main>
  );
}
