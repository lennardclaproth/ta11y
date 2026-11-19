package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// CatalogueSearchRequest asks the provider for entries matching a query and caches
// what comes back.
type CatalogueSearchRequest struct {
	Source string `json:"source"`
	Q      string `json:"q"`
	Limit  int    `json:"limit,omitempty"`
}

func (r CatalogueSearchRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if strings.TrimSpace(r.Q) == "" {
		problems["q"] = "q is required"
	}
	if !marketdata.Source(strings.TrimSpace(r.Source)).IsValid() {
		problems["source"] = "source is invalid"
	}
	if r.Limit < 0 || r.Limit > 100 {
		problems["limit"] = "limit must be between 0 and 100"
	}
	return len(problems) == 0, problems
}

// CatalogueSearchResponse returns the locally visible results after a provider
// search, together with what the provider reported.
type CatalogueSearchResponse struct {
	Pagination PaginationResponse `json:"pagination"`
	// UpstreamTotal is how many entries the provider matched in total, which is
	// routinely far more than one page.
	UpstreamTotal int `json:"upstream_total"`
	// Truncated is true when the provider matched more than the single page that
	// was fetched. Clients must say so rather than presenting the page as the
	// complete answer.
	Truncated bool `json:"truncated"`
	// Cached is how many catalogue entries this search wrote locally.
	Cached int                `json:"cached"`
	Data   []ListingSearchRow `json:"data"`
}

// CatalogueSyncRequest starts a bounded catalogue seed run.
type CatalogueSyncRequest struct {
	Source string `json:"source"`
	// Pages bounds the run. Omitted means marketdata.DefaultSeedPages.
	Pages int `json:"pages,omitempty"`
}

func (r CatalogueSyncRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if !marketdata.Source(strings.TrimSpace(r.Source)).IsValid() {
		problems["source"] = "source is invalid"
	}
	if r.Pages < 0 || r.Pages > marketdata.MaxSeedPages {
		problems["pages"] = "pages must be between 0 and 100"
	}
	return len(problems) == 0, problems
}

