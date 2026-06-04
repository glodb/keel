package arraypagination

import "sort"

type ArrayPaginationInterface interface {
	GetSortKeyData(key string) any
	GetTimeValue(timeKey string) int64
}

func SortSlice[T ArrayPaginationInterface](s []T, sortType int, sortKey string) {
	sort.Slice(s, func(i, j int) bool {
		ki := s[i].GetSortKeyData(sortKey)
		kj := s[j].GetSortKeyData(sortKey)

		var ascending bool
		switch vi := ki.(type) {
		case int:
			vj, ok := kj.(int)
			if !ok {
				panic("mixed key types: int vs non-int")
			}
			ascending = vi < vj

		case float64:
			vj, ok := kj.(float64)
			if !ok {
				panic("mixed key types: float64 vs non-float64")
			}
			ascending = vi < vj

		case string:
			vj, ok := kj.(string)
			if !ok {
				panic("mixed key types: string vs non-string")
			}
			ascending = vi < vj

		default:
			return false
		}

		// If sortType is -1 (descending), reverse the comparison
		if sortType == -1 {
			return !ascending
		}
		return ascending
	})
}
