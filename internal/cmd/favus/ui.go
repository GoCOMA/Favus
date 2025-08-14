// cmd/ui.go
package favus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/GoCOMA/Favus/internal/wsagent"
	"github.com/spf13/cobra"

	// 브라우저 자동 열기 (선택)
	"github.com/pkg/browser"
)

var (
	// favus ui 플래그
	uiAddrFlag    string // 로컬 에이전트 바인드 주소 (기본: 127.0.0.1:7777)
	uiWSEndpoint  string // 업스트림 WebSocket 서버 (예: ws://127.0.0.1:8765/ws)
	uiAPIKey      string // 선택: 서버가 요구하면 사용
	uiOpenBrowser bool   // 시작 시 브라우저 자동 오픈
	uiOpenURL     string // 강제로 열 URL 지정 (미지정 시 endpoint로부터 추정)

	uiForeground bool

	// favus stop-ui 플래그
	stopAddrFlag string
)

func init() {
	// ---- favus ui ----
	uiCmd := &cobra.Command{
		Use:   "ui",
		Short: "Start local UI bridge (WebSocket agent) to stream CLI events to the Web UI",
		Long: `Start a small local agent that:
- accepts HTTP events at /event from Favus commands
- forwards them to the upstream WebSocket server (WSS/WS)
- exposes /healthz and /shutdown for control

Run this once (foreground or background), then use other favus commands (upload/resume) in the same terminal.`,
		RunE: runUI,
	}

	uiCmd.Flags().StringVar(&uiAddrFlag, "addr", wsagent.DefaultAddr(), "Local agent bind address (host:port)")
	uiCmd.Flags().StringVar(&uiWSEndpoint, "endpoint", "", "Upstream WebSocket endpoint (e.g. ws://127.0.0.1:8765/ws)")
	uiCmd.Flags().StringVar(&uiAPIKey, "api-key", "", "Optional API key for upstream (sent as X-API-Key)")
	uiCmd.Flags().BoolVar(&uiOpenBrowser, "open", false, "Open the Web UI in your browser after connecting")
	uiCmd.Flags().StringVar(&uiOpenURL, "open-url", "", "Explicit URL to open (overrides auto-derived one)")
	uiCmd.Flags().BoolVar(&uiForeground, "foreground", false, "Run in foreground (block until Ctrl+C)")

	rootCmd.AddCommand(uiCmd)

	// ---- favus stop-ui ----
	stopCmd := &cobra.Command{
		Use:   "stop-ui",
		Short: "Stop the local UI bridge (WebSocket agent)",
		RunE:  stopUI,
	}
	stopCmd.Flags().StringVar(&stopAddrFlag, "addr", wsagent.DefaultAddr(), "Local agent bind address (host:port)")
	rootCmd.AddCommand(stopCmd)
}

