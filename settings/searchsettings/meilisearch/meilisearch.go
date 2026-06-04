package meilisearch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/searchsettings/search"

	meili "github.com/meilisearch/meilisearch-go"
)

type MeilisearchClient struct {
	client meili.ServiceManager
}

var getInstance = sync.OnceValue(func() *MeilisearchClient {
	instance := &MeilisearchClient{}

	config := configmanager.GetInstance().MeilisearchConfig
	if config.Host == "" {
		logger.Log().Error("Meilisearch host not configured")
		return nil
	}

	client := meili.New(config.Host, meili.WithAPIKey(config.ApiKey))

	instance.client = client
	logger.Log().Info("Meilisearch client initialized", logger.StringField("host", config.Host))

	// Register this instance as the global search implementation
	search.SetSearch(instance)
	logger.Log().Info("Meilisearch registered as global search instance")

	return instance
})

func GetInstance() *MeilisearchClient {
	return getInstance()
}

// Health checks if Meilisearch is healthy
func (m *MeilisearchClient) Health(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	if !m.client.IsHealthy() {
		return fmt.Errorf("meilisearch is not healthy")
	}
	return nil
}

// CreateIndex creates a new index with the given name and primary key
func (m *MeilisearchClient) CreateIndex(ctx context.Context, indexName string, primaryKey string) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	task, err := m.client.CreateIndexWithContext(ctx, &meili.IndexConfig{
		Uid:        indexName,
		PrimaryKey: primaryKey,
	})

	if err != nil {
		logger.Log().Error("Failed to create index",
			logger.StringField("index", indexName),
			logger.ErrorField("error", err))
		return err
	}

	// Wait for the task to complete
	if task != nil {
		_, err = m.client.WaitForTask(task.TaskUID, 5000*time.Millisecond)
		if err != nil {
			logger.Log().Error("Failed to wait for index creation task", logger.ErrorField("error", err))
			return err
		}
	}

	logger.Log().Info("Index created successfully", logger.StringField("index", indexName))
	return nil
}

// DeleteIndex deletes an index
func (m *MeilisearchClient) DeleteIndex(ctx context.Context, indexName string) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	task, err := m.client.DeleteIndexWithContext(ctx, indexName)
	if err != nil {
		logger.Log().Error("Failed to delete index",
			logger.StringField("index", indexName),
			logger.ErrorField("error", err))
		return err
	}

	if task != nil {
		_, err = m.client.WaitForTask(task.TaskUID, 5000*time.Millisecond)
	}

	return err
}

// GetIndex retrieves an index
func (m *MeilisearchClient) GetIndex(ctx context.Context, indexName string) (interface{}, error) {
	if m.client == nil {
		return nil, fmt.Errorf("meilisearch client not initialized")
	}

	return m.client.GetIndexWithContext(ctx, indexName)
}

// ListIndexes lists all indexes
func (m *MeilisearchClient) ListIndexes(ctx context.Context) ([]interface{}, error) {
	if m.client == nil {
		return nil, fmt.Errorf("meilisearch client not initialized")
	}

	// Note: meilisearch-go doesn't have a direct GetIndexes method
	// We would need to implement this differently or return empty
	return nil, fmt.Errorf("list indexes not implemented")
}

// AddDocuments adds documents to an index
func (m *MeilisearchClient) AddDocuments(ctx context.Context, indexName string, documents interface{}) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	index := m.client.Index(indexName)
	task, err := index.AddDocumentsWithContext(ctx, documents)
	if err != nil {
		logger.Log().Error("Failed to add documents",
			logger.StringField("index", indexName),
			logger.ErrorField("error", err))
		return err
	}

	if task != nil {
		taskResult, err := m.client.WaitForTask(task.TaskUID, 10000*time.Millisecond)
		if err != nil {
			logger.Log().Error("Failed to wait for add documents task",
				logger.StringField("index", indexName),
				logger.Int64Field("taskUID", task.TaskUID),
				logger.ErrorField("error", err))
			return err
		}

		// Check if task actually succeeded
		if taskResult.Status == meili.TaskStatusFailed {
			logger.Log().Error("MeiliSearch task failed",
				logger.StringField("index", indexName),
				logger.Int64Field("taskUID", task.TaskUID),
				logger.StringField("status", string(taskResult.Status)),
				logger.AnyField("error", taskResult.Error))
			return fmt.Errorf("meilisearch task failed: %v", taskResult.Error)
		}

		logger.Log().Debug("Documents added successfully",
			logger.StringField("index", indexName),
			logger.Int64Field("taskUID", task.TaskUID),
			logger.StringField("status", string(taskResult.Status)))
	}

	return nil
}

