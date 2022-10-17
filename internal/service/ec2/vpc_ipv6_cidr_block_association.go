package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCIPv6CIDRBlockAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCIPv6CIDRBlockAssociationCreate,
		Read:   resourceVPCIPv6CIDRBlockAssociationRead,
		Delete: resourceVPCIPv6CIDRBlockAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
			// ipv6_cidr_block can be set by a value returned from IPAM or explicitly in config.
			if diff.Id() != "" && diff.HasChange("ipv6_cidr_block") {
				// If netmask is set then ipv6_cidr_block is derived from IPAM, ignore changes.
				if diff.Get("ipv6_netmask_length") != 0 {
					return diff.Clear("ipv6_cidr_block")
				}
				return diff.ForceNew("ipv6_cidr_block")
			}
			return nil
		},
		Schema: map[string]*schema.Schema{
			"ipv6_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					verify.ValidIPv6CIDRNetworkAddress,
					validation.IsCIDRNetwork(VPCCIDRMaxIPv6, VPCCIDRMaxIPv6)),
			},
			// ipam parameters are not required by the API but other usage mechanisms are not implemented yet. TODO ipv6 options:
			// --amazon-provided-ipv6-cidr-block
			// --ipv6-pool
			"ipv6_ipam_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ipv6_netmask_length": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IntInSlice([]int{VPCCIDRMaxIPv6}),
				ConflictsWith: []string{"ipv6_cidr_block"},
				// This RequiredWith setting should be applied once L57 is completed
				// RequiredWith:  []string{"ipv6_ipam_pool_id"},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceVPCIPv6CIDRBlockAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcID := d.Get("vpc_id").(string)
	input := &ec2.AssociateVpcCidrBlockInput{
		VpcId: aws.String(vpcID),
	}

	if v, ok := d.GetOk("ipv6_cidr_block"); ok {
		input.Ipv6CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_ipam_pool_id"); ok {
		input.Ipv6IpamPoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv6_netmask_length"); ok {
		input.Ipv6NetmaskLength = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating EC2 VPC IPv6 CIDR Block Association: %s", input)
	output, err := conn.AssociateVpcCidrBlock(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 VPC (%s) IPv6 CIDR Block Association: %w", vpcID, err)
	}

	d.SetId(aws.StringValue(output.Ipv6CidrBlockAssociation.AssociationId))

	_, err = WaitVPCIPv6CIDRBlockAssociationCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for EC2 VPC (%s) IPv6 CIDR block (%s) to become associated: %w", vpcID, d.Id(), err)
	}

	return resourceVPCIPv6CIDRBlockAssociationRead(d, meta)
}

func resourceVPCIPv6CIDRBlockAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcIpv6CidrBlockAssociation, vpc, err := FindVPCIPv6CIDRBlockAssociationByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC IPv6 CIDR Block Association %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC IPv6 CIDR Block Association (%s): %w", d.Id(), err)
	}

	d.Set("ipv6_cidr_block", vpcIpv6CidrBlockAssociation.Ipv6CidrBlock)
	d.Set("vpc_id", vpc.VpcId)

	return nil
}

func resourceVPCIPv6CIDRBlockAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting VPC IPv6 CIDR Block Association: %s", d.Id())
	_, err := conn.DisassociateVpcCidrBlock(&ec2.DisassociateVpcCidrBlockInput{
		AssociationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCCIDRBlockAssociationIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 VPC IPv6 CIDR Block Association (%s): %w", d.Id(), err)
	}

	_, err = WaitVPCIPv6CIDRBlockAssociationDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for EC2 VPC IPv6 CIDR block (%s) to become disassociated: %w", d.Id(), err)
	}

	return nil
}
