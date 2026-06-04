package genericmodels

import "time"

type RateLimitInfo struct {
	Remaining int
	Reset     time.Time
	Limit     int
	Window    time.Duration
}
