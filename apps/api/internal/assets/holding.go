package assets

import (
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// Purchase is one recorded acquisition of a daily-priced item: how much was bought
// on which day, and what was paid per unit.
//
// There is no sell counterpart. A holding only grows, which is why the quantity
// held on any day is simply the sum of the purchases up to that day.
type Purchase struct {
	ID          uuid.UUID   `db:"id"`
	AccountID   uuid.UUID   `db:"account_id"`
	AssetID     uuid.UUID   `db:"item_id"`
	PurchasedOn time.Time   `db:"purchased_on"`
	Quantity    float64     `db:"quantity"`
	UnitPrice   money.Price `db:"unit_price"`
	CreatedAt   time.Time   `db:"created_at"`
}

// Paid is what this purchase cost: quantity times the price paid per unit.
func (p Purchase) Paid() money.Price {
	return money.Price(float64(p.UnitPrice) * p.Quantity)
}

// NewPurchase validates and constructs a purchase. The date is normalised to a UTC
// day because prices are daily: a purchase belongs to a day, not a moment.
func NewPurchase(
	accID, assetID uuid.UUID,
	purchasedOn time.Time,
	quantity float64,
	unitPrice money.Price,
) (*Purchase, error) {
	if quantity <= 0 {
		return nil, ErrPurchaseQuantityInvalid
	}
	if unitPrice < 0 {
		return nil, ErrPurchaseUnitPriceInvalid
	}
	day := date.StartOfDayUTC(purchasedOn)
	if day.After(date.StartOfDayUTC(time.Now().UTC())) {
		return nil, ErrPurchaseDateInFuture
	}
	return &Purchase{
		ID:          uuid.New(),
		AccountID:   accID,
		AssetID:     assetID,
		PurchasedOn: day,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// IsDailyPriced reports whether this item's worth follows a listing's daily price
// rather than a worth the user sets by hand.
func (a *Asset) IsDailyPriced() bool {
	return a != nil && a.ListingID != nil
}

// HoldingWorthPoint is one day of an item's derived worth, read back from the
// mutations the rebuild wrote.
type HoldingWorthPoint struct {
	Date  time.Time
	Worth money.Price
}

// HoldingValuePoint is one day of a holding's history: what it was worth against
// what had been paid for it by then.
type HoldingValuePoint struct {
	Date  time.Time
	Value money.Price
	Paid  money.Price
}

// HoldingSummary is a daily-priced item as the class views present it: how much is
// held, what it cost, what it is worth now, and the series behind that.
type HoldingSummary struct {
	AssetID   uuid.UUID
	ClassID   uuid.UUID
	Name      string
	ListingID uuid.UUID
	// Symbol and InstrumentName come from the listing; both are empty when the
	// listing has gone missing, which the read side reports rather than hides.
	Symbol         string
	InstrumentName string
	Quantity       float64
	Paid           money.Price
	AvgUnitPrice   money.Price
	// Price is the most recent known daily price, and PriceDate the day it belongs
	// to. PriceCarriedForward says that day is in the past, so the valuation reuses
	// the last known price -- weekends, holidays and a provider that is behind.
	Price               money.Price
	PriceDate           *time.Time
	PriceCarriedForward bool
	Value               money.Price
	Unrealized          money.Price
	// UnrealizedPct is nil when nothing has been paid yet, where a percentage has
	// no meaning.
	UnrealizedPct *float64
	Series        []HoldingValuePoint
}

// HoldingDetails is a holding with the purchases that produced it.
type HoldingDetails struct {
	HoldingSummary
	ClassName string
	Purchases []Purchase
}
