package alphavantage

import (
	"strings"

	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// Market is the quote currency every digital-currency series is requested in.
// The app does no currency conversion, so euro is the only market it asks for;
// an instrument the provider cannot quote in euro is simply not offered.
const Market = "EUR"

// SymbolSeparator joins a digital currency to its market in a listing symbol,
// for example "BTC/EUR".
const SymbolSeparator = "/"

// DigitalCurrency is one crypto currency the provider quotes daily.
type DigitalCurrency struct {
	// Code is the provider's currency code, e.g. "BTC".
	Code string
	// Name is the human-readable name, e.g. "Bitcoin".
	Name string
}

// Symbol is the listing symbol for this currency quoted in euro, e.g. "BTC/EUR".
func (c DigitalCurrency) Symbol() string {
	return c.Code + SymbolSeparator + Market
}

// digitalCurrencies is the set of crypto currencies offered as daily-priced
// holdings. The provider publishes a far longer list, but that list is a CSV
// download that would have to be fetched and cached at runtime; a fixed table
// costs no provider request, keeps the picker instant, and covers what is
// actually held. Adding a currency is a one-line change here.
var digitalCurrencies = []DigitalCurrency{
	{Code: "BTC", Name: "Bitcoin"},
	{Code: "ETH", Name: "Ethereum"},
	{Code: "SOL", Name: "Solana"},
	{Code: "XRP", Name: "XRP"},
	{Code: "ADA", Name: "Cardano"},
	{Code: "DOGE", Name: "Dogecoin"},
	{Code: "DOT", Name: "Polkadot"},
	{Code: "LTC", Name: "Litecoin"},
	{Code: "BCH", Name: "Bitcoin Cash"},
	{Code: "LINK", Name: "Chainlink"},
	{Code: "AVAX", Name: "Avalanche"},
	{Code: "ATOM", Name: "Cosmos"},
	{Code: "XLM", Name: "Stellar"},
	{Code: "ALGO", Name: "Algorand"},
	{Code: "UNI", Name: "Uniswap"},
	{Code: "XMR", Name: "Monero"},
}

// DigitalCurrencies returns the supported crypto currencies in presentation order.
func DigitalCurrencies() []DigitalCurrency {
	out := make([]DigitalCurrency, len(digitalCurrencies))
	copy(out, digitalCurrencies)
	return out
}

// Quotes exposes the supported currencies as market-data quotes, so the instrument
// picker can offer them before any of them is tracked. Every row is quoted in euro
// and therefore selectable.
func (c *Client) Quotes() []marketdata.Quote {
	out := make([]marketdata.Quote, 0, len(digitalCurrencies))
	for _, currency := range digitalCurrencies {
		out = append(out, marketdata.Quote{
			Symbol:     currency.Symbol(),
			Name:       currency.Name,
			Kind:       marketdata.QuoteKindCrypto,
			Currency:   money.CurrencyEUR,
			Selectable: true,
		})
	}
	return out
}

// DigitalCurrencyBySymbol resolves a listing symbol such as "BTC/EUR" back to the
// supported currency it names. It reports false for an unknown code or a market
// other than euro.
func DigitalCurrencyBySymbol(symbol string) (DigitalCurrency, bool) {
	code, market, found := strings.Cut(strings.ToUpper(strings.TrimSpace(symbol)), SymbolSeparator)
	if !found || market != Market {
		return DigitalCurrency{}, false
	}
	for _, currency := range digitalCurrencies {
		if currency.Code == code {
			return currency, true
		}
	}
	return DigitalCurrency{}, false
}
