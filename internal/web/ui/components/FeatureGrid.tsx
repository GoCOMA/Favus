const features = [
  {
    icon: '⚡',
    color: 'purple-600',
    title: '빠른 업로드',
    description: '대용량 파일도 빠르고 안정적으로 업로드',
  },
  {
    icon: '🔄',
    color: 'orange-600',
    title: '자동 재시도',
    description: '네트워크 오류 시 자동으로 재시도',
  },
  {
    icon: '📊',
    color: 'teal-600',
    title: '결과 확인',
    description: '업로드 결과와 다운로드 링크 확인',
  },
];

export default function FeatureGrid() {
  return (
    <div className="mt-16 bg-white rounded-lg shadow-sm p-8">
      <h2 className="text-2xl font-semibold text-gray-900 mb-6 text-center">
        주요 기능
      </h2>
      <div className="grid md:grid-cols-3 gap-6">
        {features.map(({ icon, color, title, description }) => (
          <div key={title} className="text-center">
            <div className={`text-${color} text-3xl mb-3`}>{icon}</div>
            <h3 className="font-semibold text-gray-900 mb-2">{title}</h3>
            <p className="text-gray-600 text-sm">{description}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
