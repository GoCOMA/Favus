'use client';

import CliCommandBlock from './components/CliCommandBlock';

export default function CliGuidePage() {
  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
      <div className="max-w-4xl mx-auto px-4 py-12 space-y-8">
        <header className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8">
          <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 bg-clip-text text-transparent">
            CLI 업로드 가이드
          </h1>
          <p className="mt-2 text-gray-600 text-lg">
            명령줄 인터페이스(CLI)를 사용해 파일을 업로드하는 방법을 안내합니다.
          </p>
        </header>

        <section className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 space-y-6">
          <h2 className="text-2xl font-semibold text-gray-900">1. CLI 설치</h2>
          <p className="text-gray-700">
            아래 명령어를 실행하여 CLI를 설치합니다. Go 언어 환경이 필요합니다.
          </p>
          <CliCommandBlock command="$ go install github.com/favus/cli@latest" />
        </section>

        <section className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 space-y-6">
          <h2 className="text-2xl font-semibold text-gray-900">
            2. 파일 업로드
          </h2>
          <p className="text-gray-700">
            업로드할 파일 경로를 지정하여 아래와 같이 실행합니다.
          </p>
          <CliCommandBlock command="$ favus upload ./yourfile.zip" />
        </section>

        <footer className="bg-amber-50 border border-amber-200 rounded-xl p-6">
          <p className="text-sm text-amber-800 font-medium">
            💡 업로드 전, CLI 환경변수나 설정파일을 통해 인증 토큰과 업로드
            경로를 설정했는지 확인하세요.
          </p>
        </footer>
      </div>
    </main>
  );
}
