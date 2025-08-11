'use client';

import type { UploadItem } from '../page';

interface UploadListProps {
  items: UploadItem[];
  onStart: (id: string) => void;
  onCancel: (id: string) => void;
  onRemove: (id: string) => void;
}

export default function UploadList({
  items,
  onStart,
  onCancel,
  onRemove,
}: UploadListProps) {
  if (!items.length) {
    return (
      <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-8 text-center text-gray-500">
        업로드 대기열이 비어 있습니다.
      </div>
    );
  }

  return (
    <div className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-4">
      <ul className="divide-y divide-gray-200">
        {items.map((i) => (
          <li
            key={i.id}
            className="p-4 flex flex-col gap-3 md:flex-row md:items-center md:gap-6"
          >
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <div className="truncate font-medium text-gray-900">
                  {i.name}
                </div>
                <div
                  className={`text-xs px-2 py-1 rounded border ${badgeColor(i.status)}`}
                >
                  {label(i.status)}
                </div>
              </div>
              <div className="mt-2 w-full bg-gray-200 rounded-full h-2 overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all duration-300 ${barColor(i.status)}`}
                  style={{ width: `${i.progress}%` }}
                />
              </div>
              <div className="mt-1 text-xs text-gray-500">{i.progress}%</div>
              {i.error && (
                <div className="mt-1 text-xs text-rose-600">{i.error}</div>
              )}
            </div>

            <div className="shrink-0 flex items-center gap-2">
              {i.status === 'queued' && (
                <button
                  onClick={() => onStart(i.id)}
                  className="px-3 py-1.5 rounded-lg text-white bg-blue-600 hover:bg-blue-700 text-sm"
                >
                  업로드 시작
                </button>
              )}
              {i.status === 'uploading' && (
                <button
                  onClick={() => onCancel(i.id)}
                  className="px-3 py-1.5 rounded-lg text-white bg-amber-600 hover:bg-amber-700 text-sm"
                >
                  취소
                </button>
              )}
              {(i.status === 'completed' ||
                i.status === 'failed' ||
                i.status === 'canceled') && (
                <button
                  onClick={() => onRemove(i.id)}
                  className="px-3 py-1.5 rounded-lg text-white bg-slate-700 hover:bg-slate-800 text-sm"
                >
                  제거
                </button>
              )}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}

function label(s: UploadItem['status']) {
  switch (s) {
    case 'queued':
      return '대기';
    case 'uploading':
      return '업로드 중';
    case 'completed':
      return '완료';
    case 'failed':
      return '실패';
    case 'canceled':
      return '취소됨';
  }
}

function badgeColor(s: UploadItem['status']) {
  switch (s) {
    case 'queued':
      return 'text-amber-700 bg-amber-50 border-amber-200';
    case 'uploading':
      return 'text-blue-700 bg-blue-50 border-blue-200';
    case 'completed':
      return 'text-emerald-700 bg-emerald-50 border-emerald-200';
    case 'failed':
      return 'text-rose-700 bg-rose-50 border-rose-200';
    case 'canceled':
      return 'text-gray-700 bg-gray-50 border-gray-200';
  }
}

function barColor(s: UploadItem['status']) {
  switch (s) {
    case 'queued':
      return 'bg-gradient-to-r from-amber-400 to-orange-500';
    case 'uploading':
      return 'bg-gradient-to-r from-blue-500 to-indigo-600';
    case 'completed':
      return 'bg-gradient-to-r from-emerald-500 to-green-600';
    case 'failed':
      return 'bg-gradient-to-r from-rose-500 to-red-600';
    case 'canceled':
      return 'bg-gradient-to-r from-gray-400 to-slate-500';
  }
}
