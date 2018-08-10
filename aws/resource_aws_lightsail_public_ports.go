package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsLightsailPublicPorts() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLightsailPublicPortsCreate,
		Read:   resourceAwsLightsailPublicPortsRead,
		Delete: resourceAwsLightsailPublicPortsDelete,

		Schema: map[string]*schema.Schema{
			"instance_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"port_infos": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
						"to_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								lightsail.NetworkProtocolUdp,
								lightsail.NetworkProtocolTcp,
							}, false),
						},
					},
				},
			},
		},
	}
}

func buildAwsLightsailPortInfos(info []interface{}) []*lightsail.PortInfo {
	if len(info) == 0 {
		return nil
	}

	var portInfos []*lightsail.PortInfo

	for _, v := range info {
		m := v.(map[string]interface{})
		portInfos = append(portInfos, &lightsail.PortInfo{
			FromPort: aws.Int64((int64)(m["from_port"].(int))),
			ToPort:   aws.Int64((int64)(m["to_port"].(int))),
			Protocol: aws.String(m["protocol"].(string)),
		})
	}
	return portInfos
}

func resourceAwsLightsailPublicPortsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	_, err := conn.PutInstancePublicPorts(&lightsail.PutInstancePublicPortsInput{
		InstanceName: aws.String(d.Get("instance_name").(string)),
		PortInfos:    buildAwsLightsailPortInfos(d.Get("port_infos").([]interface{})),
	})

	if err != nil {
		return err
	}

	return resourceAwsLightsailPublicPortsRead(d, meta)
}

func resourceAwsLightsailPublicPortsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	_, err := conn.GetInstancePortStates(&lightsail.GetInstancePortStatesInput{
		InstanceName: aws.String(d.Get("instance_name").(string)),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				log.Printf("[WARN] Lightsail Instance (%s) not found, removing from state", d.Get("instance_name"))
				return nil
			}
			return err
		}
		return err
	}

	return nil
}

func resourceAwsLightsailPublicPortsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	_, err := conn.PutInstancePublicPorts(&lightsail.PutInstancePublicPortsInput{
		InstanceName: aws.String(d.Get("instance_name").(string)),
	})

	if err != nil {
		return err
	}
	return nil
}
