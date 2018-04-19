package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsBatchComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsBatchComputeEnvironmentRead,

		Schema: map[string]*schema.Schema{
			"compute_environment_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ecs_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"service_role": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsBatchComputeEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	params := &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: []*string{aws.String(d.Get("compute_environment_name").(string))},
	}
	log.Printf("[DEBUG] Reading Batch Compute Environment: %s", params)
	desc, err := conn.DescribeComputeEnvironments(params)

	if err != nil {
		return err
	}

	for _, computeEnvironment := range desc.ComputeEnvironments {
		if aws.StringValue(computeEnvironment.ComputeEnvironmentName) != d.Get("compute_environment_name").(string) {
			continue
		}
		d.SetId(aws.StringValue(computeEnvironment.ComputeEnvironmentArn))
		d.Set("arn", computeEnvironment.ComputeEnvironmentArn)
		d.Set("ecs_cluster_arn", computeEnvironment.EcsClusterArn)
		d.Set("service_role", computeEnvironment.ServiceRole)
		d.Set("type", computeEnvironment.Type)
		d.Set("status", computeEnvironment.Status)
		d.Set("status_reason", computeEnvironment.StatusReason)
		d.Set("state", computeEnvironment.State)
	}

	if d.Id() == "" {
		return fmt.Errorf("compute environment with name %q not found", d.Get("compute_environment_name").(string))
	}

	return nil
}
