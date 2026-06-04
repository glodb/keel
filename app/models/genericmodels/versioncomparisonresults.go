package genericmodels

type VersionComparisonResult struct {
	IsGreaterThan bool
	IsLessThan    bool
	IsEqual       bool
	IsCompatible  bool
	Difference    string
	MajorDiff     int
	MinorDiff     int
	PatchDiff     int
}
