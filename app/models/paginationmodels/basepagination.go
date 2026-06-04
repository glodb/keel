package paginationmodels

type BasePagination struct {
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
	Sort      int    `json:"sort"`
	SortValue string `json:"sortValue"`
	RangeKey  string `json:"rangeKey"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}

func (bp *BasePagination) GetPage() int {
	return bp.Page
}
func (bp *BasePagination) GetLimit() int {
	return bp.Limit
}
func (bp *BasePagination) GetSort() int {
	return bp.Sort
}
func (bp *BasePagination) GetSortValue() string {
	return bp.SortValue
}
func (bp *BasePagination) GetRangeKey() string {
	return bp.RangeKey
}
func (bp *BasePagination) GetStartTime() int64 {
	return bp.StartTime
}
func (bp *BasePagination) GetEndTime() int64 {
	return bp.EndTime
}

func (bp *BasePagination) Set(page, limit, sort int, sortValue, rangeKey string, startTime, endTime int64) {
	bp.Page = page
	bp.Limit = limit
	bp.Sort = sort
	bp.SortValue = sortValue
	bp.RangeKey = rangeKey
	bp.StartTime = startTime
	bp.EndTime = endTime
}
