package baseinterfaces

import (
	"math"
	"strconv"
	"time"

	"github.com/glodb/keel/models/genericmodels"
	"github.com/glodb/keel/models/paginationmodels/arraypagination"
	"github.com/glodb/keel/models/paginationmodels/paginationinterface"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/customtypes"

	"github.com/gin-gonic/gin"
)

type FunctionOptions struct {
	sort       bool
	limitSet   bool
	pageSet    bool
	projection bool
	timeRange  bool
	sortLocal  bool
	sortInt    int
	sortKey    string
	limit      int
	page       int
	startTime  int64
	endTime    int64

	filter       bool
	timeRangeKey string
	sortMap      customtypes.M
	projectedMap customtypes.M
	timeRangeMap customtypes.M
}

// CreatePaginationOptions timeKey is optional
// zero index is for range key name if it's not present in the query
func CreatePaginationOptions(c *gin.Context, timeKey ...string) FunctionOptions {

	pageInt := 1
	if c.Query("page") != "" {
		if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
			pageInt = p
		}
	}

	sortInt := -1
	if c.Query("sort") != "" {
		sortInt, _ = strconv.Atoi(c.Query("sort"))
	}

	// Set the default page size limit from the configuration manager
	limitInt := configmanager.GetInstance().PageSize

	// If a limit is provided, convert the string limit to an integer
	if c.Query("limit") != "" {
		limitInt, _ = strconv.Atoi(c.Query("limit"))
	}

	// Ensure the limit does not exceed the maximum page size from the configuration
	if limitInt > configmanager.GetInstance().MaxPageSize {
		limitInt = configmanager.GetInstance().MaxPageSize
	}

	// Set the page and limit options for pagination

	sortValue := c.Query("sortValue")

	timeRangeKey := c.Query("rangeKey")
	if timeRangeKey == "" && len(timeKey) > 0 {
		timeRangeKey = timeKey[0]
	}
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")

	startTimeInt := int64(0)
	endTimeInt := int64(0)

	if startTime != "" {
		startTimeInt, _ = strconv.ParseInt(startTime, 10, 64)
	}

	if endTime != "" {
		endTimeInt, _ = strconv.ParseInt(endTime, 10, 64)
	}

	return createOptions(pageInt, limitInt, sortInt, sortValue, timeRangeKey, startTimeInt, endTimeInt)

}

func ConvertGetOptionsToPaginationOptions(c *gin.Context, inter paginationinterface.PaginationInterface, timeKey ...string) paginationinterface.PaginationInterface {

	pageInt := 1
	if c.Query("page") != "" {
		pageInt, _ = strconv.Atoi(c.Query("page"))
	}

	sortInt := -1
	if c.Query("sort") != "" {
		sortInt, _ = strconv.Atoi(c.Query("sort"))
	}

	// Set the default page size limit from the configuration manager
	limitInt := configmanager.GetInstance().PageSize

	// If a limit is provided, convert the string limit to an integer
	if c.Query("limit") != "" {
		limitInt, _ = strconv.Atoi(c.Query("limit"))
	}

	// Ensure the limit does not exceed the maximum page size from the configuration
	if limitInt > configmanager.GetInstance().MaxPageSize {
		limitInt = configmanager.GetInstance().MaxPageSize
	}

	// Set the page and limit options for pagination

	sortValue := c.Query("sortValue")

	timeRangeKey := c.Query("rangeKey")
	if timeRangeKey == "" && len(timeKey) > 0 {
		timeRangeKey = timeKey[0]
	}
	startTime := c.Query("startTime")
	endTime := c.Query("endTime")

	startTimeInt := int64(0)
	endTimeInt := int64(0)

	if startTime != "" {
		startTimeInt, _ = strconv.ParseInt(startTime, 10, 64)
	}

	if endTime != "" {
		endTimeInt, _ = strconv.ParseInt(endTime, 10, 64)
	}
	(inter).Set(pageInt, limitInt, sortInt, sortValue, timeRangeKey, startTimeInt, endTimeInt)

	return inter
}

func CreateInterfacePagination(c paginationinterface.PaginationInterface) FunctionOptions {
	limit := c.GetLimit()
	page := c.GetPage()

	if limit == 0 {
		limit = int(configmanager.GetInstance().PageSize)
	}

	if page == 0 {
		page = 1
	}

	return createOptions(page, limit, c.GetSort(), c.GetSortValue(), c.GetRangeKey(), c.GetStartTime(), c.GetEndTime())
}

func createOptions(page, limit, sort int, sortValue, timeRangeKey string, startTime, endTime int64) FunctionOptions {
	options := FunctionOptions{}

	// Set the page and limit options for pagination
	options.SetPage(page)
	options.SetLimit(limit)
	options.SetStartTime(startTime)
	options.SetEndTime(endTime)
	options.SetSortLocal(sort, sortValue)

	if sortValue != "" {
		options.SetSort(customtypes.M{sortValue: sort})
	} else {
		if sort != 0 {
			options.SetSort(customtypes.M{configmanager.GetInstance().SortKey: sort})
		}
	}

	if startTime > 0 {
		options.SetTimeRange(timeRangeKey, startTime, endTime)
	}

	return options
}

