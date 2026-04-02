package utils

import (
	"errors"
	"testing"
	"time"
)

// TestRetryExponentialBackoff 재시도 간격이 지수적으로 증가하는지 확인
func TestRetryExponentialBackoff(t *testing.T) {
	base := 100 * time.Millisecond
	attempts := 4
	var intervals []time.Duration
	var prev time.Time

	callCount := 0
	err := Retry(attempts, base, func() error {
		if callCount > 0 {
			intervals = append(intervals, time.Since(prev))
		}
		prev = time.Now()
		callCount++
		return errors.New("always fail")
	})

	if err == nil {
		t.Fatal("expected error but got nil")
	}

	t.Logf("총 호출 횟수: %d (기대: %d)", callCount, attempts)

	// 재시도 간격 출력
	for i, d := range intervals {
		expected := base * time.Duration(1<<uint(i))
		t.Logf("재시도 %d 대기: %.0fms (기대 base: %dms, jitter 포함이므로 더 클 수 있음)",
			i+1, float64(d.Milliseconds()), expected.Milliseconds())

		// 최소값: base * 2^i 이상이어야 함 (jitter는 양수)
		if d < expected {
			t.Errorf("재시도 %d 대기(%dms)가 기대값(%dms)보다 짧음", i+1, d.Milliseconds(), expected.Milliseconds())
		}
		// 최대값: base * 2^i * 1.5 + 여유 10ms (타이머 오차)
		maxExpected := expected + expected/2 + 10*time.Millisecond
		if d > maxExpected {
			t.Errorf("재시도 %d 대기(%dms)가 최대값(%dms)을 초과", i+1, d.Milliseconds(), maxExpected.Milliseconds())
		}
	}

	if callCount != attempts {
		t.Errorf("호출 횟수 불일치: got %d, want %d", callCount, attempts)
	}
}

// TestRetrySuccessOnThirdAttempt 3번째 시도에서 성공하는 케이스
func TestRetrySuccessOnThirdAttempt(t *testing.T) {
	callCount := 0
	err := Retry(5, 50*time.Millisecond, func() error {
		callCount++
		if callCount < 3 {
			return errors.New("not yet")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success but got: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
	t.Logf("3번째 시도에서 성공, 총 호출: %d", callCount)
}

// TestRetryMaxWaitCap 대기 시간이 30s 캡을 초과하지 않는지 확인 (짧은 단위로 검증)
func TestRetryMaxWaitCap(t *testing.T) {
	base := 10 * time.Millisecond
	const maxWait = 30 * time.Second

	var maxObserved time.Duration
	var prev time.Time
	callCount := 0

	Retry(10, base, func() error {
		if callCount > 0 {
			d := time.Since(prev)
			if d > maxObserved {
				maxObserved = d
			}
		}
		prev = time.Now()
		callCount++
		return errors.New("fail")
	})

	t.Logf("최대 관측 대기: %v", maxObserved)
	if maxObserved > maxWait+100*time.Millisecond {
		t.Errorf("대기 시간이 cap(30s)을 초과: %v", maxObserved)
	}
}
