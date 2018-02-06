package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsEcsContainerDefinition() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcsContainerDefinitionRead,

		Schema: map[string]*schema.Schema{
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// Computed values.
			"image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_digest": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memory_reservation": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"disable_networking": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"docker_labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     schema.TypeString,
			},
			"environment": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     schema.TypeString,
			},
			"port_mappings": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"container_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsEcsContainerDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
	}
	log.Printf("[DEBUG] Reading ECS Container Definition: %s", params)
	desc, err := conn.DescribeTaskDefinition(params)

	if err != nil {
		return err
	}

	taskDefinition := *desc.TaskDefinition
	for _, def := range taskDefinition.ContainerDefinitions {
		if aws.StringValue(def.Name) != d.Get("container_name").(string) {
			continue
		}

		d.SetId(fmt.Sprintf("%s/%s", aws.StringValue(taskDefinition.TaskDefinitionArn), d.Get("container_name").(string)))
		d.Set("image", aws.StringValue(def.Image))
		image := aws.StringValue(def.Image)
		if strings.Contains(image, ":") {
			d.Set("image_digest", strings.Split(image, ":")[1])
		}
		d.Set("cpu", aws.Int64Value(def.Cpu))
		d.Set("memory", aws.Int64Value(def.Memory))
		d.Set("memory_reservation", aws.Int64Value(def.MemoryReservation))
		d.Set("disable_networking", aws.BoolValue(def.DisableNetworking))
		d.Set("docker_labels", aws.StringValueMap(def.DockerLabels))

		var environment = map[string]string{}
		for _, envKeyValuePair := range def.Environment {
			environment[aws.StringValue(envKeyValuePair.Name)] = aws.StringValue(envKeyValuePair.Value)
		}
		d.Set("environment", environment)

		var portMappings = make([]map[string]string, len(def.PortMappings), len(def.PortMappings))
		for _, portKeyValuePair := range def.PortMappings {
			var portMapping = make(map[string]string)
			portMapping["container_port"] = fmt.Sprintf("%d", aws.Int64Value(portKeyValuePair.ContainerPort))
			portMapping["host_port"] = fmt.Sprintf("%d", aws.Int64Value(portKeyValuePair.HostPort))
			portMapping["protocol"] = aws.StringValue(portKeyValuePair.Protocol)
			portMappings = append(portMappings, portMapping)
		}
		d.Set("port_mappings", portMappings)
	}

	if d.Id() == "" {
		return fmt.Errorf("container with name %q not found in task definition %q", d.Get("container_name").(string), d.Get("task_definition").(string))
	}

	return nil
}
