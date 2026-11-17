package backup

import (
	"net/http"

	"github.com/lennardclaproth/ta11y/internal/backup"
	"github.com/lennardclaproth/ta11y/internal/logging"
)

// maxEODBackupBytes caps an uploaded EOD backup: it holds every day of every listing.
const maxEODBackupBytes = 256 << 20 // 256MB

// EODImportResponse is what an EOD import did, or on a dry run would do. A listing
// counts as new when it gained at least one day.
type EODImportResponse struct {
	DryRun   bool            `json:"dry_run"`
	Listings CountResponse   `json:"listings"`
	EODs     CountResponse   `json:"eods"`
	Missing  MissingResponse `json:"missing"`
}

func toEODImportResponse(r *backup.EODResult) EODImportResponse {
	return EODImportResponse{
		DryRun:   r.DryRun,
		Listings: toCount(r.Listings),
		EODs:     toCount(r.EODs),
		Missing:  toMissing(r.Missing),
	}
}

// ExportEODs returns the instance's end-of-day prices as a backup file.
//
// @Summary     Export the end-of-day prices
// @Description Returns the end-of-day prices of every listing that has any, with their split factors, as one JSON file. Listings are named by symbol and source.
// @Tags        backup
// @Produce     application/json
// @Success     200 {object} backup.EODFile
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /backup/eods/export [get]
func ExportEODs(log logging.Logger, queries *backup.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveExport(w, r, log, "eods", func() (*backup.EODFile, error) {
			return queries.ExportEODs(r.Context())
		})
	})
}

// ImportEODs merges an EOD backup into the instance's end-of-day prices.
//
// @Summary     Import end-of-day prices
// @Description Merges an EOD backup file: days a listing has are skipped, new ones are added with their split factor, and the whole file lands or nothing does. With dry_run the counts are reported and nothing is written. A file naming listings this instance lacks is refused with 409 and the missing ones listed.
// @Tags        backup
// @Accept      multipart/form-data
// @Produce     application/json
// @Param       file formData file true "EOD backup JSON"
// @Param       dry_run formData bool false "Report what would change without writing"
// @Success     200 {object} EODImportResponse
// @Failure     400 {object} map[string]string "Bad request"
// @Failure     409 {object} EODImportResponse "Listings missing"
// @Failure     413 {object} map[string]string "File too large"
// @Failure     422 {object} map[string]string "Not a readable EOD backup"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /backup/eods/import [post]
func ImportEODs(log logging.Logger, commands *backup.Commands) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveImport(w, r, log, "eods", maxEODBackupBytes,
			func(file *backup.EODFile, dryRun bool) (*backup.EODResult, error) {
				return commands.ImportEODs(r.Context(), file, dryRun)
			},
			toEODImportResponse,
		)
	})
}
