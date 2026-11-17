package backup

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/backup"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// maxAccountBackupBytes caps an uploaded cashflow, portfolio or assets backup.
const maxAccountBackupBytes = 64 << 20 // 64MB

// CashflowImportResponse is what a cashflow import did, or on a dry run would do.
type CashflowImportResponse struct {
	DryRun         bool            `json:"dry_run"`
	Imports        CountResponse   `json:"imports"`
	Transactions   CountResponse   `json:"transactions"`
	IgnoreRules    CountResponse   `json:"ignore_rules"`
	RecurringItems CountResponse   `json:"recurring_items"`
	RecurringLinks CountResponse   `json:"recurring_links"`
	Dismissals     CountResponse   `json:"recurring_dismissals"`
	Missing        MissingResponse `json:"missing"`
}

func toCashflowImportResponse(r *backup.CashflowResult) CashflowImportResponse {
	return CashflowImportResponse{
		DryRun:         r.DryRun,
		Imports:        toCount(r.Imports),
		Transactions:   toCount(r.Transactions),
		IgnoreRules:    toCount(r.IgnoreRules),
		RecurringItems: toCount(r.RecurringItems),
		RecurringLinks: toCount(r.RecurringLinks),
		Dismissals:     toCount(r.Dismissals),
		Missing:        toMissing(r.Missing),
	}
}

// ExportCashflow returns the signed-in account's cashflow as a backup file.
//
// @Summary     Export cashflow
// @Description Returns every cashflow transaction of the signed-in account with its tag and ignored state, the imports that brought them in, the ignore rules, and the recurring items with their links and dismissals, as one JSON file.
// @Tags        backup
// @Produce     application/json
// @Success     200 {object} backup.CashflowFile
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /backup/cashflow/export [get]
func ExportCashflow(log logging.Logger, queries *backup.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		serveExport(w, r, log, "cashflow", func() (*backup.CashflowFile, error) {
			return queries.ExportCashflow(r.Context(), accountID)
		})
	})
}

// ImportCashflow merges a cashflow backup into the signed-in account.
//
// @Summary     Import cashflow
// @Description Merges a cashflow backup file into the signed-in account: transactions, rules and recurring items the account has are skipped, new ones are added, and the whole file lands or nothing does. With dry_run the counts are reported and nothing is written. A file naming vendors this instance lacks is refused with 409 and the missing ones listed.
// @Tags        backup
// @Accept      multipart/form-data
// @Produce     application/json
// @Param       file formData file true "Cashflow backup JSON"
// @Param       dry_run formData bool false "Report what would change without writing"
// @Success     200 {object} CashflowImportResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     409 {object} CashflowImportResponse "Vendors missing"
// @Failure     413 {object} map[string]string "File too large"
// @Failure     422 {object} map[string]string "Not a readable cashflow backup"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /backup/cashflow/import [post]
func ImportCashflow(log logging.Logger, commands *backup.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		serveImport(w, r, log, "cashflow", maxAccountBackupBytes,
			func(file *backup.CashflowFile, dryRun bool) (*backup.CashflowResult, error) {
				return commands.ImportCashflow(r.Context(), accountID, file, dryRun)
			},
			toCashflowImportResponse,
		)
	})
}
