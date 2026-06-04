package configmodels

import (
	"time"

	"github.com/glodb/keel/settings/utilsdatatypes"
)

type RateLimitConfig struct {
	Enabled              bool                        `json:"enabled"`
	Window               time.Duration               `json:"window"` // Time window for rate limiting
	WindowData           string                      `json:"windowData"`
	MaxRequests          int                         `json:"maxRequests"`          // Maximum requests per window
	KeyPrefix            string                      `json:"keyPrefix"`            // Redis key prefix
	SkipPaths            []string                    `json:"skipPaths"`            // Paths to skip rate limiting
	SkipMethods          []string                    `json:"skipMethods"`          // HTTP methods to skip
	RateLimitSkipPaths   *utilsdatatypes.Set[string] `json:"rateLimitSkipPaths"`   // Paths to skip rate limiting
	RateLimitSkipMethods *utilsdatatypes.Set[string] `json:"rateLimitSkipMethods"` // HTTP methods to skip
}

func (r *RateLimitConfig) Init() {
	r.RateLimitSkipMethods = utilsdatatypes.NewSet[string]()
	r.RateLimitSkipPaths = utilsdatatypes.NewSet[string]()

	for _, skipPath := range r.SkipPaths {
		r.RateLimitSkipPaths.Add(skipPath)
	}

	for _, skipMethod := range r.SkipMethods {
		r.RateLimitSkipMethods.Add(skipMethod)
	}
}
