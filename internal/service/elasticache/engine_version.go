// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	gversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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

const (
	valkeyVersionRegexpPattern = `^[7-9]\.[[:digit:]]+$`
)

var (
	valkeyVersionRegexp = regexache.MustCompile(valkeyVersionRegexpPattern)
)

func validValkeyVersionString(v any, k string) (ws []string, errors []error) {
	value := v.(string)

	if !valkeyVersionRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf("%s: %s is invalid. For Valkey use <major>.<minor>.", k, value))
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
	// Valkey: Versions in format <major>.<minor>
	var validator schema.SchemaValidateFunc
	switch engine {
	case "", engineMemcached:
		validator = validMemcachedVersionString
	case engineRedis:
		validator = validRedisVersionString
	case engineValkey:
		validator = validValkeyVersionString
	}

	_, errs := validator(engineVersion, names.AttrEngineVersion)

	return errors.Join(errs...)
}

// customizeDiffEngineVersionForceNewOnDowngrade causes re-creation of the resource if the version is being downgraded
func customizeDiffEngineVersionForceNewOnDowngrade(_ context.Context, diff *schema.ResourceDiff, _ any) error {
	return engineVersionForceNewOnDowngrade(diff)
}

func customizeDiffEngineForceNewOnDowngrade() schema.CustomizeDiffFunc {
	return customdiff.ForceNewIf(names.AttrEngine, func(_ context.Context, diff *schema.ResourceDiff, meta any) bool {
		if _, is_global := diff.GetOk("global_replication_group_id"); is_global {
			return false
		}

		if !diff.HasChange(names.AttrEngine) {
			return false
		}
		if old, new := diff.GetChange(names.AttrEngine); old.(string) == engineRedis && new.(string) == engineValkey {
			return false
		}
		return true
	})
}

type getChangeDiffer interface {
	Get(key string) any
	GetChange(key string) (any, any)
}

