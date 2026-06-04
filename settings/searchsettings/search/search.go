package search

import "context"

// SearchResult represents a search result with score
type SearchResult struct {
	Document interface{}
	Score    float64
}

// SearchResponse represents a paginated search response
type SearchResponse struct {
	Hits               []map[string]interface{} `json:"hits"`
	EstimatedTotalHits int64                    `json:"estimatedTotalHits"`
	Offset             int                      `json:"offset"`
	Limit              int                      `json:"limit"`
	ProcessingTimeMs   int64                    `json:"processingTimeMs"`
}

// SearchOptions represents search configuration options
type SearchOptions struct {
	Limit                int
	Offset               int
	Filter               string
	Sort                 []string
	AttributesToRetrieve []string
	AttributesToSearchOn []string
}

// Search interface defines all search operations
type Search interface {
	// Index management
	CreateIndex(ctx context.Context, indexName string, primaryKey string) error
	DeleteIndex(ctx context.Context, indexName string) error
	GetIndex(ctx context.Context, indexName string) (interface{}, error)
	ListIndexes(ctx context.Context) ([]interface{}, error)

	// Document operations
	AddDocuments(ctx context.Context, indexName string, documents interface{}) error
	UpdateDocuments(ctx context.Context, indexName string, documents interface{}) error
	DeleteDocument(ctx context.Context, indexName string, documentID string) error
	DeleteDocuments(ctx context.Context, indexName string, documentIDs []string) error
	DeleteAllDocuments(ctx context.Context, indexName string) error
	GetDocument(ctx context.Context, indexName string, documentID string) (map[string]interface{}, error)

	// Search operations
	Search(ctx context.Context, indexName string, query string, options *SearchOptions) (*SearchResponse, error)

	// Settings
	UpdateFilterableAttributes(ctx context.Context, indexName string, attributes []string) error
	UpdateSortableAttributes(ctx context.Context, indexName string, attributes []string) error
	UpdateSearchableAttributes(ctx context.Context, indexName string, attributes []string) error
	UpdateRankingRules(ctx context.Context, indexName string, rules []string) error

	// Utility
	Health(ctx context.Context) error
	WaitForTask(ctx context.Context, taskUID int64) error
}
