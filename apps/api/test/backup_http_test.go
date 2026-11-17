//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"
)

// upload posts file as the multipart field "file", with dry_run set, and returns the status
// and body.
func (ts *testServer) upload(t *testing.T, path string, file []byte, dryRun bool) (int, []byte) {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", "backup.json")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(file); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := form.WriteField("dry_run", map[bool]string{true: "true", false: "false"}[dryRun]); err != nil {
		t.Fatalf("write dry_run: %v", err)
	}
	if err := form.Close(); err != nil {
		t.Fatalf("close form: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, ts.baseURL+path, &body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", form.FormDataContentType())
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return resp.StatusCode, out
}

// TestBackupOverHTTP drives the export and import routes on the real binary: a download is
// a JSON attachment, the same file uploads back as a dry run, and a file of another
// dataset is refused before anything is read into the ledger.
func TestBackupOverHTTP(t *testing.T) {
	ts := startServer(t)

	for _, dataset := range []string{"cashflow", "portfolio", "assets", "eods"} {
		t.Run("export "+dataset, func(t *testing.T) {
			resp, err := (&http.Client{Timeout: 10 * time.Second}).Get(ts.baseURL + "/backup/" + dataset + "/export")
			if err != nil {
				t.Fatalf("GET export: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("GET /backup/%s/export: want 200, got %d", dataset, resp.StatusCode)
			}
			disposition := resp.Header.Get("Content-Disposition")
			if !strings.HasPrefix(disposition, "attachment;") || !strings.Contains(disposition, "ta11y-"+dataset+"-") {
				t.Fatalf("Content-Disposition = %q", disposition)
			}
			var header struct {
				Kind string `json:"kind"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&header); err != nil || header.Kind != dataset {
				t.Fatalf("exported kind %q, err %v", header.Kind, err)
			}
		})
	}

	_, cashflowFile := ts.get(t, "/backup/cashflow/export")
	_, portfolioFile := ts.get(t, "/backup/portfolio/export")

	t.Run("import dry run", func(t *testing.T) {
		status, body := ts.upload(t, "/backup/cashflow/import", cashflowFile, true)
		if status != http.StatusOK {
			t.Fatalf("POST /backup/cashflow/import: want 200, got %d: %s", status, body)
		}
		var result struct {
			DryRun       bool `json:"dry_run"`
			Transactions struct {
				New int `json:"new"`
			} `json:"transactions"`
		}
		if err := json.Unmarshal(body, &result); err != nil || !result.DryRun {
			t.Fatalf("dry run result %s, err %v", body, err)
		}
	})

	t.Run("another dataset's file is refused", func(t *testing.T) {
		status, body := ts.upload(t, "/backup/cashflow/import", portfolioFile, false)
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("portfolio file to cashflow import: want 422, got %d: %s", status, body)
		}
	})

	t.Run("not a backup is refused", func(t *testing.T) {
		status, body := ts.upload(t, "/backup/assets/import", []byte("date,amount\n"), false)
		if status != http.StatusUnprocessableEntity {
			t.Fatalf("CSV to assets import: want 422, got %d: %s", status, body)
		}
	})
}
