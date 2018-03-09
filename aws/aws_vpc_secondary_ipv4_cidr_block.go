package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsVpcSecondaryIpv4CidrBlock() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcSecondaryIpv4CidrBlockCreate,
		Read:   resourceAwsVpcSecondaryIpv4CidrBlockRead,
		Delete: resourceAwsVpcSecondaryIpv4CidrBlockDelete,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ipv4_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.CIDRNetwork(16, 28), // The allowed block size is between a /28 netmask and /16 netmask.
			},
		},
	}
}

func resourceAwsVpcSecondaryIpv4CidrBlockCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.AssociateVpcCidrBlockInput{
		VpcId:     aws.String(d.Get("vpc_id").(string)),
		CidrBlock: aws.String(d.Get("ipv4_cidr_block").(string)),
	}
	log.Printf("[DEBUG] Creating VPC secondary IPv4 CIDR block: %#v", req)
	resp, err := conn.AssociateVpcCidrBlock(req)
	if err != nil {
		return fmt.Errorf("Error creating VPC secondary IPv4 CIDR block: %s", err.Error())
	}

	d.SetId(aws.StringValue(resp.CidrBlockAssociation.AssociationId))

	return resourceAwsVpcSecondaryIpv4CidrBlockRead(d, meta)
}

func resourceAwsVpcSecondaryIpv4CidrBlockRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcId := d.Get("vpc_id").(string)
	vpcRaw, _, err := VPCStateRefreshFunc(conn, vpcId)()
	if err != nil {
		return fmt.Errorf("Error reading VPC: %s", err.Error())
	}
	if vpcRaw == nil {
		log.Printf("[WARN] VPC (%s) not found, removing secondary IPv4 CIDR block from state", vpcId)
		d.SetId("")
		return nil
	}

	vpc := vpcRaw.(*ec2.Vpc)
	found := false
	for _, cidrAssociation := range vpc.CidrBlockAssociationSet {
		if aws.StringValue(cidrAssociation.AssociationId) == d.Id() {
			found = true
			d.Set("ipv4_cidr_block", cidrAssociation.CidrBlock)
		}
	}
	if !found {
		log.Printf("[WARN] VPC secondary IPv4 CIDR block (%s) not found, removing from state", d.Id())
		d.SetId("")
	}

	return nil
}

func resourceAwsVpcSecondaryIpv4CidrBlockDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Deleting VPC secondary IPv4 CIDR block: %s", d.Id())
	_, err := conn.DisassociateVpcCidrBlock(&ec2.DisassociateVpcCidrBlockInput{
		AssociationId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error deleting VPC secondary IPv4 CIDR block: %s", err.Error())
	}

	return nil
}
