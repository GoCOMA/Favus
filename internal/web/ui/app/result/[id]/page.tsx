// 업로드 결과 페이지

interface Props {
  params: { id: string };
}

export default function ResultPage({ params }: Props) {
  return (
    <main className="p-8">
      <h1 className="text-xl font-bold">업로드 결과</h1>
      <p className="mt-2 text-gray-600">ID: {params.id}</p>
      <div className="mt-6">[완료 메시지 / 다운로드 링크 / 메타정보 등]</div>
    </main>
  );
}
