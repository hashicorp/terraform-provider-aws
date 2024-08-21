// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ecs_task_definition", name="Task Definition")
func dataSourceTaskDefinition() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTaskDefinitionRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn_without_revision": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrExecutionRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFamily: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			"task_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTaskDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	taskDefinitionName := d.Get("task_definition").(string)
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
	}

	taskDefinition, _, err := findTaskDefinition(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Definition (%s): %s", taskDefinitionName, err)
	}

	d.SetId(aws.ToString(taskDefinition.TaskDefinitionArn))
	d.Set(names.AttrARN, taskDefinition.TaskDefinitionArn)
	d.Set("arn_without_revision", taskDefinitionARNStripRevision(aws.ToString(taskDefinition.TaskDefinitionArn)))
	d.Set(names.AttrExecutionRoleARN, taskDefinition.ExecutionRoleArn)
	d.Set(names.AttrFamily, taskDefinition.Family)
	d.Set("network_mode", taskDefinition.NetworkMode)
	d.Set("revision", taskDefinition.Revision)
	d.Set(names.AttrStatus, taskDefinition.Status)
	d.Set("task_role_arn", taskDefinition.TaskRoleArn)

	return diags
}
