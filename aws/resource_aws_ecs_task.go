package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
				Optional: true,
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
						"network_interfaces": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"attachment_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"private_ipv4_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"attachments": {
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
			},
			"health_status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
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

	task := out.Tasks[0] // DescribeTasks returns a list but we're only passing in a single task ARN

	d.SetId(taskArn)
	d.Set("arn", &taskArn)
	d.Set("cluster_arn", task.ClusterArn)
	d.Set("task_definition_arn", task.TaskDefinitionArn)
	d.Set("group", task.Group)
	d.Set("launch_type", task.LaunchType)
	d.Set("health_status", task.HealthStatus)

	if err := d.Set("containers", flattenEcsTaskContainers(task.Containers)); err != nil {
		return err
	}

	if err := d.Set("attachments", flattenEcsTaskAttachments(task.Attachments)); err != nil {
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

func flattenEcsTaskContainers(config []*ecs.Container) []map[string]interface{} {
	containers := make([]map[string]interface{}, 0, len(config))

	for _, raw := range config {
		item := make(map[string]interface{})
		item["container_arn"] = *raw.ContainerArn
		item["task_arn"] = *raw.TaskArn
		item["network_interfaces"] = flattenEcsTaskContainerNetworkInterfaces(raw.NetworkInterfaces)

		containers = append(containers, item)
	}

	return containers
}

func flattenEcsTaskContainerNetworkInterfaces(config []*ecs.NetworkInterface) []map[string]interface{} {
	networkInterfaces := make([]map[string]interface{}, 0, len(config))

	for _, raw := range config {
		item := make(map[string]interface{})
		item["attachment_id"] = *raw.AttachmentId
		item["private_ipv4_address"] = *raw.PrivateIpv4Address

		networkInterfaces = append(networkInterfaces, item)
	}

	return networkInterfaces
}

func flattenEcsTaskAttachments(config []*ecs.Attachment) []map[string]interface{} {
	attachments := make([]map[string]interface{}, 0, len(config))

	for _, raw := range config {
		item := make(map[string]interface{})
		item["id"] = *raw.Id
		item["type"] = *raw.Type
		item["details"] = flattenEcsTaskAttachmentDetails(raw.Details)

		attachments = append(attachments, item)
	}

	return attachments
}

func flattenEcsTaskAttachmentDetails(config []*ecs.KeyValuePair) []map[string]interface{} {
	details := make([]map[string]interface{}, 0, len(config))

	for _, raw := range config {
		item := make(map[string]interface{})
		item["name"] = *raw.Name
		item["value"] = *raw.Value

		details = append(details, item)
	}

	return details
}
