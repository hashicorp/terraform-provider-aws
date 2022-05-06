package elasticache

import (
	"context"
	"fmt"
	"math"
	"regexp"

	multierror "github.com/hashicorp/go-multierror"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	versionStringRegexpInternalPattern = `[[:digit:]]+(\.[[:digit:]]+){2}`
	versionStringRegexpPattern         = "^" + versionStringRegexpInternalPattern + "$"
)

var versionStringRegexp = regexp.MustCompile(versionStringRegexpPattern)

func validMemcachedVersionString(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !versionStringRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf("%s: must be a version string matching <major>.<minor>.<patch>", k))
	}

	return
}

const (
	redisVersionPreV6RegexpRaw  = `[1-5](\.[[:digit:]]+){2}`
	redisVersionPostV6RegexpRaw = `(([6-9])\.x)|([6-9]\.[[:digit:]]+)`

	redisVersionRegexpRaw = redisVersionPreV6RegexpRaw + "|" + redisVersionPostV6RegexpRaw
)

const (
	redisVersionRegexpPattern       = "^" + redisVersionRegexpRaw + "$"
	redisVersionPostV6RegexpPattern = "^" + redisVersionPostV6RegexpRaw + "$"
)

var (
	redisVersionRegexp       = regexp.MustCompile(redisVersionRegexpPattern)
	redisVersionPostV6Regexp = regexp.MustCompile(redisVersionPostV6RegexpPattern)
)

func validRedisVersionString(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !redisVersionRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf("%s: Redis versions must match <major>.<minor> when using version 6 or higher, or <major>.<minor>.<patch>", k))
	}

	return
}

// CustomizeDiffValidateClusterEngineVersion validates the correct format for `engine_version`, based on `engine`
func CustomizeDiffValidateClusterEngineVersion(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	engineVersion, ok := diff.GetOk("engine_version")
	if !ok {
		return nil
	}

	return validateClusterEngineVersion(diff.Get("engine").(string), engineVersion.(string))
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

	_, errs := validator(engineVersion, "engine_version")

	var err *multierror.Error
	err = multierror.Append(err, errs...)
	return err.ErrorOrNil()
}

// customizeDiffEngineVersionForceNewOnDowngrade causes re-creation of the resource if the version is being downgraded
func customizeDiffEngineVersionForceNewOnDowngrade(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	return engineVersionForceNewOnDowngrade(diff)
}

type forceNewDiffer interface {
	Id() string
	HasChange(key string) bool
	GetChange(key string) (interface{}, interface{})
	ForceNew(key string) error
}

func engineVersionForceNewOnDowngrade(diff forceNewDiffer) error {
	if diff.Id() == "" || !diff.HasChange("engine_version") {
		return nil
	}

	o, n := diff.GetChange("engine_version")
	oVersion, err := normalizeEngineVersion(o.(string))
	if err != nil {
		return fmt.Errorf("error parsing old engine_version: %w", err)
	}
	nVersion, err := normalizeEngineVersion(n.(string))
	if err != nil {
		return fmt.Errorf("error parsing new engine_version: %w", err)
	}

	if nVersion.GreaterThan(oVersion) || nVersion.Equal(oVersion) {
		return nil
	}

	return diff.ForceNew("engine_version")
}

// normalizeEngineVersion returns a github.com/hashicorp/go-version Version
// that can handle a regular 1.2.3 version number or either the  6.x or 6.0 version number used for
// ElastiCache Redis version 6 and higher. 6.x will sort to 6.<maxint>
func normalizeEngineVersion(version string) (*gversion.Version, error) {
	if matches := redisVersionPostV6Regexp.FindStringSubmatch(version); matches != nil {
		if matches[1] != "" {
			version = fmt.Sprintf("%s.%d", matches[2], math.MaxInt)
		} else if matches[3] != "" {
			version = matches[3]
		}
	}
	return gversion.NewVersion(version)
}
