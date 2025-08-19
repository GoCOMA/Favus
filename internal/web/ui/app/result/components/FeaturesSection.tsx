interface FeatureProps {
  emoji: string;
  title: string;
  description: string;
  color: string;
}

function FeatureCard({ emoji, title, description, color }: FeatureProps) {
  return (
    <div className="text-center">
      <div className={`${color} text-4xl mb-4`}>{emoji}</div>
      <h3 className="text-xl font-semibold text-gray-900 mb-3">{title}</h3>
      <p className="text-gray-600">{description}</p>
    </div>
  );
}

export default function FeaturesSection() {
  const features = [
    {
      emoji: '🔄',
      title: '실시간 시뮬레이션',
      description: '파일들이 하나씩 완료되는 과정을 실시간으로 시뮬레이션하여 확인할 수 있습니다.',
      color: 'text-blue-600',
    },
    {
      emoji: '📊',
      title: '개별 파일 모니터링',
      description: '각 파일의 진행률, 상태, 완료 시간을 개별적으로 확인할 수 있습니다.',
      color: 'text-emerald-600',
    },
    {
      emoji: '🎯',
      title: '한국어 인터페이스',
      description: '한국어 시간 형식과 직관적인 UI로 사용자 친화적인 경험을 제공합니다.',
      color: 'text-purple-600',
    },
  ];

  return (
    <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8">
      <h2 className="text-3xl font-bold text-gray-900 mb-8 text-center">주요 기능</h2>
      <div className="grid md:grid-cols-3 gap-8">
        {features.map((feature, index) => (
          <FeatureCard key={index} {...feature} />
        ))}
      </div>
    </div>
  );
}