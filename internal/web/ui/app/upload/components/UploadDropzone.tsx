'use client';

import { useCallback, useRef, useState } from 'react';

export default function UploadDropzone({
  onFiles,
}: {
  onFiles: (files: File[]) => void;
}) {
  const [isOver, setIsOver] = useState(false);
  const inputRef = useRef<HTMLInputElement | null>(null);

  const onDrop = useCallback(
    (e: React.DragEvent<HTMLDivElement>) => {
      e.preventDefault();
      setIsOver(false);
      const files = Array.from(e.dataTransfer.files);
      if (files.length) onFiles(files);
    },
    [onFiles],
  );

  const onBrowse = () => inputRef.current?.click();

  return (
    <section
      onDragOver={(e) => {
        e.preventDefault();
        setIsOver(true);
      }}
      onDragLeave={() => setIsOver(false)}
      onDrop={onDrop}
      className={`relative rounded-2xl border-2 border-dashed p-10 transition-colors ${
        isOver ? 'border-blue-500 bg-blue-50/60' : 'border-gray-300 bg-white/70'
      }`}
    >
      <div className="text-center space-y-3">
        <div className="text-5xl">📤</div>
        <h2 className="text-xl font-semibold text-gray-800">
          파일을 여기에 드래그 앤 드롭
        </h2>
        <p className="text-gray-600 text-sm">또는</p>
        <button
          onClick={onBrowse}
          className="px-5 py-2 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-xl hover:from-blue-700 hover:to-indigo-700 transition-all shadow"
        >
          파일 선택
        </button>
        <input
          ref={inputRef}
          type="file"
          multiple
          className="hidden"
          onChange={(e) => {
            const files = e.target.files ? Array.from(e.target.files) : [];
            if (files.length) onFiles(files);
            e.currentTarget.value = '';
          }}
        />
      </div>
    </section>
  );
}
