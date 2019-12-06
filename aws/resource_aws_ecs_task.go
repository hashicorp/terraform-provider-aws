package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	//"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsEcsTask() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcsTaskUpdate,
		Read:   resourceAwsEcsTaskRead,
		Update: resourceAwsEcsTaskUpdate,
		Delete: resourceAwsEcsTaskDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"task_definition_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"containers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"task_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						//"network_interfaces": {
						//	Type:     schema.TypeList,
						//	Computed: true,
						//	Elem: &schema.Resource{
						//		Schema: map[string]*schema.Schema{
						//			"attachment_id": {
						//				Type:     schema.TypeString,
						//				Computed: true,
						//			},
						//			"private_ipv4_address": {
						//				Type:     schema.TypeString,
						//				Computed: true,
						//			},
						//		},
						//	},
						//},
					},
				},
			},
			/*"attachments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"details": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"value": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},*/
		},
	}
}

func resourceAwsEcsTaskRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	taskArn := d.Get("arn").(string)
	clusterArn := d.Get("cluster_arn").(string)
	log.Printf("[DEBUG] Reading task %s", d.Id())
	out, err := conn.DescribeTasks(&ecs.DescribeTasksInput{
		Tasks:   []*string{aws.String(taskArn)},
		Cluster: aws.String(clusterArn),
	})
	if err != nil {
		return err
	}

	task := out.Tasks[0]	// DescribeTasks returns a list but we're only passing in a single task ARN

	d.SetId(taskArn)
	d.Set("arn", &taskArn)
	d.Set("cluster_arn", task.ClusterArn)
	d.Set("task_definition_arn", task.TaskDefinitionArn)
	d.Set("group", task.Group)
	d.Set("launch_type", task.LaunchType)

	containers, err := flattenEcsTaskContainers(task.Containers)
	if err != nil {
		return err
	}
	err = d.Set("containers", containers)
	if err != nil {
		return err
	}
	
	return nil
}

func resourceAwsEcsTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsEcsTaskDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func flattenEcsTaskContainers(config []*ecs.Containers) []map[string]interface{} {
	containers := make([]map[string]interface{}, 0, len(config))

	for _, raw := range config {
		item := make(map[string]interface{})
		item["container_arn"] = *raw.ContainerArn
		item["task_arn"] = *raw.TaskArn

		containers = append(containers, item)
	}

	return containers
}
