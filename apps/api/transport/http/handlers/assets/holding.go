package assets

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/money"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// PurchaseRequest is one recorded acquisition as the client sends it. Quantity is
// a decimal string like every money field, because a crypto amount has more
// significant digits than a JSON number round-trips comfortably.
type PurchaseRequest struct {
	Date      string `json:"date"`
	Quantity  string `json:"quantity"`
	UnitPrice string `json:"unit_price"`
}

func (r PurchaseRequest) validate(problems map[string]string, prefix string) {
	if _, err := time.Parse(time.DateOnly, r.Date); err != nil {
		problems[prefix+"date"] = "date must be in YYYY-MM-DD format"
	}
	quantity, err := strconv.ParseFloat(strings.TrimSpace(r.Quantity), 64)
	if err != nil {
		problems[prefix+"quantity"] = "quantity must be a valid decimal string"
	} else if quantity <= 0 {
		problems[prefix+"quantity"] = "quantity must be greater than zero"
	}
	unitPrice, err := money.ParsePrice(r.UnitPrice)
	if err != nil {
		problems[prefix+"unit_price"] = "unit_price must be a valid decimal string"
	} else if unitPrice < 0 {
		problems[prefix+"unit_price"] = "unit_price cannot be negative"
	}
}

// toInput converts an already validated request; the parses cannot fail here.
func (r PurchaseRequest) toInput() assets.PurchaseInput {
	day, _ := time.Parse(time.DateOnly, r.Date)
	quantity, _ := strconv.ParseFloat(strings.TrimSpace(r.Quantity), 64)
	unitPrice, _ := money.ParsePrice(r.UnitPrice)
	return assets.PurchaseInput{Date: day, Quantity: quantity, UnitPrice: unitPrice}
}

// CreateHoldingRequest adds an item whose worth follows a daily price, with the
// purchase that started it.
type CreateHoldingRequest struct {
	ClassID  uuid.UUID       `json:"class_id"`
	Name     string          `json:"name"`
	Symbol   string          `json:"symbol"`
	Purchase PurchaseRequest `json:"purchase"`
}

func (r CreateHoldingRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.ClassID == uuid.Nil {
		problems["class_id"] = "class_id is required"
	}
	name := strings.TrimSpace(r.Name)
	if name == "" {
		problems["name"] = "name is required"
	} else if len(name) > 255 {
		problems["name"] = "name cannot exceed 255 characters"
	}
	if strings.TrimSpace(r.Symbol) == "" {
		problems["symbol"] = "symbol is required"
	}
	r.Purchase.validate(problems, "purchase.")
	return len(problems) == 0, problems
}

// CreateHoldingResponse confirms the created item.
type CreateHoldingResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// PurchaseResponse is one recorded acquisition.
type PurchaseResponse struct {
	ID        uuid.UUID `json:"id"`
	Date      string    `json:"date"`
	Quantity  string    `json:"quantity"`
	UnitPrice string    `json:"unit_price"`
	Paid      string    `json:"paid"`
}

// HoldingValuePointResponse is one day of a holding's worth against what had been
// paid for it by then.
type HoldingValuePointResponse struct {
	Date  string `json:"date"`
	Value string `json:"value"`
	Paid  string `json:"paid"`
}

// HoldingResponse is a daily-priced item as the class views read it.
type HoldingResponse struct {
	ID        uuid.UUID `json:"id"`
	ClassID   uuid.UUID `json:"class_id"`
	ClassName string    `json:"class_name,omitempty"`
	Name      string    `json:"name"`
	// Symbol and instrument_name are empty when the listing behind the item has
	// gone missing; the item still reports the worth it last derived.
	Symbol         string `json:"symbol"`
	InstrumentName string `json:"instrument_name,omitempty"`
	Quantity       string `json:"quantity"`
	Paid           string `json:"paid"`
	AvgUnitPrice   string `json:"avg_unit_price"`
	Price          string `json:"price"`
	// PriceDate is absent while no price is known yet. PriceCarriedForward says the
	// known price is older than today, so the valuation reuses it.
	PriceDate           *string  `json:"price_date,omitempty"`
	PriceCarriedForward bool     `json:"price_carried_forward"`
	Value               string   `json:"value"`
	Unrealized          string   `json:"unrealized"`
	UnrealizedPct       *float64 `json:"unrealized_pct,omitempty"`

	Series    []HoldingValuePointResponse `json:"series"`
	Purchases []PurchaseResponse          `json:"purchases,omitempty"`
}

