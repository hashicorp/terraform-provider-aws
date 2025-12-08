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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rds_orderable_db_instance", name="Orderable DB Instance")
func dataSourceOrderableInstance() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOrderableInstanceRead,

		Schema: map[string]*schema.Schema{
			"availability_zone_group": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Required: true,
			},
			"engine_latest_version": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"instance_class": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"license_model": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"max_iops_per_db_instance": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"max_iops_per_gib": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"max_storage_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"min_iops_per_db_instance": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"min_iops_per_gib": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"min_storage_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"multi_az_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"outpost_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"preferred_engine_versions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"preferred_instance_classes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"read_replica_capable": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrStorageType: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"supported_engine_modes": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supported_network_types": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supports_clusters": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"supports_enhanced_monitoring": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"supports_global_databases": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"supports_iam_database_authentication": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"supports_iops": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"supports_kerberos_authentication": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"supports_multi_az": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"supports_performance_insights": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"supports_storage_autoscaling": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"supports_storage_encryption": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"vpc": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceOrderableInstanceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds.DescribeOrderableDBInstanceOptionsInput{
		MaxRecords: aws.Int32(1000),
	}

	if v, ok := d.GetOk("availability_zone_group"); ok {
		input.AvailabilityZoneGroup = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_class"); ok {
		input.DBInstanceClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngine); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("license_model"); ok {
		input.LicenseModel = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc"); ok {
		input.Vpc = aws.Bool(v.(bool))
	}

	var instanceClassResults []awstypes.OrderableDBInstanceOption

	pages := rds.NewDescribeOrderableDBInstanceOptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading RDS Orderable DB Instance Options: %s", err)
		}

		for _, instanceOption := range page.OrderableDBInstanceOptions {
			if v, ok := d.GetOk("read_replica_capable"); ok {
				if aws.ToBool(instanceOption.ReadReplicaCapable) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk(names.AttrStorageType); ok {
				if aws.ToString(instanceOption.StorageType) != v.(string) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_clusters"); ok {
				if aws.ToBool(instanceOption.SupportsClusters) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_enhanced_monitoring"); ok {
				if aws.ToBool(instanceOption.SupportsEnhancedMonitoring) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_global_databases"); ok {
				if aws.ToBool(instanceOption.SupportsGlobalDatabases) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_iam_database_authentication"); ok {
				if aws.ToBool(instanceOption.SupportsIAMDatabaseAuthentication) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_iops"); ok {
				if aws.ToBool(instanceOption.SupportsIops) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_kerberos_authentication"); ok {
				if aws.ToBool(instanceOption.SupportsKerberosAuthentication) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_multi_az"); ok {
				if aws.ToBool(instanceOption.MultiAZCapable) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_performance_insights"); ok {
				if aws.ToBool(instanceOption.SupportsPerformanceInsights) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_storage_autoscaling"); ok {
				if aws.ToBool(instanceOption.SupportsStorageAutoscaling) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_storage_encryption"); ok {
				if aws.ToBool(instanceOption.SupportsStorageEncryption) != v.(bool) {
					continue
				}
			}

			instanceClassResults = append(instanceClassResults, instanceOption)
		}
	}

	if len(instanceClassResults) == 0 {
		return sdkdiag.AppendErrorf(diags, "no RDS Orderable DB Instance Options found matching criteria; try different search")
	}

	if v, ok := d.GetOk("supported_engine_modes"); ok && len(v.([]any)) > 0 {
		var matches []awstypes.OrderableDBInstanceOption
		search := flex.ExpandStringValueList(v.([]any))

		for _, ic := range instanceClassResults {
		searchedModes:
			for _, s := range search {
				for _, mode := range ic.SupportedEngineModes {
					if mode == s {
						matches = append(matches, ic)
						break searchedModes
					}
				}
			}
		}

		if len(matches) == 0 {
			return sdkdiag.AppendErrorf(diags, "no RDS Orderable DB Instance Options found matching supported_engine_modes: %#v", search)
		}

		instanceClassResults = matches
	}

	if v, ok := d.GetOk("supported_network_types"); ok && len(v.([]any)) > 0 {
		var matches []awstypes.OrderableDBInstanceOption
		search := flex.ExpandStringValueList(v.([]any))

		for _, ic := range instanceClassResults {
		searchedNetworks:
			for _, s := range search {
				for _, netType := range ic.SupportedNetworkTypes {
					if netType == s {
						matches = append(matches, ic)
						break searchedNetworks
					}
				}
			}
		}

		if len(matches) == 0 {
			return sdkdiag.AppendErrorf(diags, "no RDS Orderable DB Instance Options found matching supported_network_types: %#v", search)
		}

		instanceClassResults = matches
	}

	prefSearch := false

	if v, ok := d.GetOk("preferred_engine_versions"); ok && len(v.([]any)) > 0 {
		var matches []awstypes.OrderableDBInstanceOption
		search := flex.ExpandStringValueList(v.([]any))

		for _, s := range search {
			for _, ic := range instanceClassResults {
				if aws.ToString(ic.EngineVersion) == s {
					matches = append(matches, ic)
				}
				// keeping all the instance classes to ensure we can match any preferred instance classes
			}
		}

		if len(matches) == 0 {
			return sdkdiag.AppendErrorf(diags, "no RDS Orderable DB Instance Options found matching preferred_engine_versions: %#v", search)
		}

		prefSearch = true
		instanceClassResults = matches
	}

	latestVersion := d.Get("engine_latest_version").(bool)

	if v, ok := d.GetOk("preferred_instance_classes"); ok && len(v.([]any)) > 0 {
		var matches []awstypes.OrderableDBInstanceOption
		search := flex.ExpandStringValueList(v.([]any))

		for _, s := range search {
			for _, ic := range instanceClassResults {
				if aws.ToString(ic.DBInstanceClass) == s {
					matches = append(matches, ic)
				}

				if !latestVersion && len(matches) > 0 {
					break
				}

				// otherwise, get all the instance classes that match the *first* preferred class (and any other criteria)
			}

			// if we have a match, we can stop searching
			if len(matches) > 0 {
				break
			}
		}

		if len(matches) == 0 {
			return sdkdiag.AppendErrorf(diags, "no RDS Orderable DB Instance Options found matching preferred_instance_classes: %#v", search)
		}

		prefSearch = true
		instanceClassResults = matches
	}

	var found *awstypes.OrderableDBInstanceOption

	if latestVersion && prefSearch {
		sortInstanceClassesByVersion(instanceClassResults)
		found = &instanceClassResults[len(instanceClassResults)-1]
	}

	if found == nil && len(instanceClassResults) > 0 && prefSearch {
		found = &instanceClassResults[0]
	}

	if found == nil && len(instanceClassResults) == 1 {
		found = &instanceClassResults[0]
	}

	if found == nil && len(instanceClassResults) > 4 {
		// there can be a LOT(!!) of results, so if there are more than this, only include the search criteria
		return sdkdiag.AppendErrorf(diags, "multiple (%d) RDS DB Instance Classes match the criteria; try a different search: %+v\nPreferred instance classes: %+v\nPreferred engine versions: %+v", len(instanceClassResults), input, d.Get("preferred_instance_classes"), d.Get("preferred_engine_versions"))
	}

	if found == nil && len(instanceClassResults) > 1 {
		// there can be a lot of results, so if there are a few, include the results and search criteria
		return sdkdiag.AppendErrorf(diags, "multiple (%d) RDS DB Instance Classes (%v) match the criteria; try a different search: %+v\nPreferred instance classes: %+v\nPreferred engine versions: %+v", len(instanceClassResults), instanceClassResults, input, d.Get("preferred_instance_classes"), d.Get("preferred_engine_versions"))
	}

	if found == nil {
		return sdkdiag.AppendErrorf(diags, "no RDS DB Instance Classes match the criteria; try a different search")
	}

	d.SetId(aws.ToString(found.DBInstanceClass))
	d.Set("availability_zone_group", found.AvailabilityZoneGroup)
	d.Set(names.AttrAvailabilityZones, tfslices.ApplyToAll(found.AvailabilityZones, func(v awstypes.AvailabilityZone) string {
		return aws.ToString(v.Name)
	}))
	d.Set(names.AttrEngine, found.Engine)
	d.Set(names.AttrEngineVersion, found.EngineVersion)
	d.Set("instance_class", found.DBInstanceClass)
	d.Set("license_model", found.LicenseModel)
	d.Set("max_iops_per_db_instance", found.MaxIopsPerDbInstance)
	d.Set("max_iops_per_gib", found.MaxIopsPerGib)
	d.Set("max_storage_size", found.MaxStorageSize)
	d.Set("min_iops_per_db_instance", found.MinIopsPerDbInstance)
	d.Set("min_iops_per_gib", found.MinIopsPerGib)
	d.Set("min_storage_size", found.MinStorageSize)
	d.Set("multi_az_capable", found.MultiAZCapable)
	d.Set("outpost_capable", found.OutpostCapable)
	d.Set("read_replica_capable", found.ReadReplicaCapable)
	d.Set(names.AttrStorageType, found.StorageType)
	d.Set("supported_engine_modes", found.SupportedEngineModes)
	d.Set("supported_network_types", found.SupportedNetworkTypes)
	d.Set("supports_clusters", found.SupportsClusters)
	d.Set("supports_enhanced_monitoring", found.SupportsEnhancedMonitoring)
	d.Set("supports_global_databases", found.SupportsGlobalDatabases)
	d.Set("supports_iam_database_authentication", found.SupportsIAMDatabaseAuthentication)
	d.Set("supports_iops", found.SupportsIops)
	d.Set("supports_kerberos_authentication", found.SupportsKerberosAuthentication)
	d.Set("supports_multi_az", found.MultiAZCapable)
	d.Set("supports_performance_insights", found.SupportsPerformanceInsights)
	d.Set("supports_storage_autoscaling", found.SupportsStorageAutoscaling)
	d.Set("supports_storage_encryption", found.SupportsStorageEncryption)
	d.Set("vpc", found.Vpc)

	return diags
}

func sortInstanceClassesByVersion(ic []awstypes.OrderableDBInstanceOption) {
	if len(ic) < 2 {
		return
	}

	sort.Slice(ic, func(i, j int) bool { // nosemgrep:ci.semgrep.stdlib.prefer-slices-sortfunc
		return version.LessThan(aws.ToString(ic[i].EngineVersion), aws.ToString(ic[j].EngineVersion))
	})
}
