package ecs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceTaskDefinition() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTaskDefinitionRead,

		Schema: map[string]*schema.Schema{
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Computed values.
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

func dataSourceTaskDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
	}
	log.Printf("[DEBUG] Reading ECS Task Definition: %s", params)
	desc, err := conn.DescribeTaskDefinition(params)

	if err != nil {
		return fmt.Errorf("Failed getting task definition %q: %w", d.Get("task_definition").(string), err)
	}

	if desc == nil || desc.TaskDefinition == nil {
		return fmt.Errorf("error reading ECS Task Definition: empty response")
	}

	taskDefinition := desc.TaskDefinition

	d.SetId(aws.StringValue(taskDefinition.TaskDefinitionArn))
	d.Set("family", taskDefinition.Family)
	d.Set("network_mode", taskDefinition.NetworkMode)
	d.Set("revision", taskDefinition.Revision)
	d.Set("status", taskDefinition.Status)
	d.Set("task_role_arn", taskDefinition.TaskRoleArn)

	if d.Id() == "" {
		return fmt.Errorf("task definition %q not found", d.Get("task_definition").(string))
	}

	return nil
}
