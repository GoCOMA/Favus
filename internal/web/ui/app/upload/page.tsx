// 웹 업로드 UI
'use client';

export default function UploadPage() {
  return (
    <main className="p-8">
      <h1 className="text-2xl font-semibold">파일 업로드</h1>
      <p className="text-gray-600 mt-2">
        파일을 드래그하거나 클릭하여 업로드하세요.
      </p>

      <div className="mt-6 border border-dashed border-gray-300 rounded-xl p-8 text-center text-gray-500">
        [UploadDropzone 컴포넌트 자리]
      </div>
    </main>
  );
}
