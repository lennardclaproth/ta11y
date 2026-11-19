package marketdata

import (
	"context"
	"iter"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/ta11y/internal/date"
)

type fakeSyncStore struct {
	listing  *Listing
	rangeEnd *time.Time
}

func (f *fakeSyncStore) Listing(context.Context, uuid.UUID) (*Listing, error) { return f.listing, nil }
func (f *fakeSyncStore) SetAccumulatedRange(_ context.Context, _ uuid.UUID, _, to *time.Time) error {
	f.rangeEnd = to
	return nil
}
func (f *fakeSyncStore) TryAcquireSyncLock(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}
func (f *fakeSyncStore) ReleaseSyncLock(context.Context, uuid.UUID) error { return nil }
func (f *fakeSyncStore) InsertEOD(context.Context, *EOD) error            { return nil }

type fakeEODFetcher struct{ to *time.Time }

func (f *fakeEODFetcher) GetEOD(_ context.Context, _ []string, _, to *time.Time) iter.Seq2[EOD, error] {
	f.to = to
	return func(func(EOD, error) bool) {}
}

func TestSyncEOD_ClampsFutureEndToLatestBusinessDate(t *testing.T) {
	store := &fakeSyncStore{listing: &Listing{ID: uuid.New(), Symbol: "X", Source: SourceMarketStack}}
	fetcher := &fakeEODFetcher{}
	syncer := NewSyncer(store, map[Source]EODFetcher{SourceMarketStack: fetcher})

	future := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := syncer.SyncEOD(context.Background(), store.listing.ID, nil, &future); err != nil {
		t.Fatalf("SyncEOD: %v", err)
	}

	want := date.LatestBusinessDate(time.Now(), time.Local)
	if fetcher.to == nil || !fetcher.to.Equal(want) {
		t.Fatalf("fetched through %v, want %v", fetcher.to, want)
	}
	if store.rangeEnd == nil || !store.rangeEnd.Equal(want) {
		t.Fatalf("recorded accumulated end %v, want %v", store.rangeEnd, want)
	}
}
