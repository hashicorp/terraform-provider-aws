package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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

	return resourceAwsVpcIpv4CidrBlockAssociationRead(d, meta)
}

func resourceAwsVpcIpv4CidrBlockAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcId := d.Get("vpc_id").(string)
	vpcRaw, _, err := VPCStateRefreshFunc(conn, vpcId)()
	if err != nil {
		return fmt.Errorf("Error reading VPC: %s", err)
	}
	if vpcRaw == nil {
		log.Printf("[WARN] VPC (%s) not found, removing IPv4 CIDR block association from state", vpcId)
		d.SetId("")
		return nil
	}

	vpc := vpcRaw.(*ec2.Vpc)
	found := false
	for _, cidrAssociation := range vpc.CidrBlockAssociationSet {
		if aws.StringValue(cidrAssociation.AssociationId) == d.Id() {
			found = true
			d.Set("cidr_block", cidrAssociation.CidrBlock)
		}
	}
	if !found {
		log.Printf("[WARN] VPC IPv4 CIDR block association (%s) not found, removing from state", d.Id())
		d.SetId("")
	}

	return nil
}

func resourceAwsVpcIpv4CidrBlockAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Deleting VPC IPv4 CIDR block association: %s", d.Id())
	_, err := conn.DisassociateVpcCidrBlock(&ec2.DisassociateVpcCidrBlockInput{
		AssociationId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting VPC IPv4 CIDR block association: %s", err)
	}

	return nil
}