// CatalogueSyncResponse describes one catalogue seed run.
type CatalogueSyncResponse struct {
	ID            uuid.UUID  `json:"id"`
	Source        string     `json:"source"`
	Status        string     `json:"status"`
	PagesFetched  int        `json:"pages_fetched"`
	RowsUpserted  int        `json:"rows_upserted"`
	UpstreamTotal *int       `json:"upstream_total,omitempty"`
	LastError     *string    `json:"last_error,omitempty"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
}

// CatalogueStatusResponse summarises the locally cached catalogue for one source.
type CatalogueStatusResponse struct {
	Source string `json:"source"`
	// Entries is how many provider entries are cached locally and therefore
	// searchable without spending a provider request.
	Entries    int                    `json:"entries"`
	LatestSync *CatalogueSyncResponse `json:"latest_sync,omitempty"`
}

// SearchProviderCatalogue searches the provider directly and caches the results.
//
// It costs one provider request per call, which is why it is an explicit action
// rather than part of listing search: the local catalogue can never be known to be
// complete for a query it has not seen, so widening a search has to be deliberate.
// Only one page is fetched -- provider results are relevance-ranked, so the
// intended instrument is on the first page.
//
// @Summary Search the provider catalogue
// @Description Run one metered provider ticker search, cache the results locally, and return the matching rows. Costs one provider request.
// @Tags listings
// @Accept json
// @Produce json
// @Param request body CatalogueSearchRequest true "Catalogue search payload"
// @Success 200 {object} CatalogueSearchResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/catalogue/search [post]
func SearchProviderCatalogue(
	log logging.Logger,
	catalogue *marketdata.Catalogue,
	queries *marketdata.Queries,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[CatalogueSearchRequest](r)
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

		query := strings.TrimSpace(req.Q)
		source := marketdata.Source(strings.TrimSpace(req.Source))

		refreshed, err := catalogue.Refresh(r.Context(), source, query)
		if err != nil {
			writeCatalogueError(w, log, r, "search provider catalogue", err)
			return
		}

		limit := req.Limit
		if limit == 0 {
			limit = 25
		}
		// Read back through the normal combined search so a provider hit that is
		// already tracked is presented as tracked rather than as a new entry.
		results, total, err := queries.SearchCatalogue(r.Context(), query, marketdata.ScopeAll, limit, 0)
		if err != nil {
			log.Error(r.Context(), "search provider catalogue: failed to read back results", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to search the provider catalogue"})
			return
		}

		data := toListingSearchRows(results)
		_ = httpx.JSONEncode(w, http.StatusOK, CatalogueSearchResponse{
			Pagination: PaginationResponse{
				Limit:  limit,
				Offset: 0,
				Count:  len(data),
				Total:  total,
			},
			UpstreamTotal: refreshed.UpstreamTotal,
			Truncated:     refreshed.Truncated,
			Cached:        refreshed.Upserted,
			Data:          data,
		})
	})
}

// StartCatalogueSync starts a bounded catalogue seed run in the background.
//
// The provider returns its catalogue in popularity order, so a small page budget
// caches the instruments most portfolios actually hold. The run is capped rather
// than mirroring the provider's full universe, which would cost thousands of
// requests and hours of wall time.
//
// @Summary Seed the provider catalogue
// @Description Start a bounded background sync that caches the provider's most-traded entries. Costs one provider request per page.
// @Tags listings
// @Accept json
// @Produce json
// @Param request body CatalogueSyncRequest true "Catalogue sync payload"
// @Success 202 {object} CatalogueSyncResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/catalogue/sync [post]
func StartCatalogueSync(
	log logging.Logger,
	catalogue *marketdata.Catalogue,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.JSONDecode[CatalogueSyncRequest](r)
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

		run, err := catalogue.StartSeed(
			r.Context(),
			marketdata.Source(strings.TrimSpace(req.Source)),
			req.Pages,
		)
		if err != nil {
			writeCatalogueError(w, log, r, "start catalogue sync", err)
			return
		}

		_ = httpx.JSONEncode(w, http.StatusAccepted, toCatalogueSyncResponse(run))
	})
}

// GetCatalogueStatusRequest selects which source to report on.
type GetCatalogueStatusRequest struct {
	Source string `query:"source"`
}

func (r GetCatalogueStatusRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if !marketdata.Source(strings.TrimSpace(r.Source)).IsValid() {
		problems["source"] = "source is invalid"
	}
	return len(problems) == 0, problems
}

// GetCatalogueStatus reports how much of a provider's catalogue is cached locally
// and when it was last seeded, so clients can warn that the cache has gone stale.
//
// @Summary Provider catalogue status
// @Description Return the cached entry count and the most recent seed run for a source.
// @Tags listings
// @Accept json
// @Produce json
// @Param source query string true "Listing source"
// @Success 200 {object} CatalogueStatusResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/catalogue/status [get]
func GetCatalogueStatus(
	log logging.Logger,
	queries *marketdata.Queries,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := httpx.DecodeQuery[GetCatalogueStatusRequest](r)
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

		status, err := queries.CatalogueStatus(r.Context(), marketdata.Source(strings.TrimSpace(req.Source)))
		if err != nil {
			log.Error(r.Context(), "catalogue status: failed to read status", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to read the catalogue status"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, CatalogueStatusResponse{
			Source:     string(status.Source),
			Entries:    status.Entries,
			LatestSync: toCatalogueSyncResponse(status.LatestSync),
		})
	})
}

// writeCatalogueError maps the catalogue's known failures to status codes and logs
// anything it does not recognise.
func writeCatalogueError(w http.ResponseWriter, log logging.Logger, r *http.Request, operation string, err error) {
	switch {
	case errors.Is(err, marketdata.ErrCatalogueSyncInProgress):
		_ = httpx.JSONEncode(w, http.StatusConflict, map[string]string{"catalogue": "a catalogue sync is already running"})
	case errors.Is(err, marketdata.ErrCatalogueSourceUnsupported):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"source": "this source does not support catalogue search"})
	case errors.Is(err, marketdata.ErrCatalogueQueryEmpty):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"q": "q is required"})
	case errors.Is(err, marketdata.ErrCatalogueSeedPagesInvalid):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"pages": err.Error()})
	default:
		log.Error(r.Context(), operation+": failed", err)
		_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to reach the market data provider"})
	}
}

func toCatalogueSyncResponse(run *marketdata.CatalogueSync) *CatalogueSyncResponse {
	if run == nil {
		return nil
	}
	return &CatalogueSyncResponse{
		ID:            run.ID,
		Source:        string(run.Source),
		Status:        string(run.Status),
		PagesFetched:  run.PagesFetched,
		RowsUpserted:  run.RowsUpserted,
		UpstreamTotal: run.UpstreamTotal,
		LastError:     run.LastError,
		StartedAt:     run.StartedAt,
		FinishedAt:    run.FinishedAt,
	}
}