// UpdateDocuments updates documents in an index
func (m *MeilisearchClient) UpdateDocuments(ctx context.Context, indexName string, documents interface{}) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	index := m.client.Index(indexName)
	task, err := index.UpdateDocumentsWithContext(ctx, documents)
	if err != nil {
		logger.Log().Error("Failed to update documents",
			logger.StringField("index", indexName),
			logger.ErrorField("error", err))
		return err
	}

	if task != nil {
		taskResult, err := m.client.WaitForTask(task.TaskUID, 10000*time.Millisecond)
		if err != nil {
			logger.Log().Error("Failed to wait for update documents task",
				logger.StringField("index", indexName),
				logger.Int64Field("taskUID", task.TaskUID),
				logger.ErrorField("error", err))
			return err
		}

		// Check if task actually succeeded
		if taskResult.Status == meili.TaskStatusFailed {
			logger.Log().Error("MeiliSearch update task failed",
				logger.StringField("index", indexName),
				logger.Int64Field("taskUID", task.TaskUID),
				logger.StringField("status", string(taskResult.Status)),
				logger.AnyField("error", taskResult.Error))
			return fmt.Errorf("meilisearch update task failed: %v", taskResult.Error)
		}

		logger.Log().Debug("Documents updated successfully",
			logger.StringField("index", indexName),
			logger.Int64Field("taskUID", task.TaskUID),
			logger.StringField("status", string(taskResult.Status)))
	}

	return nil
}

// DeleteDocument deletes a single document
func (m *MeilisearchClient) DeleteDocument(ctx context.Context, indexName string, documentID string) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	index := m.client.Index(indexName)
	task, err := index.DeleteDocumentWithContext(ctx, documentID)
	if err != nil {
		logger.Log().Error("Failed to delete document",
			logger.StringField("index", indexName),
			logger.StringField("documentID", documentID),
			logger.ErrorField("error", err))
		return err
	}

	if task != nil {
		_, err = m.client.WaitForTask(task.TaskUID, 5000*time.Millisecond)
	}

	return err
}

// DeleteDocuments deletes multiple documents
func (m *MeilisearchClient) DeleteDocuments(ctx context.Context, indexName string, documentIDs []string) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	index := m.client.Index(indexName)
	task, err := index.DeleteDocumentsWithContext(ctx, documentIDs)
	if err != nil {
		logger.Log().Error("Failed to delete documents",
			logger.StringField("index", indexName),
			logger.ErrorField("error", err))
		return err
	}

	if task != nil {
		_, err = m.client.WaitForTask(task.TaskUID, 5000*time.Millisecond)
	}

	return err
}

// DeleteAllDocuments deletes all documents in an index
func (m *MeilisearchClient) DeleteAllDocuments(ctx context.Context, indexName string) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	index := m.client.Index(indexName)
	task, err := index.DeleteAllDocumentsWithContext(ctx)
	if err != nil {
		logger.Log().Error("Failed to delete all documents",
			logger.StringField("index", indexName),
			logger.ErrorField("error", err))
		return err
	}

	if task != nil {
		_, err = m.client.WaitForTask(task.TaskUID, 10000*time.Millisecond)
	}

	return err
}

// GetDocument retrieves a single document
func (m *MeilisearchClient) GetDocument(ctx context.Context, indexName string, documentID string) (map[string]interface{}, error) {
	if m.client == nil {
		return nil, fmt.Errorf("meilisearch client not initialized")
	}

	index := m.client.Index(indexName)
	var document map[string]interface{}
	err := index.GetDocumentWithContext(ctx, documentID, nil, &document)

	return document, err
}

