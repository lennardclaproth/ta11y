package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	httpx "github.com/lennardclaproth/ta11y/transport/http"
)

// ProviderCredentialResponse is one external-provider connection record. The stored
// API key is never included: only whether one is set and a hint identifying it.
type ProviderCredentialResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	IngestionMode string  `json:"ingestion_mode"`
	BaseURI       *string `json:"base_uri,omitempty"`
	HasAPIKey     bool    `json:"has_api_key"`
	APIKeyHint    string  `json:"api_key_hint"`
	Remaining     int     `json:"remaining"`
	Used          int     `json:"used"`
	Total         int     `json:"total"`
	ResetsAt      *string `json:"resets_at,omitempty"`
}

// UpdateProviderCredentialsRequest sets a provider's API key and/or base URI.
// Omitted fields are left unchanged.
type UpdateProviderCredentialsRequest struct {
	APIKey  *string `json:"api_key,omitempty"`
	BaseURI *string `json:"base_uri,omitempty"`
}

func (r UpdateProviderCredentialsRequest) isValid() (bool, map[string]string) {
	problems := make(map[string]string)
	if r.APIKey == nil && r.BaseURI == nil {
		problems["credentials"] = marketdata.ErrNoCredentialFieldsToUpdate.Error()
	}
	if r.APIKey != nil && strings.TrimSpace(*r.APIKey) == "" {
		problems["api_key"] = "api key cannot be empty"
	}
	if r.BaseURI != nil && strings.TrimSpace(*r.BaseURI) == "" {
		problems["base_uri"] = "base uri cannot be empty"
	}
	return len(problems) == 0, problems
}

// RevealProviderAPIKeyResponse carries a provider's API key in full.
type RevealProviderAPIKeyResponse struct {
	ID     string `json:"id"`
	APIKey string `json:"api_key"`
}

// GetProviderCredentials lists the external providers and whether each has a key.
//
// @Summary List provider credentials
// @Description Return every market-data provider with its ingestion mode, base URI, quota counters, and a masked hint for its API key. The stored key itself is never returned.
// @Tags providers
// @Produce json
// @Success 200 {array} ProviderCredentialResponse
// @Failure 500 {object} map[string]string
// @Router /marketdata/providers [get]
func GetProviderCredentials(
	log logging.Logger,
	credentials *marketdata.Credentials,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		found, err := credentials.List(r.Context())
		if err != nil {
			log.Error(r.Context(), "list provider credentials: failed to list providers", err)
			_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to list provider credentials"})
			return
		}
		_ = httpx.JSONEncode(w, http.StatusOK, toProviderCredentialResponses(found))
	})
}

// UpdateProviderCredentials sets a provider's API key and/or base URI.
//
// @Summary Update provider credentials
// @Description Set a provider's API key and/or base URI. Omitted fields are left unchanged, and the response withholds the stored key.
// @Tags providers
// @Accept json
// @Produce json
// @Param provider_id path string true "Provider ID"
// @Param request body UpdateProviderCredentialsRequest true "Credential payload"
// @Success 200 {object} ProviderCredentialResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/providers/{provider_id}/credentials [patch]
func UpdateProviderCredentials(
	log logging.Logger,
	credentials *marketdata.Credentials,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerID, err := uuid.Parse(r.PathValue("provider_id"))
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"provider_id": "provider id must be a uuid"})
			return
		}

		req, err := httpx.JSONDecode[UpdateProviderCredentialsRequest](r)
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

		credential, err := credentials.Update(r.Context(), providerID, trimmedPtr(req.APIKey), trimmedPtr(req.BaseURI))
		if err != nil {
			writeCredentialError(w, log, r, "update provider credentials", err)
			return
		}

		// Deliberately logged without the key, so the audit trail records that a
		// credential changed without becoming a place secrets accumulate.
		log.Info(r.Context(), "provider credentials updated",
			"provider", credential.Name, "provider_id", credential.ID.String(),
			"api_key_changed", req.APIKey != nil, "base_uri_changed", req.BaseURI != nil)

		_ = httpx.JSONEncode(w, http.StatusOK, toProviderCredentialResponse(credential))
	})
}

// RevealProviderAPIKey returns a provider's API key in full.
//
// This is the only path that discloses a stored secret. It is a POST so it is not
// casually linked or cached, and every call is logged.
//
// @Summary Reveal a provider API key
// @Description Return the stored API key for one provider in full. Every call is logged.
// @Tags providers
// @Produce json
// @Param provider_id path string true "Provider ID"
// @Success 200 {object} RevealProviderAPIKeyResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /marketdata/providers/{provider_id}/credentials/reveal [post]
func RevealProviderAPIKey(
	log logging.Logger,
	credentials *marketdata.Credentials,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerID, err := uuid.Parse(r.PathValue("provider_id"))
		if err != nil {
			_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"provider_id": "provider id must be a uuid"})
			return
		}

		apiKey, err := credentials.Reveal(r.Context(), providerID)
		if err != nil {
			writeCredentialError(w, log, r, "reveal provider api key", err)
			return
		}

		log.Info(r.Context(), "provider api key revealed", "provider_id", providerID.String())
		_ = httpx.JSONEncode(w, http.StatusOK, RevealProviderAPIKeyResponse{
			ID:     providerID.String(),
			APIKey: apiKey,
		})
	})
}

// writeCredentialError maps credential failures to status codes. Provider errors are
// reported without echoing any stored value.
func writeCredentialError(w http.ResponseWriter, log logging.Logger, r *http.Request, operation string, err error) {
	switch {
	case errors.Is(err, marketdata.ErrProviderNotFound):
		_ = httpx.JSONEncode(w, http.StatusNotFound, map[string]string{"provider": "provider not found"})
	case errors.Is(err, marketdata.ErrCredentialsNotConfigurable):
		_ = httpx.JSONEncode(w, http.StatusConflict, map[string]string{"provider": err.Error()})
	case errors.Is(err, marketdata.ErrNoCredentialFieldsToUpdate),
		errors.Is(err, marketdata.ErrProviderAPIKeyEmpty):
		_ = httpx.JSONEncode(w, http.StatusBadRequest, map[string]string{"credentials": err.Error()})
	default:
		log.Error(r.Context(), operation+": failed", err)
		_ = httpx.JSONEncode(w, http.StatusInternalServerError, map[string]string{"error": "failed to update provider credentials"})
	}
}

// trimmedPtr normalises an optional string field, keeping nil as "leave unchanged".
func trimmedPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func toProviderCredentialResponses(credentials []*marketdata.ProviderCredential) []ProviderCredentialResponse {
	responses := make([]ProviderCredentialResponse, 0, len(credentials))
	for _, credential := range credentials {
		if credential == nil {
			continue
		}
		responses = append(responses, toProviderCredentialResponse(credential))
	}
	return responses
}

func toProviderCredentialResponse(credential *marketdata.ProviderCredential) ProviderCredentialResponse {
	if credential == nil {
		return ProviderCredentialResponse{}
	}
	return ProviderCredentialResponse{
		ID:            credential.ID.String(),
		Name:          string(credential.Name),
		IngestionMode: string(credential.IngestionMode),
		BaseURI:       credential.BaseURI,
		HasAPIKey:     credential.HasAPIKey,
		APIKeyHint:    credential.APIKeyHint,
		Remaining:     credential.Remaining,
		Used:          credential.Used,
		Total:         credential.Total,
		ResetsAt:      credential.ResetsAt,
	}
}
