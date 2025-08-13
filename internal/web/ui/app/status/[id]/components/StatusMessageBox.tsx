'use client';

interface Props {
  message?: string;
  retryCount?: number;
}

export function StatusMessageBox({ message, retryCount }: Props) {
  return (
    <>
      {message && (
        <div className="mb-6 p-4 bg-blue-50 rounded-lg">
          <p className="text-blue-800">{message}</p>
        </div>
      )}
      {retryCount !== undefined && retryCount > 0 && (
        <div className="mb-6 p-4 bg-yellow-50 rounded-lg">
          <p className="text-yellow-800">재시도 횟수: {retryCount}회</p>
        </div>
      )}
    </>
  );
}
