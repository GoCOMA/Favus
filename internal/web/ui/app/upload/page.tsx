'use client';

import { useCallback, useMemo, useRef, useState } from 'react';
import UploadDropzone from './components/UploadDropzone';
import UploadList from './components/UploadList';
import UploadToolbar from './components/UploadToolbar';

export type UploadStatus =
  | 'queued'
  | 'uploading'
  | 'completed'
  | 'failed'
  | 'canceled';

export type UploadItem = {
  id: string;
  file: File;
  name: string;
  size: number;
  status: UploadStatus;
  progress: number; // 0~100
  error?: string;
};

export default function UploadPage() {
  const [items, setItems] = useState<UploadItem[]>([]);
  const uploadingMap = useRef<Map<string, { stop: () => void }>>(new Map());

  const totalSize = useMemo(
    () => items.reduce((acc, i) => acc + i.size, 0),
    [items],
  );
  const overallProgress = useMemo(() => {
    if (!items.length) return 0;
    const weighted = items.reduce(
      (acc, i) => acc + (i.progress / 100) * i.size,
      0,
    );
    return Math.round((weighted / totalSize) * 100);
  }, [items, totalSize]);

  const addFiles = useCallback((files: File[]) => {
    setItems((prev) => {
      const next = [...prev];
      for (const f of files) {
        const id = `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
        next.push({
          id,
          file: f,
          name: f.name,
          size: f.size,
          status: 'queued',
          progress: 0,
        });
      }
      return next;
    });
  }, []);

  const clearAll = useCallback(() => {
    // stop all running uploads
    uploadingMap.current.forEach(({ stop }) => stop());
    uploadingMap.current.clear();
    setItems([]);
  }, []);

  const removeItem = useCallback((id: string) => {
    const runner = uploadingMap.current.get(id);
    if (runner) runner.stop();
    uploadingMap.current.delete(id);
    setItems((prev) => prev.filter((i) => i.id !== id));
  }, []);

  const cancelItem = useCallback((id: string) => {
    const runner = uploadingMap.current.get(id);
    if (runner) runner.stop();
    uploadingMap.current.delete(id);
    setItems((prev) =>
      prev.map((i) =>
        i.id === id ? { ...i, status: 'canceled' as const } : i,
      ),
    );
  }, []);

  const startOne = useCallback((id: string) => {
    setItems((prev) =>
      prev.map((i) =>
        i.id === id
          ? { ...i, status: 'uploading', progress: 0, error: undefined }
          : i,
      ),
    );

    // Simulated upload runner
    const tickMs = 120 + Math.floor(Math.random() * 180);
    const speed = 3 + Math.random() * 7; // % per tick
    let progress = 0;
    const interval = setInterval(() => {
      progress = Math.min(100, progress + speed);
      setItems((prev) =>
        prev.map((i) => (i.id === id ? { ...i, progress } : i)),
      );
      if (progress >= 100) {
        clearInterval(interval);
        uploadingMap.current.delete(id);
        setItems((prev) =>
          prev.map((i) => (i.id === id ? { ...i, status: 'completed' } : i)),
        );
      }
    }, tickMs);

    const stop = () => {
      clearInterval(interval);
      setItems((prev) =>
        prev.map((i) => (i.id === id ? { ...i, status: 'canceled' } : i)),
      );
    };

    uploadingMap.current.set(id, { stop });
  }, []);

  const startAll = useCallback(() => {
    const queued = items.filter((i) => i.status === 'queued').map((i) => i.id);
    // stagger starts a little to simulate real requests
    queued.forEach((id, idx) => {
      setTimeout(() => startOne(id), idx * 120);
    });
  }, [items, startOne]);

  const hasUploading = items.some((i) => i.status === 'uploading');
  const hasQueued = items.some((i) => i.status === 'queued');

  return (
    <main className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
      <div className="max-w-5xl mx-auto px-4 py-12 space-y-8">
        <header className="bg-white/70 backdrop-blur-sm rounded-2xl shadow-xl border border-white/20 p-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 bg-clip-text text-transparent">
                파일 업로드
              </h1>
              <p className="text-gray-600 mt-1">
                로컬 파일을 선택하거나 드래그 앤 드롭으로 업로드 대기열에
                추가하세요.
              </p>
            </div>
            <div className="text-sm text-gray-600">
              총 {items.length}개 • {formatBytes(totalSize)} • 진행률{' '}
              {overallProgress}%
            </div>
          </div>
          <div className="mt-4 w-full bg-gray-200 rounded-full h-3 overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-blue-500 to-indigo-600 rounded-full transition-all duration-500"
              style={{ width: `${overallProgress}%` }}
            />
          </div>
        </header>

        <UploadDropzone onFiles={addFiles} />

        <UploadToolbar
          canStart={hasQueued}
          canClear={items.length > 0 && !hasUploading}
          onStart={startAll}
          onClear={clearAll}
        />

        <UploadList
          items={items}
          onStart={startOne}
          onCancel={cancelItem}
          onRemove={removeItem}
        />
      </div>
    </main>
  );
}

function formatBytes(bytes: number) {
  if (!bytes) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`;
}
