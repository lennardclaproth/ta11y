package importer

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/importer"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// UnlinkedProductResponse is a product an import brought in that no listing matches.
type UnlinkedProductResponse struct {
	Name         string  `json:"name"`
	ISIN         *string `json:"isin"`
	Symbol       *string `json:"symbol"`
	Transactions int     `json:"transactions"`
}

// ImportResultResponse reports what one import did: its lifecycle state, the counters
// the processor produced, and the products it brought in that do not count towards
// performance yet.
//
// The import's stored StatusMsg is deliberately absent. It is a wrapped Go error that
// names upload paths, tables and driver text, and the classified Reason is what a client
// can act on. The full message is logged where the import fails.
type ImportResultResponse struct {
	ImportID   uuid.UUID `json:"import_id"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Reason     string    `json:"reason"`
	TotalRows  int       `json:"total_rows"`
	Imported   int       `json:"imported"`
	Duplicates int       `json:"duplicates"`
	Failed     int       `json:"failed"`
	// AutoIgnored is how many of the imported rows an ignore rule recognised. They
	// are counted in Imported too: they were imported, and then ignored.
	AutoIgnored int       `json:"auto_ignored"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// UnlinkedProducts is empty until a portfolio import has completed.
	UnlinkedProducts []UnlinkedProductResponse `json:"unlinked_products"`
}

// GetImport returns the outcome of one import for the signed-in account.
//
// @Summary     Get import result
// @Description Returns an import's lifecycle state, row counters, and — once a portfolio import has completed — the products it brought in that no listing matches.
// @Tags        imports
// @Produce     application/json
// @Param       import_id path string true "Import UUID"
// @Success     200 {object} ImportResultResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     404 {object} map[string]string "Not found"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /imports/{import_id} [get]
func GetImport(log logging.Logger, queries *importer.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		importID, err := uuid.Parse(r.PathValue("import_id"))
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"import_id": "import_id must be a valid UUID"})
			return
		}

		result, err := queries.Result(r.Context(), importID, accountID)
		if err != nil {
			if errors.Is(err, importer.ErrImportNotFound) {
				_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"import_id": err.Error()})
				return
			}
			log.Error(r.Context(), "get import: failed to read import result", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to read import result"})
			return
		}

		_ = httpx.JSONEncode(w, http.StatusOK, toImportResultResponse(result))
	})
}

func toImportResultResponse(result *importer.ImportResult) ImportResultResponse {
	imp := result.Import
	products := make([]UnlinkedProductResponse, 0, len(result.UnlinkedProducts))
	for _, product := range result.UnlinkedProducts {
		products = append(products, toUnlinkedProductResponse(product))
	}
	return ImportResultResponse{
		ImportID:         imp.ID,
		Type:             string(imp.Type),
		Status:           string(imp.Status),
		Reason:           importer.FailureReason(imp),
		TotalRows:        imp.TotalRows,
		Imported:         imp.Imported,
		Duplicates:       imp.Duplicates,
		Failed:           imp.Failed,
		AutoIgnored:      imp.AutoIgnored,
		CreatedAt:        imp.CreatedAt,
		UpdatedAt:        imp.UpdatedAt,
		UnlinkedProducts: products,
	}
}

func toUnlinkedProductResponse(product portfolio.ImportedProduct) UnlinkedProductResponse {
	return UnlinkedProductResponse{
		Name:         product.Name,
		ISIN:         product.ISIN,
		Symbol:       product.Symbol,
		Transactions: product.Transactions,
	}
}
