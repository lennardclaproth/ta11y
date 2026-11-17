package memorybus_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lennardclaproth/ta11y/internal/eventbus"
	memorybus "github.com/lennardclaproth/ta11y/internal/eventbus/memory"
)

type ping struct{ N int }

func TestMemoryBusPublishDeliversToSubscriber(t *testing.T) {
	bus := memorybus.NewMemoryBus(memorybus.WithWorkers(2), memorybus.WithQueueSize(8))
	defer func() { _ = bus.Close() }()

	got := make(chan eventbus.Envelope, 1)
	sub, err := bus.Subscribe("topic.ping", func(_ context.Context, env eventbus.Envelope) error {
		got <- env
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer func() { _ = sub.Close() }()

	if err := bus.Publish(context.Background(), "topic.ping", ping{N: 7}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case env := <-got:
		p, ok := env.Payload.(ping)
		if !ok {
			t.Fatalf("unexpected payload type %T", env.Payload)
		}
		if p.N != 7 {
			t.Fatalf("unexpected payload %+v", p)
		}
		if env.CorrelationID != env.MessageID.String() {
			t.Fatalf("root event correlation should equal message id, got %q", env.CorrelationID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for delivery")
	}
}

// TestMemoryBusPropagatesCausationChain verifies that an event published from
// within a handler inherits the parent's correlation ID and is marked as caused
// by the parent message.
func TestMemoryBusPropagatesCausationChain(t *testing.T) {
	bus := memorybus.NewMemoryBus(memorybus.WithWorkers(2), memorybus.WithQueueSize(8))
	defer func() { _ = bus.Close() }()

	child := make(chan eventbus.Envelope, 1)
	childSub, err := bus.Subscribe("topic.child", func(_ context.Context, env eventbus.Envelope) error {
		child <- env
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe child: %v", err)
	}
	defer func() { _ = childSub.Close() }()

	parentSub, err := bus.Subscribe("topic.parent", func(ctx context.Context, _ eventbus.Envelope) error {
		return bus.Publish(ctx, "topic.child", ping{N: 1})
	})
	if err != nil {
		t.Fatalf("subscribe parent: %v", err)
	}
	defer func() { _ = parentSub.Close() }()

	if err := bus.Publish(context.Background(), "topic.parent", ping{N: 0}); err != nil {
		t.Fatalf("publish parent: %v", err)
	}

	select {
	case env := <-child:
		if env.CausationID == "" {
			t.Fatal("expected child to be caused by the parent message")
		}
		if env.CorrelationID == env.MessageID.String() {
			t.Fatal("child should inherit the parent's correlation, not correlate to itself")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for child delivery")
	}
}

// TestMemoryBusDrainWaitsForFollowUpEvents verifies that Drain returns only once a
// handler's follow-up publish has been handled as well.
func TestMemoryBusDrainWaitsForFollowUpEvents(t *testing.T) {
	bus := memorybus.NewMemoryBus(memorybus.WithWorkers(2), memorybus.WithQueueSize(8))
	defer func() { _ = bus.Close() }()

	var handled atomic.Int32
	if _, err := bus.Subscribe("topic.first", func(ctx context.Context, _ eventbus.Envelope) error {
		time.Sleep(50 * time.Millisecond)
		return bus.Publish(ctx, "topic.second", ping{N: 2})
	}); err != nil {
		t.Fatalf("subscribe first: %v", err)
	}
	if _, err := bus.Subscribe("topic.second", func(context.Context, eventbus.Envelope) error {
		time.Sleep(50 * time.Millisecond)
		handled.Add(1)
		return nil
	}); err != nil {
		t.Fatalf("subscribe second: %v", err)
	}

	if err := bus.Publish(context.Background(), "topic.first", ping{N: 1}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := bus.Drain(ctx); err != nil {
		t.Fatalf("drain: %v", err)
	}
	if handled.Load() != 1 {
		t.Fatalf("drain returned before the follow-up event was handled")
	}
}

// TestMemoryBusDrainHonoursDeadline verifies that Drain gives up when ctx expires.
func TestMemoryBusDrainHonoursDeadline(t *testing.T) {
	bus := memorybus.NewMemoryBus(memorybus.WithWorkers(1), memorybus.WithQueueSize(8))
	defer func() { _ = bus.Close() }()

	release := make(chan struct{})
	defer close(release)
	if _, err := bus.Subscribe("topic.slow", func(context.Context, eventbus.Envelope) error {
		<-release
		return nil
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := bus.Publish(context.Background(), "topic.slow", ping{}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := bus.Drain(ctx); err == nil {
		t.Fatal("expected drain to report the expired deadline")
	}
}
