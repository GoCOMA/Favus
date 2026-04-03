package uploader

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/GoCOMA/Favus/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// NetworkProfile holds the measured RTT and the recommended upload parameters.
type NetworkProfile struct {
	RTT        time.Duration
	Workers    int
	PartSizeMB int
	Label      string
}

// MeasureNetwork measures RTT to S3 and returns recommended workers and part size.
func (u *Uploader) MeasureNetwork(ctx context.Context) (*NetworkProfile, error) {
	return tuneNetwork(ctx, u.s3Client, u.Config.Bucket)
}

// tuneNetwork measures RTT to S3 and returns recommended workers and part size.
// It sends 3 HeadBucket probes and uses the median RTT to pick a tier.
func tuneNetwork(ctx context.Context, s3Client *s3.Client, bucket string) (*NetworkProfile, error) {
	samples := make([]time.Duration, 3)
	for i := range samples {
		start := time.Now()
		_, err := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &bucket})
		if err != nil {
			return nil, fmt.Errorf("RTT probe failed: %w", err)
		}
		samples[i] = time.Since(start)
	}

	rtt := median(samples)
	profile := recommend(rtt)
	profile.RTT = rtt

	utils.Info(fmt.Sprintf(
		"[AutoTune] RTT=%.1fms → %s → workers=%d, partSize=%dMB",
		float64(rtt.Milliseconds()), profile.Label, profile.Workers, profile.PartSizeMB,
	))

	return profile, nil
}

// recommend maps an RTT value to upload parameters based on empirical benchmarks.
//
//	< 3ms  → EC2 same-region  (measured: ~1ms,  peak 499 MB/s at 32 workers)
//	3-15ms → EC2 cross-region (moderate latency)
//	> 15ms → Local / external (measured: ~40ms, peak  60 MB/s at 16 workers)
func recommend(rtt time.Duration) *NetworkProfile {
	switch {
	case rtt < 3*time.Millisecond:
		return &NetworkProfile{
			Workers:    32,
			PartSizeMB: 50,
			Label:      "Low latency (<3ms)",
		}
	case rtt < 15*time.Millisecond:
		return &NetworkProfile{
			Workers:    16,
			PartSizeMB: 20,
			Label:      "Medium latency (3~15ms)",
		}
	default:
		return &NetworkProfile{
			Workers:    4,
			PartSizeMB: 10,
			Label:      "High latency (>15ms)",
		}
	}
}

func median(d []time.Duration) time.Duration {
	cp := make([]time.Duration, len(d))
	copy(cp, d)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	return cp[len(cp)/2]
}
