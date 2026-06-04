package search

import (
	"context"
	"sync"
)

var searchInstance Search
var once sync.Once

// SetSearch sets the global search instance (called during initialization)
func SetSearch(s Search) {
	once.Do(func() {
		searchInstance = s
	})
}

// GetSearch returns the global search instance
func GetSearch() Search {
	return searchInstance
}

func GetSearchContext() context.Context {
	return context.Background()
}
