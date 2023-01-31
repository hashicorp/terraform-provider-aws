package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVPCIPv4CIDRBlockAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCIPv4CIDRBlockAssociationCreate,
		ReadWithoutTimeout:   resourceVPCIPv4CIDRBlockAssociationRead,
		DeleteWithoutTimeout: resourceVPCIPv4CIDRBlockAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
			// cidr_block can be set by a value returned from IPAM or explicitly in config.
			if diff.Id() != "" && diff.HasChange("cidr_block") {
				// If netmask is set then cidr_block is derived from IPAM, ignore changes.
				if diff.Get("ipv4_netmask_length") != 0 {
					return diff.Clear("cidr_block")
				}
				return diff.ForceNew("cidr_block")
			}
			return nil
		},

		Schema: map[string]*schema.Schema{
			"cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDRNetwork(VPCCIDRMinIPv4, VPCCIDRMaxIPv4),
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
				ValidateFunc: validation.IntBetween(VPCCIDRMinIPv4, VPCCIDRMaxIPv4),
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

func resourceVPCIPv4CIDRBlockAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	vpcID := d.Get("vpc_id").(string)
	input := &ec2.AssociateVpcCidrBlockInput{
		VpcId: aws.String(vpcID),
	}

	if v, ok := d.GetOk("cidr_block"); ok {
		input.CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_ipam_pool_id"); ok {
		input.Ipv4IpamPoolId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_netmask_length"); ok {
		input.Ipv4NetmaskLength = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating EC2 VPC IPv4 CIDR Block Association: %s", input)
	output, err := conn.AssociateVpcCidrBlockWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC (%s) IPv4 CIDR Block Association: %s", vpcID, err)
	}

	d.SetId(aws.StringValue(output.CidrBlockAssociation.AssociationId))

	_, err = WaitVPCCIDRBlockAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC (%s) IPv4 CIDR block (%s) to become associated: %s", vpcID, d.Id(), err)
	}

	return append(diags, resourceVPCIPv4CIDRBlockAssociationRead(ctx, d, meta)...)
}

func resourceVPCIPv4CIDRBlockAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	vpcCidrBlockAssociation, vpc, err := FindVPCCIDRBlockAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC IPv4 CIDR Block Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC IPv4 CIDR Block Association (%s): %s", d.Id(), err)
	}

	d.Set("cidr_block", vpcCidrBlockAssociation.CidrBlock)
	d.Set("vpc_id", vpc.VpcId)

	return diags
}

func resourceVPCIPv4CIDRBlockAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[DEBUG] Deleting EC2 VPC IPv4 CIDR Block Association: %s", d.Id())
	_, err := conn.DisassociateVpcCidrBlockWithContext(ctx, &ec2.DisassociateVpcCidrBlockInput{
		AssociationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCCIDRBlockAssociationIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPC IPv4 CIDR Block Association (%s): %s", d.Id(), err)
	}

	_, err = WaitVPCCIDRBlockAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPC IPv4 CIDR block (%s) to become disassociated: %s", d.Id(), err)
	}

	return diags
}
