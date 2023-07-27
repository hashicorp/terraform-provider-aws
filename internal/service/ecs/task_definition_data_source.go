// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_ecs_task_definition")
func DataSourceTaskDefinition() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTaskDefinitionRead,

		Schema: map[string]*schema.Schema{
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Computed values.
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn_without_revision": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"execution_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"family": {
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	taskDefinitionName := d.Get("task_definition").(string)
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
	}

	output, err := conn.DescribeTaskDefinitionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Definition (%s): %s", taskDefinitionName, err)
	}

	taskDefinition := output.TaskDefinition
	d.SetId(aws.StringValue(taskDefinition.TaskDefinitionArn))
	d.Set("arn", taskDefinition.TaskDefinitionArn)
	d.Set("arn_without_revision", StripRevision(aws.StringValue(taskDefinition.TaskDefinitionArn)))
	d.Set("execution_role_arn", taskDefinition.ExecutionRoleArn)
	d.Set("family", taskDefinition.Family)
	d.Set("network_mode", taskDefinition.NetworkMode)
	d.Set("revision", taskDefinition.Revision)
	d.Set("status", taskDefinition.Status)
	d.Set("task_role_arn", taskDefinition.TaskRoleArn)

	return diags
}
