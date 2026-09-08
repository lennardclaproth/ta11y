package portfolio

import (
	"testing"
	"time"

	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/money"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func buy(at time.Time, qty float64, unitPrice money.Price) Transaction {
	return Transaction{OccurredAt: at, Type: TxBuy, Quantity: qty, UnitPrice: unitPrice}
}

func sell(at time.Time, qty float64, unitPrice money.Price) Transaction {
	return Transaction{OccurredAt: at, Type: TxSell, Quantity: qty, UnitPrice: unitPrice}
}

func split(at time.Time, factor float64) Transaction {
	return Transaction{OccurredAt: at, Type: TxSplit, Quantity: factor}
}

// The case that motivated this: buy pre-split, hold through the split. Quantity has to
// scale with the price series, or the position silently shrinks to a quarter of itself.
func TestSplitScalesTheHeldQuantity(t *testing.T) {
	acc := PositionAcc{}

	// 10 shares at $500 -> $5,000 invested.
	if err := acc.ApplyTx(buy(day(2020, time.August, 1), 10, money.Price(50000))); err != nil {
		t.Fatalf("buy: %v", err)
	}
	costBefore := acc.CostBasis

	if err := acc.ApplyTx(split(day(2020, time.August, 31), 4)); err != nil {
		t.Fatalf("split: %v", err)
	}

	if acc.Quantity != 40 {
		t.Fatalf("quantity = %v, want 40 after a 4-for-1 split", acc.Quantity)
	}
	// A split moves no money: the same investment is simply divided into more shares.
	if acc.CostBasis != costBefore {
		t.Fatalf("cost basis = %v, want it unchanged at %v", acc.CostBasis, costBefore)
	}
	if acc.RealizedPnL != 0 {
		t.Fatalf("realized pnl = %v, want 0 -- a split realizes nothing", acc.RealizedPnL)
	}
}

// Market value is the whole point: post-split the price is a quarter, so an unscaled
// quantity would report a quarter of the real holding.
func TestSplitKeepsMarketValueContinuous(t *testing.T) {
	acc := PositionAcc{}
	if err := acc.ApplyTx(buy(day(2020, time.August, 1), 10, money.Price(50000))); err != nil {
		t.Fatalf("buy: %v", err)
	}

	preSplitValue := acc.Quantity * money.Price(50000).Float64()

	if err := acc.ApplyTx(split(day(2020, time.August, 31), 4)); err != nil {
		t.Fatalf("split: %v", err)
	}
	postSplitValue := acc.Quantity * money.Price(12500).Float64()

	if preSplitValue != postSplitValue {
		t.Fatalf("value jumped across the split: %v -> %v", preSplitValue, postSplitValue)
	}
}

// Selling the whole post-split holding must close the position at zero. Before the fix
// the sell was clamped to the pre-split quantity, leaving a phantom holding open.
func TestSellingTheFullPostSplitHoldingClosesThePosition(t *testing.T) {
	acc := PositionAcc{}
	if err := acc.ApplyTx(buy(day(2020, time.August, 1), 10, money.Price(50000))); err != nil {
		t.Fatalf("buy: %v", err)
	}
	if err := acc.ApplyTx(split(day(2020, time.August, 31), 4)); err != nil {
		t.Fatalf("split: %v", err)
	}
	if err := acc.ApplyTx(sell(day(2020, time.September, 10), 40, money.Price(15000))); err != nil {
		t.Fatalf("sell: %v", err)
	}

	if acc.Quantity != 0 {
		t.Fatalf("quantity = %v, want 0 after selling the whole holding", acc.Quantity)
	}
	// Bought for $5,000, sold 40 at $150 = $6,000, so $1,000 realized.
	if acc.RealizedPnL != money.Price(100000) {
		t.Fatalf("realized pnl = %v, want 100000 cents", acc.RealizedPnL)
	}
}

func TestSplitFactorsThatChangeNothing(t *testing.T) {
	cases := map[string]float64{
		"neutral factor":  1,
		"zero factor":     0,
		"negative factor": -2,
	}
	for name, factor := range cases {
		t.Run(name, func(t *testing.T) {
			acc := PositionAcc{}
			if err := acc.ApplyTx(buy(day(2020, time.August, 1), 10, money.Price(50000))); err != nil {
				t.Fatalf("buy: %v", err)
			}
			if err := acc.ApplyTx(split(day(2020, time.August, 31), factor)); err != nil {
				t.Fatalf("split: %v", err)
			}
			// A provider reporting nonsense must never wipe or invert a holding.
			if acc.Quantity != 10 {
				t.Fatalf("quantity = %v, want it untouched at 10", acc.Quantity)
			}
		})
	}
}