func runUI(cmd *cobra.Command, args []string) error {
	// 0) 기본값/ENV 보정
	if uiWSEndpoint == "" {
		if v := os.Getenv("FAVUS_WS_ENDPOINT"); v != "" {
			uiWSEndpoint = v
		} else {
			// 합리적 기본값
			uiWSEndpoint = "ws://127.0.0.1:8765/ws"
		}
	}
	if uiAPIKey == "" {
		uiAPIKey = os.Getenv("FAVUS_WS_API_KEY")
	}

	// 1) 이미 떠있으면 스킵 (원하면 --open 처리만 수행)
	if wsagent.IsRunningAt(uiAddrFlag) {
		fmt.Printf("🔁 UI agent already running at http://%s\n", uiAddrFlag)
		if uiOpenBrowser {
			toOpen := uiOpenURL
			if toOpen == "" {
				toOpen = deriveUIURL(uiWSEndpoint)
			}
			_ = openBrowserSafe(toOpen)
		}
		return nil
	}

	// 2) 백그라운드로 띄우는 옵션
	if !uiForeground {
		args := []string{
			"ui",
			"--addr", uiAddrFlag,
			"--endpoint", uiWSEndpoint,
			"--foreground", // 포그라운드 모드로 실제 에이전트 실행
		}
		if uiAPIKey != "" {
			args = append(args, "--api-key", uiAPIKey)
		}
		if uiOpenBrowser {
			args = append(args, "--open")
		}
		if uiOpenURL != "" {
			args = append(args, "--open-url", uiOpenURL)
		}

		// 로그 파일로 리디렉션
		logDir := filepath.Join(os.Getenv("HOME"), ".favus")
		_ = os.MkdirAll(logDir, 0o755)
		logPath := filepath.Join(logDir, "agent.log")

		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolve executable: %w", err)
		}

		c := exec.Command(exe, args...)
		lf, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		c.Stdout = lf
		c.Stderr = lf

		if err := c.Start(); err != nil {
			_ = lf.Close()
			return fmt.Errorf("failed to start UI agent in background: %w", err)
		}
		// 부모에서 핸들 닫기
		_ = lf.Close()

		// 헬스체크 폴링(최대 2초)
		deadline := time.Now().Add(2 * time.Second)
		started := false
		for time.Now().Before(deadline) {
			if wsagent.IsRunningAt(uiAddrFlag) {
				started = true
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if started {
			fmt.Printf("✅ UI agent started in background (pid %d), logs: %s\n", c.Process.Pid, logPath)
		} else {
			fmt.Printf("⚠️  UI agent start not confirmed yet. Logs: %s (pid %d)\n", logPath, c.Process.Pid)
		}
		return nil
	}

	// 3) 포그라운드 실행
	cfg := wsagent.AgentConfig{
		Addr:       uiAddrFlag,
		WSEndpoint: uiWSEndpoint,
		APIKey:     uiAPIKey,
	}

	ag, err := wsagent.Start(cfg)
	if err != nil {
		return fmt.Errorf("failed to start UI agent: %w", err)
	}
	fmt.Printf("✅ UI agent started: http://%s  → %s\n", cfg.Addr, cfg.WSEndpoint)

	// 4) 필요 시 브라우저 오픈
	if uiOpenBrowser {
		toOpen := uiOpenURL
		if toOpen == "" {
			toOpen = deriveUIURL(uiWSEndpoint)
		}
		_ = openBrowserSafe(toOpen)
	}

	// 5) Ctrl+C 또는 SIGTERM 대기 후 정리
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	fmt.Println("🔌 Press Ctrl+C to stop (or run `favus stop-ui`).")
	<-sigc

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = ag.Stop(ctx)
	fmt.Println("👋 UI agent stopped.")
	return nil
}

func stopUI(cmd *cobra.Command, args []string) error {
	// /shutdown 호출
	url := "http://" + stopAddrFlag + "/shutdown"
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	client := &http.Client{Timeout: 800 * time.Millisecond}
	res, err := client.Do(req)
	if err != nil {
		// 연결 자체가 안 되면 이미 안 떠있는 것일 수 있음
		fmt.Println("ℹ️  UI agent is not running (or unreachable).")
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode/100 == 2 || res.StatusCode == http.StatusAccepted {
		fmt.Println("✅ Stop signal sent to UI agent.")
		return nil
	}
	return fmt.Errorf("stop-ui failed: HTTP %d", res.StatusCode)
}

// ----- helpers -----

func deriveUIURL(ws string) string {
	// ws(s)://host[:port]/something → http(s)://host[:port]/
	u, err := url.Parse(ws)
	if err != nil {
		// 못 파싱하면 그냥 원문 리턴(브라우저가 못 열겠지만…)
		return ws
	}
	switch strings.ToLower(u.Scheme) {
	case "ws":
		u.Scheme = "http"
	case "wss":
		u.Scheme = "https"
	default:
		// 이미 http/https 같은 경우 그냥 둔다
	}
	u.Path = "/" // UI 루트로 유도; 필요 시 --open-url 로 정확히 지정
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func openBrowserSafe(u string) error {
	// 브라우저 열기 실패해도 치명적이지 않으니 에러는 출력만
	if u == "" {
		return errors.New("no URL to open")
	}
	if err := browser.OpenURL(u); err != nil {
		fmt.Fprintf(os.Stderr, "warn: failed to open browser: %v\n", err)
		return err
	}
	fmt.Printf("🌐 Opening UI: %s\n", u)
	return nil
}
