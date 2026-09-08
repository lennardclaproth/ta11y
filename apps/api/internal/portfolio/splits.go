package portfolio

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/ta11y/internal/marketdata"
)

// SplitSource supplies the share splits recorded against a listing.
type SplitSource interface {
	SplitsForListing(ctx context.Context, listingID uuid.UUID) ([]marketdata.Split, error)
}

// splitEvents turns a listing's splits into synthetic transactions carrying the
// instrument's identity, so they route to the same position cycle as its trades.
//
// The multiplier travels in Quantity. These rows are never persisted: they carry no
// id and no PositionID, and the transaction-to-position mapping skips anything
// without one.
func splitEvents(splits []marketdata.Split, isin, symbol *string) []Transaction {
	events := make([]Transaction, 0, len(splits))
	for _, split := range splits {
		if split.Factor <= 0 || split.Factor == 1 {
			continue
		}
		events = append(events, Transaction{
			Origin:      TransactionOriginImport,
			Source:      "corporate-action",
			OccurredAt:  split.Date.UTC(),
			ISIN:        isin,
			Symbol:      symbol,
			Type:        TxSplit,
			Quantity:    split.Factor,
			Description: fmt.Sprintf("%g-for-1 share split", split.Factor),
		})
	}
	return events
}

// mergeChronologically interleaves synthetic events into an already-sorted
// transaction stream.
//
// Splits sort before same-instant trades: a trade executed on the ex-date is already
// quoted in post-split shares, so rescaling after it would multiply the new shares a
// second time.
func mergeChronologically(transactions, events []Transaction) []Transaction {
	if len(events) == 0 {
		return transactions
	}
	merged := make([]Transaction, 0, len(transactions)+len(events))
	merged = append(merged, transactions...)
	merged = append(merged, events...)
	sort.SliceStable(merged, func(i, j int) bool {
		left, right := merged[i], merged[j]
		if left.OccurredAt.Equal(right.OccurredAt) {
			return left.Type == TxSplit && right.Type != TxSplit
		}
		return left.OccurredAt.Before(right.OccurredAt)
	})
	return merged
}

// instrument identifies one tradable thing across the transactions that mention it.
type instrument struct {
	isin   *string
	symbol *string
}

// collectInstruments returns the distinct instruments in a transaction stream, keyed
// exactly as position cycles are.
//
// Sharing canonicalPositionKey's alias promotion matters: a stream whose early rows
// carry only a symbol and whose later rows add the ISIN describes one instrument, and
// counting it twice would resolve two listings and inject the same split twice --
// scaling the holding by the square of the factor.
func collectInstruments(transactions []Transaction) map[string]instrument {
	aliases := make(map[string]string)
	found := make(map[string]instrument)
	// Symbol-keyed entries that later turn out to have an ISIN are folded into it.
	bySymbol := make(map[string]string)

	for _, tx := range transactions {
		if tx.Type == TxCash {
			continue
		}
		key, symbolKey, err := canonicalPositionKey(tx, aliases)
		if err != nil {
			continue // nothing to identify the instrument by
		}
		// Promote an earlier symbol-only entry now that the ISIN is known.
		if symbolKey != "" && key != symbolKey {
			if previousKey, ok := bySymbol[symbolKey]; ok && previousKey != key {
				if promoted, ok := found[previousKey]; ok {
					delete(found, previousKey)
					found[key] = mergeInstrument(found[key], promoted)
				}
			}
			bySymbol[symbolKey] = key
		} else if symbolKey != "" {
			bySymbol[symbolKey] = key
		}

		found[key] = mergeInstrument(found[key], instrument{isin: tx.ISIN, symbol: tx.Symbol})
	}
	return found
}

// mergeInstrument fills gaps in an instrument's identity from another sighting of it.
func mergeInstrument(into, from instrument) instrument {
	if into.isin == nil {
		into.isin = from.isin
	}
	if into.symbol == nil {
		into.symbol = from.symbol
	}
	return into
}

// splitFactorsByDay indexes splits by UTC day for the snapshot walk, which steps a day
// at a time and needs to know whether the day it is about to build carries one.
func splitFactorsByDay(splits []marketdata.Split) map[time.Time]float64 {
	if len(splits) == 0 {
		return nil
	}
	byDay := make(map[time.Time]float64, len(splits))
	for _, split := range splits {
		if split.Factor <= 0 || split.Factor == 1 {
			continue
		}
		day := time.Date(
			split.Date.UTC().Year(), split.Date.UTC().Month(), split.Date.UTC().Day(),
			0, 0, 0, 0, time.UTC,
		)
		// Two records on one day compound rather than overwrite.
		if existing, ok := byDay[day]; ok {
			byDay[day] = existing * split.Factor
			continue
		}
		byDay[day] = split.Factor
	}
	return byDay
}

// storedTransactions drops the synthetic split events, leaving only rows that exist in
// the transactions table. The position mapping updates by transaction id, so a
// synthetic row would address nothing.
func storedTransactions(transactions []Transaction) []Transaction {
	stored := make([]Transaction, 0, len(transactions))
	for _, tx := range transactions {
		if tx.Type == TxSplit {
			continue
		}
		stored = append(stored, tx)
	}
	return stored
}
