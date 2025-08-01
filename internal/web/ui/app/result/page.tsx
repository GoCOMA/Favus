'use client';

import Link from 'next/link';
import { initializeMockData } from '@/lib/api';

export default function ResultHomePage() {
  const handleInitializeMockData = () => {
    initializeMockData();
    alert('목데이터가 초기화되었습니다! 배치 처리 ID: batch1, batch2, batch3');
  };

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
      <div className="max-w-6xl mx-auto px-4 py-12">
        {/* 헤더 */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 bg-clip-text text-transparent mb-4">
            배치 처리 결과 모니터링
          </h1>
          <p className="text-xl text-gray-600 max-w-3xl mx-auto">
            실시간으로 300개 파일의 배치 처리 진행 상황을 확인하고 개별 파일 상태를 모니터링하세요.
          </p>
        </div>

        {/* 배치 처리 결과 카드들 */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
          {/* 300개 파일 배치 */}
          <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 hover:shadow-2xl transition-all duration-300 transform hover:-translate-y-2">
            <div className="text-center">
              <div className="text-6xl mb-6">📊</div>
              <h2 className="text-2xl font-bold text-gray-900 mb-4">대용량 배치</h2>
              <p className="text-gray-600 mb-6">
                300개 파일의 대규모 배치 처리 결과를 확인하세요.
              </p>
              <div className="space-y-3">
                <Link
                  href="/result/batch1"
                  className="block w-full px-6 py-3 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl hover:from-blue-700 hover:to-indigo-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium"
                >
                  300개 파일 (batch1)
                </Link>
                <div className="text-sm text-gray-500">
                  실시간 시뮬레이션 가능
                </div>
              </div>
            </div>
          </div>

          {/* 150개 파일 배치 */}
          <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 hover:shadow-2xl transition-all duration-300 transform hover:-translate-y-2">
            <div className="text-center">
              <div className="text-6xl mb-6">📈</div>
              <h2 className="text-2xl font-bold text-gray-900 mb-4">중간 규모 배치</h2>
              <p className="text-gray-600 mb-6">
                150개 파일의 중간 규모 배치 처리 결과를 확인하세요.
              </p>
              <div className="space-y-3">
                <Link
                  href="/result/batch2"
                  className="block w-full px-6 py-3 bg-gradient-to-r from-emerald-600 to-green-600 text-white rounded-xl hover:from-emerald-700 hover:to-green-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium"
                >
                  150개 파일 (batch2)
                </Link>
                <div className="text-sm text-gray-500">
                  빠른 처리 시뮬레이션
                </div>
              </div>
            </div>
          </div>

          {/* 50개 파일 배치 */}
          <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 hover:shadow-2xl transition-all duration-300 transform hover:-translate-y-2">
            <div className="text-center">
              <div className="text-6xl mb-6">⚡</div>
              <h2 className="text-2xl font-bold text-gray-900 mb-4">소규모 배치</h2>
              <p className="text-gray-600 mb-6">
                50개 파일의 소규모 배치 처리 결과를 확인하세요.
              </p>
              <div className="space-y-3">
                <Link
                  href="/result/batch3"
                  className="block w-full px-6 py-3 bg-gradient-to-r from-purple-600 to-pink-600 text-white rounded-xl hover:from-purple-700 hover:to-pink-700 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium"
                >
                  50개 파일 (batch3)
                </Link>
                <div className="text-sm text-gray-500">
                  즉시 완료 시뮬레이션
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* 테스트용 목데이터 초기화 */}
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

        {/* 기능 소개 */}
        <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8">
          <h2 className="text-3xl font-bold text-gray-900 mb-8 text-center">주요 기능</h2>
          <div className="grid md:grid-cols-3 gap-8">
            <div className="text-center">
              <div className="text-blue-600 text-4xl mb-4">🔄</div>
              <h3 className="text-xl font-semibold text-gray-900 mb-3">실시간 시뮬레이션</h3>
              <p className="text-gray-600">
                파일들이 하나씩 완료되는 과정을 실시간으로 시뮬레이션하여 확인할 수 있습니다.
              </p>
            </div>
            <div className="text-center">
              <div className="text-emerald-600 text-4xl mb-4">📊</div>
              <h3 className="text-xl font-semibold text-gray-900 mb-3">개별 파일 모니터링</h3>
              <p className="text-gray-600">
                각 파일의 진행률, 상태, 완료 시간을 개별적으로 확인할 수 있습니다.
              </p>
            </div>
            <div className="text-center">
              <div className="text-purple-600 text-4xl mb-4">🎯</div>
              <h3 className="text-xl font-semibold text-gray-900 mb-3">한국어 인터페이스</h3>
              <p className="text-gray-600">
                한국어 시간 형식과 직관적인 UI로 사용자 친화적인 경험을 제공합니다.
              </p>
            </div>
          </div>
        </div>

        {/* 홈으로 돌아가기 */}
        <div className="mt-12 text-center">
          <Link
            href="/"
            className="inline-flex items-center px-8 py-3 bg-gradient-to-r from-gray-600 to-slate-700 text-white rounded-xl hover:from-gray-700 hover:to-slate-800 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium"
          >
            ← 메인 홈으로 돌아가기
          </Link>
        </div>
      </div>
    </main>
  );
} 