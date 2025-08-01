import Link from 'next/link';

export default function CliUploadCard() {
  return (
    <div className="bg-white rounded-lg shadow-sm p-8 hover:shadow-md transition-shadow mb-8">
      <div className="text-center">
        <div className="text-green-600 text-5xl mb-4">💻</div>
        <h2 className="text-2xl font-semibold text-gray-900 mb-4">
          CLI 업로드
        </h2>
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
  );
}
