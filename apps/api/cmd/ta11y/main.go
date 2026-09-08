package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lennardclaproth/ta11y/docs"
	"github.com/lennardclaproth/ta11y/internal/account"
	"github.com/lennardclaproth/ta11y/internal/assets"
	"github.com/lennardclaproth/ta11y/internal/auth"
	"github.com/lennardclaproth/ta11y/internal/bootstrap"
	"github.com/lennardclaproth/ta11y/internal/cashflow"
	"github.com/lennardclaproth/ta11y/internal/config"
	"github.com/lennardclaproth/ta11y/internal/eventbus"
	memorybus "github.com/lennardclaproth/ta11y/internal/eventbus/memory"
	"github.com/lennardclaproth/ta11y/internal/files"
	"github.com/lennardclaproth/ta11y/internal/importer"
	importercashflow "github.com/lennardclaproth/ta11y/internal/importer/cashflow"
	cashflowparsers "github.com/lennardclaproth/ta11y/internal/importer/cashflow/parsers"
	importereod "github.com/lennardclaproth/ta11y/internal/importer/eod"
	eodparsers "github.com/lennardclaproth/ta11y/internal/importer/eod/parsers"
	importerportfolio "github.com/lennardclaproth/ta11y/internal/importer/portfolio"
	portfolioparsers "github.com/lennardclaproth/ta11y/internal/importer/portfolio/parsers"
	"github.com/lennardclaproth/ta11y/internal/logging"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/marketdata/marketstack"
	"github.com/lennardclaproth/ta11y/internal/notify"
	"github.com/lennardclaproth/ta11y/internal/portfolio"
	"github.com/lennardclaproth/ta11y/internal/storage"
	"github.com/lennardclaproth/ta11y/internal/vendor"
	"github.com/lennardclaproth/ta11y/migrations"
	apphttp "github.com/lennardclaproth/ta11y/transport/http"
	basehandlers "github.com/lennardclaproth/ta11y/transport/http/handlers"
	accounthttp "github.com/lennardclaproth/ta11y/transport/http/handlers/account"
	assethttp "github.com/lennardclaproth/ta11y/transport/http/handlers/assets"
	authhttp "github.com/lennardclaproth/ta11y/transport/http/handlers/auth"
	cashflowhttp "github.com/lennardclaproth/ta11y/transport/http/handlers/cashflow"
	importerhttp "github.com/lennardclaproth/ta11y/transport/http/handlers/importer"
	marketdatahttp "github.com/lennardclaproth/ta11y/transport/http/handlers/marketdata"
	portfoliohttp "github.com/lennardclaproth/ta11y/transport/http/handlers/portfolio"
	vendorshttp "github.com/lennardclaproth/ta11y/transport/http/handlers/vendors"
	assetsevents "github.com/lennardclaproth/ta11y/transport/messaging/handlers/assets"
	importerevents "github.com/lennardclaproth/ta11y/transport/messaging/handlers/importer"
	portfolioevents "github.com/lennardclaproth/ta11y/transport/messaging/handlers/portfolio"
	httpSwagger "github.com/swaggo/http-swagger"
)

const eventQueueSize = 128

