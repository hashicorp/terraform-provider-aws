// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_batch_job_queue", name="Job Queue")
// @Tags
func DataSourceJobQueue() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceJobQueueRead,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},

			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			"scheduling_policy_arn": {
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

			names.AttrTags: tftags.TagsSchemaComputed(),

			names.AttrPriority: {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"compute_environment_order": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_environment": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"order": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceJobQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	params := &batch.DescribeJobQueuesInput{
		JobQueues: []*string{aws.String(d.Get(names.AttrName).(string))},
	}
	log.Printf("[DEBUG] Reading Batch Job Queue: %s", params)
	desc, err := conn.DescribeJobQueuesWithContext(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Batch Job Queue (%s): %s", d.Get(names.AttrName).(string), err)
	}

	if l := len(desc.JobQueues); l == 0 {
		return sdkdiag.AppendErrorf(diags, "reading Batch Job Queue (%s): empty response", d.Get(names.AttrName).(string))
	} else if l > 1 {
		return sdkdiag.AppendErrorf(diags, "reading Batch Job Queue (%s): too many results: wanted 1, got %d", d.Get(names.AttrName).(string), l)
	}

	jobQueue := desc.JobQueues[0]
	d.SetId(aws.StringValue(jobQueue.JobQueueArn))
	d.Set(names.AttrARN, jobQueue.JobQueueArn)
	d.Set(names.AttrName, jobQueue.JobQueueName)
	d.Set("scheduling_policy_arn", jobQueue.SchedulingPolicyArn)
	d.Set(names.AttrStatus, jobQueue.Status)
	d.Set(names.AttrStatusReason, jobQueue.StatusReason)
	d.Set(names.AttrState, jobQueue.State)
	d.Set(names.AttrPriority, jobQueue.Priority)

	ceos := make([]map[string]interface{}, 0)
	for _, v := range jobQueue.ComputeEnvironmentOrder {
		ceo := map[string]interface{}{}
		ceo["compute_environment"] = aws.StringValue(v.ComputeEnvironment)
		ceo["order"] = int(aws.Int64Value(v.Order))
		ceos = append(ceos, ceo)
	}
	if err := d.Set("compute_environment_order", ceos); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting compute_environment_order: %s", err)
	}

	setTagsOut(ctx, jobQueue.Tags)

	return diags
}
