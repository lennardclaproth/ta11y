package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/money"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// CreateListingRequest creates a market listing.
type CreateListingRequest struct {
	Name        string  `json:"name"`
	Symbol      string  `json:"symbol"`
	Source      string  `json:"source"`
	Description *string `json:"description,omitempty"`
	Exchange    *string `json:"exchange,omitempty"`
	Region      *string `json:"region,omitempty"`
	Currency    *string `json:"currency,omitempty"`
	ISIN        *string `json:"isin,omitempty"`
	Ticker      *string `json:"ticker,omitempty"`
	Type        *string `json:"type,omitempty"`
	// SyncPrices controls whether creation immediately backfills price history.
	// It defaults to true so existing clients keep the original behaviour. The
	// catalogue drawer sends false when adopting a batch: backfilling is a
	// synchronous paged provider fetch, so one full history sync per adopted
	// listing would stall the request and drain the provider request budget.
	SyncPrices *bool `json:"sync_prices,omitempty"`
}

// priceSync resolves the optional sync_prices flag, defaulting to an immediate
// backfill when the client says nothing.
func (r CreateListingRequest) priceSync() marketdata.PriceSync {
	if r.SyncPrices != nil && !*r.SyncPrices {
		return marketdata.DeferPriceSync
	}
	return marketdata.SyncPricesNow
}

func (r CreateListingRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Name) == "" {
		problems["name"] = "name is required"
	}
	if strings.TrimSpace(r.Symbol) == "" {
		problems["symbol"] = "symbol is required"
	}
	if strings.TrimSpace(r.Source) == "" {
		problems["source"] = "source is required"
	}
	if r.Currency != nil {
		currency := money.Currency(strings.TrimSpace(*r.Currency))
		if !currency.IsValid() {
			problems["currency"] = "currency is invalid"
		}
	}
	return len(problems) == 0, problems
}

// UpdateListingFieldsRequest updates mutable listing metadata fields.
type UpdateListingFieldsRequest struct {
	ID          uuid.UUID `json:"id"`
	Description *string   `json:"description,omitempty"`
	Exchange    *string   `json:"exchange,omitempty"`
	Region      *string   `json:"region,omitempty"`
	Currency    *string   `json:"currency,omitempty"`
	ISIN        *string   `json:"isin,omitempty"`
	Ticker      *string   `json:"ticker,omitempty"`
	Type        *string   `json:"type,omitempty"`
}

func (r UpdateListingFieldsRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.ID == uuid.Nil {
		problems["id"] = "id is required"
	}
	if r.Description == nil &&
		r.Exchange == nil &&
		r.Region == nil &&
		r.Currency == nil &&
		r.ISIN == nil &&
		r.Ticker == nil &&
		r.Type == nil {
		problems["listing"] = marketdata.ErrNoListingFieldsToUpdate.Error()
	}
	if r.Currency != nil {
		currency := money.Currency(strings.TrimSpace(*r.Currency))
		if !currency.IsValid() {
			problems["currency"] = "currency is invalid"
		}
	}
	return len(problems) == 0, problems
}

