'use client';

import { BatchFileItem } from '@/lib/types';
import {
  getFileIcon,
  getStatusColor,
  getStatusText,
  formatFileSize,
  formatKoreanDate,
  formatKoreanDateTime,
} from '@/lib/utils';

interface FileDetailProps {
  file: BatchFileItem;
}

export function FileDetail({ file }: FileDetailProps) {
  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      alert('클립보드에 복사되었습니다!');
    } catch (err) {
      console.error('클립보드 복사 실패:', err);
    }
  };

  return (
    <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 mb-8">
      <h3 className="text-2xl font-bold text-gray-900 mb-6">파일 상세 정보</h3>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        <div className="bg-white/50 rounded-xl p-6 border border-gray-200">
          <h4 className="font-semibold text-gray-900 mb-4 text-lg">
            기본 정보
          </h4>
          <div className="space-y-4 text-sm">
            <div className="flex justify-between items-center py-2 border-b border-gray-100">
              <span className="text-gray-600 font-medium">파일명:</span>
              <div className="flex items-center gap-2">
                <span className="text-lg">{getFileIcon(file.fileName)}</span>
                <span className="font-semibold text-gray-800">
                  {file.fileName}
                </span>
              </div>
            </div>
            <div className="flex justify-between items-center py-2 border-b border-gray-100">
              <span className="text-gray-600 font-medium">파일 크기:</span>
              <span className="font-semibold text-gray-800">
                {formatFileSize(file.fileSize)}
              </span>
            </div>
            <div className="flex justify-between items-center py-2 border-b border-gray-100">
              <span className="text-gray-600 font-medium">상태:</span>
              <span
                className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-medium border ${getStatusColor(file.status)}`}
              >
                {getStatusText(file.status)}
              </span>
            </div>
            {file.startedAt && (
              <div className="flex justify-between items-center py-2 border-b border-gray-100">
                <span className="text-gray-600 font-medium">시작 시간:</span>
                <div className="text-right">
                  <div className="font-semibold text-gray-800">
                    {formatKoreanDate(file.startedAt)}
                  </div>
                  <div className="text-xs text-gray-500">
                    {formatKoreanDateTime(file.startedAt)}
                  </div>
                </div>
              </div>
            )}
            {file.completedAt && (
              <div className="flex justify-between items-center py-2">
                <span className="text-gray-600 font-medium">완료 시간:</span>
                <div className="text-right">
                  <div className="font-semibold text-gray-800">
                    {formatKoreanDate(file.completedAt)}
                  </div>
                  <div className="text-xs text-gray-500">
                    {formatKoreanDateTime(file.completedAt)}
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {file.status === 'completed' && (
          <div className="bg-white/50 rounded-xl p-6 border border-gray-200">
            <h4 className="font-semibold text-gray-900 mb-4 text-lg">
              다운로드
            </h4>
            <div className="space-y-4">
              {file.downloadUrl && (
                <a
                  href={file.downloadUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="block w-full bg-gradient-to-r from-blue-600 to-indigo-600 text-white text-center py-3 px-4 rounded-xl hover:from-blue-700 hover:to-indigo-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium"
                >
                  📥 파일 다운로드
                </a>
              )}
              {file.s3Url && (
                <button
                  onClick={() => copyToClipboard(file.s3Url!)}
                  className="w-full bg-gradient-to-r from-gray-100 to-gray-200 text-gray-700 py-3 px-4 rounded-xl hover:from-gray-200 hover:to-gray-300 transition-all duration-300 shadow-sm hover:shadow-md font-medium"
                >
                  🔗 S3 URL 복사
                </button>
              )}
            </div>
          </div>
        )}

        {file.error && (
          <div className="md:col-span-2">
            <h4 className="font-semibold text-rose-900 mb-4 text-lg">
              오류 정보
            </h4>
            <div className="bg-gradient-to-r from-rose-50 to-red-50 border border-rose-200 rounded-xl p-6">
              <p className="text-rose-800 text-sm font-medium">{file.error}</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
