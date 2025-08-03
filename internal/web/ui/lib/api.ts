import { BatchFileItem, BatchResult, UploadResult } from './types';

// 목데이터 저장소
const mockUploads = new Map<string, { result: UploadResult }>();
const mockBatchResults = new Map<string, BatchResult>();
const activeSimulations = new Map<string, NodeJS.Timeout>();

// 업로드 결과 조회 (목데이터)
export async function getUploadResult(id: string): Promise<UploadResult> {
  const upload = mockUploads.get(id);
  if (!upload) {
    throw new Error('업로드 정보를 찾을 수 없습니다.');
  }
  return upload.result;
}

// 배치 처리 결과 조회 (목데이터)
export async function getBatchResult(batchId: string): Promise<BatchResult> {
  const batch = mockBatchResults.get(batchId);
  if (!batch) {
    throw new Error('배치 처리 정보를 찾을 수 없습니다.');
  }
  return batch;
}

// 실시간 배치 처리 시뮬레이션 시작
export function startBatchSimulation(
  batchId: string,
  onUpdate?: (result: BatchResult) => void,
) {
  const batch = mockBatchResults.get(batchId);
  if (!batch) return;

  // 기존 시뮬레이션 중지
  stopBatchSimulation(batchId);
  resetBatchFiles(batch);

  let currentFileIndex = 0;
  let processingCount = 0;
  const maxConcurrent = 3; // 동시에 처리할 수 있는 파일 수

  const processNextFile = () => {
    if (currentFileIndex >= batch.totalFiles) {
      // 모든 파일 처리 완료
      if (batch.completedFiles + batch.failedFiles === batch.totalFiles) {
        batch.overallStatus = 'completed';
        batch.completedAt = new Date().toISOString();
        onUpdate?.(batch);
      }
      return;
    }

    // 동시 처리 제한 확인
    if (processingCount >= maxConcurrent) {
      setTimeout(processNextFile, 1000);
      return;
    }

    const file = batch.files[currentFileIndex];
    currentFileIndex++;
    processingCount++;

    // 파일 처리 시작
    file.status = 'processing';
    file.progress = 0;
    file.startedAt = new Date().toISOString();
    batch.pendingFiles--;
    batch.processingFiles++;
    batch.overallStatus = 'processing';

    // 진행률 시뮬레이션
    const progressInterval = setInterval(() => {
      file.progress += Math.random() * 15 + 5; // 5-20%씩 증가

      if (file.progress >= 100) {
        file.progress = 100;
        clearInterval(progressInterval);

        // 파일 완료 처리
        setTimeout(() => {
          const shouldFail = Math.random() < 0.05; // 5% 확률로 실패

          if (shouldFail) {
            file.status = 'failed';
            file.error = '처리 중 오류가 발생했습니다.';
            batch.failedFiles++;
          } else {
            file.status = 'completed';
            file.downloadUrl = `https://mock-download.example.com/files/${file.id}`;
            file.s3Url = `https://mock-s3.amazonaws.com/bucket/uploads/${file.id}`;
            batch.completedFiles++;
          }

          file.completedAt = new Date().toISOString();
          batch.processingFiles--;
          processingCount--;

          // 전체 진행률 업데이트
          batch.overallProgress =
            ((batch.completedFiles + batch.failedFiles) / batch.totalFiles) *
            100;

          // 콜백 호출
          onUpdate?.(batch);

          // 다음 파일 처리
          setTimeout(processNextFile, 500);
        }, 200);
      } else {
        // 진행률 업데이트
        onUpdate?.(batch);
      }
    }, 300); // 300ms마다 진행률 업데이트
  };

  // 첫 번째 파일 처리 시작
  setTimeout(processNextFile, 1000);

  // 시뮬레이션 ID 저장
  activeSimulations.set(
    batchId,
    setTimeout(() => {}, 0),
  ); // 더미 타이머
}

// 배치 시뮬레이션 중지
export function stopBatchSimulation(batchId: string) {
  const simulation = activeSimulations.get(batchId);
  if (simulation) {
    clearTimeout(simulation);
    activeSimulations.delete(batchId);
  }
}