// Search performs a search query
func (m *MeilisearchClient) Search(ctx context.Context, indexName string, query string, options *search.SearchOptions) (*search.SearchResponse, error) {
	if m.client == nil {
		return nil, fmt.Errorf("meilisearch client not initialized")
	}

	index := m.client.Index(indexName)

	// Build search request
	searchRequest := &meili.SearchRequest{
		Query: query,
	}

	if options != nil {
		if options.Limit > 0 {
			limit := int64(options.Limit)
			searchRequest.Limit = limit
		}
		if options.Offset > 0 {
			offset := int64(options.Offset)
			searchRequest.Offset = offset
		}
		if options.Filter != "" {
			searchRequest.Filter = options.Filter
		}
		if len(options.Sort) > 0 {
			searchRequest.Sort = options.Sort
		}
		if len(options.AttributesToRetrieve) > 0 {
			searchRequest.AttributesToRetrieve = options.AttributesToRetrieve
		}
		if len(options.AttributesToSearchOn) > 0 {
			searchRequest.AttributesToSearchOn = options.AttributesToSearchOn
		}
	}

	// Perform search
	result, err := index.SearchWithContext(ctx, query, searchRequest)
	if err != nil {
		logger.Log().Error("Search failed",
			logger.StringField("index", indexName),
			logger.StringField("query", query),
			logger.ErrorField("error", err))
		return nil, err
	}

	// Convert result to our response format
	// Convert []interface{} to []map[string]interface{}
	hits := make([]map[string]interface{}, 0, len(result.Hits))
	for _, hit := range result.Hits {
		if hitMap, ok := hit.(map[string]interface{}); ok {
			hits = append(hits, hitMap)
		}
	}

	response := &search.SearchResponse{
		Hits:               hits,
		EstimatedTotalHits: result.EstimatedTotalHits,
		Offset:             int(result.Offset),
		Limit:              int(result.Limit),
		ProcessingTimeMs:   result.ProcessingTimeMs,
	}

	return response, nil
}

// UpdateFilterableAttributes updates filterable attributes for an index
func (m *MeilisearchClient) UpdateFilterableAttributes(ctx context.Context, indexName string, attributes []string) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	index := m.client.Index(indexName)
	task, err := index.UpdateFilterableAttributesWithContext(ctx, &attributes)
	if err != nil {
		logger.Log().Error("Failed to update filterable attributes",
			logger.StringField("index", indexName),
			logger.ErrorField("error", err))
		return err
	}

	if task != nil {
		_, err = m.client.WaitForTask(task.TaskUID, 5000*time.Millisecond)
	}

	return err
}

// UpdateSortableAttributes updates sortable attributes for an index
func (m *MeilisearchClient) UpdateSortableAttributes(ctx context.Context, indexName string, attributes []string) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	index := m.client.Index(indexName)
	task, err := index.UpdateSortableAttributesWithContext(ctx, &attributes)
	if err != nil {
		logger.Log().Error("Failed to update sortable attributes",
			logger.StringField("index", indexName),
			logger.ErrorField("error", err))
		return err
	}

	if task != nil {
		_, err = m.client.WaitForTask(task.TaskUID, 5000*time.Millisecond)
	}

	return err
}

// UpdateSearchableAttributes updates searchable attributes for an index
func (m *MeilisearchClient) UpdateSearchableAttributes(ctx context.Context, indexName string, attributes []string) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	index := m.client.Index(indexName)
	task, err := index.UpdateSearchableAttributesWithContext(ctx, &attributes)
	if err != nil {
		logger.Log().Error("Failed to update searchable attributes",
			logger.StringField("index", indexName),
			logger.ErrorField("error", err))
		return err
	}

	if task != nil {
		_, err = m.client.WaitForTask(task.TaskUID, 5000*time.Millisecond)
	}

	return err
}

// UpdateRankingRules updates ranking rules for an index
func (m *MeilisearchClient) UpdateRankingRules(ctx context.Context, indexName string, rules []string) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	index := m.client.Index(indexName)
	task, err := index.UpdateRankingRulesWithContext(ctx, &rules)
	if err != nil {
		logger.Log().Error("Failed to update ranking rules",
			logger.StringField("index", indexName),
			logger.ErrorField("error", err))
		return err
	}

	if task != nil {
		_, err = m.client.WaitForTask(task.TaskUID, 5000*time.Millisecond)
	}

	return err
}

// WaitForTask waits for a task to complete
func (m *MeilisearchClient) WaitForTask(ctx context.Context, taskUID int64) error {
	if m.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}

	_, err := m.client.WaitForTask(taskUID, 10000*time.Millisecond)
	return err
}
