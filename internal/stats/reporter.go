package stats

import (
	"fmt"
	"time"

	netbridge "github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/pkg/humanize"
)

type Reporter struct{}

func (r *Reporter) FormatTable(stats netbridge.TrafficStats) string {
	return fmt.Sprintf(
		"  ↑ Upload  : %s/s  (total: %s)\n  ↓ Download: %s/s  (total: %s)\n  Uptime    : %s",
		humanize.Bytes(int64(stats.RateUp)),
		humanize.Bytes(stats.BytesUp),
		humanize.Bytes(int64(stats.RateDown)),
		humanize.Bytes(stats.BytesDown),
		humanize.Duration(stats.Uptime),
	)
}

func (r *Reporter) FormatJSON(stats netbridge.TrafficStats) map[string]any {
	return map[string]any{
		"bytes_up":    stats.BytesUp,
		"bytes_down":  stats.BytesDown,
		"rate_up":     stats.RateUp,
		"rate_down":   stats.RateDown,
		"uptime":      stats.Uptime.String(),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}
}
