package importer

import "strings"

// DedupSequencer numbers rows that carry identical content within one import file.
//
// Monthly broker exports overlap, so a row's line number says nothing about whether
// it is new: the same transaction sits on a different line in next month's export.
// What identifies a row is its content. What separates two genuinely distinct rows
// that happen to look identical is only how often that content occurs in the same
// file. The sequencer answers the second question, and its number takes the place of
// the line number in the dedup checksum — so a row that reappears in an overlapping
// export is recognised, while a row that legitimately occurs twice in one export
// stays two rows.
type DedupSequencer struct {
	seen map[string]int
}

// NewDedupSequencer creates a sequencer scoped to one import file.
func NewDedupSequencer() *DedupSequencer {
	return &DedupSequencer{seen: make(map[string]int)}
}

// Next returns the 1-based occurrence of this content within the file. The parts are
// the fields that identify the row — date, product, amount, quantity, and the
// broker's own order reference where the export carries one.
func (s *DedupSequencer) Next(parts ...string) int {
	const sep = "\x1F"
	key := strings.Join(parts, sep)
	s.seen[key]++
	return s.seen[key]
}
