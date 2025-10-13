# Favus Web UI – Developer Guide

## Overview

The `ui` folder contains the Favus web UI and supporting code.

- `lib/contexts/`
  - `WebSocketContext`: real backend WebSocket context

- `app/(home)/components/`
  - `UploadStatusList`: renders **vertical per-part progress bars** for uploads

- `app/(home)/page.tsx`
  - `HomePage`: landing page that includes `UploadStatusList`

---

## Pre-requisites

- **npm ≥ v10**
- **Node.js ≥ v20** (prefer v22)
- Next.js 15+ / React 19+ / TailwindCSS

---

## Installing npm dependencies

From the repo root (or `ui`):

```bash
npm install
# or
pnpm install
```

---

## Running a local development server

> Tip: `./favus ui --endpoint ws://127.0.0.1:8765/ws --foreground` now launches this dev server for you (make sure `npm install` has been run at least once).  
> WSL/리눅스 환경에서는 `http://<호스트 IP>:3000`으로 접근할 수 있도록 `0.0.0.0`에 바인딩됩니다.

```bash
npm run dev
# http://localhost:3000
```

---

## WebSocket configuration

Your `WebSocketProvider` in `lib/contexts/WebSocketContext.tsx`:

- **Hardcodes** the URL to `ws://127.0.0.1:8765/ws`.
- **Normalizes** incoming messages to `{ Type, RunID, Payload }` from `{ type, runId, payload }`.
- Supports two subscription keys:
  - `'*'` → global subscriber
  - `RunID` → per-run subscriber

- Maintains **one subscriber per key** (Map of `id -> callback`, so the latest call to `subscribe(id, ...)` wins).
- Provides **auto-reconnect** up to 5 times with a 3s delay.

### Provider API

```ts
type WebSocketMessage = {
  Type: string; // normalized from rawMessage.type
  RunID: string; // normalized from rawMessage.runId
  Payload: any; // normalized from rawMessage.payload
};

type WebSocketContextType = {
  subscribe: (id: string, cb: (msg: WebSocketMessage) => void) => void; // id = '*' or RunID
  unsubscribe: (id: string) => void; // removes the single callback for that id
  isConnected: boolean; // WebSocket ready state
};
```

### Usage

```tsx
import { WebSocketProvider } from '@/lib/contexts/WebSocketContext';

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html>
      <body>
        <WebSocketProvider>{children}</WebSocketProvider>
      </body>
    </html>
  );
}
```

---

## Message schema

The UI expects **normalized** messages `{ Type, RunID, Payload }`:

- `session_start`

  ```ts
  interface StartPayload {
    bucket: string;
    key: string; // display name
    uploadId: string; // file id
    total: number; // total bytes
    partMB: number; // part size in MB
  }
  ```

  → `totalParts = ceil(total / (partMB * 1024 * 1024))`

- `part_done`

  ```ts
  interface PartDonePayload {
    part: number; // 1-based
    size: number;
    etag: string;
  }
  ```

  → mark part completed (blue), update progress

- `session_done`

  ```ts
  interface DonePayload {
    success: boolean;
    uploadId: string;
  }
  ```

  → `completed` (100%) or `failed` (keep current %)

- `error`

  ```ts
  interface ErrorPayload {
    message: string;
    partNumber?: number;
  }
  ```

  → `failed`; if `partNumber` is provided, mark that part red

**Color legend:** Blue = completed, Red = failed, Gray = pending.

---

## Home integration

```tsx
'use client';

import UploadStatusList from '@/app/(home)/components/UploadStatusList';

export default function HomePage() {
  return (
    <main className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-4xl mx-auto px-4">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold">Welcome to Favus</h1>
          <p className="text-xl text-gray-600">
            Upload large files via CLI and monitor progress in real time.
          </p>
        </div>
        <UploadStatusList />
      </div>
    </main>
  );
}
```

---

## Upgrading dependencies

Update `package.json` (and workspace packages if applicable), then run `npm install` at the root.
