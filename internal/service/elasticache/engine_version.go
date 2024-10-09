// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	versionStringRegexpInternalPattern = `[[:digit:]]+(\.[[:digit:]]+){2}`
	versionStringRegexpPattern         = "^" + versionStringRegexpInternalPattern + "$"
)

var versionStringRegexp = regexache.MustCompile(versionStringRegexpPattern)

func validMemcachedVersionString(v any, k string) (ws []string, errors []error) {
	value := v.(string)

	if !versionStringRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf("%s: must be a version string matching <major>.<minor>.<patch>", k))
	}

	return
}

const (
	redisVersionPreV6RegexpPattern  = `^[1-5](\.[[:digit:]]+){2}$`
	redisVersionPostV6RegexpPattern = `^((6)\.x)|([6-9]\.[[:digit:]]+)$`

	redisVersionRegexpPattern = redisVersionPreV6RegexpPattern + "|" + redisVersionPostV6RegexpPattern
)

var (
	redisVersionRegexp       = regexache.MustCompile(redisVersionRegexpPattern)
	redisVersionPostV6Regexp = regexache.MustCompile(redisVersionPostV6RegexpPattern)
)

func validRedisVersionString(v any, k string) (ws []string, errors []error) {
	value := v.(string)

	if !redisVersionRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf("%s: %s is invalid. For Redis v6 or higher, use <major>.<minor>. For Redis v5 or lower, use <major>.<minor>.<patch>.", k, value))
	}

	return
}

// customizeDiffValidateClusterEngineVersion validates the correct format for `engine_version`, based on `engine`
func customizeDiffValidateClusterEngineVersion(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	engineVersion, ok := diff.GetOk(names.AttrEngineVersion)
	if !ok {
		return nil
	}

	return validateClusterEngineVersion(diff.Get(names.AttrEngine).(string), engineVersion.(string))
}

// validateClusterEngineVersion validates the correct format for `engine_version`, based on `engine`
func validateClusterEngineVersion(engine, engineVersion string) error {
	// Memcached: Versions in format <major>.<minor>.<patch>
	// Redis: Starting with version 6, must match <major>.<minor>, prior to version 6, <major>.<minor>.<patch>
	var validator schema.SchemaValidateFunc
	if engine == "" || engine == engineMemcached {
		validator = validMemcachedVersionString
	} else {
		validator = validRedisVersionString
	}

	_, errs := validator(engineVersion, names.AttrEngineVersion)

	return errors.Join(errs...)
}

// customizeDiffEngineVersionForceNewOnDowngrade causes re-creation of the resource if the version is being downgraded
func customizeDiffEngineVersionForceNewOnDowngrade(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	return engineVersionForceNewOnDowngrade(diff)
}

type getChangeDiffer interface {
	Get(key string) any
	GetChange(key string) (any, any)
}

func engineVersionIsDowngrade(diff getChangeDiffer) (bool, error) {
	o, n := diff.GetChange(names.AttrEngineVersion)
	if o == "6.x" {
		actual := diff.Get("engine_version_actual")
		aVersion, err := gversion.NewVersion(actual.(string))
		if err != nil {
			return false, fmt.Errorf("parsing current engine_version: %w", err)
		}
		nVersion, err := normalizeEngineVersion(n.(string))
		if err != nil {
			return false, fmt.Errorf("parsing new engine_version: %w", err)
		}

		aSegments := aVersion.Segments()
		nSegments := nVersion.Segments()

		if nSegments[0] != aSegments[0] {
			return nSegments[0] < aSegments[0], nil
		}
		return nSegments[1] < aSegments[1], nil
	}

	oVersion, err := normalizeEngineVersion(o.(string))
	if err != nil {
		return false, fmt.Errorf("parsing old engine_version: %w", err)
	}
	nVersion, err := normalizeEngineVersion(n.(string))
	if err != nil {
		return false, fmt.Errorf("parsing new engine_version: %w", err)
	}

	return nVersion.LessThan(oVersion), nil
}

type forceNewDiffer interface {
	Id() string
	Get(key string) any
	GetChange(key string) (any, any)
	HasChange(key string) bool
	ForceNew(key string) error
}

func engineVersionForceNewOnDowngrade(diff forceNewDiffer) error {
	if diff.Id() == "" || !diff.HasChange(names.AttrEngineVersion) {
		return nil
	}

	if downgrade, err := engineVersionIsDowngrade(diff); err != nil {
		return err
	} else if !downgrade {
		return nil
	}

	return diff.ForceNew(names.AttrEngineVersion)
}

// normalizeEngineVersion returns a github.com/hashicorp/go-version Version from:
// - a regular 1.2.3 version number
// - either the 6.x or 6.0 version number used for ElastiCache Redis version 6. 6.x will sort to 6.<maxint>
// - a 7.0 version number used from version 7
func normalizeEngineVersion(version string) (*gversion.Version, error) {
	if matches := redisVersionPostV6Regexp.FindStringSubmatch(version); matches != nil {
		if matches[1] != "" {
			version = fmt.Sprintf("%s.%d", matches[2], math.MaxInt)
		}
	}
	return gversion.NewVersion(version)
}

func setEngineVersionMemcached(d *schema.ResourceData, version *string) {
	d.Set(names.AttrEngineVersion, version)
	d.Set("engine_version_actual", version)
}

func setEngineVersionRedis(d *schema.ResourceData, version *string) error {
	engineVersion, err := gversion.NewVersion(aws.ToString(version))
	if err != nil {
		return fmt.Errorf("reading engine version: %w", err)
	}
	if engineVersion.Segments()[0] < 6 {
		d.Set(names.AttrEngineVersion, engineVersion.String())
	} else {
		// Handle major-only version number
		configVersion := d.Get(names.AttrEngineVersion).(string)
		if t, _ := regexp.MatchString(`[6-9]\.x`, configVersion); t {
			d.Set(names.AttrEngineVersion, fmt.Sprintf("%d.x", engineVersion.Segments()[0]))
		} else {
			d.Set(names.AttrEngineVersion, fmt.Sprintf("%d.%d", engineVersion.Segments()[0], engineVersion.Segments()[1]))
		}
	}
	d.Set("engine_version_actual", engineVersion.String())

	return nil
}

type versionDiff [3]int

// diffVersion returns a diff of the versions, component by component.
// Only reports the first diff, since subsequent segments are unimportant for us.
func diffVersion(n, o *gversion.Version) (result versionDiff) {
	if n.String() == o.String() {
		return
	}

	segmentsNew := n.Segments64()
	segmentsOld := o.Segments64()

	for i := 0; i < 3; i++ {
		lhs := segmentsNew[i]
		rhs := segmentsOld[i]
		if lhs < rhs {
			result[i] = -1
			break
		} else if lhs > rhs {
			result[i] = 1
			break
		}
	}

	return
}
