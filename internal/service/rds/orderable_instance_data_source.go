// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"sort"

	"github.com/YakDriver/go-version"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rds_orderable_db_instance")
func DataSourceOrderableInstance() *schema.Resource {
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

			"engine_latest_version": {
				Type:     schema.TypeBool,
				Optional: true,
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

			"preferred_instance_classes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"preferred_engine_versions": {
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

func dataSourceOrderableInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeOrderableDBInstanceOptionsInput{
		MaxRecords: aws.Int64(1000),
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

	var instanceClassResults []*rds.OrderableDBInstanceOption

	err := conn.DescribeOrderableDBInstanceOptionsPagesWithContext(ctx, input, func(resp *rds.DescribeOrderableDBInstanceOptionsOutput, lastPage bool) bool {
		for _, instanceOption := range resp.OrderableDBInstanceOptions {
			if instanceOption == nil {
				continue
			}

			if v, ok := d.GetOk("read_replica_capable"); ok {
				if aws.BoolValue(instanceOption.ReadReplicaCapable) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk(names.AttrStorageType); ok {
				if aws.StringValue(instanceOption.StorageType) != v.(string) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_clusters"); ok {
				if aws.BoolValue(instanceOption.SupportsClusters) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_enhanced_monitoring"); ok {
				if aws.BoolValue(instanceOption.SupportsEnhancedMonitoring) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_global_databases"); ok {
				if aws.BoolValue(instanceOption.SupportsGlobalDatabases) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_iam_database_authentication"); ok {
				if aws.BoolValue(instanceOption.SupportsIAMDatabaseAuthentication) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_iops"); ok {
				if aws.BoolValue(instanceOption.SupportsIops) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_kerberos_authentication"); ok {
				if aws.BoolValue(instanceOption.SupportsKerberosAuthentication) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_multi_az"); ok {
				if aws.BoolValue(instanceOption.MultiAZCapable) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_performance_insights"); ok {
				if aws.BoolValue(instanceOption.SupportsPerformanceInsights) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_storage_autoscaling"); ok {
				if aws.BoolValue(instanceOption.SupportsStorageAutoscaling) != v.(bool) {
					continue
				}
			}

			if v, ok := d.GetOk("supports_storage_encryption"); ok {
				if aws.BoolValue(instanceOption.SupportsStorageEncryption) != v.(bool) {
					continue
				}
			}

			instanceClassResults = append(instanceClassResults, instanceOption)
		}
		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Orderable DB Instance Options: %s", err)
	}

	if len(instanceClassResults) == 0 {
		return sdkdiag.AppendErrorf(diags, "no RDS Orderable DB Instance Options found matching criteria; try different search")
	}

	if v, ok := d.GetOk("supported_engine_modes"); ok && len(v.([]interface{})) > 0 {
		var matches []*rds.OrderableDBInstanceOption
		search := flex.ExpandStringValueList(v.([]interface{}))

		for _, ic := range instanceClassResults {
		searchedModes:
			for _, s := range search {
				for _, mode := range ic.SupportedEngineModes {
					if aws.StringValue(mode) == s {
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

	if v, ok := d.GetOk("supported_network_types"); ok && len(v.([]interface{})) > 0 {
		var matches []*rds.OrderableDBInstanceOption
		search := flex.ExpandStringValueList(v.([]interface{}))

		for _, ic := range instanceClassResults {
		searchedNetworks:
			for _, s := range search {
				for _, netType := range ic.SupportedNetworkTypes {
					if aws.StringValue(netType) == s {
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

	if v, ok := d.GetOk("preferred_engine_versions"); ok && len(v.([]interface{})) > 0 {
		var matches []*rds.OrderableDBInstanceOption
		search := flex.ExpandStringValueList(v.([]interface{}))

		for _, s := range search {
			for _, ic := range instanceClassResults {
				if aws.StringValue(ic.EngineVersion) == s {
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

	if v, ok := d.GetOk("preferred_instance_classes"); ok && len(v.([]interface{})) > 0 {
		var matches []*rds.OrderableDBInstanceOption
		search := flex.ExpandStringValueList(v.([]interface{}))

		for _, s := range search {
			for _, ic := range instanceClassResults {
				if aws.StringValue(ic.DBInstanceClass) == s {
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

	var found *rds.OrderableDBInstanceOption

	if latestVersion && prefSearch {
		sortInstanceClassesByVersion(instanceClassResults)
		found = instanceClassResults[len(instanceClassResults)-1]
	}

	if found == nil && len(instanceClassResults) > 0 && prefSearch {
		found = instanceClassResults[0]
	}

	if found == nil && len(instanceClassResults) == 1 {
		found = instanceClassResults[0]
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

	d.SetId(aws.StringValue(found.DBInstanceClass))
	d.Set("availability_zone_group", found.AvailabilityZoneGroup)
	var availabilityZones []string
	for _, v := range found.AvailabilityZones {
		availabilityZones = append(availabilityZones, aws.StringValue(v.Name))
	}
	d.Set(names.AttrAvailabilityZones, availabilityZones)
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
	d.Set("supported_engine_modes", aws.StringValueSlice(found.SupportedEngineModes))
	d.Set("supported_network_types", aws.StringValueSlice(found.SupportedNetworkTypes))
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

func sortInstanceClassesByVersion(ic []*rds.OrderableDBInstanceOption) {
	if len(ic) < 2 {
		return
	}

	sort.Slice(ic, func(i, j int) bool {
		return version.LessThan(aws.StringValue(ic[i].EngineVersion), aws.StringValue(ic[j].EngineVersion))
	})
}
