'use client';

interface UploadToolbarProps {
  canStart: boolean;
  canClear: boolean;
  onStart: () => void;
  onClear: () => void;
}

export default function UploadToolbar({
  canStart,
  canClear,
  onStart,
  onClear,
}: UploadToolbarProps) {
  return (
    <div className="flex items-center gap-3">
      <button
        disabled={!canStart}
        onClick={onStart}
        className={`px-5 py-2 rounded-xl text-white shadow transition-all ${
          canStart
            ? 'bg-emerald-600 hover:bg-emerald-700'
            : 'bg-emerald-300 cursor-not-allowed'
        }`}
      >
        모든 업로드 시작
      </button>
      <button
        disabled={!canClear}
        onClick={onClear}
        className={`px-5 py-2 rounded-xl text-white shadow transition-all ${
          canClear
            ? 'bg-slate-700 hover:bg-slate-800'
            : 'bg-slate-400 cursor-not-allowed'
        }`}
      >
        대기열 비우기
      </button>
    </div>
  );
}
