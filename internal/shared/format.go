package shared

import (
	"fmt"
	"time"
)

func FormatCount(n int64) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func FormatBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.1fGB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1fMB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1fKB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%dB", b)
	}
}

// Rate computes per-second rate from two counter snapshots.
func Rate(prev, curr int64, dt time.Duration) float64 {
	sec := dt.Seconds()
	if sec <= 0 {
		return 0
	}
	return float64(curr-prev) / sec
}

// FormatRate formats a per-second rate. For non-byte rates isBytes=false.
func FormatRate(val float64, humanize, isBytes bool) string {
	if !humanize {
		return fmt.Sprintf("%.0f/s", val)
	}
	if isBytes {
		return FormatBytes(int64(val)) + "/s"
	}
	return FormatCount(int64(val)) + "/s"
}
