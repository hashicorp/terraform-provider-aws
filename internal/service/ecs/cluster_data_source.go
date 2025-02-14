// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

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

// @SDKDataSource("aws_ecs_cluster", name="Cluster")
// @Tags
func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrClusterName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"pending_tasks_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"registered_container_instances_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"running_tasks_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"service_connect_defaults": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrNamespace: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"setting": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	clusterName := d.Get(names.AttrClusterName).(string)
	cluster, err := findClusterByNameOrARN(ctx, conn, clusterName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Cluster (%s): %s", clusterName, err)
	}

	arn := aws.ToString(cluster.ClusterArn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set("pending_tasks_count", cluster.PendingTasksCount)
	d.Set("registered_container_instances_count", cluster.RegisteredContainerInstancesCount)
	d.Set("running_tasks_count", cluster.RunningTasksCount)
	if cluster.ServiceConnectDefaults != nil {
		if err := d.Set("service_connect_defaults", []interface{}{flattenClusterServiceConnectDefaults(cluster.ServiceConnectDefaults)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting service_connect_defaults: %s", err)
		}
	} else {
		d.Set("service_connect_defaults", nil)
	}
	if err := d.Set("setting", flattenClusterSettings(cluster.Settings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting setting: %s", err)
	}
	d.Set(names.AttrStatus, cluster.Status)

	setTagsOut(ctx, cluster.Tags)

	return diags
}
