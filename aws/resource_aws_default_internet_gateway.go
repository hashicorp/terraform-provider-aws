package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDefaultInternetGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDefaultInternetGatewayCreate,
		Read:   resourceAwsInternetGatewayRead,
		Update: resourceAwsDefaultInternetGatewayUpdate,
		Delete: resourceAwsDefaultInternetGatewayDelete,

		Schema: map[string]*schema.Schema{
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsDefaultInternetGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	// The Default Internet Gateway is the sole attached gateway in the default VPC.
	descVpcReq := &ec2.DescribeVpcsInput{}
	descVpcReq.Filters = buildEC2AttributeFilterList(map[string]string{
		"isDefault": "true",
	})
	descVpcResp, err := conn.DescribeVpcs(descVpcReq)
	if err != nil {
		return err
	}
	if descVpcResp.Vpcs == nil || len(descVpcResp.Vpcs) == 0 {
		return fmt.Errorf("No default VPC found in this region")
	}

	vpcId := aws.StringValue(descVpcResp.Vpcs[0].VpcId)

	descIgwReq := &ec2.DescribeInternetGatewaysInput{}
	descIgwReq.Filters = buildEC2AttributeFilterList(map[string]string{
		"attachment.vpc-id": vpcId,
		"attachment.state":  "available",
	})
	descIgwResp, err := conn.DescribeInternetGateways(descIgwReq)
	if err != nil {
		return err
	}
	if descIgwResp.InternetGateways == nil || len(descIgwResp.InternetGateways) == 0 {
		return fmt.Errorf("No attached Internet Gateways found in VPC: %s", vpcId)
	}
	if len(descIgwResp.InternetGateways) > 1 {
		return fmt.Errorf("Multiple attached Internet Gateways found in VPC: %s", vpcId)
	}

	d.SetId(aws.StringValue(descIgwResp.InternetGateways[0].InternetGatewayId))
	return resourceAwsDefaultInternetGatewayUpdate(d, meta)
}

func resourceAwsDefaultInternetGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if err := setTags(conn, d); err != nil {
		return err
	}

	return resourceAwsInternetGatewayRead(d, meta)
}

func resourceAwsDefaultInternetGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Default Internet Gateway. Terraform will remove this resource from the state file, however resources may remain.")
	d.SetId("")
	return nil
}
