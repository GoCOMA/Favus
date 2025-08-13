package uploader

import (
	"io"
)

// ReadSeekCloserProgress wraps an io.ReadSeekCloser and calls onDelta with the net-new bytes read.
// It adjusts correctly when the underlying reader Seek()s backwards (e.g., SDK retries).
type ReadSeekCloserProgress struct {
	r        io.ReadSeekCloser
	onDelta  func(n int64)
	curPos   int64 // current position within this stream (bytes already read from start)
	reported int64
}

func NewReadSeekCloserProgress(r io.ReadSeekCloser, onDelta func(n int64)) *ReadSeekCloserProgress {
	return &ReadSeekCloserProgress{r: r, onDelta: onDelta}
}

func (p *ReadSeekCloserProgress) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.curPos += int64(n)
		// net new from last reported; normally equals n unless we were previously rewound
		delta := p.curPos - p.reported
		if delta > 0 {
			p.onDelta(delta)
			p.reported += delta
		}
	}
	return n, err
}

func (p *ReadSeekCloserProgress) Seek(offset int64, whence int) (int64, error) {
	newPos, err := p.r.Seek(offset, whence)
	if err != nil {
		return 0, err
	}
	p.curPos = newPos
	// If we rewound, prevent double counting on re-read
	if p.curPos < p.reported {
		p.reported = p.curPos
	}
	return newPos, nil
}

func (p *ReadSeekCloserProgress) Close() error {
	return p.r.Close()
}
