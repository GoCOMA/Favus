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
		parts:             make(map[int]*partTracker),
	}
}

func (r *wsReporter) send(evType string, payload any) {
	if !r.enabled {
		return
	}
	b, _ := json.Marshal(payload)
	fmt.Printf("[WS-DEBUG] send → type=%s payload=%s\n", evType, string(b))
	_ = wsagent.SendEvent(context.Background(), r.addr, wsagent.Event{
		Type:      evType,
		RunID:     r.runID,
		Timestamp: time.Now(),
		Payload:   b,
	})
}

func (r *wsReporter) start(bucket, key, uploadID string, partSizeBytes int64, extra map[string]any) {
	if !r.enabled {
		return
	}
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
	r.send("session_start", p)
}

func (r *wsReporter) progressAdd(delta int64) {
	if !r.enabled || delta <= 0 {
		return
	}
	r.uploadedBytes += delta

	// 250ms 스로틀
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
	if !r.enabled {
		return
	}
	r.uploadedBytes = bytes
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
	if !r.enabled {
		return
	}
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
	if !r.enabled || delta <= 0 {
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
	if !r.enabled {
		return
	}
	r.send("part_done", map[string]any{
		"part": part,
		"size": size,
		"etag": etag,
	})
	delete(r.parts, part)
}

func (r *wsReporter) error(msg string, partNum *int) {
	if !r.enabled {
		return
	}
	payload := map[string]any{
		"message": msg,
	}
	if partNum != nil {
		payload["part"] = *partNum
	}
	r.send("error", payload)
}

func (r *wsReporter) done(success bool, uploadID string) {
	if !r.enabled {
		return
	}
	dur := time.Since(r.started)
	r.send("session_done", map[string]any{
		"success":  success,
		"uploadId": uploadID,
		"duration": dur.String(),
		"bytes":    r.uploadedBytes,
		"total":    r.totalBytes,
	})
}
