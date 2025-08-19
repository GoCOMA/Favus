'use client';

export function StatusSpinner() {
  return (
    <div className="mt-6 p-4 bg-green-50 rounded-lg">
      <div className="flex items-center">
        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-green-600 mr-2"></div>
        <span className="text-green-800 text-sm">
          실시간으로 업데이트 중...
        </span>
      </div>
      <p className="text-xs text-green-600 mt-1">(목데이터 시뮬레이션)</p>
    </div>
  );
}
