package backup

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/backup"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// PortfolioImportResponse is what a portfolio import did, or on a dry run would do.
type PortfolioImportResponse struct {
	DryRun       bool            `json:"dry_run"`
	Imports      CountResponse   `json:"imports"`
	Transactions CountResponse   `json:"transactions"`
	Missing      MissingResponse `json:"missing"`
}

func toPortfolioImportResponse(r *backup.PortfolioResult) PortfolioImportResponse {
	return PortfolioImportResponse{
		DryRun:       r.DryRun,
		Imports:      toCount(r.Imports),
		Transactions: toCount(r.Transactions),
		Missing:      toMissing(r.Missing),
	}
}

// ExportPortfolio returns the signed-in account's portfolio transactions as a backup file.
//
// @Summary     Export portfolio
// @Description Returns every portfolio transaction of the signed-in account with the imports that brought them in, as one JSON file. Positions and snapshots are left out and rebuilt on import.
// @Tags        backup
// @Produce     application/json
// @Success     200 {object} backup.PortfolioFile
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /backup/portfolio/export [get]
func ExportPortfolio(log logging.Logger, queries *backup.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		serveExport(w, r, log, "portfolio", func() (*backup.PortfolioFile, error) {
			return queries.ExportPortfolio(r.Context(), accountID)
		})
	})
}

// ImportPortfolio merges a portfolio backup into the signed-in account.
//
// @Summary     Import portfolio
// @Description Merges a portfolio backup file into the signed-in account: transactions the account has are skipped, new ones are added, and the whole file lands or nothing does; the portfolio is rebuilt afterwards. With dry_run the counts are reported and nothing is written. A file naming vendors this instance lacks is refused with 409 and the missing ones listed.
// @Tags        backup
// @Accept      multipart/form-data
// @Produce     application/json
// @Param       file formData file true "Portfolio backup JSON"
// @Param       dry_run formData bool false "Report what would change without writing"
// @Success     200 {object} PortfolioImportResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     409 {object} PortfolioImportResponse "Vendors missing"
// @Failure     413 {object} map[string]string "File too large"
// @Failure     422 {object} map[string]string "Not a readable portfolio backup"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /backup/portfolio/import [post]
func ImportPortfolio(log logging.Logger, commands *backup.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		serveImport(w, r, log, "portfolio", maxAccountBackupBytes,
			func(file *backup.PortfolioFile, dryRun bool) (*backup.PortfolioResult, error) {
				return commands.ImportPortfolio(r.Context(), accountID, file, dryRun)
			},
			toPortfolioImportResponse,
		)
	})
}
