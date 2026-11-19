package parsers

import "errors"

// ErrMissingHeader reports a file whose header row is not the one the vendor's export
// carries. It is the only parse failure that means "wrong file": everything else a
// parser can return — a read error, a close error — is a server-side failure the user
// can retry, and answering those with "upload a different file" sends them the wrong way.
var ErrMissingHeader = errors.New("missing required header")
