package uploader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/GoCOMA/Favus/internal/wsagent"
	"github.com/google/uuid"
)

type partTracker struct {
	size        int64
	sent        int64
	started     time.Time
	lastFlushAt time.Time
}

type wsReporter struct {
	mu sync.Mutex // Protects all fields below

	enabled           bool
	addr              string
	runID             string
	started           time.Time
	totalBytes        int64
	uploadedBytes     int64
	lastProgressFlush time.Time

	lastCheck     time.Time
	checkInterval time.Duration
	lastErrorLog  time.Time

	startPayload map[string]any
	startSent    bool

	parts map[int]*partTracker
}

func agentAddr() string {
	if v := os.Getenv("FAVUS_AGENT_ADDR"); v != "" {
		return v // 사용자가 favus ui --addr 바꾸면 ENV로 맞출 수 있음
	}
	return wsagent.DefaultAddr()
}

func newWSReporter(total int64) *wsReporter {
	addr := agentAddr()
	ok := wsagent.IsRunningAt(addr)
	return &wsReporter{
		enabled:           ok,
		addr:              addr,
		runID:             uuid.NewString(),
		started:           time.Now(),
		totalBytes:        total,
		lastProgressFlush: time.Time{},
		checkInterval:     2 * time.Second,
		lastCheck:         time.Now().Add(-2 * time.Second),
		parts:             make(map[int]*partTracker),
	}
}

func (r *wsReporter) send(evType string, payload any) {
	if !r.ensureAgent() {
		return
	}
	r.emitStart()
	if !r.enabled || (r.startPayload != nil && !r.startSent) {
		return
	}
	_ = r.writeEvent(evType, payload)
}

func (r *wsReporter) start(bucket, key, uploadID string, partSizeBytes int64, extra map[string]any) {
	p := map[string]any{
		"bucket":   bucket,
		"key":      key,
		"uploadId": uploadID,
		"partMB":   float64(partSizeBytes) / (1024.0 * 1024.0),
		"total":    r.totalBytes,
	}
	for k, v := range extra {
		p[k] = v
	}
	r.startPayload = p
	r.startSent = false
	r.emitStart()
}

func (r *wsReporter) progressAdd(delta int64) {
	if delta <= 0 {
		return
	}

	r.mu.Lock()
	r.uploadedBytes += delta
	uploadedBytes := r.uploadedBytes
	totalBytes := r.totalBytes
	started := r.started
	shouldFlush := r.lastProgressFlush.IsZero() || time.Since(r.lastProgressFlush) >= 250*time.Millisecond
	r.mu.Unlock()

	if !shouldFlush || !r.ensureAgent() {
		return
	}

	// Compute outside lock
	elapsed := time.Since(started).Seconds()
	var bps float64
	if elapsed > 0 {
		bps = float64(uploadedBytes) / elapsed
	}
	var pct float64
	if totalBytes > 0 {
		pct = (float64(uploadedBytes) / float64(totalBytes)) * 100.0
	}

	r.send("total_progress", map[string]any{
		"bytes":   uploadedBytes,
		"total":   totalBytes,
		"percent": pct,
		"bps":     bps,
	})

	r.mu.Lock()
	r.lastProgressFlush = time.Now()
	r.mu.Unlock()
}

func (r *wsReporter) totalProgressImmediate(bytes int64) {
	// resume 초기 바이트 등 즉시 1회 송신
	r.mu.Lock()
	r.uploadedBytes = bytes
	uploadedBytes := r.uploadedBytes
	totalBytes := r.totalBytes
	started := r.started
	r.mu.Unlock()

	if !r.ensureAgent() {
		return
	}

	elapsed := time.Since(started).Seconds()
	var bps float64
	if elapsed > 0 {
		bps = float64(uploadedBytes) / elapsed
	}
	var pct float64
	if totalBytes > 0 {
		pct = (float64(uploadedBytes) / float64(totalBytes)) * 100.0
	}

	r.send("total_progress", map[string]any{
		"bytes":   uploadedBytes,
		"total":   totalBytes,
		"percent": pct,
		"bps":     bps,
	})

	r.mu.Lock()
	r.lastProgressFlush = time.Now()
	r.mu.Unlock()
}

