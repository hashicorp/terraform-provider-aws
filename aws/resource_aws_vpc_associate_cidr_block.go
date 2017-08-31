package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsVpcAssociateCidrBlock() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcAssociateCidrBlockCreate,
		Read:   resourceAwsVpcAssociateCidrBlockRead,
		Delete: resourceAwsVpcAssociateCidrBlockDelete,

		Schema: map[string]*schema.Schema{

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ipv4_cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},

			"assign_generated_ipv6_cidr_block": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ForceNew:      true,
				ConflictsWith: []string{"ipv4_cidr_block"},
			},

			"ipv6_cidr_block": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsVpcAssociateCidrBlockCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	params := &ec2.AssociateVpcCidrBlockInput{
		VpcId: aws.String(d.Get("vpc_id").(string)),
		AmazonProvidedIpv6CidrBlock: aws.Bool(d.Get("assign_generated_ipv6_cidr_block").(bool)),
	}

	if v, ok := d.GetOk("ipv4_cidr_block"); ok {
		params.CidrBlock = aws.String(v.(string))
	}

	resp, err := conn.AssociateVpcCidrBlock(params)
	if err != nil {
		return err
	}

	if d.Get("assign_generated_ipv6_cidr_block").(bool) == true {
		d.SetId(*resp.Ipv6CidrBlockAssociation.AssociationId)
	} else {
		d.SetId(*resp.CidrBlockAssociation.AssociationId)
	}

	return resourceAwsVpcAssociateCidrBlockRead(d, meta)
}

func resourceAwsVpcAssociateCidrBlockRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcRaw, _, err := VPCStateRefreshFunc(conn, d.Get("vpc_id").(string))()
	if err != nil {
		return err
	}

	if vpcRaw == nil {
		log.Printf("[INFO] No VPC Found for id %q", d.Get("vpc_id").(string))
		d.SetId("")
		return nil
	}

	vpc := vpcRaw.(*ec2.Vpc)
	found := false

	if d.Get("assign_generated_ipv6_cidr_block").(bool) == true {
		for _, ipv6Association := range vpc.Ipv6CidrBlockAssociationSet {
			if *ipv6Association.AssociationId == d.Id() {
				found = true
				d.Set("ipv6_cidr_block", ipv6Association.Ipv6CidrBlock)
				break
			}
		}
	} else {
		for _, cidrAssociation := range vpc.CidrBlockAssociationSet {
			if *cidrAssociation.AssociationId == d.Id() {
				found = true
				d.Set("ipv4_cidr_block", cidrAssociation.CidrBlock)
			}
		}
	}

	if !found {
		log.Printf("[INFO] No VPC CIDR Association found for ID: %q", d.Id())
	}

	return nil
}

func resourceAwsVpcAssociateCidrBlockDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	params := &ec2.DisassociateVpcCidrBlockInput{
		AssociationId: aws.String(d.Id()),
	}

	_, err := conn.DisassociateVpcCidrBlock(params)
	if err != nil {
		return err
	}

	return nil
}
