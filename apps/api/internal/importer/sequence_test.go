package importer

import "testing"

func TestDedupSequencer_RepeatsContentGetTheirOwnNumber(t *testing.T) {
	sequencer := NewDedupSequencer()

	first := sequencer.Next("20260901", "PRODUCT", "-100")
	second := sequencer.Next("20260901", "PRODUCT", "-100")
	other := sequencer.Next("20260901", "OTHER", "-100")

	if first != 1 || second != 2 {
		t.Fatalf("expected identical rows to be numbered 1 and 2, got %d and %d", first, second)
	}
	if other != 1 {
		t.Fatalf("expected different content to start at 1, got %d", other)
	}
}

func TestDedupSequencer_ShiftedExportKeepsTheSameNumbers(t *testing.T) {
	// The second export drops the oldest row and adds a newer one, so every carried-over
	// row sits on a different line. Its sequence has to survive that move, or the
	// overlap comes in as new transactions.
	september := NewDedupSequencer()
	september.Next("20260801", "OLD", "-10")
	carriedOver := september.Next("20260901", "PRODUCT", "-100")

	october := NewDedupSequencer()
	sameRowAgain := october.Next("20260901", "PRODUCT", "-100")
	october.Next("20261001", "NEW", "-20")

	if carriedOver != sameRowAgain {
		t.Fatalf("expected the same row to keep sequence %d across exports, got %d", carriedOver, sameRowAgain)
	}
}

func TestDedupSequencer_PartsAreNotConcatenatedAmbiguously(t *testing.T) {
	sequencer := NewDedupSequencer()

	if got := sequencer.Next("AB", "C"); got != 1 {
		t.Fatalf("expected first occurrence, got %d", got)
	}
	if got := sequencer.Next("A", "BC"); got != 1 {
		t.Fatalf("expected differently split parts to be a different row, got %d", got)
	}
}