func (r *wsReporter) partStart(part int, size int64, offset int64) {
	tr := &partTracker{
		size:        size,
		sent:        0,
		started:     time.Now(),
		lastFlushAt: time.Time{},
	}

	r.mu.Lock()
	r.parts[part] = tr
	r.mu.Unlock()

	r.send("part_start", map[string]any{
		"part":   part,
		"size":   size,
		"offset": offset,
	})
}

func (r *wsReporter) partProgressAdd(part int, delta int64) {
	if delta <= 0 {
		return
	}

	r.mu.Lock()
	tr, ok := r.parts[part]
	if !ok {
		r.mu.Unlock()
		return
	}
	tr.sent += delta

	shouldFlush := tr.lastFlushAt.IsZero() || time.Since(tr.lastFlushAt) >= 200*time.Millisecond
	if !shouldFlush {
		r.mu.Unlock()
		return
	}

	// Copy values for computation outside lock
	sent := tr.sent
	size := tr.size
	started := tr.started
	tr.lastFlushAt = time.Now()
	r.mu.Unlock()

	// Compute outside lock
	var pct float64
	if size > 0 {
		pct = (float64(sent) / float64(size)) * 100.0
	}
	elapsed := time.Since(started).Seconds()
	var bps float64
	if elapsed > 0 {
		bps = float64(sent) / elapsed
	}

	r.send("part_progress", map[string]any{
		"part":    part,
		"sent":    sent,
		"size":    size,
		"percent": pct,
		"bps":     bps,
	})
}

func (r *wsReporter) partDone(part int, size int64, etag string) {
	r.send("part_done", map[string]any{
		"part": part,
		"size": size,
		"etag": etag,
	})

	r.mu.Lock()
	delete(r.parts, part)
	r.mu.Unlock()
}

func (r *wsReporter) error(msg string, partNum *int) {
	payload := map[string]any{
		"message": msg,
	}
	if partNum != nil {
		payload["part"] = *partNum
	}
	r.send("error", payload)
}

func (r *wsReporter) done(success bool, uploadID string) {
	r.mu.Lock()
	started := r.started
	uploadedBytes := r.uploadedBytes
	totalBytes := r.totalBytes
	r.mu.Unlock()

	dur := time.Since(started)
	r.send("session_done", map[string]any{
		"success":  success,
		"uploadId": uploadID,
		"duration": dur.String(),
		"bytes":    uploadedBytes,
		"total":    totalBytes,
	})
}

func (r *wsReporter) ensureAgent() bool {
	r.mu.Lock()
	if r.enabled {
		r.mu.Unlock()
		return true
	}
	if time.Since(r.lastCheck) < r.checkInterval {
		r.mu.Unlock()
		return false
	}
	r.lastCheck = time.Now()
	addr := r.addr
	r.mu.Unlock()

	if wsagent.IsRunningAt(addr) {
		r.mu.Lock()
		r.enabled = true
		r.mu.Unlock()
		return true
	}
	return false
}

func (r *wsReporter) emitStart() {
	r.mu.Lock()
	if r.startPayload == nil || r.startSent {
		r.mu.Unlock()
		return
	}
	payload := r.startPayload
	r.mu.Unlock()

	if !r.ensureAgent() {
		return
	}
	_ = r.writeEvent("session_start", payload)
}

func (r *wsReporter) writeEvent(evType string, payload any) error {
	b, _ := json.Marshal(payload)
	fmt.Printf("[WS-DEBUG] send → type=%s payload=%s\n", evType, string(b))

	r.mu.Lock()
	addr := r.addr
	runID := r.runID
	r.mu.Unlock()

	err := wsagent.SendEvent(context.Background(), addr, wsagent.Event{
		Type:      evType,
		RunID:     runID,
		Timestamp: time.Now(),
		Payload:   b,
	})
	if err != nil {
		r.handleSendError(evType, err)
		return err
	}
	if evType == "session_start" {
		r.mu.Lock()
		r.startSent = true
		r.mu.Unlock()
	}
	return nil
}

func (r *wsReporter) handleSendError(evType string, err error) {
	if err == nil {
		return
	}

	r.mu.Lock()
	r.enabled = false
	r.lastCheck = time.Now()
	shouldLog := time.Since(r.lastErrorLog) >= 5*time.Second
	if shouldLog {
		r.lastErrorLog = time.Now()
	}
	r.mu.Unlock()

	if shouldLog {
		fmt.Fprintf(os.Stderr, "warn: failed to deliver WebSocket event %q: %v\n", evType, err)
	}
}
