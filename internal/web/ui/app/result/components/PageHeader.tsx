export default function PageHeader() {
  return (
    <div className="text-center mb-12">
      <h1 className="text-4xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 bg-clip-text text-transparent mb-4">
        배치 처리 결과 모니터링
      </h1>
      <p className="text-xl text-gray-600 max-w-3xl mx-auto">
        실시간으로 300개 파일의 배치 처리 진행 상황을 확인하고 개별 파일 상태를 모니터링하세요.
      </p>
    </div>
  );
}