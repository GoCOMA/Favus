export default function FeatureGrid() {
  return (
    <div className="mt-16 bg-white rounded-lg shadow-sm p-8">
      <h2 className="text-2xl font-semibold text-gray-900 mb-6 text-center">
        주요 기능
      </h2>
      <div className="grid md:grid-cols-3 gap-6">
        <div className="text-center">
          <div className="text-purple-600 text-3xl mb-3">⚡</div>
          <h3 className="font-semibold text-gray-900 mb-2">빠른 업로드</h3>
          <p className="text-gray-600 text-sm">
            대용량 파일도 빠르고 안정적으로 업로드
          </p>
        </div>
        <div className="text-center">
          <div className="text-orange-600 text-3xl mb-3">🔄</div>
          <h3 className="font-semibold text-gray-900 mb-2">자동 재시도</h3>
          <p className="text-gray-600 text-sm">
            네트워크 오류 시 자동으로 재시도
          </p>
        </div>
        <div className="text-center">
          <div className="text-teal-600 text-3xl mb-3">📊</div>
          <h3 className="font-semibold text-gray-900 mb-2">결과 확인</h3>
          <p className="text-gray-600 text-sm">
            업로드 결과와 다운로드 링크 확인
          </p>
        </div>
      </div>
    </div>
  );
}
