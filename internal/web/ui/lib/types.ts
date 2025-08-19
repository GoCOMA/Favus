export interface UploadResult {
  id: string;
  fileName: string;
  fileSize: number;
  contentType: string;
  s3Url: string;
  downloadUrl: string;
  metadata: Record<string, any>;
  createdAt: string;
  completedAt: string;
}

// 배치 처리 결과를 위한 새로운 인터페이스들
export interface BatchFileItem {
  id: string;
  fileName: string;
  fileSize: number;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  progress: number;
  error?: string;
  downloadUrl?: string;
  s3Url?: string;
  startedAt?: string;
  completedAt?: string;
}

export interface BatchResult {
  batchId: string;
  totalFiles: number;
  completedFiles: number;
  failedFiles: number;
  pendingFiles: number;
  processingFiles: number;
  overallStatus: 'pending' | 'processing' | 'completed' | 'failed';
  overallProgress: number;
  files: BatchFileItem[];
  createdAt: string;
  startedAt: string;
  completedAt?: string;
  metadata: {
    batchName?: string;
    description?: string;
    tags?: string[];
  };
}

export interface UploadStatus {
  id: string;
  status: 'pending' | 'uploading' | 'processing' | 'completed' | 'failed';
  progress: number;
  message?: string;
  retryCount?: number;
  createdAt: string;
  updatedAt: string;
  error?: string;
}
