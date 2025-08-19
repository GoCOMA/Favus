<<<<<<< HEAD
// 특정 업로드 상태 확인
=======
'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getUploadStatus } from '@/lib/api/uploadApi';
import type { UploadStatus } from '@/lib/types';
import { getStatusColor, getStatusText } from '@/lib/utils';
import { StatusProgressBar } from './components/StatusProgressBar';
import { StatusMessageBox } from './components/StatusMessageBox';
import { StatusSpinner } from './components/StatusSpinner';
import { useWebSocket } from '@/lib/contexts/WebSocketContext';
>>>>>>> parent of 5a8b1ac (Revert "feat: web/ui connect websocket")

interface Props {
  params: { id: string };
}

export default function StatusPage({ params }: Props) {
<<<<<<< HEAD
=======
  const router = useRouter();
  const { subscribe, unsubscribe } = useWebSocket();
  const [status, setStatus] = useState<UploadStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const fetchInitial = async () => {
      try {
        const res = await getUploadStatus(params.id);
        setStatus(res);
        // 업로드가 완료되면 결과 페이지로 이동
        if (res.status === 'completed') {
          setTimeout(() => {
            router.push(`/result/${params.id}`);
          }, 1000); // 1초 후 결과 페이지로 이동
          return;
        }
        // 업로드가 실패하면 에러 표시
        else if (res.status === 'failed') {
          setError(res.error || '업로드에 실패했습니다.');
          return;
        }
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : '상태를 가져오는데 실패했습니다.',
        );
      } finally {
        setLoading(false);
      }
    };

    // 초기 로드
    fetchInitial();
  }, [params.id, router]);

  useEffect(() => {
    if (
      !status ||
      !['pending', 'uploading', 'processing'].includes(status.status)
    )
      return;

    // 글로벌 WebSocket을 통한 실시간 진행률 업데이트
    const handleProgressMessage = (message: any) => {
      if (message.Type === "progress") {
        try {
          const payload = JSON.parse(message.Payload);
          console.log(`[RUN ${message.RunID}] uploaded ${payload.bytes} bytes`);
          
          // 업로드된 바이트를 기반으로 진행률 계산 (가정: 전체 크기 대비)
          const estimatedProgress = Math.min(
            Math.floor((payload.bytes / (10 * 1024 * 1024)) * 100), // 10MB 가정
            100
          );
          
          setStatus(prev => prev ? {
            ...prev,
            progress: estimatedProgress,
            message: `${(payload.bytes / 1024 / 1024).toFixed(2)}MB 업로드 완료`
          } : null);
        } catch (err) {
          console.error("Failed to parse progress payload:", err);
        }
      }
    };

    subscribe(params.id, handleProgressMessage);

    // 폴백: 2초마다 상태 업데이트 (WebSocket 메시지가 없을 때)
    const interval = setInterval(() => {
      getUploadStatus(params.id).then(setStatus).catch(console.error);
    }, 2000);

    return () => {
      unsubscribe(params.id);
      clearInterval(interval);
    };
  }, [status?.status, params.id, subscribe, unsubscribe]);

  if (loading) {
    return (
      <main className="min-h-screen bg-gray-50 py-12">
        <div className="max-w-2xl mx-auto px-4">
          <div className="bg-white rounded-lg shadow-sm p-8">
            <div className="animate-pulse">
              <div className="h-8 bg-gray-200 rounded w-1/3 mb-4"></div>
              <div className="h-4 bg-gray-200 rounded w-1/2 mb-6"></div>
              <div className="h-32 bg-gray-200 rounded"></div>
            </div>
          </div>
        </div>
      </main>
    );
  }

  if (error) {
    return (
      <main className="min-h-screen bg-gray-50 py-12">
        <div className="max-w-2xl mx-auto px-4">
          <div className="bg-white rounded-lg shadow-sm p-8">
            <div className="text-center">
              <div className="text-red-500 text-6xl mb-4">⚠️</div>
              <h1 className="text-2xl font-bold text-gray-900 mb-2">
                오류 발생
              </h1>
              <p className="text-gray-600 mb-6">{error}</p>
              <div className="space-y-3">
                <button
                  onClick={() => router.push('/upload')}
                  className="block w-full bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors"
                >
                  다시 업로드하기
                </button>
                <button
                  onClick={() => router.push('/')}
                  className="block w-full bg-gray-600 text-white px-6 py-2 rounded-lg hover:bg-gray-700 transition-colors"
                >
                  홈으로 돌아가기
                </button>
              </div>
            </div>
          </div>
        </div>
      </main>
    );
  }

  if (!status) {
    return (
      <main className="min-h-screen bg-gray-50 py-12">
        <div className="max-w-2xl mx-auto px-4">
          <div className="bg-white rounded-lg shadow-sm p-8">
            <div className="text-center">
              <div className="text-gray-400 text-6xl mb-4">🔍</div>
              <h1 className="text-2xl font-bold text-gray-900 mb-2">
                업로드 정보를 찾을 수 없습니다
              </h1>
              <p className="text-gray-600 mb-6">ID: {params.id}</p>
              <div className="space-y-3">
                <button
                  onClick={() => router.push('/upload')}
                  className="block w-full bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors"
                >
                  새로 업로드하기
                </button>
                <button
                  onClick={() => router.push('/')}
                  className="block w-full bg-gray-600 text-white px-6 py-2 rounded-lg hover:bg-gray-700 transition-colors"
                >
                  홈으로 돌아가기
                </button>
              </div>
            </div>
          </div>
        </div>
      </main>
    );
  }

>>>>>>> parent of 5a8b1ac (Revert "feat: web/ui connect websocket")
  return (
    <main className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-2xl mx-auto px-4">
        <div className="bg-white rounded-lg shadow-sm p-8">
          <div className="mb-6">
            <h1 className="text-2xl font-bold text-gray-900 mb-2">
              업로드 상태
            </h1>
            <p className="text-gray-600">ID: {status.id}</p>
          </div>

          {/* 상태 표시 */}
          <div className="mb-6">
            <div
              className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(status.status)}`}
            >
              {getStatusText(status.status)}
            </div>
          </div>

          {/* 진행률 바 */}
          <StatusProgressBar progress={status.progress} />

          {/* 메시지 */}
          <StatusMessageBox
            message={status.message}
            retryCount={status.retryCount}
          />

          {/* 시간 정보 */}
          <div className="text-sm text-gray-500 space-y-1">
            <p>생성 시간: {new Date(status.createdAt).toLocaleString()}</p>
            <p>
              마지막 업데이트: {new Date(status.updatedAt).toLocaleString()}
            </p>
          </div>

          {/* 실시간 업데이트 표시 */}
          {['pending', 'uploading', 'processing'].includes(status.status) && (
            <StatusSpinner />
          )}

          {/* 완료 대기 중 */}
          {status.status === 'completed' && (
            <div className="mt-6 p-4 bg-green-50 rounded-lg">
              <div className="flex items-center">
                <div className="text-green-600 text-xl mr-2">✅</div>
                <span className="text-green-800 text-sm">
                  업로드 완료! 결과 페이지로 이동 중...
                </span>
              </div>
            </div>
          )}
        </div>
      </div>
    </main>
  );
}
