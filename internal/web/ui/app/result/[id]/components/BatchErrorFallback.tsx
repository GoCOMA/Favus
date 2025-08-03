'use client';

import { initializeMockData } from '@/lib/api';
import { useRouter } from 'next/navigation';

export function ErrorFallback({
  id,
  router,
}: {
  id: string;
  router: ReturnType<typeof useRouter>;
}) {
  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
      <div className="max-w-7xl mx-auto px-4 py-12">
        <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8">
          <div className="text-center">
            <div className="text-6xl mb-6">🔍</div>
            <h1 className="text-3xl font-bold text-gray-900 mb-4">
              배치 처리 결과를 찾을 수 없습니다
            </h1>
            <p className="text-gray-600 mb-8 text-lg">ID: {id}</p>
            <button
              onClick={() => router.push('/')}
              className="px-8 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl hover:from-blue-700 hover:to-indigo-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5"
            >
              홈으로 돌아가기
            </button>
          </div>
        </div>
      </div>
    </main>
  );
}
