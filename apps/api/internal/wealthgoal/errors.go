package wealthgoal

import "fmt"

// ErrInvalidSharePercent is returned for a share outside 0..100.
var ErrInvalidSharePercent = fmt.Errorf("share_percent must be a whole number between 0 and 100")
