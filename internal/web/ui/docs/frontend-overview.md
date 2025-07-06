# Favus – Frontend Overview

## 프로젝트명

**Favus**

---

## 서비스 개요 (프론트엔드 관점)

**Favus**는 대용량 파일을 Amazon S3에 안정적으로 업로드할 수 있도록 지원하는 웹/CLI 기반 플랫폼입니다.

프론트엔드는 다음과 같은 역할을 담당합니다:

- 사용자에게 **간편한 대용량 업로드 UI 제공**
- **멀티파트 업로드**를 위한 presigned URL 요청 및 분할 업로드 실행
- **업로드 상태 시각화**, 재시도 및 Resume 기능 지원
- CLI 업로드 가이드 제공

---

## 기술 스택

- **Next.js (App Router)**
- **TypeScript**
- **Tailwind CSS**
- **Zustand**
- **React Query**
- **ESLint + Prettier**

---

## 디렉토리 구조

```
/internal/web/ui
├── app/
│   ├── layout.tsx
│   ├── page.tsx
│   ├── upload/page.tsx
│   ├── upload/cli/page.tsx
│   ├── status/[id]/page.tsx
│   └── result/[id]/page.tsx
├── components/
├── hooks/
├── lib/
├── styles/
├── public/
```

---

## 구현해야 할 주요 기능

### 웹 업로드 (`/upload`)

- Drag & Drop 업로드 (`UploadDropzone`)
- presigned URL 요청 및 청크별 S3 업로드
- 업로드 상태 시각화 (`UploadProgressBar`)
- 완료 후 `/status/[id]` 또는 `/result/[id]`로 이동

### CLI 업로드 안내 (`/upload/cli`)

- 설치 및 실행 가이드
- 명령어 복사 버튼 제공

### 업로드 상태 확인 (`/status/[id]`)

- 주기적 polling으로 업로드 상태 표시
- 오류/재시도 상태 안내

### 업로드 결과 확인 (`/result/[id]`)

- 업로드 완료 후 S3 링크 및 메타 정보 표시

---

## 상태 관리 전략

| 상태 종류   | 도구        | 설명                              |
| ----------- | ----------- | --------------------------------- |
| 업로드 진행 | Zustand     | 클라이언트 상태 (파일, 진행률 등) |
| 상태 조회   | React Query | 서버 상태 fetch 및 polling        |

---

## API 연동 지점

| 목적               | 경로 예시                          |
| ------------------ | ---------------------------------- |
| presigned URL 요청 | `POST /api/generate-presigned-url` |
| 업로드 상태 확인   | `GET /api/status/[id]`             |
| 업로드 결과 조회   | `GET /api/result/[id]`             |

---

## 초기 개발 체크리스트

- [x] `layout.tsx` 및 페이지 라우팅 구성
- [x] `/upload/page.tsx` 기본 UI 생성
- [ ] `UploadDropzone` 컴포넌트 구현
- [ ] `lib/s3Client.ts` presigned URL 요청 함수
- [ ] Zustand로 업로드 상태 관리
- [ ] `/status/[id]` – polling 처리
- [ ] `/result/[id]` – 결과 정보 표시

---

## 커밋 & 브랜치 컨벤션

### 브랜치 규칙

```
<type>/<short-description>
예: feat/upload-ui, chore/setup-eslint
```

### 커밋 메시지 예시

```
feat: implement UploadDropzone component
chore: setup Zustand store for upload state
fix: handle upload resume after failure
```

---

## 기대 효과 (프론트엔드 기준)

- 멀티파트 업로드를 **단순 UI로 추상화**
- 네트워크 오류에도 **Resume 가능**
- 사용자는 **진행률/결과를 시각적으로 인지** 가능
- 개발자는 별도 로직 없이 S3에 **안정적인 대용량 업로드 제공 가능**

---

## 공유 및 참고

- 이 문서는 `/docs/frontend-overview.md`로 저장되어야 함
- 다른 팀원이나 계정에서 프로젝트 이해용으로 활용 가능
- 지속해서 업데이트 필요
