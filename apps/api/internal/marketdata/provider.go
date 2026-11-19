package marketdata

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// ProviderIngestionMode defines how provider data is ingested.
type ProviderIngestionMode string

const (
	ProviderIngestionModeAPI    ProviderIngestionMode = "API"
	ProviderIngestionModeManual ProviderIngestionMode = "MANUAL"
)

// Provider stores provider connection metadata and token state.
type Provider struct {
	ID            uuid.UUID             `db:"id"`
	Name          ProviderName          `db:"name"`
	IngestionMode ProviderIngestionMode `db:"ingestion_mode"`
	BaseURI       *string               `db:"base_uri"`
	ApiKey        *string               `db:"api_key"`
	Remaining     int                   `db:"remaining"`
	Used          int                   `db:"used"`
	Total         int                   `db:"total"`
	ResetsAt      *string               `db:"resets_at"`
}

var (
	// ErrProviderNameEmpty indicates missing provider name.
	ErrProviderNameEmpty = fmt.Errorf("provider name cannot be empty")
	// ErrProviderBaseURIEmpty indicates missing provider base URI for API providers.
	ErrProviderBaseURIEmpty = fmt.Errorf("provider base URI cannot be empty")
	// ErrProviderBaseURIInvalid indicates a base URI that is not a public https address.
	ErrProviderBaseURIInvalid = fmt.Errorf("provider base URI must be an https URL on a public host")
	// ErrProviderResponseTooLarge indicates a provider response over MaxProviderResponseBytes.
	ErrProviderResponseTooLarge = fmt.Errorf("provider response exceeds the size limit")
	// ErrProviderAPIKeyEmpty indicates missing provider API key for API providers.
	ErrProviderAPIKeyEmpty = fmt.Errorf("provider API key cannot be empty")
	// ErrProviderIngestionModeEmpty indicates missing ingestion mode.
	ErrProviderIngestionModeEmpty = fmt.Errorf("provider ingestion mode cannot be empty")
	// ErrProviderNotFound indicates that the provider does not exist.
	ErrProviderNotFound = fmt.Errorf("provider not found")
	// ErrProviderSourceNotMapped indicates unknown listing source to provider mapping.
	ErrProviderSourceNotMapped = fmt.Errorf("provider source not mapped")
	// ErrProviderManualNotSupported indicates invalid manual-provider field population.
	ErrProviderManualNotSupported = fmt.Errorf("provider does not support manual ingestion")
	// ErrProviderAutomaticNotAllowed indicates invalid API-provider field population.
	ErrProviderAutomaticNotAllowed = fmt.Errorf("provider does not support automatic ingestion")
)

// ProviderName is the canonical provider identifier.
type ProviderName string

const (
	ProviderMarketStack  ProviderName = "marketstack"
	ProviderAlphaVantage ProviderName = "alphavantage"
	ProviderBrandNewDay  ProviderName = "brandnewday"
)

// NewAPIProviderWithAPIKey builds an API-ingestion provider.
func NewAPIProviderWithAPIKey(name ProviderName, baseURI, apiKey string) (*Provider, error) {
	if name == "" {
		return nil, ErrProviderNameEmpty
	}
	if baseURI == "" {
		return nil, ErrProviderBaseURIEmpty
	}
	if apiKey == "" {
		return nil, ErrProviderAPIKeyEmpty
	}
	emptyResetsAt := ""

	return &Provider{
		ID:            uuid.New(),
		Name:          name,
		IngestionMode: ProviderIngestionModeAPI,
		BaseURI:       &baseURI,
		ApiKey:        &apiKey,
		ResetsAt:      &emptyResetsAt,
	}, nil
}

// NewManualProvider builds a manual-ingestion provider.
func NewManualProvider(name ProviderName) (*Provider, error) {
	if name == "" {
		return nil, ErrProviderNameEmpty
	}

	return &Provider{
		ID:            uuid.New(),
		Name:          name,
		IngestionMode: ProviderIngestionModeManual,
		BaseURI:       nil,
		ApiKey:        nil,
		ResetsAt:      nil,
	}, nil
}

