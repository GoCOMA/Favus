package uploader

import (
	"context"
	"encoding/json"
	"time"

	"github.com/GoCOMA/Favus/internal/wsagent"
)

// WSTracker wraps UploadStatus and emits WebSocket events in addition
// to updating the local status file.
type WSTracker struct {
	*UploadStatus
	wsEnabled bool
	wsAddr    string
	runID     string
}

// NewWSTracker creates a new WSTracker wrapper.
// It auto-detects if the local wsagent is running.
func NewWSTracker(us *UploadStatus) *WSTracker {
	addr := wsagent.DefaultAddr()
	return &WSTracker{
		UploadStatus: us,
		wsEnabled:    wsagent.IsRunningAt(addr),
		wsAddr:       addr,
		runID:        us.UploadID,
	}
}

// AddCompletedPart updates status and emits a WS event.
func (wt *WSTracker) AddCompletedPart(partNumber int, eTag string) {
	wt.UploadStatus.AddCompletedPart(partNumber, eTag)

	if wt.wsEnabled {
		payload := map[string]any{
			"part": partNumber,
			"etag": eTag,
		}
		_ = wsagent.SendEvent(context.Background(), wt.wsAddr, wsagent.Event{
			Type:      "tracker_part_done",
			RunID:     wt.runID,
			Timestamp: time.Now(),
			Payload:   mustJSON(payload),
		})
	}
}

// SaveStatus saves to disk and emits a WS event.
func (wt *WSTracker) SaveStatus(statusFilePath string) error {
	err := wt.UploadStatus.SaveStatus(statusFilePath)
	if err != nil {
		return err
	}

	if wt.wsEnabled {
		payload := map[string]any{
			"statusFile": statusFilePath,
			"partsDone":  len(wt.UploadStatus.CompletedParts),
			"totalParts": wt.UploadStatus.TotalParts,
		}
		_ = wsagent.SendEvent(context.Background(), wt.wsAddr, wsagent.Event{
			Type:      "tracker_save",
			RunID:     wt.runID,
			Timestamp: time.Now(),
			Payload:   mustJSON(payload),
		})
	}
	return nil
}

// helper: panic 없는 RawMessage 변환
func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
