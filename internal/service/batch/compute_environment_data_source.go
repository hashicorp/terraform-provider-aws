// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

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

// @SDKDataSource("aws_batch_compute_environment", name="Compute Environment")
// @Tags
// @Testing(tagsIdentifierAttribute="arn")
func dataSourceComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceComputeEnvironmentRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_environment_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ecs_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrServiceRole: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusReason: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"update_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"job_execution_timeout_minutes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"terminate_jobs_on_update": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceComputeEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	name := d.Get("compute_environment_name").(string)
	computeEnvironment, err := findComputeEnvironmentDetailByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Batch Compute Environment (%s): %s", name, err)
	}

	arn := aws.ToString(computeEnvironment.ComputeEnvironmentArn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set("compute_environment_name", computeEnvironment.ComputeEnvironmentName)
	d.Set("ecs_cluster_arn", computeEnvironment.EcsClusterArn)
	d.Set(names.AttrServiceRole, computeEnvironment.ServiceRole)
	d.Set(names.AttrState, computeEnvironment.State)
	d.Set(names.AttrStatus, computeEnvironment.Status)
	d.Set(names.AttrStatusReason, computeEnvironment.StatusReason)
	d.Set(names.AttrType, computeEnvironment.Type)
	if err := d.Set("update_policy", flattenComputeEnvironmentUpdatePolicy(computeEnvironment.UpdatePolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting update_policy: %s", err)
	}

	setTagsOut(ctx, computeEnvironment.Tags)

	return diags
}
