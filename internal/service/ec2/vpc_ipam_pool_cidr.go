// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	intretry "github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_ipam_pool_cidr", name="IPAM Pool CIDR")
func resourceIPAMPoolCIDR() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMPoolCIDRCreate,
		ReadWithoutTimeout:   resourceIPAMPoolCIDRRead,
		DeleteWithoutTimeout: resourceIPAMPoolCIDRDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			// Allocations release are eventually consistent with a max time of 20m.
			Delete: schema.DefaultTimeout(32 * time.Minute),
		},

		CustomizeDiff: resourceIPAMPoolCIDRCustomizeDiff,

		Schema: map[string]*schema.Schema{
			"cidr": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.Any(
					verify.ValidIPv4CIDRNetworkAddress,
					verify.ValidIPv6CIDRNetworkAddress,
				),
			},
			"cidr_authorization_context": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMessage: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"signature": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			// This resource's ID is a concatenated id of `<cidr>_<poolid>`
			// ipam_pool_cidr_id was not part of the initial feature release
			"ipam_pool_cidr_id": {
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
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.IntBetween(0, 128),
				ConflictsWith: []string{"cidr"},
			},
		},
	}
}

func resourceIPAMPoolCIDRCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	poolID := d.Get("ipam_pool_id").(string)
	input := &ec2.ProvisionIpamPoolCidrInput{
		IpamPoolId: aws.String(poolID),
	}

	if v, ok := d.GetOk("cidr"); ok {
		input.Cidr = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cidr_authorization_context"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.CidrAuthorizationContext = expandIPAMCIDRAuthorizationContext(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("netmask_length"); ok {
		input.NetmaskLength = aws.Int32(int32(v.(int)))
	}

	output, err := conn.ProvisionIpamPoolCidr(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM Pool (%s) CIDR: %s", poolID, err)
	}

	// its possible that cidr is computed based on netmask_length
	cidrBlock := aws.ToString(output.IpamPoolCidr.Cidr)
	poolCidrID := aws.ToString(output.IpamPoolCidr.IpamPoolCidrId)

	ipamPoolCidr, err := waitIPAMPoolCIDRCreated(ctx, conn, poolCidrID, poolID, cidrBlock, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Pool CIDR (%s) create: %s", poolCidrID, err)
	}

	// This resource's ID is a concatenated id of `<cidr>_<poolid>`
	// ipam_pool_cidr_id was not part of the initial feature release
	d.SetId(ipamPoolCIDRCreateResourceID(aws.ToString(ipamPoolCidr.Cidr), poolID))

	return append(diags, resourceIPAMPoolCIDRRead(ctx, d, meta)...)
}

func resourceIPAMPoolCIDRRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	cidrBlock, poolID, err := ipamPoolCIDRParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findIPAMPoolCIDRByTwoPartKey(ctx, conn, cidrBlock, poolID)

	if !d.IsNewResource() && intretry.NotFound(err) {
		log.Printf("[WARN] IPAM Pool CIDR (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Pool CIDR (%s): %s", d.Id(), err)
	}

	d.Set("cidr", output.Cidr)
	d.Set("ipam_pool_cidr_id", output.IpamPoolCidrId)
	d.Set("ipam_pool_id", poolID)
	d.Set("netmask_length", output.NetmaskLength)

	return diags
}

func resourceIPAMPoolCIDRDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	cidrBlock, poolID, err := ipamPoolCIDRParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// VPC / Subnet allocations take upto 20m to be released after resource deletion.
	log.Printf("[DEBUG] Checking for allocations from CIDR %s in IPAM Pool (%s) that need to be released", cidrBlock, poolID)

	ipamPoolAllocationsInput := &ec2.GetIpamPoolAllocationsInput{
		IpamPoolId: aws.String(poolID),
	}

	allocations, err := findIPAMPoolAllocations(ctx, conn, ipamPoolAllocationsInput)
	if intretry.NotFound(err) {
		log.Printf("[DEBUG] IPAM Pool (%s) not found, skipping allocation checks", poolID)
		allocations = nil
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing IPAM Pool (%s) allocations: %s", poolID, err)
	}

	var allocationsToTrack []awstypes.IpamPoolAllocation
	for _, allocation := range allocations {
		allocationCIDR := aws.ToString(allocation.Cidr)

		if !types.CIDRBlocksOverlap(cidrBlock, allocationCIDR) {
			continue
		}

		resourceType := allocation.ResourceType
		if resourceType == awstypes.IpamPoolAllocationResourceTypeVpc || resourceType == awstypes.IpamPoolAllocationResourceTypeSubnet {
			allocationsToTrack = append(allocationsToTrack, allocation)
		}
	}

	if len(allocationsToTrack) > 0 {
		log.Printf("[DEBUG] Found %d VPC/Subnet allocation(s) from CIDR %s that need to be released", len(allocationsToTrack), cidrBlock)

		for _, allocation := range allocationsToTrack {
			resourceID := aws.ToString(allocation.ResourceId)
			allocationCIDR := aws.ToString(allocation.Cidr)
			resourceType := allocation.ResourceType

			switch resourceType {
			case awstypes.IpamPoolAllocationResourceTypeVpc:
				_, err := findVPCByID(ctx, conn, resourceID)
				if err == nil {
					return sdkdiag.AppendErrorf(diags, "VPC %s (CIDR: %s) must be deleted before IPAM Pool CIDR can be deprovisioned", resourceID, allocationCIDR)
				}
				log.Printf("[DEBUG] VPC %s already deleted, waiting for allocation (CIDR: %s) to be released from IPAM Pool %s", resourceID, allocationCIDR, poolID)

			case awstypes.IpamPoolAllocationResourceTypeSubnet:
				_, err := findSubnetByID(ctx, conn, resourceID)
				if err == nil {
					return sdkdiag.AppendErrorf(diags, "subnet %s (CIDR: %s) must be deleted before IPAM Pool CIDR can be deprovisioned", resourceID, allocationCIDR)
				}
				log.Printf("[DEBUG] Subnet %s already deleted, waiting for allocation (CIDR: %s) to be released from IPAM Pool %s", resourceID, allocationCIDR, poolID)
			}
		}

		log.Printf("[DEBUG] Waiting for IPAM to release %d allocation(s) from CIDR %s", len(allocationsToTrack), cidrBlock)

		_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func(ctx context.Context) (any, error) {
			allocations, err := findIPAMPoolAllocations(ctx, conn, ipamPoolAllocationsInput)
			if intretry.NotFound(err) {
				log.Printf("[DEBUG] IPAM Pool (%s) deleted during wait, allocations released", poolID)
				return nil, &retry.NotFoundError{}
			}
			if err != nil {
				return nil, err
			}

			for _, allocation := range allocations {
				allocationCIDR := aws.ToString(allocation.Cidr)

				if !types.CIDRBlocksOverlap(cidrBlock, allocationCIDR) {
					continue
				}

				resourceType := allocation.ResourceType
				if resourceType == awstypes.IpamPoolAllocationResourceTypeVpc || resourceType == awstypes.IpamPoolAllocationResourceTypeSubnet {
					return allocation, nil
				}
			}

			return nil, &retry.NotFoundError{}
		})

		if tfresource.TimedOut(err) {
			return sdkdiag.AppendErrorf(diags, "timeout waiting for IPAM Pool (%s) allocations to be released after %s", poolID, d.Timeout(schema.TimeoutDelete))
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IPAM Pool (%s) allocations to be released: %s", poolID, err)
		}
	} else {
		log.Printf("[DEBUG] No VPC/Subnet allocations found for CIDR %s, proceeding with deprovision", cidrBlock)
	}

	log.Printf("[DEBUG] Deleting IPAM Pool CIDR: %s", d.Id())
	input := ec2.DeprovisionIpamPoolCidrInput{
		Cidr:       aws.String(cidrBlock),
		IpamPoolId: aws.String(poolID),
	}
	_, err = conn.DeprovisionIpamPoolCidr(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) {
		return diags
	}

	// IncorrectState error can mean: State = "deprovisioned" || State = "pending-deprovision".
	if err != nil && !tfawserr.ErrCodeEquals(err, errCodeIncorrectState) {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM Pool CIDR (%s): %s", d.Id(), err)
	}

	if _, err := waitIPAMPoolCIDRDeleted(ctx, conn, d.Get("ipam_pool_cidr_id").(string), poolID, cidrBlock, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Pool CIDR (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const ipamPoolCIDRIDSeparator = "_"

func ipamPoolCIDRCreateResourceID(cidrBlock, poolID string) string {
	parts := []string{cidrBlock, poolID}
	id := strings.Join(parts, ipamPoolCIDRIDSeparator)

	return id
}

func ipamPoolCIDRParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, ipamPoolCIDRIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected cidr%[2]spool-id", id, ipamPoolCIDRIDSeparator)
	}

	return parts[0], parts[1], nil
}

func expandIPAMCIDRAuthorizationContext(tfMap map[string]any) *awstypes.IpamCidrAuthorizationContext {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.IpamCidrAuthorizationContext{}

	if v, ok := tfMap[names.AttrMessage].(string); ok && v != "" {
		apiObject.Message = aws.String(v)
	}

	if v, ok := tfMap["signature"].(string); ok && v != "" {
		apiObject.Signature = aws.String(v)
	}

	return apiObject
}

func resourceIPAMPoolCIDRCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v any) error {
	// cidr can be set by a value returned from IPAM or explicitly in config.
	if diff.Id() != "" && diff.HasChange("cidr") {
		// If netmask is set then cidr is derived from IPAM, ignore changes.
		if diff.Get("netmask_length") != 0 {
			return diff.Clear("cidr")
		}
		return diff.ForceNew("cidr")
	}

	return nil
}
