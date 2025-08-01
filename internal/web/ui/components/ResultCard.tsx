import Link from 'next/link';

export default function ResultCard() {
  return (
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
  );
}