func engineVersionIsDowngrade(diff getChangeDiffer) (bool, error) {
	o, n := diff.GetChange(names.AttrEngineVersion)
	if o == "6.x" || o == "7.x" {
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
	GetOk(key string) (any, bool)
	GetChange(key string) (any, any)
	HasChange(key string) bool
	ForceNew(key string) error
}

func engineVersionForceNewOnDowngrade(diff forceNewDiffer) error {
	if _, is_global := diff.GetOk("global_replication_group_id"); is_global {
		return nil
	}

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

func setEngineVersionValkey(d *schema.ResourceData, version *string) error {
	engineVersion, err := gversion.NewVersion(aws.ToString(version))
	if err != nil {
		return fmt.Errorf("reading engine version: %w", err)
	}
	d.Set(names.AttrEngineVersion, fmt.Sprintf("%d.%d", engineVersion.Segments()[0], engineVersion.Segments()[1]))
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

	for i := range 3 {
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

// redisMajorVersionWildcardRegexp matches the "<major>.x" convention (e.g. "6.x")
// that ElastiCache Redis accepts for engine_version to mean "the latest minor of
// that major version".
var redisMajorVersionWildcardRegexp = regexache.MustCompile(`^([6-9])\.x$`)

// customizeDiffValidateEngineVersion validates, at plan time, that the configured
// engine_version:
//   - is actually available for the engine in the target region, and
//   - is compatible with the family of the referenced parameter_group_name
//     (e.g. engine_version "7.0" with parameter_group_name "default.redis6.x").
//
// Both checks require calling the ElastiCache API (DescribeCacheEngineVersions,
// and for custom parameter groups DescribeCacheParameterGroups). To avoid breaking
// plans for callers that previously planned offline or without
// elasticache:DescribeCacheEngineVersions / elasticache:DescribeCacheParameterGroups
// permissions, any API error is treated as "skip" (fail open); only a definitive
// mismatch produces a plan-time error. Values that are unset or not yet known
// (e.g. interpolated from another resource) are also skipped.
func customizeDiffValidateEngineVersion(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
	engine, engineVersion, ok := configuredEngineVersion(diff)
	if !ok {
		return nil
	}

	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	available, err := findCacheEngineVersions(ctx, conn, engine)
	if err != nil {
		tflog.Debug(ctx, "skipping ElastiCache engine_version validation", map[string]any{
			names.AttrEngine:        engine,
			names.AttrEngineVersion: engineVersion,
			"error":                 err.Error(),
		})
		return nil //nolint:nilerr // fail open: skip plan-time validation when the API is unavailable (e.g. missing IAM permissions)
	}
	if len(available) == 0 {
		return nil
	}

	matched, found := findAvailableCacheEngineVersion(engine, engineVersion, available)
	if !found {
		return fmt.Errorf("engine_version %q is not available for engine %q in this region; available versions: %s",
			engineVersion, engine, strings.Join(availableEngineVersionStrings(available), ", "))
	}

	// Parameter group compatibility. Only checked when parameter_group_name is set,
	// known, and its family can be resolved.
	parameterGroupName, ok := configuredParameterGroupName(diff)
	if !ok {
		return nil
	}

	family, ok, err := findCacheParameterGroupFamily(ctx, conn, parameterGroupName)
	if err != nil {
		tflog.Debug(ctx, "skipping ElastiCache parameter group compatibility validation", map[string]any{
			names.AttrParameterGroupName: parameterGroupName,
			"error":                      err.Error(),
		})
		return nil //nolint:nilerr // fail open: skip plan-time validation when the API is unavailable (e.g. missing IAM permissions)
	}
	if !ok {
		return nil
	}

	if versionFamily := aws.ToString(matched.CacheParameterGroupFamily); !strings.EqualFold(versionFamily, family) {
		return fmt.Errorf("engine_version %q is not compatible with parameter_group_name %q: engine version family %q does not match parameter group family %q",
			engineVersion, parameterGroupName, versionFamily, family)
	}

	return nil
}

// configuredEngineVersion returns the configured engine and engine_version when
// both are set and known. It returns ok=false when engine_version is unset, empty,
// or not yet known (e.g. interpolated from another resource), or when the engine
// cannot be determined.
func configuredEngineVersion(diff *schema.ResourceDiff) (engine, engineVersion string, ok bool) {
	v, exists := diff.GetOk(names.AttrEngineVersion)
	if !exists {
		return "", "", false
	}
	engineVersion, _ = v.(string)
	if engineVersion == "" || !rawConfigAttrKnown(diff, names.AttrEngineVersion) {
		return "", "", false
	}

	engine, _ = diff.Get(names.AttrEngine).(string)
	if engine == "" {
		return "", "", false
	}

	return engine, engineVersion, true
}

// configuredParameterGroupName returns the configured parameter_group_name when it
// is set and known.
func configuredParameterGroupName(diff *schema.ResourceDiff) (string, bool) {
	v, exists := diff.GetOk(names.AttrParameterGroupName)
	if !exists {
		return "", false
	}
	name, _ := v.(string)
	if name == "" || !rawConfigAttrKnown(diff, names.AttrParameterGroupName) {
		return "", false
	}

	return name, true
}

// rawConfigAttrKnown reports whether the top-level attribute is known in the raw
// configuration. Unknown values (interpolated from another resource that has not
// yet been applied) cannot be validated at plan time.
func rawConfigAttrKnown(diff *schema.ResourceDiff, attr string) bool {
	rawConfig := diff.GetRawConfig()
	if rawConfig.IsNull() || !rawConfig.IsKnown() {
		return false
	}

	return rawConfig.GetAttr(attr).IsKnown()
}

// findAvailableCacheEngineVersion returns the CacheEngineVersion matching the
// configured engine_version, if the version is available.
func findAvailableCacheEngineVersion(engine, engineVersion string, available []awstypes.CacheEngineVersion) (awstypes.CacheEngineVersion, bool) {
	for _, v := range available {
		if engineVersionMatches(engine, engineVersion, aws.ToString(v.EngineVersion)) {
			return v, true
		}
	}

	return awstypes.CacheEngineVersion{}, false
}

// engineVersionMatches reports whether a configured engine_version matches a
// version returned by the API, accounting for the Redis "<major>.x" convention.
func engineVersionMatches(engine, configVersion, apiVersion string) bool {
	if apiVersion == "" {
		return false
	}
	if configVersion == apiVersion {
		return true
	}

	// Redis accepts a "<major>.x" convention (e.g. "6.x") that maps to any minor
	// version of that major (e.g. the API's "6.0" or "6.2").
	if engine == engineRedis {
		if m := redisMajorVersionWildcardRegexp.FindStringSubmatch(configVersion); m != nil {
			return strings.HasPrefix(apiVersion, m[1]+".")
		}
	}

	return false
}

// availableEngineVersionStrings returns the sorted, de-duplicated set of engine
// version strings from the supplied CacheEngineVersions, for use in diagnostics.
func availableEngineVersionStrings(available []awstypes.CacheEngineVersion) []string {
	seen := make(map[string]struct{}, len(available))
	versions := make([]string, 0, len(available))
	for _, v := range available {
		ev := aws.ToString(v.EngineVersion)
		if ev == "" {
			continue
		}
		if _, dup := seen[ev]; dup {
			continue
		}
		seen[ev] = struct{}{}
		versions = append(versions, ev)
	}
	slices.Sort(versions)

	return versions
}

// findCacheParameterGroupFamily resolves the parameter group family for the given
// parameter group name. For AWS default parameter groups the family is derived from
// the name without an API call; for custom groups it is looked up via the API.
// It returns ok=false (without error) when the family cannot be determined, such as
// when a custom parameter group does not yet exist.
func findCacheParameterGroupFamily(ctx context.Context, conn *elasticache.Client, name string) (string, bool, error) {
	if family, ok := defaultCacheParameterGroupFamily(name); ok {
		return family, true, nil
	}

	pg, err := findCacheParameterGroupByName(ctx, conn, name)
	if err != nil {
		if retry.NotFound(err) {
			return "", false, nil
		}
		return "", false, err
	}

	return aws.ToString(pg.CacheParameterGroupFamily), true, nil
}

// defaultCacheParameterGroupFamily derives the family of an AWS-managed default
// parameter group from its name (e.g. "default.redis6.x" -> "redis6.x",
// "default.redis7.cluster.on" -> "redis7"). It returns ok=false for names that are
// not default parameter groups.
func defaultCacheParameterGroupFamily(name string) (string, bool) {
	const prefix = "default."
	if !strings.HasPrefix(name, prefix) {
		return "", false
	}

	family := strings.TrimPrefix(name, prefix)
	family = strings.TrimSuffix(family, ".cluster.on")

	return family, true
}

// findCacheEngineVersions returns all cache engine versions available for the given
// engine in the current region.
func findCacheEngineVersions(ctx context.Context, conn *elasticache.Client, engine string) ([]awstypes.CacheEngineVersion, error) {
	input := &elasticache.DescribeCacheEngineVersionsInput{
		Engine: aws.String(engine),
	}

	var output []awstypes.CacheEngineVersion
	pages := elasticache.NewDescribeCacheEngineVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.CacheEngineVersions...)
	}

	return output, nil
}
