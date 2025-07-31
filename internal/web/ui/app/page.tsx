'use client';

import Link from 'next/link';
import { initializeMockData } from '@/lib/api';

//홈 화면
export default function HomePage() {
  const handleInitializeMockData = () => {
    initializeMockData();
    alert('목데이터가 초기화되었습니다! 샘플 업로드 ID: sample1, sample2, sample3');
  };

  return (
    <main className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-4xl mx-auto px-4">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">Favus에 오신 걸 환영합니다</h1>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            CLI를 통해 대용량 파일을 안정적으로 업로드하고 결과를 확인하세요.
          </p>
        </div>

        <div className="max-w-2xl mx-auto">
          {/* CLI 업로드 카드 */}
          <div className="bg-white rounded-lg shadow-sm p-8 hover:shadow-md transition-shadow mb-8">
            <div className="text-center">
              <div className="text-green-600 text-5xl mb-4">💻</div>
              <h2 className="text-2xl font-semibold text-gray-900 mb-4">CLI 업로드</h2>
              <p className="text-gray-600 mb-6">
                명령줄에서 고급 기능과 함께 빠르게 파일을 업로드하세요.
              </p>
              <Link
                href="/upload/cli"
                className="inline-flex items-center px-6 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
              >
                CLI 사용법 보기
              </Link>
            </div>
          </div>

          {/* 결과 조회 카드 */}
          <div className="bg-white rounded-lg shadow-sm p-8 hover:shadow-md transition-shadow mb-8">
            <div className="text-center">
              <div className="text-blue-600 text-5xl mb-4">📊</div>
              <h2 className="text-2xl font-semibold text-gray-900 mb-4">결과 조회</h2>
              <p className="text-gray-600 mb-6">
                업로드된 파일의 결과와 다운로드 링크를 확인하세요.
              </p>
              <div className="space-y-3">
                <Link
                  href="/result/sample1"
                  className="inline-flex items-center px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                >
                  샘플 결과 보기
                </Link>
                <p className="text-sm text-gray-500">
                  샘플 ID: sample1, sample2, sample3
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* 테스트용 목데이터 초기화 */}
        <div className="mt-8 bg-yellow-50 border border-yellow-200 rounded-lg p-6 max-w-2xl mx-auto">
          <h3 className="text-lg font-semibold text-yellow-800 mb-2">🧪 테스트용 목데이터</h3>
          <p className="text-yellow-700 mb-4">
            API가 없으므로 테스트용 샘플 데이터를 생성할 수 있습니다.
          </p>
          <div className="flex gap-3">
            <button
              onClick={handleInitializeMockData}
              className="px-4 py-2 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700 transition-colors"
            >
              목데이터 초기화
            </button>
            <div className="text-sm text-yellow-600">
              샘플 ID: sample1, sample2, sample3
            </div>
          </div>
        </div>

        {/* 기능 소개 */}
        <div className="mt-16 bg-white rounded-lg shadow-sm p-8">
          <h2 className="text-2xl font-semibold text-gray-900 mb-6 text-center">주요 기능</h2>
          <div className="grid md:grid-cols-3 gap-6">
            <div className="text-center">
              <div className="text-purple-600 text-3xl mb-3">⚡</div>
              <h3 className="font-semibold text-gray-900 mb-2">빠른 업로드</h3>
              <p className="text-gray-600 text-sm">대용량 파일도 빠르고 안정적으로 업로드</p>
            </div>
            <div className="text-center">
              <div className="text-orange-600 text-3xl mb-3">🔄</div>
              <h3 className="font-semibold text-gray-900 mb-2">자동 재시도</h3>
              <p className="text-gray-600 text-sm">네트워크 오류 시 자동으로 재시도</p>
            </div>
            <div className="text-center">
              <div className="text-teal-600 text-3xl mb-3">📊</div>
              <h3 className="font-semibold text-gray-900 mb-2">결과 확인</h3>
              <p className="text-gray-600 text-sm">업로드 결과와 다운로드 링크 확인</p>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
