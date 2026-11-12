package marketdata

import (
	"errors"
	"net/url"
	"strings"
	"testing"
)

func TestValidateBaseURI(t *testing.T) {
	for _, tc := range []struct {
		raw string
		ok  bool
	}{
		{"https://api.marketstack.com/v2", true},
		{"https://www.alphavantage.co", true},
		{"http://api.marketstack.com/v2", false},
		{"https://169.254.169.254/latest", false},
		{"https://10.0.0.5", false},
		{"https://127.0.0.1:8443", false},
		{"https://[::1]/x", false},
		{"https://localhost/x", false},
		{"https://user:pw@api.marketstack.com", false},
		{"https://api.marketstack.com/v2?access_key=x", false},
		{"https:///v2", false},
		{"not a url", false},
	} {
		err := ValidateBaseURI(tc.raw)
		if tc.ok && err != nil {
			t.Errorf("ValidateBaseURI(%q) = %v, want nil", tc.raw, err)
		}
		if !tc.ok && !errors.Is(err, ErrProviderBaseURIInvalid) {
			t.Errorf("ValidateBaseURI(%q) = %v, want ErrProviderBaseURIInvalid", tc.raw, err)
		}
	}
}

func TestRedactRequestError_StripsQuery(t *testing.T) {
	err := &url.Error{Op: "Get", URL: "https://api.marketstack.com/v2/eod?symbols=X&access_key=SECRET", Err: errors.New("timeout")}
	got := RedactRequestError(err).Error()
	if strings.Contains(got, "SECRET") || strings.Contains(got, "access_key") {
		t.Fatalf("redacted error still carries the key: %s", got)
	}
	if !strings.Contains(got, "api.marketstack.com/v2/eod") {
		t.Fatalf("redacted error lost the endpoint: %s", got)
	}
	if strings.Contains(err.Error(), "SECRET") == false {
		t.Fatal("original error was mutated")
	}
}

func TestReadProviderBody_RefusesOversizedResponse(t *testing.T) {
	_, err := ReadProviderBody(strings.NewReader(strings.Repeat("x", MaxProviderResponseBytes+1)))
	if !errors.Is(err, ErrProviderResponseTooLarge) {
		t.Fatalf("err = %v, want ErrProviderResponseTooLarge", err)
	}
}
