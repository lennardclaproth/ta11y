package assets_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/money"
)

func TestNewPurchaseValidates(t *testing.T) {
	accID, assetID := uuid.New(), uuid.New()
	yesterday := time.Now().UTC().AddDate(0, 0, -1)

	tests := []struct {
		name      string
		date      time.Time
		quantity  float64
		unitPrice money.Price
		wantErr   error
	}{
		{name: "valid", date: yesterday, quantity: 0.5, unitPrice: money.Price(100_00)},
		{name: "zero quantity", date: yesterday, quantity: 0, unitPrice: money.Price(100_00), wantErr: assets.ErrPurchaseQuantityInvalid},
		{name: "negative quantity", date: yesterday, quantity: -1, unitPrice: money.Price(100_00), wantErr: assets.ErrPurchaseQuantityInvalid},
		{name: "negative unit price", date: yesterday, quantity: 1, unitPrice: money.Price(-1), wantErr: assets.ErrPurchaseUnitPriceInvalid},
		{name: "future date", date: time.Now().UTC().AddDate(0, 0, 1), quantity: 1, unitPrice: money.Price(100_00), wantErr: assets.ErrPurchaseDateInFuture},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			purchase, err := assets.NewPurchase(accID, assetID, test.date, test.quantity, test.unitPrice)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("expected %v, got %v", test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !purchase.PurchasedOn.Equal(purchase.PurchasedOn.Truncate(24 * time.Hour)) {
				t.Errorf("expected the date to be normalised to a UTC day, got %s", purchase.PurchasedOn)
			}
		})
	}
}

// TestPurchasePaid covers the amount every "paid against worth" figure is built
// from.
func TestPurchasePaid(t *testing.T) {
	purchase := assets.Purchase{Quantity: 0.25, UnitPrice: money.Price(40_00)}
	if got := purchase.Paid(); got != money.Price(10_00) {
		t.Errorf("expected 10.00 paid, got %s", got)
	}
}

// TestIsDailyPriced separates the two kinds of item the assets feature now has.
func TestIsDailyPriced(t *testing.T) {
	listingID := uuid.New()
	if (&assets.Asset{}).IsDailyPriced() {
		t.Error("an item without a listing should be manual")
	}
	if !(&assets.Asset{ListingID: &listingID}).IsDailyPriced() {
		t.Error("an item with a listing should be daily priced")
	}
}
