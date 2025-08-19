'use client';

export function LoadingFallback() {
  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
      <div className="max-w-7xl mx-auto px-4 py-12">
        <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8">
          <div className="animate-pulse space-y-6">
            <div className="h-8 bg-gradient-to-r from-gray-200 to-gray-300 rounded-lg w-1/3"></div>
            <div className="h-4 bg-gradient-to-r from-gray-200 to-gray-300 rounded w-1/2"></div>
            <div className="h-32 bg-gradient-to-r from-gray-200 to-gray-300 rounded-xl"></div>
          </div>
        </div>
      </div>
    </main>
  );
}