// CreateHolding adds a daily-priced item to a manual class.
//
// @Summary Create daily-priced asset item
// @Description Adds an asset item linked to a daily-priced instrument together with its first purchase. The item's worth is derived from its purchases and the instrument's daily price, so it carries no worth of its own and cannot be set or adjusted by hand.
// @Tags assets
// @Accept json
// @Produce json
// @Param request body CreateHoldingRequest true "Create holding payload"
// @Success 201 {object} CreateHoldingResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/holdings [post]
func CreateHolding(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		req, err := httpx.JSONDecode[CreateHoldingRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		isValid, problems := req.isValid()
		if !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		asset, err := commands.CreateDailyPricedAsset(
			r.Context(),
			accountID,
			req.ClassID,
			strings.TrimSpace(req.Name),
			strings.TrimSpace(req.Symbol),
			req.Purchase.toInput(),
		)
		switch {
		case errors.Is(err, assets.ErrClassNotFound) || errors.Is(err, assets.ErrClassAccountMismatch):
			_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"error": "asset class not found"})
			return
		case errors.Is(err, marketdata.ErrQuoteNotFound):
			_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"symbol": "no daily-priced instrument with that symbol"})
			return
		case errors.Is(err, marketdata.ErrQuoteNotSelectable):
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"symbol": "instrument cannot be tracked with a daily price"})
			return
		case errors.Is(err, assets.ErrClassNotManual) || errors.Is(err, assets.ErrClassReserved):
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "asset class cannot hold a daily-priced item"})
			return
		case errors.Is(err, assets.ErrPurchaseQuantityInvalid) ||
			errors.Is(err, assets.ErrPurchaseUnitPriceInvalid) ||
			errors.Is(err, assets.ErrPurchaseDateInFuture):
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"purchase": err.Error()})
			return
		case err != nil:
			log.Error(r.Context(), "create holding: failed to create daily-priced asset", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create asset item"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusCreated, CreateHoldingResponse{ID: asset.ID, Name: asset.Name})
	})
}

// AddPurchase records another acquisition of a daily-priced item.
//
// @Summary Add purchase to daily-priced item
// @Description Records another acquisition of an item linked to a daily price. The item's whole worth history is rebuilt from its purchases afterwards.
// @Tags assets
// @Accept json
// @Produce json
// @Param asset_id path string true "Asset item ID"
// @Param request body PurchaseRequest true "Purchase payload"
// @Success 201 {object} PurchaseResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/{asset_id}/purchases [post]
func AddPurchase(log logging.Logger, commands assets.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		assetID, err := uuid.Parse(r.PathValue("asset_id"))
		if err != nil || assetID == uuid.Nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"asset_id": "asset_id must be a valid UUID"})
			return
		}
		req, err := httpx.JSONDecode[PurchaseRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
			return
		}
		problems := make(map[string]string)
		req.validate(problems, "")
		if len(problems) > 0 {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		purchase, err := commands.AddPurchase(r.Context(), accountID, assetID, req.toInput())
		switch {
		// The same not-found answer covers an unknown item and someone else's item,
		// so a request cannot probe which asset IDs exist.
		case errors.Is(err, assets.ErrAssetNotFound) || errors.Is(err, assets.ErrClassAccountMismatch):
			_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"error": "asset not found"})
			return
		case errors.Is(err, assets.ErrAssetNotDailyPriced):
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "asset is not linked to a daily price"})
			return
		case errors.Is(err, assets.ErrPurchaseQuantityInvalid) ||
			errors.Is(err, assets.ErrPurchaseUnitPriceInvalid) ||
			errors.Is(err, assets.ErrPurchaseDateInFuture):
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"purchase": err.Error()})
			return
		case err != nil:
			log.Error(r.Context(), "add purchase: failed to record purchase", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to record purchase"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusCreated, toPurchaseResponse(*purchase))
	})
}

