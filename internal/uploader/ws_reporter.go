package uploader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	r.uploadedBytes += delta

	// 250ms 스로틀
	if !r.ensureAgent() {
		return
	}
	now := time.Now()
	if r.lastProgressFlush.IsZero() || now.Sub(r.lastProgressFlush) >= 250*time.Millisecond {
		elapsed := now.Sub(r.started).Seconds()
		var bps float64
		if elapsed > 0 {
			bps = float64(r.uploadedBytes) / elapsed
		}
		var pct float64
		if r.totalBytes > 0 {
			pct = (float64(r.uploadedBytes) / float64(r.totalBytes)) * 100.0
		}
		r.send("total_progress", map[string]any{
			"bytes":   r.uploadedBytes,
			"total":   r.totalBytes,
			"percent": pct,
			"bps":     bps,
		})
		r.lastProgressFlush = now
	}
}

func (r *wsReporter) totalProgressImmediate(bytes int64) {
	// resume 초기 바이트 등 즉시 1회 송신
	r.uploadedBytes = bytes
	if !r.ensureAgent() {
		return
	}
	elapsed := time.Since(r.started).Seconds()
	var bps float64
	if elapsed > 0 {
		bps = float64(r.uploadedBytes) / elapsed
	}
	var pct float64
	if r.totalBytes > 0 {
		pct = (float64(r.uploadedBytes) / float64(r.totalBytes)) * 100.0
	}
	r.send("total_progress", map[string]any{
		"bytes":   r.uploadedBytes,
		"total":   r.totalBytes,
		"percent": pct,
		"bps":     bps,
	})
	r.lastProgressFlush = time.Now()
}

func (r *wsReporter) partStart(part int, size int64, offset int64) {
	tr := &partTracker{
		size:        size,
		sent:        0,
		started:     time.Now(),
		lastFlushAt: time.Time{},
	}
	r.parts[part] = tr
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
	tr, ok := r.parts[part]
	if !ok {
		return
	}
	tr.sent += delta

	now := time.Now()
	if tr.lastFlushAt.IsZero() || now.Sub(tr.lastFlushAt) >= 200*time.Millisecond {
		var pct float64
		if tr.size > 0 {
			pct = (float64(tr.sent) / float64(tr.size)) * 100.0
		}
		elapsed := now.Sub(tr.started).Seconds()
		var bps float64
		if elapsed > 0 {
			bps = float64(tr.sent) / elapsed
		}
		r.send("part_progress", map[string]any{
			"part":    part,
			"sent":    tr.sent,
			"size":    tr.size,
			"percent": pct,
			"bps":     bps,
		})
		tr.lastFlushAt = now
	}
}

func (r *wsReporter) partDone(part int, size int64, etag string) {
	r.send("part_done", map[string]any{
		"part": part,
		"size": size,
		"etag": etag,
	})
	delete(r.parts, part)
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
	dur := time.Since(r.started)
	r.send("session_done", map[string]any{
		"success":  success,
		"uploadId": uploadID,
		"duration": dur.String(),
		"bytes":    r.uploadedBytes,
		"total":    r.totalBytes,
	})
}

func (r *wsReporter) ensureAgent() bool {
	if r.enabled {
		return true
	}
	if time.Since(r.lastCheck) < r.checkInterval {
		return false
	}
	r.lastCheck = time.Now()
	if wsagent.IsRunningAt(r.addr) {
		r.enabled = true
		return true
	}
	return false
}

func (r *wsReporter) emitStart() {
	if r.startPayload == nil || r.startSent {
		return
	}
	if !r.ensureAgent() {
		return
	}
	_ = r.writeEvent("session_start", r.startPayload)
}

func (r *wsReporter) writeEvent(evType string, payload any) error {
	b, _ := json.Marshal(payload)
	fmt.Printf("[WS-DEBUG] send → type=%s payload=%s\n", evType, string(b))
	err := wsagent.SendEvent(context.Background(), r.addr, wsagent.Event{
		Type:      evType,
		RunID:     r.runID,
		Timestamp: time.Now(),
		Payload:   b,
	})
	if err != nil {
		r.handleSendError(evType, err)
		return err
	}
	if evType == "session_start" {
		r.startSent = true
	}
	return nil
}

func (r *wsReporter) handleSendError(evType string, err error) {
	if err == nil {
		return
	}
	r.enabled = false
	r.lastCheck = time.Now()
	if time.Since(r.lastErrorLog) >= 5*time.Second {
		fmt.Fprintf(os.Stderr, "warn: failed to deliver WebSocket event %q: %v\n", evType, err)
		r.lastErrorLog = time.Now()
	}
}
