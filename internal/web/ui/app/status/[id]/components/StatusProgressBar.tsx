'use client';

interface Props {
  progress: number;
}

export function StatusProgressBar({ progress }: Props) {
  const getProgressColor = (progress: number) => {
    if (progress < 30) return 'bg-red-500';
    if (progress < 60) return 'bg-yellow-500';
    if (progress < 90) return 'bg-blue-600';
    return 'bg-green-500';
  };

  const getProgressText = (progress: number) => {
    if (progress === 0) return '준비 중...';
    if (progress < 100) return `${progress}% 완료`;
    return '업로드 완료!';
  };

  return (
    <div className="mb-6">
      <div className="flex justify-between items-center mb-2">
        <span className="text-sm font-medium text-gray-700">진행률</span>
        <span className="text-sm text-gray-500 font-medium">
          {getProgressText(progress)}
        </span>
      </div>
      <div className="w-full bg-gray-200 rounded-full h-3 shadow-inner">
        <div
          className={`${getProgressColor(progress)} h-3 rounded-full transition-all duration-500 ease-out shadow-sm`}
          style={{ width: `${Math.min(progress, 100)}%` }}
        >
          {progress > 10 && (
            <div className="h-full bg-gradient-to-r from-transparent via-white/20 to-transparent rounded-full"></div>
          )}
        </div>
      </div>
      {progress > 0 && progress < 100 && (
        <div className="mt-1 flex justify-center">
          <div className="flex space-x-1">
            {[0, 1, 2].map((i) => (
              <div
                key={i}
                className={`w-1 h-1 rounded-full bg-blue-400 animate-pulse`}
                style={{ animationDelay: `${i * 0.2}s` }}
              ></div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
