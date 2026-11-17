package backup

import "errors"

var (
	// ErrInvalidFile is a file that is not a readable ta11y backup.
	ErrInvalidFile = errors.New("backup: invalid file")
	// ErrUnsupportedVersion is a backup written in a format version this build cannot read.
	ErrUnsupportedVersion = errors.New("backup: unsupported version")
	// ErrWrongKind is an EOD file offered as an account backup, or the other way round.
	ErrWrongKind = errors.New("backup: wrong kind of file")
	// ErrMissingReferences is a file that names vendors or listings this instance does
	// not have. Nothing is written; the result names what is missing.
	ErrMissingReferences = errors.New("backup: vendors or listings missing")
)
