package assets

import (
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/date"
	"github.com/lennardclaproth/ta11y/internal/money"
)

func growthPctFromBouds(bound ClassBounds) float64 {
	if bound.First != nil && bound.Last != nil {
		if pct := growthPctFromInception(bound.First.ClassTotalWorth, bound.Last.ClassTotalWorth); pct != nil {
			return *pct
		}
	}
	return 0.0
}

func growthPctFromInception(inceptionWorth, latestWorth money.Price) *float64 {
	inception := inceptionWorth.Float64()
	if inception == 0 {
		return nil
	}
	latest := latestWorth.Float64()
	value := ((latest - inception) / math.Abs(inception)) * 100
	return &value
}

// worthSeriesByAsset collapses mutations into one worth per item per day, oldest
// first. A day with several mutations keeps the last one, which is the worth the
// item ended that day on.
func worthSeriesByAsset(mutations []Mutation) map[uuid.UUID][]HoldingWorthPoint {
	type dayEntry struct {
		effectiveDate time.Time
		worth         money.Price
	}
	byAsset := make(map[uuid.UUID]map[string]dayEntry)

	for _, mutation := range mutations {
		days := byAsset[mutation.AssetID]
		if days == nil {
			days = make(map[string]dayEntry)
			byAsset[mutation.AssetID] = days
		}
		key := date.StartOfDayUTC(mutation.EffectiveDate).Format(time.DateOnly)
		if existing, ok := days[key]; ok && !mutation.EffectiveDate.After(existing.effectiveDate) {
			continue
		}
		days[key] = dayEntry{effectiveDate: mutation.EffectiveDate, worth: mutation.NewWorth}
	}

	out := make(map[uuid.UUID][]HoldingWorthPoint, len(byAsset))
	for assetID, days := range byAsset {
		keys := make([]string, 0, len(days))
		for key := range days {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		points := make([]HoldingWorthPoint, 0, len(keys))
		for _, key := range keys {
			day, err := time.ParseInLocation(time.DateOnly, key, time.UTC)
			if err != nil {
				continue
			}
			points = append(points, HoldingWorthPoint{Date: day, Worth: days[key].worth})
		}
		out[assetID] = points
	}
	return out
}

// buildHoldingSeries pairs each day of derived worth with what had been paid by
// then, which is the "value against paid" the holding is read by.
func buildHoldingSeries(worth []HoldingWorthPoint, purchases []*Purchase) []HoldingValuePoint {
	points := make([]HoldingValuePoint, 0, len(worth))
	for _, point := range worth {
		paid := money.Price(0)
		for _, purchase := range purchases {
			if purchase == nil || purchase.PurchasedOn.After(point.Date) {
				continue
			}
			paid += purchase.Paid()
		}
		points = append(points, HoldingValuePoint{Date: point.Date, Value: point.Worth, Paid: paid})
	}
	return points
}

func growthPointsFromMutations(mutations []Mutation) []GrowthPoint {
	type dailyPoint struct {
		date          time.Time
		effectiveDate time.Time
		totalWorth    money.Price
	}

	byDate := make(map[string]dailyPoint, len(mutations))

	for _, mutation := range mutations {
		date := date.StartOfDayUTC(mutation.EffectiveDate)
		key := date.Format(time.DateOnly)

		existing, exists := byDate[key]
		if exists && !mutation.EffectiveDate.After(existing.effectiveDate) {
			continue
		}

		byDate[key] = dailyPoint{
			date:          date,
			effectiveDate: mutation.EffectiveDate,
			totalWorth:    mutation.ClassTotalWorth,
		}
	}

	keys := make([]string, 0, len(byDate))
	for key := range byDate {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	points := make([]GrowthPoint, 0, len(keys))
	for _, key := range keys {
		point := byDate[key]

		points = append(points, GrowthPoint{
			Date:       point.date,
			TotalWorth: point.totalWorth,
		})
	}

	return points
}
