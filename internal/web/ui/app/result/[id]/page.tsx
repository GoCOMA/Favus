'use client';

import { use, useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  getBatchResult,
  startBatchSimulation,
  stopBatchSimulation,
  initializeMockData,
} from '@/lib/api';
import { BatchResult, BatchFileItem } from '@/lib/types';
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
  const [batchResult, setBatchResult] = useState<BatchResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedFile, setSelectedFile] = useState<BatchFileItem | null>(null);
  const [isSimulationRunning, setIsSimulationRunning] = useState(false);

  useEffect(() => {
    const fetchBatchResult = async () => {
      try {
        initializeMockData();
        await new Promise((resolve) => setTimeout(resolve, 100));
        const resultData = await getBatchResult(id);
        setBatchResult(resultData);
        setError(null);
      } catch (err) {
        try {
          const mockData = createDirectMockData(id);
          setBatchResult(mockData);
          setError(null);
        } catch (directErr) {
          setError(
            '배치 처리 정보를 찾을 수 없습니다. 목데이터를 초기화해주세요.',
          );
        }
      } finally {
        setLoading(false);
      }
    };
    fetchBatchResult();
  }, [id]);

  const createDirectMockData = (batchId: string): BatchResult => {
    let totalFiles = 50;

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
      console.log(fileId);
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

  const startSimulation = () => {
    if (!batchResult || isSimulationRunning) return;

    setIsSimulationRunning(true);
    startBatchSimulation(id, (updatedResult) => {
      setBatchResult({ ...updatedResult });

      if (updatedResult.overallStatus === 'completed') {
        setIsSimulationRunning(false);
      }
    });
  };

  const stopSimulation = () => {
    stopBatchSimulation(id);
    setIsSimulationRunning(false);
  };

  useEffect(() => {
    return () => {
      stopBatchSimulation(id);
    };
  }, [id]);

  if (loading) return <LoadingFallback />;
  if (error) return <ErrorFallback error={error} id={id} router={router} />;
  if (!batchResult) return <BatchErrorFallback id={id} router={router} />;

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
      <div className="max-w-7xl mx-auto px-4 py-12">
        <HeaderSection
          batchResult={batchResult}
          isSimulationRunning={isSimulationRunning}
          startSimulation={startSimulation}
          stopSimulation={stopSimulation}
        />
        <SummarySection batchResult={batchResult} />
        <FileList
          batchResult={batchResult}
          selectedFile={selectedFile}
          setSelectedFile={setSelectedFile}
        />
        {selectedFile && <FileDetail file={selectedFile} />}
        <TimeInfo batchResult={batchResult} />
        <div className="mb-8 p-6 bg-gradient-to-r from-amber-50 to-orange-50 rounded-2xl border border-amber-200">
          <p className="text-amber-800 text-sm font-medium">
            💡 현재는 목데이터로 표시됩니다. 실제 API 연동 시 실제 데이터가
            표시됩니다.
          </p>
        </div>
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
