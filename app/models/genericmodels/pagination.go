package genericmodels

import "math"

type Pagination struct {
	TotalDocuments int64 `json:"total_documents"`
	TotalPages     int64 `json:"total_pages"`
	CurrentPage    int64 `json:"current_page"`
	Limit          int64 `json:"limit"`
}

func NewPagination(count int64, pageSize int, pageNumber int) (pr Pagination) {
	pr.TotalDocuments = count
	pr.TotalPages = int64(math.Ceil((float64(count) / float64(pageSize))))
	pr.Limit = int64(pageSize)
	pr.CurrentPage = int64(pageNumber)
	return
}
