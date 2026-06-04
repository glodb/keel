package versioning

import (
	"fmt"
	"sync"

	"github.com/glodb/keel/app/models/genericmodels"
	"github.com/glodb/keel/settings/cachesettings/cache"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
)

type Versioning struct {
	Version *genericmodels.Version
}

var getInstance = sync.OnceValue(func() *Versioning {
	cfg := configmanager.GetInstance()
	return &Versioning{
		Version: &cfg.ServiceVersion,
	}
})

func GetInstance() *Versioning {
	return getInstance()
}

// CompareVersions compares two versions and returns detailed comparison results
func (v *Versioning) CompareVersions(version1, version2 *genericmodels.Version) *genericmodels.VersionComparisonResult {
	result := &genericmodels.VersionComparisonResult{
		MajorDiff: version1.Major - version2.Major,
		MinorDiff: version1.Minor - version2.Minor,
		PatchDiff: version1.Patch - version2.Patch,
	}

	// Determine comparison results
	if result.MajorDiff > 0 {
		result.IsGreaterThan = true
		result.Difference = "major"
	} else if result.MajorDiff < 0 {
		result.IsLessThan = true
		result.Difference = "major"
	} else if result.MinorDiff > 0 {
		result.IsGreaterThan = true
		result.Difference = "minor"
	} else if result.MinorDiff < 0 {
		result.IsLessThan = true
		result.Difference = "minor"
	} else if result.PatchDiff > 0 {
		result.IsGreaterThan = true
		result.Difference = "patch"
	} else if result.PatchDiff < 0 {
		result.IsLessThan = true
		result.Difference = "patch"
	} else {
		result.IsEqual = true
		result.Difference = "none"
	}

	// Determine compatibility (same major version is generally compatible)
	result.IsCompatible = version1.Major == version2.Major

	return result
}

// CompareVersionStrings compares two version strings and returns comparison results
func (v *Versioning) CompareVersionStrings(version1Str, version2Str string) (*genericmodels.VersionComparisonResult, error) {
	var version1 genericmodels.Version
	err := version1.Parse(version1Str)
	if err != nil {
		return nil, fmt.Errorf("error parsing first version '%s': %w", version1Str, err)
	}

	var version2 genericmodels.Version
	err = version2.Parse(version2Str)
	if err != nil {
		return nil, fmt.Errorf("error parsing second version '%s': %w", version2Str, err)
	}

	return v.CompareVersions(&version1, &version2), nil
}

// IsCompatibleWith checks if this version is backward compatible with another version
func (v *Versioning) IsCompatibleWith(version1, version2 *genericmodels.Version) bool {
	result := v.CompareVersions(version1, version2)
	return result.IsCompatible
}

// IsCompatibleWithString checks if a version string is compatible with another
func (v *Versioning) IsCompatibleWithString(version1Str, version2Str string) (bool, error) {
	result, err := v.CompareVersionStrings(version1Str, version2Str)
	if err != nil {
		return false, err
	}
	return result.IsCompatible, nil
}

// GetVersionDifference returns a human-readable description of the difference between versions
func (v *Versioning) GetVersionDifference(version1, version2 *genericmodels.Version) string {
	result := v.CompareVersions(version1, version2)

	if result.IsEqual {
		return "Versions are identical"
	}

	version1Str := version1.String()
	version2Str := version2.String()

	if result.IsGreaterThan {
		return fmt.Sprintf("Version %s is %s than %s", version1Str, result.Difference, version2Str)
	} else {
		return fmt.Sprintf("Version %s is %s than %s", version1Str, result.Difference, version2Str)
	}
}

// IsGreaterThan checks if version1 is greater than version2
func (v *Versioning) IsGreaterThan(version1, version2 *genericmodels.Version) bool {
	result := v.CompareVersions(version1, version2)
	return result.IsGreaterThan
}

// IsLessThan checks if version1 is less than version2
func (v *Versioning) IsLessThan(version1, version2 *genericmodels.Version) bool {
	result := v.CompareVersions(version1, version2)
	return result.IsLessThan
}

// IsEqual checks if version1 is equal to version2
func (v *Versioning) IsEqual(version1, version2 *genericmodels.Version) bool {
	result := v.CompareVersions(version1, version2)
	return result.IsEqual
}

func (v *Versioning) LoadServiceVersion(serviceName string, env string) {

	versionKey := "service_version_" + serviceName + "_" + env
	versionFile, err := cache.GetCache().GetString(cache.GetCacheContext(), versionKey)
	if err != nil {
		logger.Log().Error("Error getting service version",
			logger.StringField("service", serviceName),
			logger.StringField("env", env),
			logger.ErrorField("error", err),
		)
		cache.GetCache().SetString(cache.GetCacheContext(), versionKey, "1.0.0")
		configmanager.GetInstance().SetVersion(genericmodels.Version{Major: 1, Minor: 0, Patch: 0})
	}

	var version genericmodels.Version
	err = version.Parse(versionFile)
	if err != nil {
		logger.Log().Error("Error parsing service version",
			logger.StringField("service", serviceName),
			logger.StringField("env", env),
			logger.ErrorField("error", err),
		)
	}
	version.Patch++
	if version.Patch > 999 {
		version.Patch = 0
		version.Minor++
	}
	if version.Minor > 99 {
		version.Minor = 0
		version.Major++
	}

	cache.GetCache().SetString(cache.GetCacheContext(), versionKey, version.String())
	configmanager.GetInstance().SetVersion(version)
}

// IsVersionInRange checks if a version string falls within the specified range (inclusive)
func (v *Versioning) IsVersionInRange(versionStr string, minVersion, maxVersion *genericmodels.Version) (bool, error) {
	// Parse the target version
	var targetVersion genericmodels.Version
	err := targetVersion.Parse(versionStr)
	if err != nil {
		return false, fmt.Errorf("error parsing target version '%s': %w", versionStr, err)
	}

	// Check if target version is greater than or equal to minimum version
	isGreaterThanOrEqualMin := v.CompareVersions(&targetVersion, minVersion)
	if isGreaterThanOrEqualMin.IsLessThan {
		return false, nil
	}

	// Check if target version is less than or equal to maximum version
	isLessThanOrEqualMax := v.CompareVersions(&targetVersion, maxVersion)
	if isLessThanOrEqualMax.IsGreaterThan {
		return false, nil
	}

	// Both conditions are met
	return true, nil
}
