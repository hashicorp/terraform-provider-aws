package aws

import (
	"context"
	"fmt"
	"regexp"

	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	redisVersionPreV6RegexpRaw  = `[1-5](\.[[:digit:]]+){2}`
	redisVersionPostV6RegexpRaw = `([6-9]|[[:digit:]]{2})\.x`

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

func ValidateElastiCacheRedisVersionString(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !redisVersionRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf("%s: Redis versions must match <major>.x when using version 6 or higher, or <major>.<minor>.<bug-fix>", k))
	}

	return
}

// NormalizeElastiCacheEngineVersion returns a github.com/hashicorp/go-version Version
// that can handle a regular 1.2.3 version number or a 6.x version number used for
// ElastiCache Redis version 6 and higher
func NormalizeElastiCacheEngineVersion(version string) (*gversion.Version, error) {
	if matches := redisVersionPostV6Regexp.FindStringSubmatch(version); matches != nil {
		version = matches[1]
	}
	return gversion.NewVersion(version)
}

// CustomizeDiffElastiCacheEngineVersion causes re-creation of the resource if the version is being downgraded
func CustomizeDiffElastiCacheEngineVersion(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() == "" || !diff.HasChange("engine_version") {
		return nil
	}

	o, n := diff.GetChange("engine_version")
	oVersion, err := NormalizeElastiCacheEngineVersion(o.(string))
	if err != nil {
		return fmt.Errorf("error parsing old engine_version: %w", err)
	}
	nVersion, err := NormalizeElastiCacheEngineVersion(n.(string))
	if err != nil {
		return fmt.Errorf("error parsing new engine_version: %w", err)
	}

	if nVersion.GreaterThan(oVersion) {
		return nil
	}

	return diff.ForceNew("engine_version")
}
