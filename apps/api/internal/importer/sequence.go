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

// Next returns the 1-based occurrence of this content within the file. The parts have to
// be exactly the fields the row's checksum digests — callers pass the row's own DedupKey.
// A finer key separates rows the checksum cannot separate, which gives them the same
// number and therefore the same checksum, and one of them is dropped as a duplicate.
func (s *DedupSequencer) Next(parts ...string) int {
	const sep = "\x1F"
	key := strings.Join(parts, sep)
	s.seen[key]++
	return s.seen[key]
}
