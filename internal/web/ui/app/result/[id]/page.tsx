'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { use } from 'react';
import { getBatchResult, BatchResult, BatchFileItem, startBatchSimulation, stopBatchSimulation, initializeMockData } from '@/lib/api';

interface Props {
  params: Promise<{ id: string }>;
}

export default function ResultPage({ params }: Props) {
  const router = useRouter();
  const { id } = use(params);
  const [batchResult, setBatchResult] = useState<BatchResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedFile, setSelectedFile] = useState<BatchFileItem | null>(null);
  const [isSimulationRunning, setIsSimulationRunning] = useState(false);

  useEffect(() => {
    const fetchBatchResult = async () => {
      try {
        // 먼저 목데이터 초기화 시도
        initializeMockData();
        
        // 잠시 대기 후 데이터 조회 (초기화 완료 대기)
        await new Promise(resolve => setTimeout(resolve, 100));
        
        const resultData = await getBatchResult(id);
        setBatchResult(resultData);
        setError(null);
      } catch (err) {
        console.error('배치 결과 조회 실패:', err);
        
        // 목데이터가 없는 경우 직접 생성
        try {
          console.log('🔄 직접 배치 데이터 생성 시도...');
          const mockData = createDirectMockData(id);
          setBatchResult(mockData);
          setError(null);
        } catch (directErr) {
          console.error('직접 데이터 생성도 실패:', directErr);
          setError('배치 처리 정보를 찾을 수 없습니다. 목데이터를 초기화해주세요.');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchBatchResult();
  }, [id]);

  // 직접 목데이터 생성 함수
  const createDirectMockData = (batchId: string): BatchResult => {
    let totalFiles = 50; // 기본값
    
    if (batchId === 'batch1') totalFiles = 300;
    else if (batchId === 'batch2') totalFiles = 150;
    else if (batchId === 'batch3') totalFiles = 50;
    else if (batchId === 'sample1') totalFiles = 100;
    else if (batchId === 'sample2') totalFiles = 75;
    else if (batchId === 'sample3') totalFiles = 25;
    
    const files: BatchFileItem[] = [];
    
    for (let i = 0; i < totalFiles; i++) {
      const fileId = `${batchId}_file_${i + 1}`;
      files.push({
        id: fileId,
        fileName: `file_${i + 1}.txt`,
        fileSize: Math.floor(Math.random() * 10 + 1) * 1024 * 1024,
        status: 'pending',
        progress: 0,
      });
    }

    const now = new Date();
    return {
      batchId,
      totalFiles,
      completedFiles: 0,
      failedFiles: 0,
      pendingFiles: totalFiles,
      processingFiles: 0,
      overallStatus: 'pending',
      overallProgress: 0,
      files,
      createdAt: new Date(now.getTime() - 600000).toISOString(),
      startedAt: new Date(now.getTime() - 300000).toISOString(),
      metadata: {
        batchName: `배치 처리 ${batchId}`,
        description: `${totalFiles}개 파일 처리`,
        tags: ['batch', 'processing'],
      },
    };
  };

  // 실시간 시뮬레이션 시작
  const startSimulation = () => {
    if (!batchResult || isSimulationRunning) return;
    
    setIsSimulationRunning(true);
    startBatchSimulation(id, (updatedResult) => {
      setBatchResult({ ...updatedResult });
      
      // 시뮬레이션 완료 시
      if (updatedResult.overallStatus === 'completed') {
        setIsSimulationRunning(false);
      }
    });
  };

  // 시뮬레이션 중지
  const stopSimulation = () => {
    stopBatchSimulation(id);
    setIsSimulationRunning(false);
  };

  // 컴포넌트 언마운트 시 시뮬레이션 정리
  useEffect(() => {
    return () => {
      stopBatchSimulation(id);
    };
  }, [id]);

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getFileIcon = (fileName: string) => {
    const extension = fileName.split('.').pop()?.toLowerCase();
    switch (extension) {
      case 'pdf':
        return '📄';
      case 'doc':
      case 'docx':
        return '📝';
      case 'xls':
      case 'xlsx':
        return '📊';
      case 'ppt':
      case 'pptx':
        return '📈';
      case 'jpg':
      case 'jpeg':
      case 'png':
      case 'gif':
      case 'svg':
        return '🖼️';
      case 'mp4':
      case 'avi':
      case 'mov':
        return '🎥';
      case 'mp3':
      case 'wav':
      case 'flac':
        return '🎵';
      case 'zip':
      case 'rar':
      case '7z':
        return '📦';
      case 'txt':
        return '📄';
      default:
        return '📁';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'text-amber-600 bg-gradient-to-r from-amber-50 to-orange-50 border-amber-200';
      case 'processing':
        return 'text-blue-600 bg-gradient-to-r from-blue-50 to-indigo-50 border-blue-200';
      case 'completed':
        return 'text-emerald-600 bg-gradient-to-r from-emerald-50 to-green-50 border-emerald-200';
      case 'failed':
        return 'text-rose-600 bg-gradient-to-r from-rose-50 to-red-50 border-rose-200';
      default:
        return 'text-gray-600 bg-gradient-to-r from-gray-50 to-slate-50 border-gray-200';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending':
        return '대기 중';
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

  const formatKoreanDateTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60));
    
    if (diffInMinutes < 1) {
      return '방금 전';
    } else if (diffInMinutes < 60) {
      return `${diffInMinutes}분 전`;
    } else if (diffInMinutes < 1440) {
      const hours = Math.floor(diffInMinutes / 60);
      return `${hours}시간 전`;
    } else {
      const days = Math.floor(diffInMinutes / 1440);
      return `${days}일 전`;
    }
  };

  const formatKoreanDate = (dateString: string) => {
    const date = new Date(dateString);
    const year = date.getFullYear();
    const month = date.getMonth() + 1;
    const day = date.getDate();
    const hours = date.getHours().toString().padStart(2, '0');
    const minutes = date.getMinutes().toString().padStart(2, '0');
    
    return `${year}년 ${month}월 ${day}일 ${hours}시 ${minutes}분`;
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      alert('클립보드에 복사되었습니다!');
    } catch (err) {
      console.error('클립보드 복사 실패:', err);
    }
  };

  if (loading) {
    return (
      <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
        <div className="max-w-7xl mx-auto px-4 py-12">
          <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8">
            <div className="animate-pulse space-y-6">
              <div className="h-8 bg-gradient-to-r from-gray-200 to-gray-300 rounded-lg w-1/3"></div>
              <div className="h-4 bg-gradient-to-r from-gray-200 to-gray-300 rounded w-1/2"></div>
              <div className="h-32 bg-gradient-to-r from-gray-200 to-gray-300 rounded-xl"></div>
            </div>
          </div>
        </div>
      </main>
    );
  }

  if (error) {
    return (
      <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
        <div className="max-w-7xl mx-auto px-4 py-12">
          <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8">
            <div className="text-center">
              <div className="text-6xl mb-6">⚠️</div>
              <h1 className="text-3xl font-bold text-gray-900 mb-4">오류 발생</h1>
              <p className="text-gray-600 mb-8 text-lg">{error}</p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
                <button
                  onClick={() => {
                    initializeMockData();
                    window.location.reload();
                  }}
                  className="px-8 py-3 bg-gradient-to-r from-amber-600 to-orange-600 text-white rounded-xl hover:from-amber-700 hover:to-orange-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5"
                >
                  목데이터 초기화
                </button>
                <button
                  onClick={() => router.push('/')}
                  className="px-8 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl hover:from-blue-700 hover:to-indigo-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5"
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

  if (!batchResult) {
    return (
      <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
        <div className="max-w-7xl mx-auto px-4 py-12">
          <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8">
            <div className="text-center">
              <div className="text-6xl mb-6">🔍</div>
              <h1 className="text-3xl font-bold text-gray-900 mb-4">배치 처리 결과를 찾을 수 없습니다</h1>
              <p className="text-gray-600 mb-8 text-lg">ID: {id}</p>
              <button
                onClick={() => router.push('/')}
                className="px-8 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl hover:from-blue-700 hover:to-indigo-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5"
              >
                홈으로 돌아가기
              </button>
            </div>
          </div>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
      <div className="max-w-7xl mx-auto px-4 py-12">
        {/* 헤더 섹션 */}
        <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 mb-8">
          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-4xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 bg-clip-text text-transparent mb-3">
                {batchResult.metadata.batchName || `배치 처리 ${batchResult.batchId}`}
              </h1>
              <p className="text-gray-600 text-lg">{batchResult.metadata.description}</p>
            </div>
            <div className="flex items-center gap-4">
              <div className={`inline-flex items-center px-4 py-2 rounded-full text-sm font-medium border ${getStatusColor(batchResult.overallStatus)}`}>
                {getStatusText(batchResult.overallStatus)}
              </div>
              {isSimulationRunning && (
                <div className="flex items-center text-blue-600 bg-blue-50 px-4 py-2 rounded-full">
                  <div className="animate-spin rounded-full h-4 w-4 border-2 border-blue-600 border-t-transparent mr-2"></div>
                  <span className="text-sm font-medium">실시간 처리 중...</span>
                </div>
              )}
            </div>
          </div>

          {/* 전체 진행률 */}
          <div className="mb-8">
            <div className="flex justify-between items-center mb-3">
              <span className="text-xl font-semibold text-gray-700">전체 진행률</span>
              <span className="text-xl font-bold text-blue-600">{Math.round(batchResult.overallProgress)}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-4 overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-blue-500 to-indigo-600 rounded-full transition-all duration-500 ease-out shadow-lg"
                style={{ width: `${batchResult.overallProgress}%` }}
              ></div>
            </div>
          </div>

          {/* 통계 정보 */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6 mb-8">
            <div className="text-center p-6 bg-gradient-to-br from-emerald-50 to-green-50 rounded-xl border border-emerald-200 shadow-sm hover:shadow-md transition-all duration-300">
              <div className="text-3xl font-bold text-emerald-600 mb-1">{batchResult.completedFiles}</div>
              <div className="text-sm text-emerald-700 font-medium">완료</div>
            </div>
            <div className="text-center p-6 bg-gradient-to-br from-blue-50 to-indigo-50 rounded-xl border border-blue-200 shadow-sm hover:shadow-md transition-all duration-300">
              <div className="text-3xl font-bold text-blue-600 mb-1">{batchResult.processingFiles}</div>
              <div className="text-sm text-blue-700 font-medium">처리 중</div>
            </div>
            <div className="text-center p-6 bg-gradient-to-br from-amber-50 to-orange-50 rounded-xl border border-amber-200 shadow-sm hover:shadow-md transition-all duration-300">
              <div className="text-3xl font-bold text-amber-600 mb-1">{batchResult.pendingFiles}</div>
              <div className="text-sm text-amber-700 font-medium">대기 중</div>
            </div>
            <div className="text-center p-6 bg-gradient-to-br from-rose-50 to-red-50 rounded-xl border border-rose-200 shadow-sm hover:shadow-md transition-all duration-300">
              <div className="text-3xl font-bold text-rose-600 mb-1">{batchResult.failedFiles}</div>
              <div className="text-sm text-rose-700 font-medium">실패</div>
            </div>
          </div>

          {/* 시뮬레이션 컨트롤 */}
          <div className="flex gap-4">
            {!isSimulationRunning && batchResult.overallStatus !== 'completed' && (
              <button
                onClick={startSimulation}
                className="px-8 py-3 bg-gradient-to-r from-emerald-500 to-green-600 text-white rounded-xl hover:from-emerald-600 hover:to-green-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium"
              >
                실시간 시뮬레이션 시작
              </button>
            )}
            {isSimulationRunning && (
              <button
                onClick={stopSimulation}
                className="px-8 py-3 bg-gradient-to-r from-rose-500 to-red-600 text-white rounded-xl hover:from-rose-600 hover:to-red-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium"
              >
                ⏹️ 시뮬레이션 중지
              </button>
            )}
            {batchResult.overallStatus === 'completed' && (
              <div className="px-8 py-3 bg-gradient-to-r from-emerald-100 to-green-100 text-emerald-800 rounded-xl border border-emerald-200 font-medium">
                ✅ 모든 파일 처리 완료!
              </div>
            )}
          </div>
        </div>

        {/* 파일 목록 */}
        <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 mb-8">
          <div className="flex items-center justify-between mb-8">
            <h2 className="text-2xl font-bold text-gray-900">파일 목록 ({batchResult.totalFiles}개)</h2>
            <div className="text-sm text-gray-500 bg-gray-100 px-4 py-2 rounded-full">
              완료: {batchResult.completedFiles} / {batchResult.totalFiles}
            </div>
          </div>

          {/* 파일 그리드 */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 max-h-96 overflow-y-auto pr-2">
            {batchResult.files.map((file) => (
              <div
                key={file.id}
                className={`p-6 border-2 rounded-xl cursor-pointer transition-all duration-300 hover:shadow-lg transform hover:-translate-y-1 ${
                  selectedFile?.id === file.id 
                    ? 'border-blue-500 bg-gradient-to-br from-blue-50 to-indigo-50 shadow-lg' 
                    : 'border-gray-200 bg-white/50 hover:border-gray-300'
                }`}
                onClick={() => setSelectedFile(file)}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-lg">{getFileIcon(file.fileName)}</span>
                    <span className="font-semibold text-sm truncate text-gray-800">{file.fileName}</span>
                  </div>
                  <div className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-medium border ${getStatusColor(file.status)}`}>
                    {getStatusText(file.status)}
                  </div>
                </div>
                
                <div className="text-xs text-gray-500 mb-3 font-medium">
                  {formatFileSize(file.fileSize)}
                </div>

                {/* 개별 파일 진행률 */}
                {file.status === 'processing' && (
                  <div className="mb-3">
                    <div className="flex justify-between text-xs text-gray-500 mb-2 font-medium">
                      <span>진행률</span>
                      <span>{Math.round(file.progress)}%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2 overflow-hidden">
                      <div
                        className="h-full bg-gradient-to-r from-blue-500 to-indigo-600 rounded-full transition-all duration-300 ease-out"
                        style={{ width: `${file.progress}%` }}
                      ></div>
                    </div>
                  </div>
                )}

                {file.status === 'completed' && (
                  <div className="text-xs text-emerald-600 mb-3 font-medium flex items-center">
                    <span className="mr-1">✅</span> 완료됨
                  </div>
                )}

                {file.error && (
                  <div className="text-xs text-rose-600 mt-2 font-medium">
                    {file.error}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* 선택된 파일 상세 정보 */}
        {selectedFile && (
          <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 mb-8">
            <h3 className="text-2xl font-bold text-gray-900 mb-6">파일 상세 정보</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              <div className="bg-white/50 rounded-xl p-6 border border-gray-200">
                <h4 className="font-semibold text-gray-900 mb-4 text-lg">기본 정보</h4>
                <div className="space-y-4 text-sm">
                  <div className="flex justify-between items-center py-2 border-b border-gray-100">
                    <span className="text-gray-600 font-medium">파일명:</span>
                    <div className="flex items-center gap-2">
                      <span className="text-lg">{getFileIcon(selectedFile.fileName)}</span>
                      <span className="font-semibold text-gray-800">{selectedFile.fileName}</span>
                    </div>
                  </div>
                  <div className="flex justify-between items-center py-2 border-b border-gray-100">
                    <span className="text-gray-600 font-medium">파일 크기:</span>
                    <span className="font-semibold text-gray-800">{formatFileSize(selectedFile.fileSize)}</span>
                  </div>
                  <div className="flex justify-between items-center py-2 border-b border-gray-100">
                    <span className="text-gray-600 font-medium">상태:</span>
                    <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-medium border ${getStatusColor(selectedFile.status)}`}>
                      {getStatusText(selectedFile.status)}
                    </span>
                  </div>
                  {selectedFile.startedAt && (
                    <div className="flex justify-between items-center py-2 border-b border-gray-100">
                      <span className="text-gray-600 font-medium">시작 시간:</span>
                      <div className="text-right">
                        <div className="font-semibold text-gray-800">{formatKoreanDate(selectedFile.startedAt)}</div>
                        <div className="text-xs text-gray-500">{formatKoreanDateTime(selectedFile.startedAt)}</div>
                      </div>
                    </div>
                  )}
                  {selectedFile.completedAt && (
                    <div className="flex justify-between items-center py-2">
                      <span className="text-gray-600 font-medium">완료 시간:</span>
                      <div className="text-right">
                        <div className="font-semibold text-gray-800">{formatKoreanDate(selectedFile.completedAt)}</div>
                        <div className="text-xs text-gray-500">{formatKoreanDateTime(selectedFile.completedAt)}</div>
                      </div>
                    </div>
                  )}
                </div>
              </div>

              {selectedFile.status === 'completed' && (
                <div className="bg-white/50 rounded-xl p-6 border border-gray-200">
                  <h4 className="font-semibold text-gray-900 mb-4 text-lg">다운로드</h4>
                  <div className="space-y-4">
                    {selectedFile.downloadUrl && (
                      <a
                        href={selectedFile.downloadUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="block w-full bg-gradient-to-r from-blue-600 to-indigo-600 text-white text-center py-3 px-4 rounded-xl hover:from-blue-700 hover:to-indigo-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium"
                      >
                        📥 파일 다운로드
                      </a>
                    )}
                    {selectedFile.s3Url && (
                      <button
                        onClick={() => copyToClipboard(selectedFile.s3Url!)}
                        className="w-full bg-gradient-to-r from-gray-100 to-gray-200 text-gray-700 py-3 px-4 rounded-xl hover:from-gray-200 hover:to-gray-300 transition-all duration-300 shadow-sm hover:shadow-md font-medium"
                      >
                        🔗 S3 URL 복사
                      </button>
                    )}
                  </div>
                </div>
              )}

              {selectedFile.error && (
                <div className="md:col-span-2">
                  <h4 className="font-semibold text-rose-900 mb-4 text-lg">오류 정보</h4>
                  <div className="bg-gradient-to-r from-rose-50 to-red-50 border border-rose-200 rounded-xl p-6">
                    <p className="text-rose-800 text-sm font-medium">{selectedFile.error}</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* 시간 정보 */}
        <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 mb-8">
          <h3 className="text-2xl font-bold text-gray-900 mb-6">처리 시간</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="bg-white/50 rounded-xl p-6 border border-gray-200">
              <span className="text-gray-600 font-medium text-sm">생성 시간:</span>
              <div className="mt-2">
                <p className="font-semibold text-gray-800">{formatKoreanDate(batchResult.createdAt)}</p>
                <p className="text-xs text-gray-500">{formatKoreanDateTime(batchResult.createdAt)}</p>
              </div>
            </div>
            <div className="bg-white/50 rounded-xl p-6 border border-gray-200">
              <span className="text-gray-600 font-medium text-sm">시작 시간:</span>
              <div className="mt-2">
                <p className="font-semibold text-gray-800">{formatKoreanDate(batchResult.startedAt)}</p>
                <p className="text-xs text-gray-500">{formatKoreanDateTime(batchResult.startedAt)}</p>
              </div>
            </div>
            {batchResult.completedAt && (
              <div className="bg-white/50 rounded-xl p-6 border border-gray-200">
                <span className="text-gray-600 font-medium text-sm">완료 시간:</span>
                <div className="mt-2">
                  <p className="font-semibold text-gray-800">{formatKoreanDate(batchResult.completedAt)}</p>
                  <p className="text-xs text-gray-500">{formatKoreanDateTime(batchResult.completedAt)}</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* 목데이터 안내 */}
        <div className="mb-8 p-6 bg-gradient-to-r from-amber-50 to-orange-50 rounded-2xl border border-amber-200">
          <p className="text-amber-800 text-sm font-medium">
            💡 현재는 목데이터로 표시됩니다. 실제 API 연동 시 실제 데이터가 표시됩니다.
          </p>
        </div>

        {/* 액션 버튼들 */}
        <div className="flex gap-4">
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
