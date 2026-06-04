package paginationinterface

type PaginationInterface interface {
	GetPage() int
	GetLimit() int
	GetSort() int
	GetSortValue() string
	GetRangeKey() string
	GetStartTime() int64
	GetEndTime() int64
	Set(int, int, int, string, string, int64, int64)
}
