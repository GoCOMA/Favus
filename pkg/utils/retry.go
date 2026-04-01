package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// Retry executes fn up to attempts times.
// On failure it waits base * 2^i + jitter before the next attempt (Exponential Backoff with Jitter).
// sleep is the base wait duration. Maximum wait per attempt is capped at 30s.
func Retry(attempts int, sleep time.Duration, fn func() error) error {
	const maxWait = 30 * time.Second

	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		// No sleep after the last attempt
		if i == attempts-1 {
			break
		}

		// backoff = base * 2^i  (2s, 4s, 8s, 16s ...)
		backoff := sleep * time.Duration(1<<uint(i))

		// jitter = random value in [0, backoff/2)
		jitter := time.Duration(rand.Int63n(int64(backoff / 2)))

		wait := backoff + jitter
		if wait > maxWait {
			wait = maxWait
		}

		fmt.Printf("Retrying (%d/%d) after %.1fs: %v\n", i+1, attempts, wait.Seconds(), err)
		time.Sleep(wait)
	}
	return fmt.Errorf("all retries failed: %w", err)
}
