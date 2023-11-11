// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_memorydb_snapshot")
func DataSourceSnapshot() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSnapshotRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"maintenance_window": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"node_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"num_shards": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"parameter_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"snapshot_retention_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"snapshot_window": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"topic_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MemoryDBConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	snapshot, err := FindSnapshotByName(ctx, conn, name)

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("MemoryDB Snapshot", err))
	}

	d.SetId(aws.StringValue(snapshot.Name))

	d.Set("arn", snapshot.ARN)
	if err := d.Set("cluster_configuration", flattenClusterConfiguration(snapshot.ClusterConfiguration)); err != nil {
		return diag.Errorf("failed to set cluster_configuration for MemoryDB Snapshot (%s): %s", d.Id(), err)
	}
	d.Set("cluster_name", snapshot.ClusterConfiguration.Name)
	d.Set("kms_key_arn", snapshot.KmsKeyId)
	d.Set("name", snapshot.Name)
	d.Set("source", snapshot.Source)

	tags, err := listTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("listing tags for MemoryDB Snapshot (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
