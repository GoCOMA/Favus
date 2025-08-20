'use client';

import { BatchResult } from '@/lib/types';

interface SummarySectionProps {
  batchResult: BatchResult;
}

export function SummarySection({ batchResult }: SummarySectionProps) {
  return (
    <div className="mb-8">
      <div className="flex justify-between items-center mb-3">
        <span className="text-xl font-semibold text-gray-700">전체 진행률</span>
        <span className="text-xl font-bold text-blue-600">
          {Math.round(batchResult.overallProgress)}%
        </span>
      </div>
      <div className="w-full bg-gray-200 rounded-full h-4 overflow-hidden">
        <div
          className="h-full bg-gradient-to-r from-blue-500 to-indigo-600 rounded-full transition-all duration-500 ease-out shadow-lg"
          style={{ width: `${batchResult.overallProgress}%` }}
        ></div>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-6 mt-8">
        <div className="text-center p-6 bg-gradient-to-br from-emerald-50 to-green-50 rounded-xl border border-emerald-200 shadow-sm hover:shadow-md transition-all duration-300">
          <div className="text-3xl font-bold text-emerald-600 mb-1">
            {batchResult.completedFiles}
          </div>
          <div className="text-sm text-emerald-700 font-medium">완료</div>
        </div>
        <div className="text-center p-6 bg-gradient-to-br from-blue-50 to-indigo-50 rounded-xl border border-blue-200 shadow-sm hover:shadow-md transition-all duration-300">
          <div className="text-3xl font-bold text-blue-600 mb-1">
            {batchResult.processingFiles}
          </div>
          <div className="text-sm text-blue-700 font-medium">처리 중</div>
        </div>
        <div className="text-center p-6 bg-gradient-to-br from-amber-50 to-orange-50 rounded-xl border border-amber-200 shadow-sm hover:shadow-md transition-all duration-300">
          <div className="text-3xl font-bold text-amber-600 mb-1">
            {batchResult.pendingFiles}
          </div>
          <div className="text-sm text-amber-700 font-medium">대기 중</div>
        </div>
        <div className="text-center p-6 bg-gradient-to-br from-rose-50 to-red-50 rounded-xl border border-rose-200 shadow-sm hover:shadow-md transition-all duration-300">
          <div className="text-3xl font-bold text-rose-600 mb-1">
            {batchResult.failedFiles}
          </div>
          <div className="text-sm text-rose-700 font-medium">실패</div>
        </div>
      </div>
    </div>
  );
}
