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

// customizeDiffValidateEngineVersionAvailable errors at plan time when engine_version is unavailable for the engine in the Region; any API error is skipped (fail open) so offline or unpermissioned plans still work.
func customizeDiffValidateEngineVersionAvailable(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
	engine, engineVersion, ok := configuredEngineVersion(diff)
	if !ok {
		return nil
	}

	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	available, err := findCacheEngineVersions(ctx, conn, engine)
	if err != nil {
		tflog.Debug(ctx, "skipping ElastiCache engine_version availability validation", map[string]any{
			names.AttrEngine:        engine,
			names.AttrEngineVersion: engineVersion,
			"error":                 err.Error(),
		})
		return nil //nolint:nilerr // fail open: skip plan-time validation when the API is unavailable (e.g. missing IAM permissions)
	}
	if len(available) == 0 {
		return nil
	}

	if !findAvailableCacheEngineVersion(engine, engineVersion, available) {
		return fmt.Errorf("engine_version %q is not available for engine %q in this region; available versions: %s",
			engineVersion, engine, strings.Join(availableEngineVersionStrings(available), ", "))
	}

	return nil
}

// configuredEngineVersion returns engine and an explicitly-set, known engine_version; a null (computed/carried-over) engine_version is skipped, an unset engine defaults to redis, and set-but-unknown values yield ok=false.
func configuredEngineVersion(diff *schema.ResourceDiff) (engine, engineVersion string, ok bool) {
	rawConfig := diff.GetRawConfig()
	if rawConfig.IsNull() || !rawConfig.IsKnown() {
		return "", "", false
	}

	// Only validate an engine_version the user explicitly set; a null value here is the computed version carried from state (e.g. when only engine changes) and is not the user's intent.
	ev := rawConfig.GetAttr(names.AttrEngineVersion)
	if ev.IsNull() || !ev.IsKnown() {
		return "", "", false
	}
	engineVersion = ev.AsString()
	if engineVersion == "" {
		return "", "", false
	}

	// engine is Optional+Computed: unset defaults to redis; set-but-unknown cannot be validated yet.
	e := rawConfig.GetAttr(names.AttrEngine)
	if !e.IsKnown() {
		return "", "", false
	}
	if e.IsNull() {
		engine = engineRedis
	} else {
		engine = e.AsString()
	}

	return engine, engineVersion, true
}

func findAvailableCacheEngineVersion(engine, engineVersion string, available []awstypes.CacheEngineVersion) bool {
	for _, v := range available {
		if engineVersionMatches(engine, engineVersion, aws.ToString(v.EngineVersion)) {
			return true
		}
	}

	return false
}

// engineVersionMatches handles the Redis "<major>.x" convention (only Redis, and only "6.x" is reachable since format validation rejects "7.x"-"9.x" first), reusing redisVersionPostV6Regexp instead of a second divergent regex.
func engineVersionMatches(engine, configVersion, apiVersion string) bool {
	if apiVersion == "" {
		return false
	}
	if configVersion == apiVersion {
		return true
	}

	if engine == engineRedis {
		// m[1] is the "<major>.x" wildcard group; m[2] is its major (e.g. "6" from "6.x").
		if m := redisVersionPostV6Regexp.FindStringSubmatch(configVersion); m != nil && m[1] != "" {
			return strings.HasPrefix(apiVersion, m[2]+".")
		}
	}

	return false
}

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
