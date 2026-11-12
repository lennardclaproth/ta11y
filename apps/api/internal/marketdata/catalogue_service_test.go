package marketdata

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeCatalogueStore struct {
	CatalogueStore
	latest  map[Source]*CatalogueSync
	updated []*CatalogueSync
	created []*CatalogueSync
}

func (s *fakeCatalogueStore) LatestCatalogueSync(_ context.Context, source Source) (*CatalogueSync, error) {
	return s.latest[source], nil
}

func (s *fakeCatalogueStore) UpdateCatalogueSync(_ context.Context, sync *CatalogueSync) error {
	s.updated = append(s.updated, sync)
	return nil
}

func (s *fakeCatalogueStore) CreateCatalogueSync(_ context.Context, sync *CatalogueSync) error {
	s.created = append(s.created, sync)
	return nil
}

type idleSearcher struct{}

func (idleSearcher) SearchTickers(context.Context, string, int, int) (TickerPage, error) {
	return TickerPage{}, nil
}

func TestRecoverInterrupted_FailsRunningSeeds(t *testing.T) {
	running := &CatalogueSync{ID: uuid.New(), Source: SourceMarketStack, Status: CatalogueSyncRunning, StartedAt: time.Now()}
	store := &fakeCatalogueStore{latest: map[Source]*CatalogueSync{SourceMarketStack: running}}
	catalogue := NewCatalogue(store, map[Source]TickerSearcher{SourceMarketStack: idleSearcher{}})

	recovered, err := catalogue.RecoverInterrupted(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if recovered != 1 || len(store.updated) != 1 {
		t.Fatalf("expected one recovered run, got %d (updates %d)", recovered, len(store.updated))
	}
	if running.Status != CatalogueSyncFailed || running.FinishedAt == nil || running.LastError == nil {
		t.Fatalf("expected the run to be failed and finished, got %+v", running)
	}
	if *running.LastError != ErrCatalogueSeedInterrupted.Error() {
		t.Fatalf("unexpected last error %q", *running.LastError)
	}
}

func TestRecoverInterrupted_LeavesFinishedSeedsAlone(t *testing.T) {
	done := &CatalogueSync{ID: uuid.New(), Source: SourceMarketStack, Status: CatalogueSyncCompleted}
	store := &fakeCatalogueStore{latest: map[Source]*CatalogueSync{SourceMarketStack: done}}
	catalogue := NewCatalogue(store, map[Source]TickerSearcher{SourceMarketStack: idleSearcher{}})

	recovered, err := catalogue.RecoverInterrupted(context.Background())
	if err != nil || recovered != 0 || len(store.updated) != 0 {
		t.Fatalf("expected nothing recovered, got %d, %v, %d updates", recovered, err, len(store.updated))
	}
}

func TestStartSeed_HonoursTheConfiguredPageCap(t *testing.T) {
	store := &fakeCatalogueStore{latest: map[Source]*CatalogueSync{}}
	catalogue := NewCatalogue(store, map[Source]TickerSearcher{SourceMarketStack: idleSearcher{}}, WithMaxSeedPages(5))

	if _, err := catalogue.StartSeed(context.Background(), SourceMarketStack, 6); !errors.Is(err, ErrCatalogueSeedPagesInvalid) {
		t.Fatalf("expected the page cap to reject 6 pages, got %v", err)
	}
	if len(store.created) != 0 {
		t.Fatal("a rejected seed must not record a run")
	}
}