// GetHolding returns one daily-priced item with its purchases.
//
// @Summary Get daily-priced asset item
// @Description Returns a daily-priced item with its quantity, what was paid, the latest known price, the value-against-paid series since the first purchase, and the purchases behind it.
// @Tags assets
// @Accept json
// @Produce json
// @Param asset_id path string true "Asset item ID"
// @Success 200 {object} HoldingResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/holdings/{asset_id} [get]
func GetHolding(log logging.Logger, queries assets.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		assetID, err := uuid.Parse(r.PathValue("asset_id"))
		if err != nil || assetID == uuid.Nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"asset_id": "asset_id must be a valid UUID"})
			return
		}

		details, err := queries.HoldingDetails(r.Context(), assetID, accountID)
		switch {
		case errors.Is(err, assets.ErrAssetNotFound) ||
			errors.Is(err, assets.ErrClassAccountMismatch) ||
			errors.Is(err, assets.ErrAssetNotDailyPriced):
			_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"error": "asset not found"})
			return
		case err != nil:
			log.Error(r.Context(), "get holding: failed to get holding details", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to get asset item"})
			return
		}

		res := toHoldingResponse(details.HoldingSummary)
		res.ClassName = details.ClassName
		res.Purchases = make([]PurchaseResponse, 0, len(details.Purchases))
		for _, purchase := range details.Purchases {
			res.Purchases = append(res.Purchases, toPurchaseResponse(purchase))
		}
		_ = httpx.JSONEncode(w, http.StatusOK, res)
	})
}

func toHoldingResponse(summary assets.HoldingSummary) HoldingResponse {
	res := HoldingResponse{
		ID:                  summary.AssetID,
		ClassID:             summary.ClassID,
		Name:                summary.Name,
		Symbol:              summary.Symbol,
		InstrumentName:      summary.InstrumentName,
		Quantity:            formatQuantity(summary.Quantity),
		Paid:                summary.Paid.String(),
		AvgUnitPrice:        summary.AvgUnitPrice.String(),
		Price:               summary.Price.String(),
		PriceCarriedForward: summary.PriceCarriedForward,
		Value:               summary.Value.String(),
		Unrealized:          summary.Unrealized.String(),
		UnrealizedPct:       summary.UnrealizedPct,
		Series:              make([]HoldingValuePointResponse, 0, len(summary.Series)),
	}
	if summary.PriceDate != nil {
		day := summary.PriceDate.Format(time.DateOnly)
		res.PriceDate = &day
	}
	for _, point := range summary.Series {
		res.Series = append(res.Series, HoldingValuePointResponse{
			Date:  point.Date.Format(time.DateOnly),
			Value: point.Value.String(),
			Paid:  point.Paid.String(),
		})
	}
	return res
}

func toPurchaseResponse(purchase assets.Purchase) PurchaseResponse {
	return PurchaseResponse{
		ID:        purchase.ID,
		Date:      purchase.PurchasedOn.Format(time.DateOnly),
		Quantity:  formatQuantity(purchase.Quantity),
		UnitPrice: purchase.UnitPrice.String(),
		Paid:      purchase.Paid().String(),
	}
}

// formatQuantity renders a held amount without an exponent and without trailing
// zeros, so a fraction of a bitcoin survives the round trip and a whole unit does
// not read as "1.00000000".
func formatQuantity(quantity float64) string {
	return strconv.FormatFloat(quantity, 'f', -1, 64)
}
