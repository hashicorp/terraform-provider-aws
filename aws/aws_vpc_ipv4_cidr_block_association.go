package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

const (
	VpcCidrBlockStateCodeDeleted = "deleted"
)

func resourceAwsVpcIpv4CidrBlockAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcIpv4CidrBlockAssociationCreate,
		Read:   resourceAwsVpcIpv4CidrBlockAssociationRead,
		Delete: resourceAwsVpcIpv4CidrBlockAssociationDelete,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.CIDRNetwork(16, 28), // The allowed block size is between a /28 netmask and /16 netmask.
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsVpcIpv4CidrBlockAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.AssociateVpcCidrBlockInput{
		VpcId:     aws.String(d.Get("vpc_id").(string)),
		CidrBlock: aws.String(d.Get("cidr_block").(string)),
	}
	log.Printf("[DEBUG] Creating VPC IPv4 CIDR block association: %#v", req)
	resp, err := conn.AssociateVpcCidrBlock(req)
	if err != nil {
		return fmt.Errorf("Error creating VPC IPv4 CIDR block association: %s", err)
	}

	d.SetId(aws.StringValue(resp.CidrBlockAssociation.AssociationId))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociating},
		Target:     []string{ec2.VpcCidrBlockStateCodeAssociated},
		Refresh:    vpcIpv4CidrBlockAssociationStateRefresh(conn, d.Get("vpc_id").(string), d.Id()),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for IPv4 CIDR block association (%s) to become available: %s", d.Id(), err)
	}

	return resourceAwsVpcIpv4CidrBlockAssociationRead(d, meta)
}

func resourceAwsVpcIpv4CidrBlockAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	assocRaw, state, err := vpcIpv4CidrBlockAssociationStateRefresh(conn, d.Get("vpc_id").(string), d.Id())()
	if err != nil {
		return fmt.Errorf("Error reading IPv4 CIDR block association: %s", err)
	}
	if state == VpcCidrBlockStateCodeDeleted {
		log.Printf("[WARN] IPv4 CIDR block association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	assoc := assocRaw.(*ec2.VpcCidrBlockAssociation)
	d.Set("cidr_block", assoc.CidrBlock)

	return nil
}

func resourceAwsVpcIpv4CidrBlockAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Deleting VPC IPv4 CIDR block association: %s", d.Id())
	_, err := conn.DisassociateVpcCidrBlock(&ec2.DisassociateVpcCidrBlockInput{
		AssociationId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, "InvalidVpcID.NotFound", "") {
			return nil
		}
		return fmt.Errorf("Error deleting VPC IPv4 CIDR block association: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeDisassociating},
		Target:     []string{ec2.VpcCidrBlockStateCodeDisassociated, VpcCidrBlockStateCodeDeleted},
		Refresh:    vpcIpv4CidrBlockAssociationStateRefresh(conn, d.Get("vpc_id").(string), d.Id()),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for VPC IPv4 CIDR block association (%s) to be deleted: %s", d.Id(), err.Error())
	}

	return nil
}

func vpcIpv4CidrBlockAssociationStateRefresh(conn *ec2.EC2, vpcId, assocId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpc, err := vpcDescribe(conn, vpcId)
		if err != nil {
			return nil, "", err
		}

		if vpc != nil {
			for _, cidrAssociation := range vpc.CidrBlockAssociationSet {
				if aws.StringValue(cidrAssociation.AssociationId) == assocId {
					return cidrAssociation, aws.StringValue(cidrAssociation.CidrBlockState.State), nil
				}
			}
		}

		return "", VpcCidrBlockStateCodeDeleted, nil
	}
}
