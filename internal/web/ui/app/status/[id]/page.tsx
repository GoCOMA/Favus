// 특정 업로드 상태 확인

interface Props {
  params: { id: string };
}

export default function StatusPage({ params }: Props) {
  return (
    <main className="p-8">
      <h1 className="text-xl font-bold">업로드 상태</h1>
      <p className="mt-2 text-gray-600">ID: {params.id}</p>
      <div className="mt-6">[업로드 진행 상황 표시 컴포넌트 자리]</div>
    </main>
  );
}
