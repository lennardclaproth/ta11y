package portfolio

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/sorting"
)

// Queries exposes portfolio read-side use cases.
type Queries struct {
	qs  QueryStore
	mdq listingResolver
}

// QueryStore reads portfolio snapshots and positions.
type QueryStore interface {
	SnapshotsForAccount(ctx context.Context, accountID uuid.UUID, limit, offset *int, from, to *time.Time, sort *sorting.Direction) ([]*PortfolioSnapshot, error)
	PositionsWithLatestSnapshot(ctx context.Context, accID uuid.UUID, includeClosed bool) ([]*PositionWithLatestSnapshot, error)
	ImportedProducts(ctx context.Context, accountID, importID uuid.UUID) ([]ImportedProduct, error)
}

// listingResolver answers the same question the rebuild asks when it links a position:
// which listing, exactly, is this instrument?
type listingResolver interface {
	ListingByIdentity(ctx context.Context, isin, symbol string) (*marketdata.Listing, error)
}

// ImportedProduct is one distinct instrument an import contributed, with how many of
// the import's rows mentioned it.
type ImportedProduct struct {
	Name         string  `db:"name"`
	ISIN         *string `db:"isin"`
	Symbol       *string `db:"symbol"`
	Transactions int     `db:"transactions"`
}

// NewQueries creates portfolio read-side use cases.
func NewQueries(qs QueryStore, mdq listingResolver) *Queries {
	return &Queries{
		qs:  qs,
		mdq: mdq,
	}
}

// UnlinkedProducts returns the instruments an import brought in that no listing
// matches, so a caller can say which products are held but do not yet count towards
// performance. It resolves against the same exact identity the rebuild links on, so a
// product disappears from this list once its listing exists and the portfolio is
// rebuilt.
func (q *Queries) UnlinkedProducts(ctx context.Context, accountID, importID uuid.UUID) ([]ImportedProduct, error) {
	products, err := q.qs.ImportedProducts(ctx, accountID, importID)
	if err != nil {
		return nil, fmt.Errorf("portfolio unlinked products: %w", err)
	}
	unlinked := make([]ImportedProduct, 0, len(products))
	for _, product := range products {
		listing, err := q.mdq.ListingByIdentity(ctx, normalizedID(product.ISIN), normalizedID(product.Symbol))
		if err != nil {
			return nil, fmt.Errorf("portfolio unlinked products: resolve listing: %w", err)
		}
		if listing == nil {
			unlinked = append(unlinked, product)
		}
	}
	return unlinked, nil
}

// SnapshotsForAccount returns portfolio snapshots for the given account.
// When limit, offset, from and to are nil, it returns all snapshots using the store's default ordering.
func (q *Queries) SnapshotsForAccount(
	ctx context.Context,
	accountID uuid.UUID,
	limit, offset *int,
	from, to *time.Time,
	sort *sorting.Direction,
) ([]*PortfolioSnapshot, error) {
	return q.qs.SnapshotsForAccount(ctx, accountID, limit, offset, from, to, sort)
}

// PositionsForAccount returns account positions with their latest snapshot metadata.
func (q *Queries) PositionsForAccount(
	ctx context.Context,
	accountID uuid.UUID,
	includeClosed bool,
) ([]*PositionWithLatestSnapshot, error) {
	positions, err := q.qs.PositionsWithLatestSnapshot(ctx, accountID, includeClosed)
	if err != nil {
		return nil, fmt.Errorf("portfolio positions: %w", err)
	}
	if positions == nil {
		return []*PositionWithLatestSnapshot{}, nil
	}
	return positions, nil
}
