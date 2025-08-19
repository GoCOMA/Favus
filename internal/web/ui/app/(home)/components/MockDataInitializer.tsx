'use client';

import { useState } from 'react';
import { initializeMockData } from '@/lib/api';

export default function MockDataInitializer() {
  const [initialized, setInitialized] = useState(false);

  const handleInitializeMockData = () => {
    initializeMockData();
    setInitialized(true);
  };

  return (
    <div className="mt-8 bg-yellow-50 border border-yellow-200 rounded-lg p-6 max-w-2xl mx-auto">
      <h3 className="text-lg font-semibold text-yellow-800 mb-2">
        🧪 테스트용 목데이터
      </h3>
      <p className="text-yellow-700 mb-4">
        API가 없으므로 테스트용 샘플 데이터를 생성할 수 있습니다.
      </p>
      <div className="flex gap-3">
        <button
          type="button"
          onClick={handleInitializeMockData}
          className="px-4 py-2 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700 transition-colors"
        >
          목데이터 초기화
        </button>
        <div className="text-sm text-yellow-600">
          샘플 ID: sample1, sample2, sample3
        </div>
      </div>
      {initialized && (
        <p className="text-green-600 text-sm mt-2">
          목데이터가 초기화되었습니다!
        </p>
      )}
    </div>
  );
}
