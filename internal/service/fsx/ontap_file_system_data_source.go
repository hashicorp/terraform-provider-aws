// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_fsx_ontap_file_system", name="ONTAP File System")
// @Tags
func dataSourceONTAPFileSystem() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceONTAPFileSystemRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic_backup_retention_days": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"daily_automatic_backup_start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_iops_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_ip_address_range": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"intercluster": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"management": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"ha_pairs": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"preferred_subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"route_table_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"storage_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"storage_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"throughput_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"throughput_capacity_per_ha_pair": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weekly_maintenance_start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceONTAPFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	id := d.Get(names.AttrID).(string)
	filesystem, err := findONTAPFileSystemByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx for NetApp ONTAP File System (%s): %s", id, err)
	}

	ontapConfig := filesystem.OntapConfiguration

	d.SetId(aws.StringValue(filesystem.FileSystemId))
	d.Set(names.AttrARN, filesystem.ResourceARN)
	d.Set("automatic_backup_retention_days", ontapConfig.AutomaticBackupRetentionDays)
	d.Set("daily_automatic_backup_start_time", ontapConfig.DailyAutomaticBackupStartTime)
	d.Set("deployment_type", ontapConfig.DeploymentType)
	if err := d.Set("disk_iops_configuration", flattenOntapFileDiskIopsConfiguration(ontapConfig.DiskIopsConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting disk_iops_configuration: %s", err)
	}
	d.Set("dns_name", filesystem.DNSName)
	d.Set("endpoint_ip_address_range", ontapConfig.EndpointIpAddressRange)
	if err := d.Set("endpoints", flattenOntapFileSystemEndpoints(ontapConfig.Endpoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoints: %s", err)
	}
	haPairs := aws.Int64Value(ontapConfig.HAPairs)
	d.Set("ha_pairs", haPairs)
	d.Set(names.AttrKMSKeyID, filesystem.KmsKeyId)
	d.Set("network_interface_ids", aws.StringValueSlice(filesystem.NetworkInterfaceIds))
	d.Set(names.AttrOwnerID, filesystem.OwnerId)
	d.Set("preferred_subnet_id", ontapConfig.PreferredSubnetId)
	d.Set("route_table_ids", aws.StringValueSlice(ontapConfig.RouteTableIds))
	d.Set("storage_capacity", filesystem.StorageCapacity)
	d.Set("storage_type", filesystem.StorageType)
	d.Set(names.AttrSubnetIDs, aws.StringValueSlice(filesystem.SubnetIds))
	if aws.StringValue(ontapConfig.DeploymentType) == fsx.OntapDeploymentTypeSingleAz2 {
		d.Set("throughput_capacity", nil)
		d.Set("throughput_capacity_per_ha_pair", ontapConfig.ThroughputCapacityPerHAPair)
	} else {
		d.Set("throughput_capacity", ontapConfig.ThroughputCapacity)
		d.Set("throughput_capacity_per_ha_pair", ontapConfig.ThroughputCapacityPerHAPair)
	}
	d.Set(names.AttrVPCID, filesystem.VpcId)
	d.Set("weekly_maintenance_start_time", ontapConfig.WeeklyMaintenanceStartTime)

	setTagsOut(ctx, filesystem.Tags)

	return diags
}
