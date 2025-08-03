export const getStatusColor = (status: string) => {
  const map: Record<string, string> = {
    pending:
      'text-amber-600 bg-gradient-to-r from-amber-50 to-orange-50 border-amber-200',
    processing:
      'text-blue-600 bg-gradient-to-r from-blue-50 to-indigo-50 border-blue-200',
    completed:
      'text-emerald-600 bg-gradient-to-r from-emerald-50 to-green-50 border-emerald-200',
    failed:
      'text-rose-600 bg-gradient-to-r from-rose-50 to-red-50 border-rose-200',
  };
  return (
    map[status] ||
    'text-gray-600 bg-gradient-to-r from-gray-50 to-slate-50 border-gray-200'
  );
};

export const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '대기 중',
    processing: '처리 중',
    uploading: '업로드 중',
    completed: '완료',
    failed: '실패',
  };
  return map[status] || '알 수 없음';
};
