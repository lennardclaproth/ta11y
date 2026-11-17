package backup

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/backup"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// AssetsImportResponse is what an assets import did, or on a dry run would do.
type AssetsImportResponse struct {
	DryRun    bool            `json:"dry_run"`
	Classes   CountResponse   `json:"classes"`
	Items     CountResponse   `json:"items"`
	Mutations CountResponse   `json:"mutations"`
	Purchases CountResponse   `json:"purchases"`
	Missing   MissingResponse `json:"missing"`
}

func toAssetsImportResponse(r *backup.AssetsResult) AssetsImportResponse {
	return AssetsImportResponse{
		DryRun:    r.DryRun,
		Classes:   toCount(r.Classes),
		Items:     toCount(r.Items),
		Mutations: toCount(r.Mutations),
		Purchases: toCount(r.Purchases),
		Missing:   toMissing(r.Missing),
	}
}

// ExportAssets returns the signed-in account's asset classes as a backup file.
//
// @Summary     Export asset classes
// @Description Returns the asset classes the signed-in account keeps by hand, with their items, worth changes and purchases, as one JSON file. The portfolio-sourced class is left out and rebuilt from the portfolio.
// @Tags        backup
// @Produce     application/json
// @Success     200 {object} backup.AssetsFile
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /backup/assets/export [get]
func ExportAssets(log logging.Logger, queries *backup.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		serveExport(w, r, log, "assets", func() (*backup.AssetsFile, error) {
			return queries.ExportAssets(r.Context(), accountID)
		})
	})
}

// ImportAssets merges an assets backup into the signed-in account.
//
// @Summary     Import asset classes
// @Description Merges an assets backup file into the signed-in account: classes are matched by name and items by class and name; an item the account already holds keeps its own history, new ones come in with theirs. The whole file lands or nothing does; asset snapshots are rebuilt afterwards. With dry_run the counts are reported and nothing is written. A file naming listings this instance lacks is refused with 409 and the missing ones listed.
// @Tags        backup
// @Accept      multipart/form-data
// @Produce     application/json
// @Param       file formData file true "Assets backup JSON"
// @Param       dry_run formData bool false "Report what would change without writing"
// @Success     200 {object} AssetsImportResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     409 {object} AssetsImportResponse "Listings missing"
// @Failure     413 {object} map[string]string "File too large"
// @Failure     422 {object} map[string]string "Not a readable assets backup"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /backup/assets/import [post]
func ImportAssets(log logging.Logger, commands *backup.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountID, ok := httpx.AccountID(w, r)
		if !ok {
			return
		}
		serveImport(w, r, log, "assets", maxAccountBackupBytes,
			func(file *backup.AssetsFile, dryRun bool) (*backup.AssetsResult, error) {
				return commands.ImportAssets(r.Context(), accountID, file, dryRun)
			},
			toAssetsImportResponse,
		)
	})
}
