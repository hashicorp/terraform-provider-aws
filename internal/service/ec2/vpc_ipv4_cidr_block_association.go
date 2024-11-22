// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_ipv4_cidr_block_association", name="VPC IPV4 CIDR Block Association")
func resourceVPCIPv4CIDRBlockAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCIPv4CIDRBlockAssociationCreate,
		ReadWithoutTimeout:   resourceVPCIPv4CIDRBlockAssociationRead,
		DeleteWithoutTimeout: resourceVPCIPv4CIDRBlockAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
			// cidr_block can be set by a value returned from IPAM or explicitly in config.
			if diff.Id() != "" && diff.HasChange(names.AttrCIDRBlock) {
				// If netmask is set then cidr_block is derived from IPAM, ignore changes.
				if diff.Get("ipv4_netmask_length") != 0 {
					return diff.Clear(names.AttrCIDRBlock)
				}
				return diff.ForceNew(names.AttrCIDRBlock)
			}
			return nil
		},

		Schema: map[string]*schema.Schema{
			names.AttrCIDRBlock: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDRNetwork(vpcCIDRMinIPv4Netmask, vpcCIDRMaxIPv4Netmask),
			},
			"ipv4_ipam_pool_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ipv4_netmask_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(vpcCIDRMinIPv4Netmask, vpcCIDRMaxIPv4Netmask),
			},
			names.AttrVPCID: {
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

func resourceVPCIPv4CIDRBlockAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcID := d.Get(names.AttrVPCID).(string)
	input := &ec2.AssociateVpcCidrBlockInput{
		VpcId: aws.String(vpcID),
	}

	if v, ok := d.GetOk(names.AttrCIDRBlock); ok {
		input.CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_ipam_pool_id"); ok {
		input.Ipv4IpamPoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_netmask_length"); ok {
		input.Ipv4NetmaskLength = aws.Int32(int32(v.(int)))
	}

	output, err := conn.AssociateVpcCidrBlock(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC (%s) IPv4 CIDR Block Association: %s", vpcID, err)
	}

	d.SetId(aws.ToString(output.CidrBlockAssociation.AssociationId))

	if _, err := waitVPCCIDRBlockAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC (%s) IPv4 CIDR block (%s) to become associated: %s", vpcID, d.Id(), err)
	}

	return append(diags, resourceVPCIPv4CIDRBlockAssociationRead(ctx, d, meta)...)
}

func resourceVPCIPv4CIDRBlockAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcCidrBlockAssociation, vpc, err := findVPCCIDRBlockAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC IPv4 CIDR Block Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC IPv4 CIDR Block Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCIDRBlock, vpcCidrBlockAssociation.CidrBlock)
	d.Set(names.AttrVPCID, vpc.VpcId)

	return diags
}

func resourceVPCIPv4CIDRBlockAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 VPC IPv4 CIDR Block Association: %s", d.Id())
	_, err := conn.DisassociateVpcCidrBlock(ctx, &ec2.DisassociateVpcCidrBlockInput{
		AssociationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCCIDRBlockAssociationIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPC IPv4 CIDR Block Association (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCCIDRBlockAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC IPv4 CIDR block (%s) to become disassociated: %s", d.Id(), err)
	}

	return diags
}
