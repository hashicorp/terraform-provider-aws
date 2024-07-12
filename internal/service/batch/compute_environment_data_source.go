// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_batch_compute_environment", name="Compute Environment")
// @Tags
func DataSourceComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceComputeEnvironmentRead,

		Schema: map[string]*schema.Schema{
			"compute_environment_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ecs_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrServiceRole: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrTags: tftags.TagsSchemaComputed(),

			names.AttrType: {
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

			names.AttrState: {
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

func dataSourceComputeEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	params := &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: []*string{aws.String(d.Get("compute_environment_name").(string))},
	}
	desc, err := conn.DescribeComputeEnvironmentsWithContext(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Batch Compute Environment (%s): %s", d.Get("compute_environment_name").(string), err)
	}

	if l := len(desc.ComputeEnvironments); l == 0 {
		return sdkdiag.AppendErrorf(diags, "reading Batch Compute Environment (%s): empty response", d.Get("compute_environment_name").(string))
	} else if l > 1 {
		return sdkdiag.AppendErrorf(diags, "reading Batch Compute Environment (%s): too many results: wanted 1, got %d", d.Get("compute_environment_name").(string), l)
	}

	computeEnvironment := desc.ComputeEnvironments[0]
	d.SetId(aws.StringValue(computeEnvironment.ComputeEnvironmentArn))
	d.Set(names.AttrARN, computeEnvironment.ComputeEnvironmentArn)
	d.Set("compute_environment_name", computeEnvironment.ComputeEnvironmentName)
	d.Set("ecs_cluster_arn", computeEnvironment.EcsClusterArn)
	d.Set(names.AttrServiceRole, computeEnvironment.ServiceRole)
	d.Set(names.AttrType, computeEnvironment.Type)
	d.Set(names.AttrStatus, computeEnvironment.Status)
	d.Set(names.AttrStatusReason, computeEnvironment.StatusReason)
	d.Set(names.AttrState, computeEnvironment.State)

	if err := d.Set("update_policy", flattenComputeEnvironmentUpdatePolicy(computeEnvironment.UpdatePolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting update_policy: %s", err)
	}

	setTagsOut(ctx, computeEnvironment.Tags)

	return diags
}
