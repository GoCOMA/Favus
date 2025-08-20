'use client';

import { BatchResult, BatchFileItem } from '@/lib/types';
import {
  getFileIcon,
  getStatusColor,
  getStatusText,
  formatFileSize,
} from '@/lib/utils';

interface FileListProps {
  batchResult: BatchResult;
  selectedFile: BatchFileItem | null;
  setSelectedFile: (file: BatchFileItem) => void;
}

export function FileList({
  batchResult,
  selectedFile,
  setSelectedFile,
}: FileListProps) {
  return (
    <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 mb-8">
      <div className="flex items-center justify-between mb-8">
        <h2 className="text-2xl font-bold text-gray-900">
          파일 목록 ({batchResult.totalFiles}개)
        </h2>
        <div className="text-sm text-gray-500 bg-gray-100 px-4 py-2 rounded-full">
          완료: {batchResult.completedFiles} / {batchResult.totalFiles}
        </div>
      </div>

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
                <span className="font-semibold text-sm truncate text-gray-800">
                  {file.fileName}
                </span>
              </div>
              <div
                className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-medium border ${getStatusColor(file.status)}`}
              >
                {getStatusText(file.status)}
              </div>
            </div>

            <div className="text-xs text-gray-500 mb-3 font-medium">
              {formatFileSize(file.fileSize)}
            </div>

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
  );
}
