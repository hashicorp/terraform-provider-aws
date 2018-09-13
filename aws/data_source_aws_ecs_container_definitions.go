package aws

import (
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsEcsContainerDefinitions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcsContainerDefinitionsRead,

		Schema: map[string]*schema.Schema{
			"container_definitions": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"image": {
							Type:     schema.TypeString,
							Required: true,
						},
						"memory": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"port_mappings": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_port": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"host_port": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"protocol": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"command": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
					},
				},
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEcsContainerDefinitionsRead(d *schema.ResourceData, meta interface{}) error {
	awsDefinitions := []ecs.ContainerDefinition{}
	for _, rawDefinition := range d.Get("container_definitions").(*schema.Set).List() {
		definition := rawDefinition.(map[string]interface{})
		awsDefinition := ecs.ContainerDefinition{}
		awsDefinition.Name = aws.String(definition["name"].(string))
		awsDefinition.Image = aws.String(definition["image"].(string))
		for _, rawMapping := range definition["port_mappings"].(*schema.Set).List() {
			mapping := rawMapping.(map[string]interface{})
			awsMapping := &ecs.PortMapping{}
			awsMapping.ContainerPort = aws.Int64(int64(mapping["container_port"].(int)))
			if hostPort, ok := mapping["host_port"].(int); ok {
				awsMapping.HostPort = aws.Int64(int64(hostPort))
			}
			if protocol, ok := mapping["protocol"].(string); ok {
				awsMapping.Protocol = aws.String(protocol)
			}
			awsDefinition.PortMappings = append(awsDefinition.PortMappings, awsMapping)
		}
		for _, rawCommand := range definition["command"].([]interface{}) {
			awsDefinition.Command = append(awsDefinition.Command, aws.String(rawCommand.(string)))
		}
		awsDefinitions = append(awsDefinitions, awsDefinition)
	}

	jsonBytes, err := json.Marshal(awsDefinitions)
	if err != nil {
		return err
	}

	jsonString := string(jsonBytes)

	d.Set("json", jsonString)
	d.SetId(strconv.Itoa(hashcode.String(jsonString)))
	return nil
}
