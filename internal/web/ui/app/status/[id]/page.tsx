'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getUploadStatus, UploadStatus } from '@/lib/api';

interface Props {
  params: { id: string };
}

export default function StatusPage({ params }: Props) {
  const router = useRouter();
  const [status, setStatus] = useState<UploadStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let intervalId: NodeJS.Timeout;

    const fetchStatus = async () => {
      try {
        const statusData = await getUploadStatus(params.id);
        setStatus(statusData);
        setError(null);

        // 업로드가 완료되면 결과 페이지로 이동
        if (statusData.status === 'completed') {
          setTimeout(() => {
            router.push(`/result/${params.id}`);
          }, 1000); // 1초 후 결과 페이지로 이동
          return;
        }

        // 업로드가 실패하면 에러 표시
        if (statusData.status === 'failed') {
          setError(statusData.error || '업로드에 실패했습니다.');
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
    fetchStatus();

    // 2초마다 상태 업데이트 (pending, uploading, processing 상태일 때만)
    intervalId = setInterval(() => {
      if (
        status &&
        ['pending', 'uploading', 'processing'].includes(status.status)
      ) {
        fetchStatus();
      }
    }, 2000);

    return () => {
      if (intervalId) {
        clearInterval(intervalId);
      }
    };
  }, [params.id, router, status?.status]);

  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending':
        return '대기 중';
      case 'uploading':
        return '업로드 중';
      case 'processing':
        return '처리 중';
      case 'completed':
        return '완료';
      case 'failed':
        return '실패';
      default:
        return '알 수 없음';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'text-yellow-600 bg-yellow-100';
      case 'uploading':
      case 'processing':
        return 'text-blue-600 bg-blue-100';
      case 'completed':
        return 'text-green-600 bg-green-100';
      case 'failed':
        return 'text-red-600 bg-red-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };

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
          <div className="mb-6">
            <div className="flex justify-between items-center mb-2">
              <span className="text-sm font-medium text-gray-700">진행률</span>
              <span className="text-sm text-gray-500">{status.progress}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                style={{ width: `${status.progress}%` }}
              ></div>
            </div>
          </div>

          {/* 메시지 */}
          {status.message && (
            <div className="mb-6 p-4 bg-blue-50 rounded-lg">
              <p className="text-blue-800">{status.message}</p>
            </div>
          )}

          {/* 재시도 정보 */}
          {status.retryCount !== undefined && status.retryCount > 0 && (
            <div className="mb-6 p-4 bg-yellow-50 rounded-lg">
              <p className="text-yellow-800">
                재시도 횟수: {status.retryCount}회
              </p>
            </div>
          )}

          {/* 시간 정보 */}
          <div className="text-sm text-gray-500 space-y-1">
            <p>생성 시간: {new Date(status.createdAt).toLocaleString()}</p>
            <p>
              마지막 업데이트: {new Date(status.updatedAt).toLocaleString()}
            </p>
          </div>

          {/* 실시간 업데이트 표시 */}
          {['pending', 'uploading', 'processing'].includes(status.status) && (
            <div className="mt-6 p-4 bg-green-50 rounded-lg">
              <div className="flex items-center">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-green-600 mr-2"></div>
                <span className="text-green-800 text-sm">
                  실시간으로 업데이트 중...
                </span>
              </div>
              <p className="text-xs text-green-600 mt-1">
                (목데이터 시뮬레이션)
              </p>
            </div>
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
