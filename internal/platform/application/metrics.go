package application

import (
	"fmt"
	"time"
)

func (s *Store) PrometheusMetrics(now time.Time) string {
	pending, failed, oldest := s.OutboxStats(now)
	return fmt.Sprintf(
		"# HELP openrevenue_outbox_pending Pending unpublished outbox records.\n"+
			"# TYPE openrevenue_outbox_pending gauge\n"+
			"openrevenue_outbox_pending %d\n"+
			"# HELP openrevenue_outbox_failed Pending records with a failed publication attempt.\n"+
			"# TYPE openrevenue_outbox_failed gauge\n"+
			"openrevenue_outbox_failed %d\n"+
			"# HELP openrevenue_outbox_oldest_age_seconds Age of the oldest unpublished record.\n"+
			"# TYPE openrevenue_outbox_oldest_age_seconds gauge\n"+
			"openrevenue_outbox_oldest_age_seconds %.0f\n",
		pending, failed, max(0, oldest.Seconds()),
	)
}
