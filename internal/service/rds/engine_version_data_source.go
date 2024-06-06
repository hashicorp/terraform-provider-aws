// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"sort"

	"github.com/YakDriver/go-version"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rds_engine_version")
func DataSourceEngineVersion() *schema.Resource {
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
				Set:      schema.HashString,
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
				Set:      schema.HashString,
			},

			"supported_feature_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"supported_modes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"supported_timezones": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"supports_global_databases": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"supports_limitless_database": {
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
				Set:      schema.HashString,
			},

			"valid_minor_targets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},

			"valid_upgrade_targets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
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

func dataSourceEngineVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

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

	log.Printf("[DEBUG] Reading RDS engine versions: %v", input)
	var engineVersions []*rds.DBEngineVersion

	err := conn.DescribeDBEngineVersionsPagesWithContext(ctx, input, func(resp *rds.DescribeDBEngineVersionsOutput, lastPage bool) bool {
		for _, engineVersion := range resp.DBEngineVersions {
			if engineVersion == nil {
				continue
			}

			engineVersions = append(engineVersions, engineVersion)
		}
		return !lastPage
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS engine versions: %s", err)
	}

	if len(engineVersions) == 0 {
		return sdkdiag.AppendErrorf(diags, "no RDS engine versions found: %+v", input)
	}

	prefSearch := false

	// preferred versions
	if l := d.Get("preferred_versions").([]interface{}); len(l) > 0 {
		var preferredVersions []*rds.DBEngineVersion

		for _, elem := range l {
			preferredVersion, ok := elem.(string)

			if !ok {
				continue
			}

			for _, engineVersion := range engineVersions {
				if preferredVersion == aws.StringValue(engineVersion.EngineVersion) {
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
	if l := d.Get("preferred_upgrade_targets").([]interface{}); len(l) > 0 {
		var prefUTs []*rds.DBEngineVersion

	engineVersionsLoop:
		for _, engineVersion := range engineVersions {
			for _, upgradeTarget := range engineVersion.ValidUpgradeTarget {
				for _, elem := range l {
					prefUT, ok := elem.(string)
					if !ok {
						continue
					}

					if prefUT == aws.StringValue(upgradeTarget.EngineVersion) {
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
	if l := d.Get("preferred_major_targets").([]interface{}); len(l) > 0 {
		var prefMTs []*rds.DBEngineVersion

	majorsLoop:
		for _, engineVersion := range engineVersions {
			for _, upgradeTarget := range engineVersion.ValidUpgradeTarget {
				for _, elem := range l {
					prefMT, ok := elem.(string)
					if !ok {
						continue
					}

					if prefMT == aws.StringValue(upgradeTarget.EngineVersion) && aws.BoolValue(upgradeTarget.IsMajorVersionUpgrade) {
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
		var wMinor []*rds.DBEngineVersion

	hasMinorLoop:
		for _, engineVersion := range engineVersions {
			for _, upgradeTarget := range engineVersion.ValidUpgradeTarget {
				if !aws.BoolValue(upgradeTarget.IsMajorVersionUpgrade) {
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
		var wMajor []*rds.DBEngineVersion

	hasMajorLoop:
		for _, engineVersion := range engineVersions {
			for _, upgradeTarget := range engineVersion.ValidUpgradeTarget {
				if aws.BoolValue(upgradeTarget.IsMajorVersionUpgrade) {
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

	var found *rds.DBEngineVersion

	if v, ok := d.GetOk("latest"); ok && v.(bool) {
		sortEngineVersions(engineVersions)
		found = engineVersions[len(engineVersions)-1]
	}

	if found == nil && len(engineVersions) == 1 {
		found = engineVersions[0]
	}

	if found == nil && len(engineVersions) > 0 && prefSearch {
		found = engineVersions[0]
	}

	if found == nil && len(engineVersions) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple RDS engine versions (%v) match the criteria: %+v", engineVersions, input)
	}

	if found == nil {
		return sdkdiag.AppendErrorf(diags, "no RDS engine versions match the criteria: %+v", input)
	}

	d.SetId(aws.StringValue(found.EngineVersion))

	if found.DefaultCharacterSet != nil {
		d.Set("default_character_set", found.DefaultCharacterSet.CharacterSetName)
	}

	d.Set(names.AttrEngine, found.Engine)
	d.Set("engine_description", found.DBEngineDescription)
	d.Set("exportable_log_types", found.ExportableLogTypes)
	d.Set("parameter_group_family", found.DBParameterGroupFamily)
	d.Set(names.AttrStatus, found.Status)

	var characterSets []string
	for _, cs := range found.SupportedCharacterSets {
		characterSets = append(characterSets, aws.StringValue(cs.CharacterSetName))
	}
	d.Set("supported_character_sets", characterSets)

	d.Set("supported_feature_names", found.SupportedFeatureNames)
	d.Set("supported_modes", found.SupportedEngineModes)

	var timezones []string
	for _, tz := range found.SupportedTimezones {
		timezones = append(timezones, aws.StringValue(tz.TimezoneName))
	}
	d.Set("supported_timezones", timezones)

	d.Set("supports_global_databases", found.SupportsGlobalDatabases)
	d.Set("supports_limitless_database", found.SupportsLimitlessDatabase)
	d.Set("supports_log_exports_to_cloudwatch", found.SupportsLogExportsToCloudwatchLogs)
	d.Set("supports_parallel_query", found.SupportsParallelQuery)
	d.Set("supports_read_replica", found.SupportsReadReplica)

	var upgradeTargets []string
	var minorTargets []string
	var majorTargets []string
	for _, ut := range found.ValidUpgradeTarget {
		upgradeTargets = append(upgradeTargets, aws.StringValue(ut.EngineVersion))

		if aws.BoolValue(ut.IsMajorVersionUpgrade) {
			majorTargets = append(majorTargets, aws.StringValue(ut.EngineVersion))
			continue
		}

		minorTargets = append(minorTargets, aws.StringValue(ut.EngineVersion))
	}
	d.Set("valid_upgrade_targets", upgradeTargets)
	d.Set("valid_minor_targets", minorTargets)
	d.Set("valid_major_targets", majorTargets)

	d.Set(names.AttrVersion, found.EngineVersion)
	d.Set("version_actual", found.EngineVersion)
	d.Set("version_description", found.DBEngineVersionDescription)

	return diags
}

func sortEngineVersions(engineVersions []*rds.DBEngineVersion) {
	if len(engineVersions) < 2 {
		return
	}

	sort.Slice(engineVersions, func(i, j int) bool {
		return version.LessThanWithTime(engineVersions[i].CreateTime, engineVersions[j].CreateTime, aws.StringValue(engineVersions[i].EngineVersion), aws.StringValue(engineVersions[j].EngineVersion))
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