// 목데이터 초기화 (테스트용)
export function initializeMockData() {
  // console.log('�� initializeMockData 호출됨');
  // 기존 데이터 초기화
  mockUploads.clear();
  mockBatchResults.clear();

  // 기존 시뮬레이션 중지
  activeSimulations.forEach((_, batchId) => {
    stopBatchSimulation(batchId);
  });

  // 단일 파일 샘플 데이터 추가
  const sampleIds = ['sample1', 'sample2', 'sample3'];

  sampleIds.forEach((id, index) => {
    const result: UploadResult = {
      id,
      fileName: `sample-file-${index + 1}.txt`,
      fileSize: (index + 1) * 1024 * 1024, // 1MB, 2MB, 3MB
      contentType: 'text/plain',
      s3Url: `https://mock-s3.amazonaws.com/bucket/uploads/${id}/sample-file-${index + 1}.txt`,
      downloadUrl: `https://mock-download.example.com/files/${id}`,
      metadata: {
        uploadedBy: 'user@example.com',
        originalName: `sample-file-${index + 1}.txt`,
        checksum: `checksum${index + 1}`,
        tags: ['sample', 'test'],
        description: `샘플 파일 ${index + 1}`,
      },
      createdAt: new Date(Date.now() - (index + 1) * 60000).toISOString(),
      completedAt: new Date(Date.now() - index * 60000).toISOString(),
    };

    mockUploads.set(id, { result });
  });

  // 배치 처리 샘플 데이터 생성
  createMockBatchResult('batch1', 300);
  createMockBatchResult('batch2', 150);
  createMockBatchResult('batch3', 50);

  // sample 배치도 추가 (기존 코드와의 호환성)
  createMockBatchResult('sample1', 100);
  createMockBatchResult('sample2', 75);
  createMockBatchResult('sample3', 25);

  // console.log('✅ initializeMockData 완료');
}

function resetBatchFiles(batch: BatchResult) {
  batch.files.forEach((file) => {
    file.status = 'pending';
    file.progress = 0;
    file.startedAt = undefined;
    file.completedAt = undefined;
    file.error = undefined;
    file.downloadUrl = undefined;
    file.s3Url = undefined;
  });

  batch.overallStatus = 'pending';
  batch.overallProgress = 0;
  batch.completedFiles = 0;
  batch.failedFiles = 0;
  batch.pendingFiles = batch.totalFiles;
  batch.processingFiles = 0;
  batch.completedAt = undefined;
}

// 배치 처리 목데이터 생성 함수
function createMockBatchResult(batchId: string, totalFiles: number) {
  const files: BatchFileItem[] = Array.from({ length: totalFiles }).map(
    (_, i) => ({
      id: `${batchId}_file_${i + 1}`,
      fileName: `file_${i + 1}.txt`,
      fileSize: Math.floor(Math.random() * 10 + 1) * 1024 * 1024, // 1-10MB
      status: 'pending',
      progress: 0,
    }),
  );

  const now = new Date();
  const batchResult: BatchResult = {
    batchId,
    totalFiles,
    completedFiles: 0,
    failedFiles: 0,
    pendingFiles: totalFiles,
    processingFiles: 0,
    overallStatus: 'pending',
    overallProgress: 0,
    files,
    createdAt: new Date(now.getTime() - 600000).toISOString(), // 10분 전 생성
    startedAt: new Date(now.getTime() - 300000).toISOString(), // 5분 전 시작
    metadata: {
      batchName: `배치 처리 ${batchId}`,
      description: `${totalFiles}개 파일 처리`,
      tags: ['batch', 'processing'],
    },
  };

  mockBatchResults.set(batchId, batchResult);
}

// 랜덤 상태 생성 (더 현실적인 분포)
// function getRandomStatus(
//   index: number,
//   total: number,
// ): 'pending' | 'processing' | 'completed' | 'failed' {
//   const progress = index / total;

//   if (progress < 0.7) {
//     return 'completed';
//   } else if (progress < 0.85) {
//     return 'processing';
//   } else if (progress < 0.95) {
//     return 'pending';
//   } else {
//     return 'failed';
//   }
// }
