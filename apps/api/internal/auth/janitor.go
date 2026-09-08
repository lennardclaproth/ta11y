package auth

import (
	"context"
	"time"

	"github.com/lennardclaproth/ta11y/internal/logging"
)

// SessionSweepInterval is how often lapsed sessions are removed.
const SessionSweepInterval = 6 * time.Hour

// ExpiredSessionDeleter removes sessions that lapsed before a cutoff.
type ExpiredSessionDeleter interface {
	DeleteExpiredSessions(ctx context.Context, before time.Time) (int64, error)
}

// SweepExpiredSessions periodically deletes lapsed sessions until ctx is cancelled.
// Expiry is already enforced when a session is resolved, so this is housekeeping rather
// than a security control -- a failed sweep is logged and retried, never fatal.
func SweepExpiredSessions(ctx context.Context, deleter ExpiredSessionDeleter, log logging.Logger) {
	ticker := time.NewTicker(SessionSweepInterval)
	defer ticker.Stop()

	sweep := func() {
		removed, err := deleter.DeleteExpiredSessions(ctx, time.Now().UTC())
		if err != nil {
			log.Warn(ctx, "failed sweeping expired sessions", "error", err.Error())
			return
		}
		if removed > 0 {
			log.Info(ctx, "swept expired sessions", "removed", removed)
		}
	}

	sweep()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sweep()
		}
	}
}
