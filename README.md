# Favus — S3 Multipart Upload CLI & Web UI

## Table of Contents

- [Overview](#overview)
- [Why Favus](#why-favus)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [CLI Usage (Quick Peek)](#cli-usage-quick-peek)
- [Web UI & Realtime Monitoring](#web-ui--realtime-monitoring)
- [Message Schema (WebSocket)](#message-schema-websocket)
- [License](#license)

---

## Overview

**Favus** is a Go/React–based tool that makes **large-file uploads to Amazon S3** fast, reliable, and observable.

- **CLI** performs intelligent **multipart uploads**, automatic **resume**, and **orphan part** cleanup.
- **Web UI** (Next.js/React) shows **per-part progress in real time** via WebSocket—complete with part-level status (done/failed/pending).

> Problem we solve: **“ghost/orphan parts”** left behind by interrupted uploads silently accrue storage costs and clutter buckets. Favus detects, visualizes, and cleans them—reducing manual ops toil and optimizing cloud spend.

---

## Why Favus

Multipart upload to S3 is efficient, but partial failures often leave **invisible, billable artifacts** in your buckets. Over time they hurt both **budget** and **operational hygiene**. Favus:

- **Prevents waste** by discovering and removing orphan parts.
- **Reduces risk** with robust resume & retries.
- **Increases transparency** through a clean, real-time UI.

---

## Key Features

### High-volume, resilient transfer

- **Smart chunking:** Splits extremely large files into parts and uploads them concurrently for speed and throughput.
- **Dual progress bars:** Terminal shows overall & per-part progress; Web UI shows **vertical, per-part bars** (blue=done, red=failed, gray=pending).
- **Auto-resume & recovery:** Uses a JSON state file to pick up exactly where it left off after interruptions. **Exponential backoff** on transient errors.

### Realtime monitoring

- **WebSocket streaming:** CLI reports events (start, part_done, error, done) to the UI in real time.
- **Stable 3-tier reporting:** **CLI Reporter → Local Agent → Python WS server**; enables smooth monitoring for **many concurrent sessions**.

### S3 management tooling

- **Orphan discovery & cleanup:** Find incompletely uploaded parts (e.g., `ls-orphans`) and remove them safely.
- **List/inspect active uploads:** Keep track of ongoing sessions.

### Enterprise-minded design

- **Layered configuration:** YAML, env vars, and flags for flexible deployment.
- **Data integrity:** ETag validation; **atomic** file update semantics where relevant.

---

## Architecture

```
flowchart LR
    U[User (CLI/Web)] --> C[Go CLI]
    C -- Events --> L[Local Agent]
    L -- WS --> P[Python WebSocket Server]
    P -- push --> W[Web UI (Next.js/React)]
    C -- Multipart PUT --> S[(Amazon S3)]
```

**Components**

- **User:** interacts via CLI or Web UI
- **Go CLI:** core multipart logic & reporting
- **Python WS server:** relays realtime events
- **Web UI:** visualizes runs and part-level progress
- **S3:** durable storage backend

---

## Tech Stack

- **Go** `1.24.1` (CLI / core)
- **Python** `3.10.12` (WebSocket relay)
- **TypeScript** `5.x`, **React** `19`, **Next.js** `15`, **Tailwind CSS** `4`
- Build & tooling: VS Code, Git/GitHub

---

## Getting Started

### Prerequisites

- **npm ≥ 10**, **Node.js ≥ 20** (prefer 22)
- AWS credentials with permission to perform multipart uploads

### Install dependencies

```bash
# from repo root (or internal/web/ui/)
npm install
# or
pnpm install
```

### Configure AWS

Set environment variables or use your `~/.aws` profile:

```bash
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=ap-northeast-2
# optional:
export AWS_PROFILE=default
```

### Build & run (Web UI)

```bash
npm run dev        # http://localhost:3000
# production
npm run build && npm run start
```

---

## CLI Usage (Quick Peek)

> Exact command names/flags may evolve. See `favus --help` in your build for the latest.

```bash
# Upload a file
favus upload ./bigfile.mov s3://your-bucket/path/bigfile.mov

# Resume a stopped upload (state file is created automatically)
favus resume ./bigfile.mov.state.json

# List orphan parts
favus ls-orphans s3://your-bucket/path/

# Remove orphan parts
favus kill-orphans s3://your-bucket/path/
```

---

## Web UI & Realtime Monitoring

The Web UI subscribes to a WebSocket for events emitted by the CLI/agent chain.

### WebSocket provider

The default provider uses `ws://127.0.0.1:8765/ws` and **normalizes** messages to:

```ts
type WebSocketMessage = {
  Type: string; // from rawMessage.type
  RunID: string; // from rawMessage.runId
  Payload: any; // from rawMessage.payload
};
```

It supports:

- `subscribe('*', cb)` for global messages
- `subscribe(RunID, cb)` for run-scoped messages
- `unsubscribe(id)`
- `isConnected`
- Auto-reconnect (5 tries, 3s delay; configurable in code)

> Prefer an env-based URL? Replace the hardcoded string with:
> `const wsUrl = process.env.NEXT_PUBLIC_WS_URL ?? "ws://127.0.0.1:8765/ws";`
>
> `.env.local`:
>
> ```bash
> NEXT_PUBLIC_WS_URL=ws://localhost:8765/ws
> ```

### UI component

- `UploadStatusList` renders one card per run (filename, status, overall %, **vertical part list**).
- Colors: **Blue** (completed), **Red** (failed), **Gray** (pending).

---

## Message Schema (WebSocket)

The UI expects `{ Type, RunID, Payload }`:

- **`session_start`**

  ```ts
  interface StartPayload {
    bucket: string;
    key: string; // display name / S3 key
    uploadId: string; // file identifier
    total: number; // total bytes
    partMB: number; // part size in MB
  }
  ```

  `totalParts = ceil(total / (partMB * 1024 * 1024))`

- **`part_done`**

  ```ts
  interface PartDonePayload {
    part: number; // 1-based
    size: number;
    etag: string;
  }
  ```

- **`session_done`**

  ```ts
  interface DonePayload {
    success: boolean;
    uploadId: string;
  }
  ```

- **`error`**

  ```ts
  interface ErrorPayload {
    message: string;
    partNumber?: number;
  }
  ```

---

## License

MIT, see [LICENSE](https://github.com/GoCOMA/Favus/blob/main/LICENSE).
