package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsEcsService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcsServiceRead,

		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"desired_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"launch_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"task_definition": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAwsEcsServiceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	params := &ecs.DescribeServicesInput{
		Cluster:  aws.String(d.Get("cluster_arn").(string)),
		Services: []*string{aws.String(d.Get("service_name").(string))},
	}

	log.Printf("[DEBUG] Reading ECS Service: %s", params)
	desc, err := conn.DescribeServices(params)

	if err != nil {
		return err
	}

	for _, service := range desc.Services {
		if aws.StringValue(service.ClusterArn) != d.Get("cluster_arn").(string) {
			continue
		}
		if aws.StringValue(service.ServiceName) != d.Get("service_name").(string) {
			continue
		}
		d.SetId(aws.StringValue(service.ServiceArn))
		d.Set("service_name", service.ServiceName)
		d.Set("arn", service.ServiceArn)
		d.Set("cluster_arn", service.ClusterArn)
		d.Set("desired_count", service.DesiredCount)
		d.Set("launch_type", service.LaunchType)
		d.Set("task_definition", service.TaskDefinition)
	}

	if d.Id() == "" {
		return fmt.Errorf("service with name %q in cluster %q not found", d.Get("service_name").(string), d.Get("cluster_arn").(string))
	}

	return nil
}
