package assets_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/money"
)

func TestNewMutationAppliesTheChangeToThePreviousWorth(t *testing.T) {
	increase := assets.ChangeDirectionIncrease
	decrease := assets.ChangeDirectionDecrease
	cases := []struct {
		name       string
		changeType assets.ChangeType
		direction  *assets.ChangeDirection
		amount     money.Price
		wantWorth  money.Price
	}{
		{"set replaces the worth", assets.ChangeTypeSet, nil, eur(1_200), eur(1_200)},
		{"increase adds to the worth", assets.ChangeTypeAdjust, &increase, eur(300), eur(1_300)},
		{"decrease takes from the worth", assets.ChangeTypeAdjust, &decrease, eur(300), eur(700)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mutation, err := assets.NewMutation(
				uuid.New(), uuid.New(), uuid.New(),
				tc.changeType, tc.direction,
				tc.amount, eur(1_000), 0,
				time.Now(), nil,
			)
			if err != nil {
				t.Fatalf("new mutation: %v", err)
			}
			if mutation.NewWorth != tc.wantWorth {
				t.Errorf("new worth: want %v, got %v", tc.wantWorth, mutation.NewWorth)
			}
			if mutation.Amount != tc.amount {
				t.Errorf("amount: want the change itself (%v), got %v", tc.amount, mutation.Amount)
			}
		})
	}
}

func eur(units int64) money.Price {
	return money.Price(units * money.Scale)
}
