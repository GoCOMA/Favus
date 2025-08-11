'use client';

export default function CliCommandBlock({ command }: { command: string }) {
  return (
    <pre className="bg-gray-900 text-green-300 p-4 rounded-lg text-sm overflow-x-auto shadow-inner">
      <code>{command}</code>
    </pre>
  );
}
