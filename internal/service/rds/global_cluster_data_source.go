// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_rds_global_cluster")
func DataSourceGlobalCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGlobalClusterRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
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
			"source_db_cluster_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceGlobalClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	globalClusterIdentifier := d.Get("global_cluster_identifier").(string)

	resp, err := conn.DescribeGlobalClusters(&rds.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(globalClusterIdentifier)})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Global RDS Cluster (%s): %s", globalClusterIdentifier, err)
	}

	if resp == nil {
		return sdkdiag.AppendErrorf(diags, "reading Global RDS Cluster (%s): empty response", globalClusterIdentifier)
	}

	var globalCluster *rds.GlobalCluster
	for _, c := range resp.GlobalClusters {
		if aws.StringValue(c.GlobalClusterIdentifier) == globalClusterIdentifier {
			globalCluster = c
			break
		}
	}

	if globalCluster == nil {
		return sdkdiag.AppendErrorf(diags, "reading Global RDS Cluster (%s): cluster not found", globalClusterIdentifier)
	}

	d.SetId(aws.StringValue(globalCluster.GlobalClusterIdentifier))

	d.Set("arn", globalCluster.GlobalClusterArn)
	d.Set("database_name", globalCluster.DatabaseName)
	d.Set("deletion_protection", globalCluster.DeletionProtection)
	d.Set("engine", globalCluster.Engine)
	d.Set("engine_version", globalCluster.EngineVersion)
	d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)

	var gcmList []interface{}
	for _, gcm := range globalCluster.GlobalClusterMembers {
		gcmMap := map[string]interface{}{
			"db_cluster_arn": aws.StringValue(gcm.DBClusterArn),
			"is_writer":      aws.BoolValue(gcm.IsWriter),
		}

		gcmList = append(gcmList, gcmMap)
	}
	if err := d.Set("global_cluster_members", gcmList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting global_cluster_members: %s", err)
	}

	if err := d.Set("global_cluster_members", flattenGlobalClusterMembers(globalCluster.GlobalClusterMembers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting global_cluster_members: %s", err)
	}

	d.Set("global_cluster_resource_id", globalCluster.GlobalClusterResourceId)
	d.Set("storage_encrypted", globalCluster.StorageEncrypted)

	return diags
}
