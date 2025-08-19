import type { UploadStatus } from '@/lib/types';
const uploadStatusStore = new Map<string, UploadStatus>();

export function initializeUploadMockData() {
  const now = Date.now();
  ['upload1', 'upload2', 'upload3'].forEach((id, index) => {
    uploadStatusStore.set(id, {
      id,
      status: 'processing',
      progress: index * 20 + 30,
      message: `업로드 ${index + 1} 진행 중...`,
      retryCount: index,
      createdAt: new Date(now - 100000).toISOString(),
      updatedAt: new Date(now).toISOString(),
    });
  });
}

export async function getUploadStatus(id: string): Promise<UploadStatus> {
  const data = uploadStatusStore.get(id);
  if (!data) {
    throw new Error('업로드 상태를 찾을 수 없습니다.');
  }

  // 상태 업데이트 시뮬레이션
  if (['pending', 'uploading', 'processing'].includes(data.status)) {
    const nextProgress = Math.min(
      data.progress + Math.floor(Math.random() * 10 + 5),
      100,
    );
    const nextStatus = nextProgress >= 100 ? 'completed' : data.status;

    const updated: UploadStatus = {
      ...data,
      progress: nextProgress,
      updatedAt: new Date().toISOString(),
      status: nextStatus,
    };
    uploadStatusStore.set(id, updated);
    return updated;
  }

  return data;
}
