package uploader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/GoCOMA/Favus/internal/wsagent"
)

// WSTracker wraps UploadStatus and emits WebSocket events in addition
// to updating the local status file.
type WSTracker struct {
	*UploadStatus
	wsEnabled     bool
	wsAddr        string
	runID         string
	lastCheck     time.Time
	checkInterval time.Duration
	lastErrorLog  time.Time
}

// NewWSTracker creates a new WSTracker wrapper.
// It auto-detects if the local wsagent is running.
func NewWSTracker(us *UploadStatus) *WSTracker {
	addr := wsagent.DefaultAddr()
	enabled := wsagent.IsRunningAt(addr)
	return &WSTracker{
		UploadStatus:  us,
		wsEnabled:     enabled,
		wsAddr:        addr,
		runID:         us.UploadID,
		checkInterval: 2 * time.Second,
		lastCheck:     time.Now().Add(-2 * time.Second),
	}
}

// AddCompletedPart updates status and emits a WS event.
func (wt *WSTracker) AddCompletedPart(partNumber int, eTag string) {
	wt.UploadStatus.AddCompletedPart(partNumber, eTag)

	if wt.ensureAgent() {
		payload := map[string]any{
			"key":  wt.UploadStatus.Key,
			"part": partNumber,
			"etag": eTag,
		}
		err := wsagent.SendEvent(context.Background(), wt.wsAddr, wsagent.Event{
			Type:      "tracker_part_done",
			RunID:     wt.runID,
			Timestamp: time.Now(),
			Payload:   mustJSON(payload),
		})
		wt.handleSendError(err, "tracker_part_done")
	}
}

// SaveStatus saves to disk and emits a WS event.
func (wt *WSTracker) SaveStatus(statusFilePath string) error {
	err := wt.UploadStatus.SaveStatus(statusFilePath)
	if err != nil {
		return err
	}

	if wt.ensureAgent() {
		payload := map[string]any{
			"key":        wt.UploadStatus.Key,
			"statusFile": statusFilePath,
			"partsDone":  len(wt.UploadStatus.CompletedParts),
			"totalParts": wt.UploadStatus.TotalParts,
		}
		err = wsagent.SendEvent(context.Background(), wt.wsAddr, wsagent.Event{
			Type:      "tracker_save",
			RunID:     wt.runID,
			Timestamp: time.Now(),
			Payload:   mustJSON(payload),
		})
		wt.handleSendError(err, "tracker_save")
	}
	return nil
}

func (wt *WSTracker) ensureAgent() bool {
	if wt.wsEnabled {
		return true
	}
	if time.Since(wt.lastCheck) < wt.checkInterval {
		return false
	}
	wt.lastCheck = time.Now()
	if wsagent.IsRunningAt(wt.wsAddr) {
		wt.wsEnabled = true
		return true
	}
	return false
}

func (wt *WSTracker) handleSendError(err error, context string) {
	if err == nil {
		return
	}
	wt.wsEnabled = false
	wt.lastCheck = time.Now()
	if time.Since(wt.lastErrorLog) >= 5*time.Second {
		fmt.Fprintf(os.Stderr, "warn: UI tracker event (%s) delivery failed: %v\n", context, err)
		wt.lastErrorLog = time.Now()
	}
}

// helper: panic 없는 RawMessage 변환
func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
