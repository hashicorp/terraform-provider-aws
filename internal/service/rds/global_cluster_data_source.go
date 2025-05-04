// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_rds_global_cluster", name="Global Cluster")
func DataSourceGlobalCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGlobalClusterRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDeletionProtection: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_lifecycle_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version_actual": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"global_cluster_members": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"db_cluster_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_writer": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"global_cluster_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceGlobalClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	globalClusterID := d.Get("global_cluster_identifier").(string)
	globalCluster, err := findGlobalClusterByID(ctx, conn, globalClusterID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Global Cluster (%s): %s", globalClusterID, err)
	}

	d.SetId(aws.ToString(globalCluster.GlobalClusterIdentifier))
	d.Set(names.AttrARN, globalCluster.GlobalClusterArn)
	d.Set(names.AttrDatabaseName, globalCluster.DatabaseName)
	d.Set(names.AttrDeletionProtection, globalCluster.DeletionProtection)
	d.Set(names.AttrEndpoint, globalCluster.Endpoint)
	d.Set(names.AttrEngine, globalCluster.Engine)
	d.Set("engine_lifecycle_support", globalCluster.EngineLifecycleSupport)
	d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
	if err := d.Set("global_cluster_members", flattenGlobalClusterMembers(globalCluster.GlobalClusterMembers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting global_cluster_members: %s", err)
	}
	d.Set("global_cluster_resource_id", globalCluster.GlobalClusterResourceId)
	d.Set(names.AttrStorageEncrypted, globalCluster.StorageEncrypted)

	oldEngineVersion, newEngineVersion := d.Get(names.AttrEngineVersion).(string), aws.ToString(globalCluster.EngineVersion)

	// For example a configured engine_version of "5.6.10a" and a returned engine_version of "5.6.global_10a".
	if oldParts, newParts := strings.Split(oldEngineVersion, "."), strings.Split(newEngineVersion, "."); len(oldParts) == 3 &&
		len(newParts) == 3 &&
		oldParts[0] == newParts[0] &&
		oldParts[1] == newParts[1] &&
		strings.HasSuffix(newParts[2], oldParts[2]) {
		d.Set(names.AttrEngineVersion, oldEngineVersion)
		d.Set("engine_version_actual", newEngineVersion)
	} else {
		d.Set(names.AttrEngineVersion, newEngineVersion)
		d.Set("engine_version_actual", newEngineVersion)
	}

	// Use the same approach as the resource for setting tags
	setTagsOut(ctx, globalCluster.TagList)

	return diags
}
