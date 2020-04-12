package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEcsTaskDefinition() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcsTaskDefinitionRead,

		Schema: map[string]*schema.Schema{
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"execution_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"memory": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"docker_volume_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"scope": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"autoprovision": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"driver": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"driver_opts": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Computed: true,
									},
									"labels": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Computed: true,
									},
								},
							},
						},
						"efs_volume_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"file_system_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"root_directory": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
				Set: resourceAwsEcsTaskDefinitionVolumeHash,
			},
			"placement_constraints": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expression": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"requires_compatibilities": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"proxy_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"properties": {
							Type:     schema.TypeMap,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsEcsTaskDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
		Include:        []*string{aws.String(ecs.TaskDefinitionFieldTags)},
	}
	log.Printf("[DEBUG] Reading ECS Task Definition: %s", params)
	desc, err := conn.DescribeTaskDefinition(params)

	if err != nil {
		return fmt.Errorf("Failed getting task definition %s %q", err, d.Get("task_definition").(string))
	}

	taskDefinition := *desc.TaskDefinition

	d.SetId(aws.StringValue(taskDefinition.TaskDefinitionArn))
	d.Set("family", aws.StringValue(taskDefinition.Family))
	d.Set("network_mode", aws.StringValue(taskDefinition.NetworkMode))
	d.Set("revision", aws.Int64Value(taskDefinition.Revision))
	d.Set("status", aws.StringValue(taskDefinition.Status))
	d.Set("task_role_arn", aws.StringValue(taskDefinition.TaskRoleArn))
	d.Set("execution_role_arn", taskDefinition.ExecutionRoleArn)
	d.Set("cpu", taskDefinition.Cpu)
	d.Set("memory", taskDefinition.Memory)
	if err := d.Set("tags", keyvaluetags.EcsKeyValueTags(desc.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("volume", flattenEcsVolumes(taskDefinition.Volumes)); err != nil {
		return fmt.Errorf("error setting volume: %s", err)
	}

	if err := d.Set("placement_constraints", flattenPlacementConstraints(taskDefinition.PlacementConstraints)); err != nil {
		log.Printf("[ERR] Error setting placement_constraints for (%s): %s", d.Id(), err)
	}

	if err := d.Set("requires_compatibilities", flattenStringList(taskDefinition.RequiresCompatibilities)); err != nil {
		return fmt.Errorf("error setting requires_compatibilities: %s", err)
	}

	if err := d.Set("proxy_configuration", flattenProxyConfiguration(taskDefinition.ProxyConfiguration)); err != nil {
		return fmt.Errorf("error setting proxy_configuration: %s", err)
	}

	if d.Id() == "" {
		return fmt.Errorf("task definition %q not found", d.Get("task_definition").(string))
	}

	return nil
}
