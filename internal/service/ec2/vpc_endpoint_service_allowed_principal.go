package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
)

func ResourceVPCEndpointServiceAllowedPrincipal() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCEndpointServiceAllowedPrincipalCreate,
		Read:   resourceVPCEndpointServiceAllowedPrincipalRead,
		Delete: resourceVPCEndpointServiceAllowedPrincipalDelete,

		Schema: map[string]*schema.Schema{
			"vpc_endpoint_service_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"principal_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCEndpointServiceAllowedPrincipalCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	svcId := d.Get("vpc_endpoint_service_id").(string)
	arn := d.Get("principal_arn").(string)

	_, err := findResourceVPCEndpointServiceAllowedPrincipals(conn, svcId)
	if err != nil {
		return err
	}

	_, err = conn.ModifyVpcEndpointServicePermissions(&ec2.ModifyVpcEndpointServicePermissionsInput{
		ServiceId:            aws.String(svcId),
		AddAllowedPrincipals: aws.StringSlice([]string{arn}),
	})
	if err != nil {
		return fmt.Errorf("Error creating VPC Endpoint Service allowed principal: %s", err.Error())
	}

	d.SetId(vpcEndpointServiceIdPrincipalARNHash(svcId, arn))

	return resourceVPCEndpointServiceAllowedPrincipalRead(d, meta)
}

func resourceVPCEndpointServiceAllowedPrincipalRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	svcId := d.Get("vpc_endpoint_service_id").(string)
	arn := d.Get("principal_arn").(string)

	principals, err := findResourceVPCEndpointServiceAllowedPrincipals(conn, svcId)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, "InvalidVpcEndpointServiceId.NotFound") {
			log.Printf("[WARN]VPC Endpoint Service (%s) not found, removing VPC Endpoint Service allowed principal (%s) from state", svcId, d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	found := false
	for _, principal := range principals {
		if aws.StringValue(principal.Principal) == arn {
			found = true
			break
		}
	}
	if !found {
		log.Printf("[WARN] VPC Endpoint Service allowed principal (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}

func resourceVPCEndpointServiceAllowedPrincipalDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	svcId := d.Get("vpc_endpoint_service_id").(string)
	arn := d.Get("principal_arn").(string)

	_, err := conn.ModifyVpcEndpointServicePermissions(&ec2.ModifyVpcEndpointServicePermissionsInput{
		ServiceId:               aws.String(svcId),
		RemoveAllowedPrincipals: aws.StringSlice([]string{arn}),
	})
	if err != nil {
		if !tfawserr.ErrCodeEquals(err, "InvalidVpcEndpointServiceId.NotFound") {
			return fmt.Errorf("Error deleting VPC Endpoint Service allowed principal: %s", err.Error())
		}
	}

	return nil
}

func findResourceVPCEndpointServiceAllowedPrincipals(conn *ec2.EC2, id string) ([]*ec2.AllowedPrincipal, error) {
	resp, err := conn.DescribeVpcEndpointServicePermissions(&ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(id),
	})
	if err != nil {
		return nil, err
	}

	return resp.AllowedPrincipals, nil
}

func vpcEndpointServiceIdPrincipalARNHash(svcId, arn string) string {
	return fmt.Sprintf("a-%s%d", svcId, create.StringHashcode(arn))
}