func (u *FunctionOptions) SetLimit(limit int) {
	u.limit = limit
	u.limitSet = true
}

func (u *FunctionOptions) SetPage(page int) {
	u.page = page
	u.pageSet = true
}

func (u *FunctionOptions) SetStartTime(startTime int64) {
	u.startTime = startTime
}

func (u *FunctionOptions) SetEndTime(endTime int64) {
	u.endTime = endTime
}

func (u *FunctionOptions) GetStartTime() int64 {
	return u.startTime
}

func (u *FunctionOptions) GetEndTime() int64 {
	return u.endTime
}
func (u *FunctionOptions) SetSort(sortedMap customtypes.M) {
	u.sort = true
	u.sortMap = sortedMap
}

func (u *FunctionOptions) SetSortLocal(sortInt int, sortKey string) {
	u.sortLocal = true
	u.sortInt = sortInt
	u.sortKey = sortKey
}

func (u *FunctionOptions) SetProjection(projectedMap customtypes.M) {
	u.projection = true
	u.projectedMap = projectedMap
}

func (u *FunctionOptions) GetLimit() int {
	if u.limitSet {
		return u.limit
	}
	return int(configmanager.GetInstance().PageSize)
}

func (u *FunctionOptions) IsLimitSet() bool {
	return u.limitSet
}

func (u *FunctionOptions) GetPage() int {
	if u.pageSet {
		return u.page
	}
	return 1
}

func (u *FunctionOptions) IsSortSet() bool {
	return u.sort
}

func (u *FunctionOptions) GetSort() customtypes.M {
	return u.sortMap
}

func (u *FunctionOptions) IsProjectionSet() bool {
	return u.projection
}

func (u *FunctionOptions) GetProjection() customtypes.M {
	return u.projectedMap
}

func (u *FunctionOptions) IsTimeRangeSet() bool {
	return u.timeRange
}

func (u *FunctionOptions) SetTimeRange(key string, startTime int64, endTime int64) {
	u.timeRange = true
	u.timeRangeKey = key

	if endTime <= 0 {
		endTime = time.Now().Unix()
	}

	if u.timeRangeKey == "" {
		u.timeRangeKey = configmanager.GetInstance().TimeRangeKey
	}

	u.timeRangeMap = customtypes.M{
		"$gte": startTime,
		"$lte": endTime,
	}
}

func (u *FunctionOptions) GetTimeRange() customtypes.M {
	return u.timeRangeMap
}

func (u *FunctionOptions) GetTimeRangeKey() string {
	return u.timeRangeKey
}

func (u *FunctionOptions) PaginateArray(data []arraypagination.ArrayPaginationInterface) genericmodels.PaginationResults {

	filteredData := []arraypagination.ArrayPaginationInterface{}
	if u.page == 0 {
		u.page = 1
	}

	firstIndex := (u.GetPage() - 1) * u.GetLimit()
	if len(data) == 0 {
		return genericmodels.PaginationResults{
			Pagination: genericmodels.Pagination{
				TotalDocuments: 0,
				TotalPages:     0,
				CurrentPage:    0,
				Limit:          0,
			},
		}
	}

	if firstIndex > len(data) {
		return genericmodels.PaginationResults{
			Pagination: genericmodels.Pagination{
				TotalDocuments: int64(len(data)),
				TotalPages:     int64(math.Ceil(float64(len(data)) / float64(u.GetLimit()))),
				CurrentPage:    int64(u.GetPage()),
				Limit:          int64(u.GetLimit()),
			},
		}
	}

	if u.timeRangeKey == "" {
		u.timeRangeKey = configmanager.GetInstance().TimeRangeKey
	}

	if u.startTime > 0 || u.endTime > 0 {
		for _, item := range data {
			if u.startTime > 0 {
				if item.GetTimeValue(u.timeRangeKey) < u.startTime {
					continue
				}
			}
			if u.endTime > 0 {
				if item.GetTimeValue(u.timeRangeKey) > u.endTime {
					continue
				}
			}
			filteredData = append(filteredData, item)
		}
	} else {
		filteredData = data
	}

	if u.sortLocal {
		if u.sortKey == "" {
			u.sortKey = configmanager.GetInstance().SortKey
		}
		arraypagination.SortSlice(filteredData, u.sortInt, u.sortKey)
	} else {
		arraypagination.SortSlice(filteredData, -1, configmanager.GetInstance().SortKey)
	}

	lastIndex := firstIndex + u.GetLimit()
	if lastIndex > len(filteredData) {
		lastIndex = len(filteredData)
	}
	dataToSend := filteredData[firstIndex:lastIndex]

	return genericmodels.PaginationResults{
		Pagination: genericmodels.Pagination{
			TotalDocuments: int64(len(data)),
			TotalPages:     int64(math.Ceil(float64(len(data)) / float64(u.GetLimit()))),
			CurrentPage:    int64(u.GetPage()),
			Limit:          int64(u.GetLimit()),
		},
		Data: dataToSend,
	}
}
