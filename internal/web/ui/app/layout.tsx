// 공통 레이아웃
import './globals.css';
import Link from 'next/link';
import { ReactNode } from 'react';
import { WebSocketProvider } from '@/lib/context/WebSocketContext';

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="ko">
      <body className="min-h-screen bg-white text-gray-900">
        <WebSocketProvider>
          <header className="p-4 border-b">
            <nav className="flex gap-4">
              <Link href="/">홈</Link>
              <Link href="/upload">웹 업로드</Link>
              <Link href="/upload/cli">CLI 안내</Link>
            </nav>
          </header>
          <main>{children}</main>
        </WebSocketProvider>
      </body>
    </html>
  );
}
