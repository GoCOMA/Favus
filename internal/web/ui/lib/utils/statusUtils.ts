export const getStatusColor = (status: string) => {
  switch (status) {
    case 'pending':
      return 'text-amber-600 bg-gradient-to-r from-amber-50 to-orange-50 border-amber-200';
    case 'processing':
      return 'text-blue-600 bg-gradient-to-r from-blue-50 to-indigo-50 border-blue-200';
    case 'completed':
      return 'text-emerald-600 bg-gradient-to-r from-emerald-50 to-green-50 border-emerald-200';
    case 'failed':
      return 'text-rose-600 bg-gradient-to-r from-rose-50 to-red-50 border-rose-200';
    default:
      return 'text-gray-600 bg-gradient-to-r from-gray-50 to-slate-50 border-gray-200';
  }
};
export const getStatusText = (status: string) => {
  switch (status) {
    case 'pending':
      return '대기 중';
    case 'processing':
      return '처리 중';
    case 'completed':
      return '완료';
    case 'failed':
      return '실패';
    default:
      return '알 수 없음';
  }
};
