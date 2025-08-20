'use client';

import { BatchResult } from '@/lib/types';
import { getStatusColor, getStatusText } from '@/lib/utils';

interface HeaderSectionProps {
  batchResult: BatchResult;
  isSimulationRunning: boolean;
  startSimulation: () => void;
  stopSimulation: () => void;
}

export function HeaderSection({
  batchResult,
  isSimulationRunning,
  startSimulation,
  stopSimulation,
}: HeaderSectionProps) {
  return (
    <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 mb-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-4xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 bg-clip-text text-transparent mb-3">
            {batchResult.metadata.batchName ||
              `배치 처리 ${batchResult.batchId}`}
          </h1>
          <p className="text-gray-600 text-lg">
            {batchResult.metadata.description}
          </p>
        </div>
        <div className="flex items-center gap-4">
          <div
            className={`inline-flex items-center px-4 py-2 rounded-full text-sm font-medium border ${getStatusColor(batchResult.overallStatus)}`}
          >
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
  );
}
