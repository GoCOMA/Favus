'use client';

import UploadStatusList from '@/app/(home)/components/UploadStatusList';
import FeatureGrid from '@/app/(home)/components/FeatureGrid';

//홈 화면
export default function HomePage() {
  return (
    <main className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-4xl mx-auto px-4">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            Favus에 오신 걸 환영합니다
          </h1>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            CLI를 통해 대용량 파일을 안정적으로 업로드하고 결과를 확인하세요.
          </p>
        </div>

        <div className="max-w-4xl mx-auto">
          <UploadStatusList />
        </div>

        {/* <FeatureGrid /> */}
      </div>
    </main>
  );
}
