package marketdata

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/money"
)

// QuoteKind labels what sort of instrument a quote is, so the picker can group
// and explain its rows.
type QuoteKind string

const (
	// QuoteKindCrypto is a digital currency quoted daily against a market currency.
	QuoteKindCrypto QuoteKind = "CRYPTO"
	// QuoteKindMetal is a precious metal. No configured provider quotes one in euro,
	// so every metal row is offered unselectable with the reason attached.
	QuoteKindMetal QuoteKind = "METAL"
)

// Quote is one instrument offered as a daily-priced holding.
//
// ListingID, Price and PriceDate are nil until the instrument is tracked: a quote
// nobody holds yet has no listing and therefore no price history, and fetching one
// just to fill a search row would spend a provider request per keystroke.
type Quote struct {
	Symbol     string
	Name       string
	Kind       QuoteKind
	Currency   money.Currency
	ListingID  *uuid.UUID
	Price      *money.Price
	PriceDate  *time.Time
	Selectable bool
	// Reason explains an unselectable row in words fit for display.
	Reason string
}

// blockedQuotes are instruments the user asks for that no configured provider
// delivers in euro. They are listed rather than hidden: a picker that silently
// omits gold reads as "gold is not supported", while a blocked row with its reason
// says what to do instead. Converting currencies is out of scope by decision, so
// these stay manual items until a euro source exists.
var blockedQuotes = []Quote{
	{
		Symbol:     "XAU",
		Name:       "Gold (troy ounce)",
		Kind:       QuoteKindMetal,
		Currency:   money.CurrencyUSD,
		Selectable: false,
		Reason:     "Priced in US dollars. Track gold as a manual item for now.",
	},
	{
		Symbol:     "XAG",
		Name:       "Silver (troy ounce)",
		Kind:       QuoteKindMetal,
		Currency:   money.CurrencyUSD,
		Selectable: false,
		Reason:     "Priced in US dollars. Track silver as a manual item for now.",
	},
}

// QuoteCatalogue supplies the daily-priced instruments one provider offers.
type QuoteCatalogue interface {
	// Quotes returns every instrument the provider quotes daily in a usable currency.
	Quotes() []Quote
}

// QuoteStore reads the tracked listing behind a quote.
type QuoteStore interface {
	GetBySymbol(ctx context.Context, symbol string) (*Listing, error)
}

// Quotes answers "what can I track with a daily price?" and turns a chosen answer
// into a tracked listing.
type Quotes struct {
	qs         QuoteStore
	queries    *Queries
	commands   *Commands
	catalogues map[Source]QuoteCatalogue
}

// NewQuotes constructs the quote service over the catalogues of the sources that
// can actually fetch prices.
func NewQuotes(qs QuoteStore, queries *Queries, commands *Commands, catalogues map[Source]QuoteCatalogue) *Quotes {
	if catalogues == nil {
		catalogues = map[Source]QuoteCatalogue{}
	}
	return &Quotes{qs: qs, queries: queries, commands: commands, catalogues: catalogues}
}

// Search returns the instruments matching a free-text query, selectable ones first
// and alphabetically within each group. An empty query returns everything, which is
// what the picker shows before the user types.
func (q *Quotes) Search(ctx context.Context, query string, limit int) ([]Quote, error) {
	if limit <= 0 || limit > quoteSearchMaxLimit {
		limit = quoteSearchMaxLimit
	}
	needle := strings.ToLower(strings.TrimSpace(query))

	candidates := make([]Quote, 0, len(blockedQuotes))
	for _, catalogue := range q.catalogues {
		candidates = append(candidates, catalogue.Quotes()...)
	}
	candidates = append(candidates, blockedQuotes...)

	matches := make([]Quote, 0, len(candidates))
	for _, candidate := range candidates {
		if !quoteMatches(candidate, needle) {
			continue
		}
		if candidate.Selectable {
			if err := q.attachTracked(ctx, &candidate); err != nil {
				return nil, err
			}
		}
		matches = append(matches, candidate)
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Selectable != matches[j].Selectable {
			return matches[i].Selectable
		}
		return matches[i].Symbol < matches[j].Symbol
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return matches, nil
}

// quoteSearchMaxLimit caps one picker page. The catalogue is a fixed, small table,
// so this exists to bound the response rather than to paginate.
const quoteSearchMaxLimit = 50

// Track returns the listing for a selectable quote, creating it the first time
// anyone adopts that instrument. Listings are global, so the second account to
// hold bitcoin reuses the first one's listing and its accumulated price history.
//
// The price backfill is deferred: the caller knows the window it needs (a holding
// reaches back to its first purchase, not to the beginning of the series), so it
// triggers the sync itself rather than paying for a full history here.
func (q *Quotes) Track(ctx context.Context, symbol string) (*Listing, error) {
	quote, source, ok := q.find(symbol)
	if !ok {
		return nil, fmt.Errorf("track quote %q: %w", symbol, ErrQuoteNotFound)
	}
	if !quote.Selectable {
		return nil, fmt.Errorf("track quote %q: %w: %s", symbol, ErrQuoteNotSelectable, quote.Reason)
	}

	existing, err := q.qs.GetBySymbol(ctx, quote.Symbol)
	if err != nil {
		return nil, fmt.Errorf("track quote %q: look up listing: %w", symbol, err)
	}
	if existing != nil {
		return existing, nil
	}

	listing, err := q.commands.CreateListing(
		ctx,
		quote.Symbol,
		quote.Name,
		source,
		DeferPriceSync,
		ListingWithCurrency(quote.Currency),
		ListingWithType(string(quote.Kind)),
		ListingWithTicker(strings.Split(quote.Symbol, "/")[0]),
	)
	if err != nil {
		return nil, fmt.Errorf("track quote %q: %w", symbol, err)
	}
	return listing, nil
}

// find resolves a symbol to its catalogue entry and the source that supplies it.
func (q *Quotes) find(symbol string) (Quote, Source, bool) {
	wanted := strings.ToUpper(strings.TrimSpace(symbol))
	for source, catalogue := range q.catalogues {
		for _, quote := range catalogue.Quotes() {
			if strings.ToUpper(quote.Symbol) == wanted {
				return quote, source, true
			}
		}
	}
	for _, quote := range blockedQuotes {
		if quote.Symbol == wanted {
			return quote, "", true
		}
	}
	return Quote{}, "", false
}

// attachTracked fills in the listing id and latest known price for a quote that is
// already tracked, and leaves them nil for one that is not.
func (q *Quotes) attachTracked(ctx context.Context, quote *Quote) error {
	listing, err := q.qs.GetBySymbol(ctx, quote.Symbol)
	if err != nil {
		return fmt.Errorf("search quotes: look up listing %s: %w", quote.Symbol, err)
	}
	if listing == nil {
		return nil
	}
	quote.ListingID = &listing.ID

	if q.queries == nil {
		return nil
	}
	latest, err := q.queries.LatestEOD(ctx, listing.ID)
	if err != nil {
		return fmt.Errorf("search quotes: latest price for %s: %w", quote.Symbol, err)
	}
	if latest == nil {
		return nil
	}
	price := latest.Close
	day := latest.Date
	quote.Price = &price
	quote.PriceDate = &day
	return nil
}

func quoteMatches(quote Quote, needle string) bool {
	if needle == "" {
		return true
	}
	haystack := strings.ToLower(quote.Symbol + " " + quote.Name + " " + string(quote.Kind))
	return strings.Contains(haystack, needle)
}
