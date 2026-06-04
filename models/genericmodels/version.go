package genericmodels

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents an API version
type Version struct {
	Major      int    `json:"major"`
	Minor      int    `json:"minor"`
	Patch      int    `json:"patch"`
	PreRelease string `json:"pre_release,omitempty"`
	Build      string `json:"build,omitempty"`
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) Parse(version string) error {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Regular expression to parse semantic version
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9\-\.]+))?(?:\+([a-zA-Z0-9\-\.]+))?$`)
	matches := re.FindStringSubmatch(version)

	if len(matches) < 4 {
		return fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return fmt.Errorf("invalid major version: %s", matches[1])
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return fmt.Errorf("invalid minor version: %s", matches[2])
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return fmt.Errorf("invalid patch version: %s", matches[3])
	}

	v.Major = major
	v.Minor = minor
	v.Patch = patch

	if len(matches) > 4 && matches[4] != "" {
		v.PreRelease = matches[4]
	}

	if len(matches) > 5 && matches[5] != "" {
		v.Build = matches[5]
	}

	return nil
}
