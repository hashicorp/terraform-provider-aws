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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_batch_job_queue", name="Job Queue")
// @Tags
// @Testing(tagsIdentifierAttribute="arn")
func dataSourceJobQueue() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceJobQueueRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
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
			"job_state_time_limit_action": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAction: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"max_time_seconds": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"reason": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrState: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrPriority: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"scheduling_policy_arn": {
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
		},
	}
}

func dataSourceJobQueueRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchClient(ctx)

	jobQueue, err := findJobQueueByID(ctx, conn, d.Get(names.AttrName).(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Batch Job Queue", err))
	}

	arn := aws.ToString(jobQueue.JobQueueArn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrName, jobQueue.JobQueueName)
	d.Set(names.AttrPriority, jobQueue.Priority)
	d.Set("scheduling_policy_arn", jobQueue.SchedulingPolicyArn)
	d.Set(names.AttrState, jobQueue.State)
	d.Set(names.AttrStatus, jobQueue.Status)
	d.Set(names.AttrStatusReason, jobQueue.StatusReason)

	tfList := make([]any, 0)
	for _, apiObject := range jobQueue.ComputeEnvironmentOrder {
		tfMap := map[string]any{}
		tfMap["compute_environment"] = aws.ToString(apiObject.ComputeEnvironment)
		tfMap["order"] = aws.ToInt32(apiObject.Order)
		tfList = append(tfList, tfMap)
	}
	if err := d.Set("compute_environment_order", tfList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting compute_environment_order: %s", err)
	}

	tfList = make([]any, 0)
	for _, apiObject := range jobQueue.JobStateTimeLimitActions {
		tfMap := map[string]any{}
		tfMap[names.AttrAction] = apiObject.Action
		tfMap["max_time_seconds"] = aws.ToInt32(apiObject.MaxTimeSeconds)
		tfMap["reason"] = aws.ToString(apiObject.Reason)
		tfMap[names.AttrState] = apiObject.State
		tfList = append(tfList, tfMap)
	}
	if err := d.Set("job_state_time_limit_action", tfList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting job_state_time_limit_action: %s", err)
	}

	setTagsOut(ctx, jobQueue.Tags)

	return diags
}
