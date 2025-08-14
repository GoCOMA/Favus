// internal/wsagent/agent.go
package wsagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ===================== Public types / helpers =====================

type AgentConfig struct {
	// 로컬 HTTP 바인딩 주소 (예: "127.0.0.1:7777")
	Addr string

	// 업스트림 WebSocket 엔드포인트 (예: "ws://127.0.0.1:8765/ws")
	WSEndpoint string

	// 필요 시 API 키를 헤더로 보냄 (X-API-Key)
	APIKey string
}

// 이벤트 공용 포맷(권장). 자유 필드가 필요하면 Payload를 쓰면 됨.
type Event struct {
	Type      string          `json:"type"`    // e.g. "session_start", "part_done", ...
	RunID     string          `json:"runId"`   // CLI가 생성한 업로드 세션 ID
	Timestamp time.Time       `json:"ts"`      // 자동으로 안 넣는 경우, CLI에서 채워도 OK
	Payload   json.RawMessage `json:"payload"` // 자유 JSON
}

// CLI가 에이전트에 이벤트를 넘길 때 사용할 헬퍼.
// addr: 보통 DefaultAddr() 또는 사용자가 지정한 --addr
// ev:   Event 또는 임의의 구조체(map[string]any 등) — JSON으로 직렬화됨
func SendEvent(ctx context.Context, addr string, ev any) error {
	b, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("wsagent: marshal event: %w", err)
	}
	url := "http://" + addr + "/event"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("wsagent: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("wsagent: post /event: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("wsagent: /event status %d: %s", res.StatusCode, string(body))
	}
	return nil
}

func DefaultAddr() string { return "127.0.0.1:7777" }

// 에이전트가 떠있는지 간단히 확인(healthz) — 주소 지정 버전
func IsRunningAt(addr string) bool {
	// TCP 레벨로 먼저 열려있는지 확인
	d := &net.Dialer{Timeout: 250 * time.Millisecond}
	conn, err := d.Dial("tcp", addr)
	if err != nil {
		return false
	}
	_ = conn.Close()

	// /healthz 확인
	client := &http.Client{Timeout: 300 * time.Millisecond}
	resp, err := client.Get("http://" + addr + "/healthz")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// 기본 주소용 (하위호환)
func IsRunning() bool { return IsRunningAt(DefaultAddr()) }

// ===================== Agent =====================

type Agent struct {
	cfg AgentConfig

	mu     sync.Mutex // wsConn writes 보호
	wsConn *websocket.Conn

	httpSrv  *http.Server
	started  chan struct{}
	stopping chan struct{}
}

func Start(cfg AgentConfig) (*Agent, error) {
	if cfg.Addr == "" {
		cfg.Addr = DefaultAddr()
	}
	if cfg.WSEndpoint == "" {
		return nil, errors.New("wsagent: WSEndpoint is empty")
	}

	// 1) 업스트림 WS 연결
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	wsHeader := http.Header{}
	if cfg.APIKey != "" {
		wsHeader.Set("X-API-Key", cfg.APIKey)
	}

	var conn *websocket.Conn
	var err error
	// 간단한 재시도(최대 5회)
	for i := 0; i < 5; i++ {
		conn, _, err = dialer.Dial(cfg.WSEndpoint, wsHeader)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		return nil, fmt.Errorf("wsagent: connect WS failed: %w", err)
	}

	ag := &Agent{
		cfg:      cfg,
		wsConn:   conn,
		started:  make(chan struct{}),
		stopping: make(chan struct{}),
	}

	// 2) 로컬 HTTP 서버
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", ag.handleHealth)
	mux.HandleFunc("/event", ag.handleEvent)
	mux.HandleFunc("/shutdown", ag.handleStop)

	ag.httpSrv = &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// 3) PID 파일
	if err := writePID(pidFilePath()); err != nil {
		// PID 파일 실패는致命적이진 않으니 경고만
		fmt.Fprintf(os.Stderr, "wsagent: write pid warn: %v\n", err)
	}

	// 4) WS 수명관리(ping/pong + read-loop)
	ag.wsConn.SetPongHandler(func(string) error {
		return ag.wsConn.SetReadDeadline(time.Now().Add(30 * time.Second))
	})
	go ag.readLoop()
	go ag.pingLoop()

	// 5) HTTP 서버 기동
	go func() {
		close(ag.started)
		if err := ag.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "wsagent: http server error: %v\n", err)
		}
	}()

	return ag, nil
}

func (a *Agent) Stop(ctx context.Context) error {
	select {
	case <-a.stopping:
		return nil // 이미 중지 중/완료
	default:
		close(a.stopping)
	}

	// 1) HTTP 서버 graceful shutdown
	if a.httpSrv != nil {
		_ = a.httpSrv.Shutdown(ctx)
	}

	// 2) WS 종료
	a.mu.Lock()
	if a.wsConn != nil {
		_ = a.wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
		_ = a.wsConn.Close()
		a.wsConn = nil
	}
	a.mu.Unlock()

	// 3) PID 파일 제거
	_ = os.Remove(pidFilePath())

	return nil
}

// ===================== HTTP handlers =====================

func (a *Agent) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (a *Agent) handleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20)) // 4MB safety
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	// 업스트림 WS로 그대로 전달(서버가 Event 스키마를 검증한다고 가정)
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wsConn == nil {
		http.Error(w, "ws not connected", http.StatusServiceUnavailable)
		return
	}
	if err := a.wsConn.WriteMessage(websocket.TextMessage, body); err != nil {
		http.Error(w, "ws write failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *Agent) handleStop(w http.ResponseWriter, r *http.Request) {
	go func() {
		// 약간의 딜레이 후 종료(응답을 먼저 보낸 뒤)
		time.Sleep(100 * time.Millisecond)
		_ = a.Stop(context.Background())
	}()
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("stopping"))
}

// ===================== WS loops =====================

func (a *Agent) readLoop() {
	defer func() {
		// 연결이 끊기면 Stop을 호출해 정리
		_ = a.Stop(context.Background())
	}()

	a.wsConn.SetReadLimit(8 << 20)
	_ = a.wsConn.SetReadDeadline(time.Now().Add(30 * time.Second))

	for {
		_, _, err := a.wsConn.ReadMessage()
		if err != nil {
			return
		}
		// 서버에서 오는 메시지를 별도로 처리할 게 없으면 discard
	}
}

func (a *Agent) pingLoop() {
	t := time.NewTicker(15 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-a.stopping:
			return
		case <-t.C:
			a.mu.Lock()
			if a.wsConn != nil {
				_ = a.wsConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				_ = a.wsConn.WriteMessage(websocket.PingMessage, []byte("ping"))
			}
			a.mu.Unlock()
		}
	}
}

// ===================== PID helpers =====================

func pidFilePath() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "."
	}
	dir := filepath.Join(home, ".favus")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "agent.pid")
}

func writePID(path string) error {
	pid := os.Getpid()
	return os.WriteFile(path, []byte(fmt.Sprint(pid)), 0o644)
}
