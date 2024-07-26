// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_ipam_pool_cidr_allocation", name="IPAM Pool CIDR Allocation")
func resourceIPAMPoolCIDRAllocation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMPoolCIDRAllocationCreate,
		ReadWithoutTimeout:   resourceIPAMPoolCIDRAllocationRead,
		DeleteWithoutTimeout: resourceIPAMPoolCIDRAllocationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cidr": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ConflictsWith: []string{"netmask_length"},
				ValidateFunc: validation.Any(
					verify.ValidIPv4CIDRNetworkAddress,
					verify.ValidIPv6CIDRNetworkAddress,
				),
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"disallowed_cidrs": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.Any(
						verify.ValidIPv4CIDRNetworkAddress,
						// Follow the numbers used for netmask_length
						validation.IsCIDRNetwork(0, 128),
					),
				},
			},
			"ipam_pool_allocation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"netmask_length": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IntBetween(0, 128),
				ConflictsWith: []string{"cidr"},
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceIPAMPoolCIDRAllocationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	ipamPoolID := d.Get("ipam_pool_id").(string)
	input := &ec2.AllocateIpamPoolCidrInput{
		ClientToken: aws.String(id.UniqueId()),
		IpamPoolId:  aws.String(ipamPoolID),
	}

	if v, ok := d.GetOk("cidr"); ok {
		input.Cidr = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disallowed_cidrs"); ok && v.(*schema.Set).Len() > 0 {
		input.DisallowedCidrs = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v := d.Get("netmask_length"); v != 0 {
		input.NetmaskLength = aws.Int32(int32(v.(int)))
	}

	output, err := conn.AllocateIpamPoolCidr(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM Pool CIDR Allocation: %s", err)
	}

	allocationID := aws.ToString(output.IpamPoolAllocation.IpamPoolAllocationId)
	d.SetId(ipamPoolCIDRAllocationCreateResourceID(allocationID, ipamPoolID))

	_, err = tfresource.RetryWhenNotFound(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return findIPAMPoolAllocationByTwoPartKey(ctx, conn, allocationID, ipamPoolID)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Pool CIDR Allocation (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceIPAMPoolCIDRAllocationRead(ctx, d, meta)...)
}

func resourceIPAMPoolCIDRAllocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	allocationID, poolID, err := ipamPoolCIDRAllocationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	allocation, err := findIPAMPoolAllocationByTwoPartKey(ctx, conn, allocationID, poolID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Pool CIDR Allocation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Pool CIDR Allocation (%s): %s", d.Id(), err)
	}

	d.Set("cidr", allocation.Cidr)
	d.Set("ipam_pool_allocation_id", allocation.IpamPoolAllocationId)
	d.Set("ipam_pool_id", poolID)
	d.Set(names.AttrResourceID, allocation.ResourceId)
	d.Set(names.AttrResourceOwner, allocation.ResourceOwner)
	d.Set(names.AttrResourceType, allocation.ResourceType)

	return diags
}

func resourceIPAMPoolCIDRAllocationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	allocationID, poolID, err := ipamPoolCIDRAllocationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting IPAM Pool CIDR Allocation: %s", d.Id())
	_, err = conn.ReleaseIpamPoolAllocation(ctx, &ec2.ReleaseIpamPoolAllocationInput{
		Cidr:                 aws.String(d.Get("cidr").(string)),
		IpamPoolAllocationId: aws.String(allocationID),
		IpamPoolId:           aws.String(poolID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) || tfawserr.ErrMessageContains(err, errCodeInvalidParameterCombination, "No allocation found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM Pool CIDR Allocation (%s): %s", d.Id(), err)
	}

	return diags
}

const ipamPoolCIDRAllocationIDSeparator = "_"

func ipamPoolCIDRAllocationCreateResourceID(allocationID, poolID string) string {
	parts := []string{allocationID, poolID}
	id := strings.Join(parts, ipamPoolCIDRAllocationIDSeparator)

	return id
}

func ipamPoolCIDRAllocationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, ipamPoolCIDRAllocationIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected allocation-id%[2]spool-id", id, ipamPoolCIDRAllocationIDSeparator)
	}

	return parts[0], parts[1], nil
}