// SearchListingsRequest contains search and pagination inputs for listings.
type SearchListingsRequest struct {
	Q      string `query:"q"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
	// Scope selects which sides of the catalogue to search: "tracked" (default),
	// "catalogue" for cached provider entries not yet tracked, or "all" for both.
	// The default keeps clients that predate the catalogue on their original
	// result set.
	Scope string `query:"scope"`
}

// scope resolves the optional scope parameter to its default.
func (r SearchListingsRequest) scope() marketdata.CatalogueScope {
	if strings.TrimSpace(r.Scope) == "" {
		return marketdata.ScopeTracked
	}
	return marketdata.CatalogueScope(strings.TrimSpace(r.Scope))
}

func (r SearchListingsRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Q) == "" {
		problems["q"] = "q is required"
	}
	if r.Limit < 0 {
		problems["limit"] = "limit must be greater than or equal to 0"
	}
	if r.Limit > 100 {
		problems["limit"] = "limit must be less than or equal to 100"
	}
	if r.Offset < 0 {
		problems["offset"] = "offset must be greater than or equal to 0"
	}
	if !r.scope().IsValid() {
		problems["scope"] = "scope must be one of tracked, catalogue or all"
	}
	return len(problems) == 0, problems
}

// ListingResponse represents one market listing record.
type ListingResponse struct {
	ID          uuid.UUID `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Source      string    `json:"source"`
	Description *string   `json:"description,omitempty"`
	Exchange    *string   `json:"exchange,omitempty"`
	Region      *string   `json:"region,omitempty"`
	Currency    *string   `json:"currency,omitempty"`
	ISIN        *string   `json:"isin,omitempty"`
	Ticker      *string   `json:"ticker,omitempty"`
	Type        *string   `json:"type,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PaginationResponse describes offset/limit paging metadata.
type PaginationResponse struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Total  int `json:"total"`
}

// ListingSearchRow is one listing search result: an instrument the caller tracks,
// or a cached provider-catalogue entry that could become one.
//
// For a tracked row every field carries the same value listing search has always
// returned. Catalogue rows omit what provider ticker search cannot supply rather
// than sending zero values, so an untracked row never implies metadata we do not
// have.
type ListingSearchRow struct {
	ID          uuid.UUID `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        *string   `json:"name"`
	Source      string    `json:"source"`
	Description *string   `json:"description,omitempty"`
	Exchange    *string   `json:"exchange,omitempty"`
	ExchangeMIC *string   `json:"exchange_mic,omitempty"`
	Region      *string   `json:"region,omitempty"`
	Currency    *string   `json:"currency,omitempty"`
	ISIN        *string   `json:"isin,omitempty"`
	Ticker      *string   `json:"ticker,omitempty"`
	Type        *string   `json:"type,omitempty"`
	// Tracked marks rows that are already listings rather than catalogue entries.
	Tracked bool `json:"tracked"`
	// HasEOD reports whether the provider holds end-of-day history for the symbol.
	HasEOD bool `json:"has_eod"`
	// Adoptable is false when a row cannot become a listing, with AdoptableReason
	// explaining why, so clients can disable the row instead of failing on submit.
	Adoptable       bool       `json:"adoptable"`
	AdoptableReason *string    `json:"adoptable_reason,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

// ListingsSearchResponse returns paginated listing search results.
type ListingsSearchResponse struct {
	Pagination PaginationResponse `json:"pagination"`
	Data       []ListingSearchRow `json:"data"`
}

// CreateListing creates a new market-data listing.
//
// @Summary Create listing
// @Description Create a listing and trigger async end-of-day data accumulation.
// @Tags listings
// @Accept json
// @Produce json
// @Param request body CreateListingRequest true "Listing payload"
// @Success 200 {object} ListingResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/listing [post]
func CreateListing(
	log logging.Logger,
	commands *marketdata.Commands,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[CreateListingRequest](r)
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

		listing, err := commands.CreateListing(
			r.Context(),
			strings.TrimSpace(req.Symbol),
			strings.TrimSpace(req.Name),
			marketdata.Source(strings.TrimSpace(req.Source)),
			req.priceSync(),
			listingOptions(req)...,
		)
		if err != nil {
			switch {
			case errors.Is(err, marketdata.ErrListingAlreadyExists):
				_ = httpx.JSONEncode(w, http.StatusConflict, map[string]string{"listing": "listing already exists"})
			case isListingValidationError(err):
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"listing": err.Error()})
			default:
				log.Error(r.Context(), "create listing: failed to create listing", err)
				_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to create listing"})
			}
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, toListingResponse(listing))
	})
}

// UpdateListingFields updates specific fields for an existing listing.
//
// @Summary Update listing fields
// @Description Patch selected listing fields; omitted fields remain unchanged.
// @Tags listings
// @Accept json
// @Produce json
// @Param request body UpdateListingFieldsRequest true "Listing patch payload"
// @Success 200 {object} ListingResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/listing [patch]
func UpdateListingFields(
	log logging.Logger,
	commands *marketdata.Commands,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[UpdateListingFieldsRequest](r)
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

		listing, err := commands.UpdateListingFields(
			r.Context(),
			req.ID,
			req.Description,
			req.Exchange,
			req.Region,
			req.Currency,
			req.ISIN,
			req.Ticker,
			req.Type,
		)
		if err != nil {
			switch {
			case errors.Is(err, marketdata.ErrListingNotFound):
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"listing": "listing not found"})
			case errors.Is(err, marketdata.ErrNoListingFieldsToUpdate) ||
				errors.Is(err, marketdata.ErrInvalidListingCurrency):
				_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"listing": err.Error()})
			default:
				log.Error(r.Context(), "update listing fields: failed to update listing", err)
				_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to update listing"})
			}
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, toListingResponse(listing))
	})
}

// GetListings returns all listings.
//
// @Summary List listings
// @Description Return all market-data listings ordered by symbol.
// @Tags listings
// @Accept json
// @Produce json
// @Success 200 {array} ListingResponse
// @Failure 500 {object} map[string]string
// @Router /marketdata/listings [get]
func GetListings(
	log logging.Logger,
	queries *marketdata.Queries,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		listings, err := queries.ListListings(r.Context())
		if err != nil {
			log.Error(r.Context(), "list listings: failed to list listings", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to list listings"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, toListingResponses(listings))
	})
}

// SearchListings searches tracked listings and the cached provider catalogue.
//
// @Summary Search listings
// @Description Search market-data listings using a case-insensitive partial query over symbol, name and isin. The optional scope parameter widens the search to cached provider-catalogue entries that are not tracked yet; tracked results always sort first. This never calls the provider.
// @Tags listings
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Page size (max 100, default 25)"
// @Param offset query int false "Offset"
// @Param scope query string false "tracked (default), catalogue or all"
// @Success 200 {object} ListingsSearchResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/listings/search [get]
func SearchListings(
	log logging.Logger,
	queries *marketdata.Queries,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[SearchListingsRequest](r)
		if err != nil {
			if httpx.WriteDecodeError(w, err) {
				return
			}
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
			return
		}
		isValid, problems := req.isValid()
		if !isValid {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, problems)
			return
		}

		limit := req.Limit
		if limit == 0 {
			limit = 25
		}

		results, total, err := queries.SearchCatalogue(r.Context(), strings.TrimSpace(req.Q), req.scope(), limit, req.Offset)
		if err != nil {
			log.Error(r.Context(), "search listings: failed to search listings", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to search listings"})
			return
		}

		data := toListingSearchRows(results)
		_ = httpx.JSONEncode(w, http.StatusOK, ListingsSearchResponse{
			Pagination: PaginationResponse{
				Limit:  limit,
				Offset: req.Offset,
				Count:  len(data),
				Total:  total,
			},
			Data: data,
		})
	})
}

func toListingSearchRows(results []*marketdata.CatalogueSearchResult) []ListingSearchRow {
	rows := make([]ListingSearchRow, 0, len(results))
	for _, result := range results {
		if result == nil {
			continue
		}
		rows = append(rows, toListingSearchRow(result))
	}
	return rows
}

func toListingSearchRow(result *marketdata.CatalogueSearchResult) ListingSearchRow {
	var currency *string
	if result.Currency != nil {
		value := string(*result.Currency)
		currency = &value
	}

	adoptable, reason := result.Adoptable()
	var adoptableReason *string
	if reason != "" {
		adoptableReason = &reason
	}

	return ListingSearchRow{
		ID:              result.ID,
		Symbol:          result.Symbol,
		Name:            result.Name,
		Source:          string(result.Source),
		Description:     result.Description,
		Exchange:        result.Exchange,
		ExchangeMIC:     result.ExchangeMIC,
		Region:          result.Region,
		Currency:        currency,
		ISIN:            result.ISIN,
		Ticker:          result.Ticker,
		Type:            result.Type,
		Tracked:         result.Tracked,
		HasEOD:          result.HasEOD,
		Adoptable:       adoptable,
		AdoptableReason: adoptableReason,
		CreatedAt:       result.CreatedAt,
		UpdatedAt:       result.UpdatedAt,
	}
}

func listingOptions(req CreateListingRequest) []marketdata.ListingOption {
	var options []marketdata.ListingOption
	if req.ISIN != nil {
		options = append(options, marketdata.ListingWithISIN(*req.ISIN))
	}
	if req.Exchange != nil {
		options = append(options, marketdata.ListingWithExchange(*req.Exchange))
	}
	if req.Currency != nil {
		options = append(options, marketdata.ListingWithCurrency(money.Currency(strings.TrimSpace(*req.Currency))))
	}
	if req.Description != nil {
		options = append(options, marketdata.ListingWithDescription(*req.Description))
	}
	if req.Region != nil {
		options = append(options, marketdata.ListingWithRegion(*req.Region))
	}
	if req.Type != nil {
		options = append(options, marketdata.ListingWithType(*req.Type))
	}
	if req.Ticker != nil {
		options = append(options, marketdata.ListingWithTicker(*req.Ticker))
	}
	return options
}

func isListingValidationError(err error) bool {
	return errors.Is(err, marketdata.ErrListingNameEmpty) ||
		errors.Is(err, marketdata.ErrListingSymbolEmpty) ||
		errors.Is(err, marketdata.ErrListingSourceEmpty) ||
		errors.Is(err, marketdata.ErrInvalidListingCurrency)
}

func toListingResponses(listings []*marketdata.Listing) []ListingResponse {
	response := make([]ListingResponse, 0, len(listings))
	for _, listing := range listings {
		if listing == nil {
			continue
		}
		response = append(response, toListingResponse(listing))
	}
	return response
}

func toListingResponse(listing *marketdata.Listing) ListingResponse {
	if listing == nil {
		return ListingResponse{}
	}

	var currency *string
	if listing.Currency != nil {
		s := string(*listing.Currency)
		currency = &s
	}

	return ListingResponse{
		ID:          listing.ID,
		Symbol:      listing.Symbol,
		Name:        listing.Name,
		Source:      string(listing.Source),
		Description: listing.Description,
		Exchange:    listing.Exchange,
		Region:      listing.Region,
		Currency:    currency,
		ISIN:        listing.ISIN,
		Ticker:      listing.Ticker,
		Type:        listing.Type,
		CreatedAt:   listing.CreatedAt,
		UpdatedAt:   listing.UpdatedAt,
	}
}
