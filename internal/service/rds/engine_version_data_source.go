// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"sort"

	"github.com/YakDriver/go-version"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/namevaluesfilters"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rds_engine_version", name="Engine Version")
func dataSourceEngineVersion() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEngineVersionRead,

		Schema: map[string]*schema.Schema{
			"default_character_set": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_only": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Required: true,
			},
			"engine_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"exportable_log_types": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			names.AttrFilter: namevaluesfilters.Schema(),
			"has_major_target": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"has_minor_target": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"include_all": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"latest": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"parameter_group_family": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"preferred_major_targets": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"preferred_upgrade_targets": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"preferred_versions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"supported_character_sets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"supported_feature_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"supported_modes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"supported_timezones": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"supports_certificate_rotation_without_restart": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_global_databases": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_integrations": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_limitless_database": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_local_write_forwarding": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_log_exports_to_cloudwatch": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_parallel_query": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"supports_read_replica": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"valid_major_targets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"valid_minor_targets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"valid_upgrade_targets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"version_actual": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceEngineVersionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds.DescribeDBEngineVersionsInput{
		ListSupportedCharacterSets: aws.Bool(true),
		ListSupportedTimezones:     aws.Bool(true),
	}

	if v, ok := d.GetOk(names.AttrEngine); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).RDSFilters()
	}

	if v, ok := d.GetOk("parameter_group_family"); ok {
		input.DBParameterGroupFamily = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("include_all"); ok {
		input.IncludeAll = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("default_only"); ok {
		input.DefaultOnly = aws.Bool(v.(bool))
	}

	// Make sure any optional arguments in the schema are in this list except for "default_only"
	if _, ok := d.GetOk("default_only"); !ok && !criteriaSet(d, []string{
		names.AttrFilter,
		"has_major_target",
		"has_minor_target",
		"include_all",
		"latest",
		"preferred_major_targets",
		"preferred_upgrade_targets",
		"preferred_versions",
		names.AttrVersion,
	}) {
		input.DefaultOnly = aws.Bool(true)
	}

	var engineVersions []awstypes.DBEngineVersion

	pages := rds.NewDescribeDBEngineVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading RDS engine versions: %s", err)
		}

		engineVersions = append(engineVersions, page.DBEngineVersions...)
	}

	if len(engineVersions) == 0 {
		return sdkdiag.AppendErrorf(diags, "no RDS engine versions found: %+v", input)
	}

	prefSearch := false

	// preferred versions
	if l := d.Get("preferred_versions").([]any); len(l) > 0 {
		var preferredVersions []awstypes.DBEngineVersion

		for _, elem := range l {
			preferredVersion, ok := elem.(string)

			if !ok {
				continue
			}

			for _, engineVersion := range engineVersions {
				if preferredVersion == aws.ToString(engineVersion.EngineVersion) {
					preferredVersions = append(preferredVersions, engineVersion)
				}
			}
		}

		if len(preferredVersions) == 0 {
			return sdkdiag.AppendErrorf(diags, "no RDS engine versions match the criteria and preferred versions: %v\n%v", input, l)
		}

		prefSearch = true
		engineVersions = preferredVersions
	}

	// preferred upgrade targets
	if l := d.Get("preferred_upgrade_targets").([]any); len(l) > 0 {
		var prefUTs []awstypes.DBEngineVersion

	engineVersionsLoop:
		for _, engineVersion := range engineVersions {
			for _, upgradeTarget := range engineVersion.ValidUpgradeTarget {
				for _, elem := range l {
					prefUT, ok := elem.(string)
					if !ok {
						continue
					}

					if prefUT == aws.ToString(upgradeTarget.EngineVersion) {
						prefUTs = append(prefUTs, engineVersion)
						continue engineVersionsLoop
					}
				}
			}
		}

		if len(prefUTs) == 0 {
			return sdkdiag.AppendErrorf(diags, "no RDS engine versions match the criteria and preferred upgrade targets: %+v\n%v", input, l)
		}

		prefSearch = true
		engineVersions = prefUTs
	}

	// preferred major targets
	if l := d.Get("preferred_major_targets").([]any); len(l) > 0 {
		var prefMTs []awstypes.DBEngineVersion

	majorsLoop:
		for _, engineVersion := range engineVersions {
			for _, upgradeTarget := range engineVersion.ValidUpgradeTarget {
				for _, elem := range l {
					prefMT, ok := elem.(string)
					if !ok {
						continue
					}

					if prefMT == aws.ToString(upgradeTarget.EngineVersion) && aws.ToBool(upgradeTarget.IsMajorVersionUpgrade) {
						prefMTs = append(prefMTs, engineVersion)
						continue majorsLoop
					}
				}
			}
		}

		if len(prefMTs) == 0 {
			return sdkdiag.AppendErrorf(diags, "no RDS engine versions match the criteria and preferred major targets: %+v\n%v", input, l)
		}

		prefSearch = true
		engineVersions = prefMTs
	}

	if v, ok := d.GetOk("has_minor_target"); ok && v.(bool) {
		var wMinor []awstypes.DBEngineVersion

	hasMinorLoop:
		for _, engineVersion := range engineVersions {
			for _, upgradeTarget := range engineVersion.ValidUpgradeTarget {
				if !aws.ToBool(upgradeTarget.IsMajorVersionUpgrade) {
					wMinor = append(wMinor, engineVersion)
					continue hasMinorLoop
				}
			}
		}

		if len(wMinor) == 0 {
			return sdkdiag.AppendErrorf(diags, "no RDS engine versions match the criteria and have a minor target: %+v", input)
		}

		engineVersions = wMinor
	}

	if v, ok := d.GetOk("has_major_target"); ok && v.(bool) {
		var wMajor []awstypes.DBEngineVersion

	hasMajorLoop:
		for _, engineVersion := range engineVersions {
			for _, upgradeTarget := range engineVersion.ValidUpgradeTarget {
				if aws.ToBool(upgradeTarget.IsMajorVersionUpgrade) {
					wMajor = append(wMajor, engineVersion)
					continue hasMajorLoop
				}
			}
		}

		if len(wMajor) == 0 {
			return sdkdiag.AppendErrorf(diags, "no RDS engine versions match the criteria and have a major target: %+v", input)
		}

		engineVersions = wMajor
	}

	var found *awstypes.DBEngineVersion

	if v, ok := d.GetOk("latest"); ok && v.(bool) {
		sortEngineVersions(engineVersions)
		found = &engineVersions[len(engineVersions)-1]
	}

	if found == nil && len(engineVersions) == 1 {
		found = &engineVersions[0]
	}

	if found == nil && len(engineVersions) > 0 && prefSearch {
		found = &engineVersions[0]
	}

	if found == nil && len(engineVersions) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple RDS engine versions (%v) match the criteria: %+v", engineVersions, input)
	}

	if found == nil {
		return sdkdiag.AppendErrorf(diags, "no RDS engine versions match the criteria: %+v", input)
	}

	d.SetId(aws.ToString(found.EngineVersion))
	if found.DefaultCharacterSet != nil {
		d.Set("default_character_set", found.DefaultCharacterSet.CharacterSetName)
	}
	d.Set(names.AttrEngine, found.Engine)
	d.Set("engine_description", found.DBEngineDescription)
	d.Set("exportable_log_types", found.ExportableLogTypes)
	d.Set("parameter_group_family", found.DBParameterGroupFamily)
	d.Set(names.AttrStatus, found.Status)
	d.Set("supported_character_sets", tfslices.ApplyToAll(found.SupportedCharacterSets, func(v awstypes.CharacterSet) string {
		return aws.ToString(v.CharacterSetName)
	}))
	d.Set("supported_feature_names", found.SupportedFeatureNames)
	d.Set("supported_modes", found.SupportedEngineModes)
	d.Set("supported_timezones", tfslices.ApplyToAll(found.SupportedTimezones, func(v awstypes.Timezone) string {
		return aws.ToString(v.TimezoneName)
	}))
	d.Set("supports_certificate_rotation_without_restart", found.SupportsCertificateRotationWithoutRestart)
	d.Set("supports_global_databases", found.SupportsGlobalDatabases)
	d.Set("supports_integrations", found.SupportsIntegrations)
	d.Set("supports_limitless_database", found.SupportsLimitlessDatabase)
	d.Set("supports_local_write_forwarding", found.SupportsLocalWriteForwarding)
	d.Set("supports_log_exports_to_cloudwatch", found.SupportsLogExportsToCloudwatchLogs)
	d.Set("supports_parallel_query", found.SupportsParallelQuery)
	d.Set("supports_read_replica", found.SupportsReadReplica)

	var upgradeTargets []string
	var minorTargets []string
	var majorTargets []string
	for _, ut := range found.ValidUpgradeTarget {
		upgradeTargets = append(upgradeTargets, aws.ToString(ut.EngineVersion))

		if aws.ToBool(ut.IsMajorVersionUpgrade) {
			majorTargets = append(majorTargets, aws.ToString(ut.EngineVersion))
			continue
		}

		minorTargets = append(minorTargets, aws.ToString(ut.EngineVersion))
	}
	d.Set("valid_upgrade_targets", upgradeTargets)
	d.Set("valid_minor_targets", minorTargets)
	d.Set("valid_major_targets", majorTargets)

	d.Set(names.AttrVersion, found.EngineVersion)
	d.Set("version_actual", found.EngineVersion)
	d.Set("version_description", found.DBEngineVersionDescription)

	return diags
}

func sortEngineVersions(engineVersions []awstypes.DBEngineVersion) {
	if len(engineVersions) < 2 {
		return
	}

	sort.Slice(engineVersions, func(i, j int) bool { // nosemgrep:ci.semgrep.stdlib.prefer-slices-sortfunc
		return version.LessThanWithTime(engineVersions[i].CreateTime, engineVersions[j].CreateTime, aws.ToString(engineVersions[i].EngineVersion), aws.ToString(engineVersions[j].EngineVersion))
	})
}

// criteriaSet returns true if any of the given criteria are set. "set" means that, in the config,
// a bool is set and true, a list is set and not empty, or a string is set and not empty.
func criteriaSet(d *schema.ResourceData, args []string) bool {
	for _, arg := range args {
		val := d.GetRawConfig().GetAttr(arg)

		switch {
		case val.CanIterateElements():
			if !val.IsNull() && val.IsKnown() && val.LengthInt() > 0 {
				return true
			}
		case val.Equals(cty.True) == cty.True:
			return true

		case val.Type() == cty.String && !val.IsNull() && val.IsKnown():
			return true
		}
	}

	return false
}
