package portfolio

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/logging"
)

const (
	// rebuildTimeout bounds one background rebuild, including the time spent waiting for a
	// rebuild started elsewhere, so one that hangs cannot hold the build lock for the
	// lifetime of the process.
	rebuildTimeout = 30 * time.Minute
	// rebuildRetryDelay is how long a queued rebuild waits before trying again while a
	// rebuild started elsewhere (an import, the rebuild action) holds the account's lock.
	rebuildRetryDelay = time.Second
)

// rebuildQueue runs the rebuilds manual writes ask for in the background, so a write
// answers without waiting for one. It runs at most one rebuild per account at a time and
// folds every request that arrives meanwhile into a single follow-up run, so the last
// rebuild always starts after the last write and its result includes every one of them.
type rebuildQueue struct {
	rb         rebuilder
	log        logging.Logger
	retryDelay time.Duration

	mu sync.Mutex
	// pending is keyed by the accounts with a rebuild running; the value says whether
	// another request arrived since that run started.
	pending map[uuid.UUID]bool
}

func newRebuildQueue(rb rebuilder, log logging.Logger) *rebuildQueue {
	return &rebuildQueue{rb: rb, log: log, retryDelay: rebuildRetryDelay, pending: make(map[uuid.UUID]bool)}
}

// request schedules a rebuild of the account. The rebuild runs on a context detached from
// ctx, because Build replaces the account's positions and a client that navigates away must
// not cut that short.
func (q *rebuildQueue) request(ctx context.Context, accountID uuid.UUID) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, running := q.pending[accountID]; running {
		q.pending[accountID] = true
		return
	}
	q.pending[accountID] = false
	go q.run(context.WithoutCancel(ctx), accountID)
}

func (q *rebuildQueue) run(ctx context.Context, accountID uuid.UUID) {
	for {
		q.buildOnce(ctx, accountID)

		q.mu.Lock()
		if !q.pending[accountID] {
			delete(q.pending, accountID)
			q.mu.Unlock()
			return
		}
		q.pending[accountID] = false
		q.mu.Unlock()
	}
}

// buildOnce runs one rebuild. A rebuild started elsewhere may have read the transactions
// before the write that asked for this one, so it is waited out rather than counted as done.
func (q *rebuildQueue) buildOnce(ctx context.Context, accountID uuid.UUID) {
	ctx, cancel := context.WithTimeout(ctx, rebuildTimeout)
	defer cancel()
	for {
		err := q.rb.Build(ctx, accountID)
		switch {
		case err == nil, errors.Is(err, ErrPortfolioNoSnapshots), errors.Is(err, ErrAccountNotFound):
			return
		case errors.Is(err, ErrBuildInProgress):
			select {
			case <-ctx.Done():
				q.logFailure(ctx, accountID, ctx.Err())
				return
			case <-time.After(q.retryDelay):
			}
		default:
			q.logFailure(ctx, accountID, err)
			return
		}
	}
}

// logFailure records a background rebuild that did not finish. The write that asked for
// it has already answered, so the log is the only place the reason survives.
func (q *rebuildQueue) logFailure(ctx context.Context, accountID uuid.UUID, err error) {
	if q.log == nil {
		return
	}
	q.log.Error(ctx, "portfolio: background rebuild after a manual write failed", err, "account_id", accountID.String())
}
