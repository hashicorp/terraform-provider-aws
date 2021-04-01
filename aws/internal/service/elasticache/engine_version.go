package elasticache

import (
	"fmt"
	"regexp"

	gversion "github.com/hashicorp/go-version"
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

// normalizeElastiCacheEngineVersion returns a github.com/hashicorp/go-version Version
// that can handle a regular 1.2.3 version number or a 6.x version number used for
// ElastiCache Redis version 6 and higher
func NormalizeElastiCacheEngineVersion(version string) (*gversion.Version, error) {
	if matches := redisVersionPostV6Regexp.FindStringSubmatch(version); matches != nil {
		version = matches[1]
	}
	return gversion.NewVersion(version)
}
