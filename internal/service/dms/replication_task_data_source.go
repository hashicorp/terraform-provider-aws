// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dms_replication_task", name="Replication Task")
func dataSourceReplicationTask() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReplicationTaskRead,

		Schema: map[string]*schema.Schema{
			"cdc_start_position": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cdc_start_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"migration_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_instance_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_task_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_task_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"replication_task_settings": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_endpoint_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"start_replication_task": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"table_mappings": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"target_endpoint_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceReplicationTaskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DMSClient(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	taskID := d.Get("replication_task_id").(string)
	task, err := findReplicationTaskByID(ctx, conn, taskID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Task (%s): %s", taskID, err)
	}

	d.SetId(aws.ToString(task.ReplicationTaskIdentifier))
	d.Set("cdc_start_position", task.CdcStartPosition)
	d.Set("migration_type", task.MigrationType)
	d.Set("replication_instance_arn", task.ReplicationInstanceArn)
	d.Set("replication_task_arn", task.ReplicationTaskArn)
	d.Set("replication_task_id", task.ReplicationTaskIdentifier)
	d.Set("replication_task_settings", task.ReplicationTaskSettings)
	d.Set("source_endpoint_arn", task.SourceEndpointArn)
	d.Set(names.AttrStatus, task.Status)
	d.Set("table_mappings", task.TableMappings)
	d.Set("target_endpoint_arn", task.TargetEndpointArn)

	tags, err := listTags(ctx, conn, aws.ToString(task.ReplicationTaskArn))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing DMS Replication Task (%s) tags: %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set(names.AttrTags, tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
