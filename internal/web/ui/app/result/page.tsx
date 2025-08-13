import PageHeader from './components/PageHeader';
import BatchCard from './components/BatchCard';
import MockDataSection from './components/MockDataSection';
import FeaturesSection from './components/FeaturesSection';
import HomeButton from './components/HomeButton';

export default function ResultHomePage() {
  const batchCards = [
    {
      emoji: '📊',
      title: '대용량 배치',
      description: '300개 파일의 대규모 배치 처리 결과를 확인하세요.',
      href: '/result/batch1',
      buttonText: '300개 파일 (batch1)',
      buttonColor: 'from-blue-600 to-indigo-600',
      hoverColor: 'from-blue-700 to-indigo-700',
      subtitle: '실시간 시뮬레이션 가능',
    },
    {
      emoji: '📈',
      title: '중간 규모 배치',
      description: '150개 파일의 중간 규모 배치 처리 결과를 확인하세요.',
      href: '/result/batch2',
      buttonText: '150개 파일 (batch2)',
      buttonColor: 'from-emerald-600 to-green-600',
      hoverColor: 'from-emerald-700 to-green-700',
      subtitle: '빠른 처리 시뮬레이션',
    },
    {
      emoji: '⚡',
      title: '소규모 배치',
      description: '50개 파일의 소규모 배치 처리 결과를 확인하세요.',
      href: '/result/batch3',
      buttonText: '50개 파일 (batch3)',
      buttonColor: 'from-purple-600 to-pink-600',
      hoverColor: 'from-purple-700 to-pink-700',
      subtitle: '즉시 완료 시뮬레이션',
    },
  ];

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
      <div className="max-w-6xl mx-auto px-4 py-12">
        <PageHeader />

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
          {batchCards.map((card, index) => (
            <BatchCard key={index} {...card} />
          ))}
        </div>

        <MockDataSection />
        <FeaturesSection />
        <HomeButton />
      </div>
    </main>
  );
} 