type application struct {
	log logging.Logger
	hub *notify.Hub

	accountCommands *account.Commands
	accountQueries  *account.Queries
	authCommands    *auth.Commands
	authQueries     *auth.Queries
	authRegistry    *auth.Registry
	authFallback    *auth.Principal
	authStore       *storage.SQLXAuthStore
	authCookies     apphttp.CookieSettings
	frontendURL     string
	vendorCommands  *vendor.Commands
	vendorQueries   *vendor.Queries

	cashflowCommands *cashflow.Commands
	cashflowQueries  *cashflow.Queries

	portfolioCommands *portfolio.Commands
	portfolioQueries  *portfolio.Queries
	portfolioBuilder  *portfolio.Builder
	portfolioStore    *storage.SQLXPortfolioStore

	assetsCommands *assets.Commands
	assetsQueries  *assets.Queries
	assetsBuilder  *assets.Builder
	assetsSyncer   *assets.Syncer

	importerCommands   *importer.Commands
	marketDataCommands *marketdata.Commands
	marketDataQueries  *marketdata.Queries
	marketDataCatalog  *marketdata.Catalogue
	marketDataCreds    *marketdata.Credentials
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ta11y: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	log := logging.NewSlogLogger(cfg.Logging.GetLogLevel())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connType := storage.ConnectionType(cfg.Database.Type)
	db := storage.NewDB(cfg.Database.ConnStr, connType)
	defer closeDB(log, db)

	migrator := migrations.NewMigrator(db, connType, log)
	if err := migrator.EnsureDBExists(ctx, cfg.Database.ConnStr); err != nil {
		return fmt.Errorf("ensure database exists: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	if err := migrator.RunMigrations(ctx, db, connType); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	bus := memorybus.NewMemoryBus(
		memorybus.WithQueueSize(eventQueueSize),
		memorybus.WithBackpressure(memorybus.BackpressureError),
	)
	defer closeBus(log, bus)

	app := buildApplication(cfg, log, bus, db)
	defer closeHub(log, app.hub)

	// Provider discovery reaches the network, so a misconfigured provider fails the
	// boot rather than the first sign-in attempt.
	registry, err := buildAuthRegistry(ctx, cfg, log)
	if err != nil {
		return err
	}
	app.authRegistry = registry

	if err := registerEventHandlers(bus, app); err != nil {
		return err
	}

	bootstrap.Vendors(ctx, app.vendorCommands, log)
	bootstrap.Providers(ctx, app.marketDataCommands, cfg.Providers, log)
	bootstrap.Accounts(ctx, app.accountCommands, app.accountQueries, cfg.Auth.BootstrapAdminEmail, log)

	// With sign-in switched off the API keeps behaving as the single bootstrapped
	// account, so local development and the test suite need no session. Config refuses
	// this combination outside development.
	if !cfg.Auth.Enabled {
		fallback, err := bootstrap.FallbackPrincipal(ctx, app.accountQueries)
		if err != nil {
			return fmt.Errorf("resolve unauthenticated fallback account: %w", err)
		}
		app.authFallback = fallback
		log.Warn(ctx, "authentication is disabled; every request runs as the bootstrapped account",
			"account_id", fallback.AccountID.String())
	}

	go auth.SweepExpiredSessions(ctx, app.authStore, log)

	router := apphttp.NewRouter()
	registerRoutes(router, app)

	server := apphttp.NewServer(fmt.Sprintf(":%d", cfg.Server.Port), router, log, cfg.Server.AllowedOrigins())
	return server.Run(ctx)
}

// buildAuthRegistry discovers the configured identity providers. With auth disabled it
// returns an empty registry, so the sign-in routes answer 404 instead of the server
// refusing to start without provider credentials.
func buildAuthRegistry(ctx context.Context, cfg *config.Config, log logging.Logger) (*auth.Registry, error) {
	if !cfg.Auth.Enabled {
		log.Warn(ctx, "authentication is disabled; the API is unauthenticated")
		return auth.NewEmptyRegistry(), nil
	}

	configs := make([]auth.ProviderConfig, 0, len(cfg.Auth.Providers))
	for slug, provider := range cfg.Auth.Providers {
		configs = append(configs, auth.ProviderConfig{
			Slug:         slug,
			Issuer:       provider.Issuer,
			ClientID:     provider.ClientID,
			ClientSecret: provider.ClientSecret,
			RedirectURL:  provider.RedirectURL,
		})
	}

	registry, err := auth.NewRegistry(ctx, configs)
	if err != nil {
		return nil, fmt.Errorf("build auth registry: %w", err)
	}
	log.Info(ctx, "authentication enabled", "providers", registry.Slugs())
	return registry, nil
}

func buildApplication(
	cfg *config.Config,
	log logging.Logger,
	bus eventbus.Bus,
	db *storage.DB,
) *application {
	accountStore := storage.NewSQLXAccountStore(db)
	authStore := storage.NewSQLXAuthStore(db)
	vendorStore := storage.NewSQLXVendorStore(db)
	cashflowStore := storage.NewSQLXCashflowStore(db)
	portfolioStore := storage.NewSQLXPortfolioStore(db)
	assetsStore := storage.NewSQLXAssetsStore(db)
	importerStore := storage.NewSQLXImporterStore(db)
	marketDataStore := storage.NewSQLXMarketDataStore(db)
	fileStore := files.NewDisk(cfg.DiskStorage.BasePath)

	marketStackClient := marketstack.NewMarketStackClient(marketDataStore, marketdata.ProviderMarketStack)
	marketDataSyncer := marketdata.NewSyncer(marketDataStore, map[marketdata.Source]marketdata.EODFetcher{
		marketdata.SourceMarketStack: marketStackClient,
	})
	marketDataCommands := marketdata.NewCommands(marketDataStore, marketDataSyncer)
	marketDataQueries := marketdata.NewQueries(marketDataStore, marketDataSyncer)
	marketDataCatalog := marketdata.NewCatalogue(marketDataStore, map[marketdata.Source]marketdata.TickerSearcher{
		marketdata.SourceMarketStack: marketStackClient,
	})
	marketDataCreds := marketdata.NewCredentials(marketDataStore)

	accountCommands := account.NewCommands(accountStore, bus)
	accountQueries := account.NewQueries(accountStore)
	authCommands := auth.NewCommands(authStore, authStore, accountQueries, accountCommands, cfg.Auth.SessionTTL, cfg.Auth.IsEmailAllowed)
	authQueries := auth.NewQueries(authStore, accountQueries)
	vendorCommands := vendor.NewCommands(vendorStore)
	vendorQueries := vendor.NewQueries(vendorStore)
	cashflowCommands := cashflow.NewCommands(cashflowStore, cashflowStore, accountQueries)
	cashflowQueries := cashflow.NewQueries(cashflowStore)
	portfolioCommands := portfolio.NewCommands(portfolioStore, *marketDataQueries, *vendorQueries)
	portfolioQueries := portfolio.NewQueries(portfolioStore)
	portfolioBuilder := portfolio.NewBuilder(marketDataQueries, portfolioStore, portfolioStore, portfolioStore, portfolioStore, bus)
	assetsQueries := assets.NewQueries(assetsStore)
	assetsBuilder := assets.NewBuilder(assetsStore, assetsStore)
	assetsSyncer := assets.NewSyncer(portfolioQueries, assetsBuilder, assetsStore, assetsStore)
	assetsCommands := assets.NewCommands(assetsStore, assetsStore, *accountQueries, assetsStore, assetsStore, bus)
	fileQueries := files.NewQueries(fileStore)

	cashflowProcessor := importercashflow.NewProcessor(vendorQueries, fileQueries, cashflowparsers.CreateCsvParser, cashflowCommands)
	portfolioProcessor := importerportfolio.NewProcessor(vendorQueries, fileQueries, portfolioparsers.CreateCsvParser, portfolioCommands)
	eodProcessor := importereod.NewProcessor(marketDataQueries, fileQueries, eodparsers.CreateEODParser, marketDataCommands)
	importerCommands := importer.NewCommands(
		importerStore,
		fileStore,
		fileStore,
		*vendorQueries,
		*accountQueries,
		*marketDataQueries,
		bus,
		importer.WithProcessors(cashflowProcessor, portfolioProcessor, eodProcessor),
	)

	return &application{
		log: log,
		hub: notify.NewHub(log),

		accountCommands: accountCommands,
		accountQueries:  accountQueries,
		authCommands:    authCommands,
		authQueries:     authQueries,
		authStore:       authStore,
		authCookies: apphttp.CookieSettings{
			Domain: cfg.Auth.CookieDomain,
			Secure: cfg.Auth.IsCookieSecure(cfg.Server.Environment),
		},
		frontendURL:    cfg.Auth.FrontendURL,
		vendorCommands: vendorCommands,
		vendorQueries:  vendorQueries,

		cashflowCommands: cashflowCommands,
		cashflowQueries:  cashflowQueries,

		portfolioCommands: portfolioCommands,
		portfolioQueries:  portfolioQueries,
		portfolioBuilder:  portfolioBuilder,
		portfolioStore:    portfolioStore,

		assetsCommands: assetsCommands,
		assetsQueries:  assetsQueries,
		assetsBuilder:  assetsBuilder,
		assetsSyncer:   assetsSyncer,

		importerCommands:   importerCommands,
		marketDataCommands: marketDataCommands,
		marketDataQueries:  marketDataQueries,
		marketDataCatalog:  marketDataCatalog,
		marketDataCreds:    marketDataCreds,
	}
}

func registerEventHandlers(bus eventbus.Bus, app *application) error {
	if err := subscribe(bus, account.TopicCreated, portfolioevents.NewAccountCreatedHandler(app.portfolioCommands, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, account.TopicCreated, assetsevents.NewAccountCreatedHandler(app.assetsCommands, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, importer.TopicAccepted, importerevents.NewAcceptedHandler(app.importerCommands, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, importer.TopicCompleted, portfolioevents.NewImportCompletedHandler(app.portfolioBuilder, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, importer.TopicCompleted, notify.NewImportCompletedHandler(app.hub).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, portfolio.TopicRebuilt, assetsevents.NewPortfolioRebuiltHandler(app.assetsSyncer, app.assetsBuilder, bus, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, portfolio.TopicRebuilt, notify.NewPortfolioRebuiltHandler(app.hub).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, assets.TopicSnapshotsRebuildRequested, assetsevents.NewSnapshotsRebuildRequestedHandler(app.assetsBuilder, bus, app.log).Handle); err != nil {
		return err
	}
	if err := subscribe(bus, assets.TopicSnapshotsRebuilt, notify.NewAssetsSnapshotsRebuiltHandler(app.hub).Handle); err != nil {
		return err
	}
	return nil
}

func subscribe[T any](bus eventbus.Bus, topic string, handler eventbus.Handler[T]) error {
	if _, err := eventbus.Subscribe(bus, topic, handler); err != nil {
		return fmt.Errorf("subscribe %s: %w", topic, err)
	}
	return nil
}

// registerRoutes binds every route through one of three tiers. The default is
// `protected`, so a route added without thought requires a session rather than silently
// being public; `public` and `adminOnly` are the deliberate exceptions.
func registerRoutes(router *apphttp.Router, app *application) {
	withRequestLogging := apphttp.WithRequestLogging(app.log)
	withAuth := apphttp.WithAuthentication(app.authQueries, app.authFallback, app.log)

	public := func(pattern string, handler http.Handler) {
		router.HandleWithMiddleware(pattern, handler, withRequestLogging)
	}
	protected := func(pattern string, handler http.Handler) {
		router.HandleWithMiddleware(pattern, handler, withRequestLogging, withAuth)
	}
	adminOnly := func(pattern string, handler http.Handler) {
		router.HandleWithMiddleware(pattern, apphttp.RequireAdmin(handler), withRequestLogging, withAuth)
	}

	// Unauthenticated by necessity: health checks, API docs, and the sign-in endpoints
	// themselves. /auth/me and /auth/logout resolve the session on their own so they can
	// answer "not signed in" rather than being rejected by middleware.
	public("GET /health", basehandlers.HealthHandler(app.log))
	public("GET /swagger/", httpSwagger.WrapHandler)
	public("GET /auth/providers", authhttp.Providers(app.authRegistry))
	public("GET /auth/me", authhttp.Me(app.log, app.authQueries))
	public("POST /auth/logout", authhttp.Logout(app.log, app.authCommands, app.authCookies))
	public("GET /auth/{provider}/login", authhttp.Login(app.log, app.authRegistry, app.authCookies, app.frontendURL))
	public("GET /auth/{provider}/callback", authhttp.Callback(app.log, app.authRegistry, app.authCommands, app.authCookies, app.frontendURL))

	protected("GET /ws/accounts/{account_id}", app.hub.Handler())

	// Account records describe every user, so listing or creating them is an
	// administrator's job. Ordinary clients learn their own account from /auth/me.
	adminOnly("GET /accounts", accounthttp.List(app.log, app.accountQueries))
	adminOnly("POST /accounts", accounthttp.Create(app.log, *app.accountCommands))

	protected("GET /vendors", vendorshttp.List(app.log, app.vendorQueries))

	protected("POST /imports/cashflow", importerhttp.ImportCashflow(app.log, app.importerCommands))
	protected("POST /imports/portfolio", importerhttp.ImportPortfolio(app.log, app.importerCommands))
	adminOnly("POST /imports/eod", importerhttp.ImportEOD(app.log, app.importerCommands))

	// Market data is shared reference data rather than account data: everyone reads it,
	// only administrators curate it.
	protected("GET /marketdata/listings", marketdatahttp.GetListings(app.log, app.marketDataQueries))
	protected("GET /marketdata/listings/search", marketdatahttp.SearchListings(app.log, app.marketDataQueries))
	protected("GET /marketdata/eods", marketdatahttp.GetEOD(app.log, app.marketDataQueries))
	adminOnly("POST /marketdata/listing", marketdatahttp.CreateListing(app.log, app.marketDataCommands))
	adminOnly("PATCH /marketdata/listing", marketdatahttp.UpdateListingFields(app.log, app.marketDataCommands))
	adminOnly("POST /marketdata/catalogue/search", marketdatahttp.SearchProviderCatalogue(app.log, app.marketDataCatalog, app.marketDataQueries))
	adminOnly("POST /marketdata/catalogue/sync", marketdatahttp.StartCatalogueSync(app.log, app.marketDataCatalog))
	adminOnly("GET /marketdata/catalogue/status", marketdatahttp.GetCatalogueStatus(app.log, app.marketDataQueries))
	adminOnly("GET /marketdata/providers", marketdatahttp.GetProviderCredentials(app.log, app.marketDataCreds))
	adminOnly("PATCH /marketdata/providers/{provider_id}/credentials", marketdatahttp.UpdateProviderCredentials(app.log, app.marketDataCreds))
	adminOnly("POST /marketdata/providers/{provider_id}/credentials/reveal", marketdatahttp.RevealProviderAPIKey(app.log, app.marketDataCreds))

	protected("GET /cashflow/transactions", cashflowhttp.GetTransactions(app.log, app.cashflowQueries))
	protected("POST /cashflow/transactions/manual", cashflowhttp.CreateTransactions(app.log, app.cashflowCommands))
	protected("GET /cashflow/analytics/monthly", cashflowhttp.GetMonthlyAnalytics(app.log, app.cashflowQueries))
	protected("GET /cashflow/analytics/tags", cashflowhttp.GetCashflowTagDistribution(app.log, app.cashflowQueries))
	protected("POST /cashflow/transactions/tag", cashflowhttp.TagTransaction(app.log, app.cashflowCommands))
	protected("POST /cashflow/transactions/tag/selection", cashflowhttp.TagTransactionsBySelection(app.log, app.cashflowCommands))
	protected("POST /cashflow/transactions/tag/filter", cashflowhttp.TagTransactionsByFilter(app.log, app.cashflowCommands))
	protected("POST /cashflow/transactions/ignore/selection", cashflowhttp.IgnoreTransactionsBySelection(app.log, app.cashflowCommands))
	protected("POST /cashflow/transactions/ignore/filter", cashflowhttp.IgnoreTransactionsByFilter(app.log, app.cashflowCommands))

	protected("GET /portfolio/positions", portfoliohttp.GetPortfolioPositions(app.log, app.portfolioQueries))
	protected("GET /portfolio/snapshots", portfoliohttp.GetPortfolioSnapshots(app.log, app.accountQueries, app.portfolioQueries))
	protected("GET /portfolio/transactions", portfoliohttp.GetPortfolioTransactions(app.log, app.accountQueries, app.portfolioStore))
	protected("POST /portfolio/transactions/manual", portfoliohttp.CreateManualPortfolioTransaction(app.log, app.portfolioCommands))
	protected("POST /portfolio/rebuild", portfoliohttp.RebuildPortfolio(app.log, app.portfolioBuilder))

	protected("GET /assets/classes", assethttp.GetClasses(app.log, *app.assetsQueries))
	protected("POST /assets/classes", assethttp.CreateClass(app.log, *app.assetsCommands))
	protected("PATCH /assets/classes", assethttp.UpdateClass(app.log, *app.assetsCommands))
	protected("GET /assets/classes/{class_id}", assethttp.GetClassDetails(app.log, *app.assetsQueries))
	protected("DELETE /assets/classes/{class_id}", assethttp.DeleteClass(app.log, *app.assetsCommands))
	protected("POST /assets", assethttp.CreateAsset(app.log, *app.assetsCommands))
	protected("PUT /assets/{asset_id}/worth", assethttp.SetAssetWorth(app.log, *app.assetsCommands))
	protected("PUT /assets/{asset_id}/adjust", assethttp.AdjustAssetWorth(app.log, *app.assetsCommands))
	protected("GET /assets/snapshots", assethttp.GetSnapshots(app.log, *app.assetsQueries))
}

func closeDB(log logging.Logger, db *storage.DB) {
	if err := db.Close(); err != nil {
		log.Warn(context.Background(), "failed closing database", "error", err.Error())
	}
}

func closeBus(log logging.Logger, bus eventbus.Bus) {
	if err := bus.Close(); err != nil {
		log.Warn(context.Background(), "failed closing event bus", "error", err.Error())
	}
}

func closeHub(log logging.Logger, hub *notify.Hub) {
	if err := hub.Close(); err != nil {
		log.Warn(context.Background(), "failed closing websocket hub", "error", err.Error())
	}
}