// ProviderNameFromSource maps a listing source to its provider name.
func ProviderNameFromSource(source Source) (ProviderName, error) {
	switch source {
	case SourceMarketStack:
		return ProviderMarketStack, nil
	case SourceAlphaVantage:
		return ProviderAlphaVantage, nil
	case SourceBrandNewDay:
		return ProviderBrandNewDay, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrProviderSourceNotMapped, source)
	}
}

// IsManualIngestion reports whether the provider uses manual ingestion.
func (p *Provider) IsManualIngestion() bool {
	return p != nil && p.IngestionMode == ProviderIngestionModeManual
}

// IsAPIIngestion reports whether the provider uses API ingestion.
func (p *Provider) IsAPIIngestion() bool {
	return p != nil && p.IngestionMode == ProviderIngestionModeAPI
}

// Validate checks provider fields for ingestion-mode compatibility.
func (p *Provider) Validate() error {
	if p == nil {
		return fmt.Errorf("provider cannot be nil")
	}
	if p.Name == "" {
		return ErrProviderNameEmpty
	}
	if p.IngestionMode == "" {
		return ErrProviderIngestionModeEmpty
	}
	switch p.IngestionMode {
	case ProviderIngestionModeManual:
		if p.BaseURI != nil || p.ApiKey != nil || p.ResetsAt != nil {
			return ErrProviderManualNotSupported
		}
		return nil
	case ProviderIngestionModeAPI:
		if p.BaseURI == nil || *p.BaseURI == "" {
			return ErrProviderBaseURIEmpty
		}
		if p.ApiKey == nil || *p.ApiKey == "" {
			return ErrProviderAPIKeyEmpty
		}
		if p.ResetsAt == nil {
			return ErrProviderAutomaticNotAllowed
		}
		return nil
	default:
		return ErrProviderIngestionModeEmpty
	}
}

// MaxProviderResponseBytes bounds how much of one provider response is read into memory.
const MaxProviderResponseBytes = 16 << 20

// maxProviderErrorBody bounds how much of a failed response is quoted in an error, since
// those errors are logged and persisted.
const maxProviderErrorBody = 512

// ValidateBaseURI accepts only an https URL naming a host, without credentials, query or
// fragment, and refuses loopback, private and link-local addresses. Every provider request
// carries the API key to this address, so it must not be pointable at the API's own network.
func ValidateBaseURI(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil ||
		u.RawQuery != "" || u.Fragment != "" {
		return ErrProviderBaseURIInvalid
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return ErrProviderBaseURIInvalid
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
			ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return ErrProviderBaseURIInvalid
		}
	}
	return nil
}

// RedactRequestError strips the query string from the URL a failed request reports.
// Provider API keys travel as query parameters, and *url.Error prints the full URL.
func RedactRequestError(err error) error {
	var urlErr *url.Error
	if !errors.As(err, &urlErr) {
		return err
	}
	redacted := *urlErr
	if u, parseErr := url.Parse(urlErr.URL); parseErr == nil {
		u.RawQuery = ""
		u.User = nil
		redacted.URL = u.String()
	} else {
		redacted.URL = "[redacted]"
	}
	return &redacted
}

// ReadProviderBody reads a provider response, refusing one larger than
// MaxProviderResponseBytes.
func ReadProviderBody(r io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r, MaxProviderResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > MaxProviderResponseBytes {
		return nil, ErrProviderResponseTooLarge
	}
	return body, nil
}

// ErrorBodySnippet returns the start of a failed response, for quoting in an error.
func ErrorBodySnippet(body []byte) string {
	if len(body) <= maxProviderErrorBody {
		return string(body)
	}
	return string(body[:maxProviderErrorBody]) + "...(truncated)"
}
