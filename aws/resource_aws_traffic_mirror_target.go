package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsTrafficMirrorTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsTrafficMirrorTargetCreate,
		Read:   resourceAwsTrafficMirrorTargetRead,
		Delete: resourceAwsTrafficMirrorTargetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ConflictsWith: []string{
					"network_load_balancer_arn",
				},
			},
			"network_load_balancer_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ConflictsWith: []string{
					"network_interface_id",
				},
			},
		},
	}
}

func resourceAwsTrafficMirrorTargetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.CreateTrafficMirrorTargetInput{}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_interface_id"); ok {
		input.NetworkInterfaceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_load_balancer_arn"); ok {
		input.NetworkLoadBalancerArn = aws.String(v.(string))
	}

	out, err := conn.CreateTrafficMirrorTarget(input)
	if err != nil {
		return fmt.Errorf("Error creating traffic mirror target %v", err)
	}

	d.SetId(*out.TrafficMirrorTarget.TrafficMirrorTargetId)

	return resourceAwsTrafficMirrorTargetRead(d, meta)
}

func resourceAwsTrafficMirrorTargetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	targetId := d.Id()
	input := &ec2.DescribeTrafficMirrorTargetsInput{
		TrafficMirrorTargetIds: []*string{&targetId},
	}

	out, err := conn.DescribeTrafficMirrorTargets(input)
	if err != nil {
		d.SetId("")
		return fmt.Errorf("Error describing traffic mirror target %v", targetId)
	}

	if 1 != len(out.TrafficMirrorTargets) {
		d.SetId("")
		return fmt.Errorf("Not able to find find target %v", targetId)
	}

	target := out.TrafficMirrorTargets[0]
	if nil != target.Description {
		d.Set("description", *target.Description)
	}

	switch *target.Type {
	case "network-interface":
		d.Set("network_interface_id", target.NetworkInterfaceId)
	case "network-load-balancer":
		d.Set("network_load_balancer_arn", target.NetworkLoadBalancerArn)
	}

	return nil
}

func resourceAwsTrafficMirrorTargetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	targetId := d.Id()
	input := &ec2.DeleteTrafficMirrorTargetInput{
		TrafficMirrorTargetId: &targetId,
	}

	_, err := conn.DeleteTrafficMirrorTarget(input)
	if nil != err {
		return fmt.Errorf("Error deleting traffic mirror target %v", targetId)
	}

	d.SetId("")
	return nil
}
