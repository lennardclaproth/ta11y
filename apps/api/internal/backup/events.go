package backup

import "github.com/google/uuid"

const (
	// TopicPortfolioRestored is published after a portfolio file brought transactions in.
	TopicPortfolioRestored = "backup.portfolio.restored"
	// TopicAssetsRestored is published after an assets file brought items in.
	TopicAssetsRestored = "backup.assets.restored"
)

// Restored asks the feature that derives data from restored records — portfolio
// positions and snapshots, or asset snapshots — to rebuild it for the account. A
// cashflow file restores nothing that is derived, so it publishes nothing.
type Restored struct {
	AccID uuid.UUID
}
