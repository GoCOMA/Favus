'use client';

import { BatchResult } from '@/lib/api';
import { formatKoreanDate, formatKoreanDateTime } from '@/lib/utils';

interface TimeInfoProps {
  batchResult: BatchResult;
}

export function TimeInfo({ batchResult }: TimeInfoProps) {
  return (
    <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 mb-8">
      <h3 className="text-2xl font-bold text-gray-900 mb-6">처리 시간</h3>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white/50 rounded-xl p-6 border border-gray-200">
          <span className="text-gray-600 font-medium text-sm">생성 시간:</span>
          <div className="mt-2">
            <p className="font-semibold text-gray-800">
              {formatKoreanDate(batchResult.createdAt)}
            </p>
            <p className="text-xs text-gray-500">
              {formatKoreanDateTime(batchResult.createdAt)}
            </p>
          </div>
        </div>
        <div className="bg-white/50 rounded-xl p-6 border border-gray-200">
          <span className="text-gray-600 font-medium text-sm">시작 시간:</span>
          <div className="mt-2">
            <p className="font-semibold text-gray-800">
              {formatKoreanDate(batchResult.startedAt)}
            </p>
            <p className="text-xs text-gray-500">
              {formatKoreanDateTime(batchResult.startedAt)}
            </p>
          </div>
        </div>
        {batchResult.completedAt && (
          <div className="bg-white/50 rounded-xl p-6 border border-gray-200">
            <span className="text-gray-600 font-medium text-sm">
              완료 시간:
            </span>
            <div className="mt-2">
              <p className="font-semibold text-gray-800">
                {formatKoreanDate(batchResult.completedAt)}
              </p>
              <p className="text-xs text-gray-500">
                {formatKoreanDateTime(batchResult.completedAt)}
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
