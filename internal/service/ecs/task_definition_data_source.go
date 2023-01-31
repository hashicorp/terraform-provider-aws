package ecs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

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
	conn := meta.(*conns.AWSClient).ECSConn()

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
	}
	log.Printf("[DEBUG] Reading ECS Task Definition: %s", params)
	desc, err := conn.DescribeTaskDefinitionWithContext(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting task definition %q: %s", d.Get("task_definition").(string), err)
	}

	if desc == nil || desc.TaskDefinition == nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Definition: empty response")
	}

	taskDefinition := desc.TaskDefinition

	d.SetId(aws.StringValue(taskDefinition.TaskDefinitionArn))
	d.Set("arn", taskDefinition.TaskDefinitionArn)
	d.Set("family", taskDefinition.Family)
	d.Set("network_mode", taskDefinition.NetworkMode)
	d.Set("revision", taskDefinition.Revision)
	d.Set("status", taskDefinition.Status)
	d.Set("task_role_arn", taskDefinition.TaskRoleArn)

	if d.Id() == "" {
		return sdkdiag.AppendErrorf(diags, "task definition %q not found", d.Get("task_definition").(string))
	}

	return diags
}
