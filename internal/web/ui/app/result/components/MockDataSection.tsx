'use client';

import { initializeMockData } from '@/lib/api';

export default function MockDataSection() {
  const handleInitializeMockData = () => {
    initializeMockData();
    alert('목데이터가 초기화되었습니다! 배치 처리 ID: batch1, batch2, batch3');
  };

  return (
    <div className="bg-gradient-to-r from-amber-50 to-orange-50 rounded-2xl border border-amber-200 p-8 mb-12">
      <div className="text-center">
        <h3 className="text-2xl font-bold text-amber-800 mb-4">🧪 테스트용 목데이터</h3>
        <p className="text-amber-700 mb-6 text-lg">
          API가 없으므로 테스트용 샘플 데이터를 생성할 수 있습니다.
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
          <button
            onClick={handleInitializeMockData}
            className="px-8 py-3 bg-gradient-to-r from-amber-600 to-orange-600 text-white rounded-xl hover:from-amber-700 hover:to-orange-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium"
          >
            목데이터 초기화
          </button>
          <div className="text-sm text-amber-600 bg-amber-100 px-4 py-2 rounded-lg">
            배치 ID: batch1, batch2, batch3
          </div>
        </div>
      </div>
    </div>
  );
}