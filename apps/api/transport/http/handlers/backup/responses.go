// Package backup serves the backup files and accepts them back.
package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/lennardclaproth/ta11y/internal/backup"
	"github.com/lennardclaproth/ta11y/internal/logging"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// CountResponse is how many records of one kind were new to the target, and how many
// it already had.
type CountResponse struct {
	New      int `json:"new"`
	Existing int `json:"existing"`
}

func toCount(c backup.Count) CountResponse { return CountResponse{New: c.New, Existing: c.Existing} }

// ListingRefResponse names a listing by symbol and source.
type ListingRefResponse struct {
	Symbol string `json:"symbol"`
	Source string `json:"source"`
}

// MissingResponse names the vendors and listings a file needs that this instance lacks.
type MissingResponse struct {
	Vendors  []string             `json:"vendors"`
	Listings []ListingRefResponse `json:"listings"`
}

func toMissing(m backup.Missing) MissingResponse {
	out := MissingResponse{Vendors: m.Vendors, Listings: make([]ListingRefResponse, 0, len(m.Listings))}
	if out.Vendors == nil {
		out.Vendors = []string{}
	}
	for _, l := range m.Listings {
		out.Listings = append(out.Listings, ListingRefResponse{Symbol: l.Symbol, Source: string(l.Source)})
	}
	return out
}

// importRequest is the multipart payload of a backup import.
type importRequest struct {
	File   multipart.File `multipart:"file"`
	DryRun bool           `form:"dry_run"`
}

// maxUploadMemory is how much of an upload is held in memory; the rest is buffered to
// a temporary file while it is parsed.
const maxUploadMemory = 8 << 20 // 8MB

// decodeUpload reads the uploaded file into dest. It writes the error response and
// returns false when the upload is missing, too large, or not JSON of the right shape.
func decodeUpload(w http.ResponseWriter, r *http.Request, log logging.Logger, maxBytes int64, dest any) (dryRun, ok bool) {
	req, err := httpx.DecodeMultipartFile[importRequest](r, httpx.MultipartFileDecoderOptions{
		FieldName: "file",
		MaxBytes:  maxBytes,
		MaxMemory: maxUploadMemory,
	})
	if req.File != nil {
		defer closeUpload(r.Context(), log, req.File)
	}
	if err != nil {
		if httpx.WriteDecodeError(w, err) {
			return false, false
		}
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"error": "invalid upload"})
		return false, false
	}
	if err := json.NewDecoder(req.File).Decode(dest); err != nil {
		_ = httpx.JSONEncode(w, http.StatusUnprocessableEntity, map[string]string{"file": "not a readable ta11y backup"})
		return false, false
	}
	return req.DryRun, true
}

func closeUpload(ctx context.Context, log logging.Logger, file multipart.File) {
	if err := file.Close(); err != nil {
		log.Error(ctx, "backup import: failed to close upload", err)
	}
}

// writeFileError maps a file the backup feature refused to a response. It returns false
// for any other error.
func writeFileError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, backup.ErrInvalidFile),
		errors.Is(err, backup.ErrUnsupportedVersion),
		errors.Is(err, backup.ErrWrongKind):
		_ = httpx.JSONEncode(w, http.StatusUnprocessableEntity, map[string]string{"file": err.Error()})
		return true
	}
	return false
}

// serveExport writes a backup file as a download named after what it holds.
func serveExport[F any](w http.ResponseWriter, r *http.Request, log logging.Logger, name string, export func() (*F, error)) {
	file, err := export()
	if err != nil {
		log.Error(r.Context(), "backup export: failed to export "+name, err)
		_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to export " + name})
		return
	}
	attach(w, "ta11y-"+name)
	_ = httpx.JSONEncode(w, http.StatusOK, file)
}

// serveImport decodes an uploaded backup file, merges it, and maps the outcome: the
// counts on success, the counts with what is missing on 409, a refused file on 422.
func serveImport[F, R, D any](
	w http.ResponseWriter,
	r *http.Request,
	log logging.Logger,
	name string,
	maxBytes int64,
	merge func(file *F, dryRun bool) (*R, error),
	respond func(*R) D,
) {
	var file F
	dryRun, ok := decodeUpload(w, r, log, maxBytes, &file)
	if !ok {
		return
	}
	result, err := merge(&file, dryRun)
	if errors.Is(err, backup.ErrMissingReferences) {
		_ = httpx.JSONEncode(w, http.StatusConflict, respond(result))
		return
	}
	if err != nil {
		if !writeFileError(w, err) {
			log.Error(r.Context(), "backup import: failed to import "+name, err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to import " + name})
		}
		return
	}
	_ = httpx.JSONEncode(w, http.StatusOK, respond(result))
}

// attach marks the response as a file to save under name, dated today.
func attach(w http.ResponseWriter, name string) {
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s-%s.json"`, name, time.Now().UTC().Format("2006-01-02")))
}
