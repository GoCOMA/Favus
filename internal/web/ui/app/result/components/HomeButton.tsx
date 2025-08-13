import Link from 'next/link';

export default function HomeButton() {
  return (
    <div className="mt-12 text-center">
      <Link
        href="/"
        className="inline-flex items-center px-8 py-3 bg-gradient-to-r from-gray-600 to-slate-700 text-white rounded-xl hover:from-gray-700 hover:to-slate-800 transition-all duration-300 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 font-medium"
      >
        ← 메인 홈으로 돌아가기
      </Link>
    </div>
  );
}