func TestMergeChronologicallyOrdersSplitsBeforeSameInstantTrades(t *testing.T) {
	exDate := day(2020, time.August, 31)
	transactions := []Transaction{
		buy(day(2020, time.August, 1), 10, money.Price(50000)),
		buy(exDate, 5, money.Price(12500)),
	}
	events := []Transaction{split(exDate, 4)}

	merged := mergeChronologically(transactions, events)

	if len(merged) != 3 {
		t.Fatalf("merged length = %d, want 3", len(merged))
	}
	if merged[1].Type != TxSplit {
		t.Fatalf("order = %v, want the split before the same-day buy", []TransactionType{
			merged[0].Type, merged[1].Type, merged[2].Type,
		})
	}

	// Replaying the merged stream: 10 pre-split shares become 40, then 5 more are
	// bought at the post-split price, for 45. Applying the split last would give 60.
	acc := PositionAcc{}
	for _, tx := range merged {
		if err := acc.ApplyTx(tx); err != nil {
			t.Fatalf("apply %s: %v", tx.Type, err)
		}
	}
	if acc.Quantity != 45 {
		t.Fatalf("quantity = %v, want 45", acc.Quantity)
	}
}

func TestMergeChronologicallyWithNoSplitsIsIdentity(t *testing.T) {
	transactions := []Transaction{
		buy(day(2020, time.August, 1), 10, money.Price(50000)),
		sell(day(2020, time.September, 1), 4, money.Price(60000)),
	}
	merged := mergeChronologically(transactions, nil)
	if len(merged) != 2 || merged[0].Type != TxBuy || merged[1].Type != TxSell {
		t.Fatal("a stream with no splits must come back unchanged")
	}
}

func TestSplitEventsCarryTheInstrumentIdentity(t *testing.T) {
	isin := "US0378331005"
	symbol := "AAPL"
	splits := []marketdata.Split{
		{Date: day(2014, time.June, 9), Factor: 7},
		{Date: day(2020, time.August, 31), Factor: 4},
		{Date: day(2021, time.January, 4), Factor: 1}, // ordinary day, must be dropped
	}

	events := splitEvents(splits, &isin, &symbol)

	if len(events) != 2 {
		t.Fatalf("events = %d, want 2 (the neutral factor is not an event)", len(events))
	}
	for _, event := range events {
		if event.Type != TxSplit {
			t.Fatalf("type = %s, want SPLIT", event.Type)
		}
		if event.ISIN == nil || *event.ISIN != isin {
			t.Fatal("split must carry the instrument's ISIN so it routes to the right cycle")
		}
		// Synthetic events must never look persistable.
		if event.PositionID != nil {
			t.Fatal("a synthetic split must not carry a position mapping")
		}
	}
}

func TestStoredTransactionsDropsSyntheticSplits(t *testing.T) {
	stream := []Transaction{
		buy(day(2020, time.August, 1), 10, money.Price(50000)),
		split(day(2020, time.August, 31), 4),
		sell(day(2020, time.September, 1), 40, money.Price(15000)),
	}

	stored := storedTransactions(stream)

	if len(stored) != 2 {
		t.Fatalf("stored = %d, want 2 -- the split is not a stored row", len(stored))
	}
	for _, tx := range stored {
		if tx.Type == TxSplit {
			t.Fatal("a synthetic split reached the persistence path")
		}
	}
}

func TestSplitFactorsByDayCompoundsSameDayFactors(t *testing.T) {
	splits := []marketdata.Split{
		{Date: time.Date(2020, time.August, 31, 9, 30, 0, 0, time.UTC), Factor: 2},
		{Date: time.Date(2020, time.August, 31, 16, 0, 0, 0, time.UTC), Factor: 3},
		{Date: day(2021, time.March, 1), Factor: 1},
	}

	byDay := splitFactorsByDay(splits)

	if got := byDay[day(2020, time.August, 31)]; got != 6 {
		t.Fatalf("factor = %v, want 6 (2 then 3 compound)", got)
	}
	if _, ok := byDay[day(2021, time.March, 1)]; ok {
		t.Fatal("a neutral factor must not be indexed as a split")
	}
}

func TestCollectInstrumentsMergesPartialIdentities(t *testing.T) {
	symbol := "AAPL"
	isin := "US0378331005"
	transactions := []Transaction{
		{OccurredAt: day(2020, time.August, 1), Type: TxBuy, Symbol: &symbol},
		{OccurredAt: day(2020, time.August, 2), Type: TxBuy, Symbol: &symbol, ISIN: &isin},
		{OccurredAt: day(2020, time.August, 3), Type: TxCash},
	}

	found := collectInstruments(transactions)

	// One instrument, not two: the symbol-only entry is promoted onto the ISIN key once
	// the ISIN appears, exactly as position cycles are promoted. Two entries here would
	// mean the split gets injected twice and the holding scaled by the factor squared.
	if len(found) != 1 {
		t.Fatalf("instruments = %v, want exactly 1", found)
	}
	inst, ok := found[isin]
	if !ok {
		t.Fatalf("instruments = %v, want the entry keyed by ISIN once it is known", found)
	}
	if inst.isin == nil || *inst.isin != isin {
		t.Fatal("the instrument should carry the ISIN")
	}
	if inst.symbol == nil || *inst.symbol != symbol {
		t.Fatal("the symbol from the earlier row should have been merged in")
	}
}
