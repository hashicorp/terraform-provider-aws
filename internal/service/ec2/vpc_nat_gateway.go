// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_nat_gateway", name="NAT Gateway")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceNATGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNATGatewayCreate,
		ReadWithoutTimeout:   resourceNATGatewayRead,
		UpdateWithoutTimeout: resourceNATGatewayUpdate,
		DeleteWithoutTimeout: resourceNATGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"allocation_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrAssociationID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connectivity_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.ConnectivityTypePublic,
				ValidateDiagFunc: enum.Validate[awstypes.ConnectivityType](),
			},
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secondary_allocation_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"secondary_private_ip_address_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"secondary_private_ip_addresses"},
			},
			"secondary_private_ip_addresses": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"secondary_private_ip_address_count"},
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.All(
			resourceNATGatewayCustomizeDiff,
			verify.SetTagsDiff,
		),
	}
}

func resourceNATGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateNatGatewayInput{
		ClientToken:       aws.String(id.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeNatgateway),
	}

	if v, ok := d.GetOk("allocation_id"); ok {
		input.AllocationId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("connectivity_type"); ok {
		input.ConnectivityType = awstypes.ConnectivityType(v.(string))
	}

	if v, ok := d.GetOk("private_ip"); ok {
		input.PrivateIpAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("secondary_allocation_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SecondaryAllocationIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("secondary_private_ip_address_count"); ok {
		input.SecondaryPrivateIpAddressCount = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("secondary_private_ip_addresses"); ok && v.(*schema.Set).Len() > 0 {
		input.SecondaryPrivateIpAddresses = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrSubnetID); ok {
		input.SubnetId = aws.String(v.(string))
	}

	output, err := conn.CreateNatGateway(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 NAT Gateway: %s", err)
	}

	d.SetId(aws.ToString(output.NatGateway.NatGatewayId))

	if _, err := waitNATGatewayCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 NAT Gateway (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceNATGatewayRead(ctx, d, meta)...)
}

func resourceNATGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	ng, err := findNATGatewayByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 NAT Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 NAT Gateway (%s): %s", d.Id(), err)
	}

	var secondaryAllocationIDs, secondaryPrivateIPAddresses []string

	for _, address := range ng.NatGatewayAddresses {
		// Length check guarantees the attributes are always set (#30865).
		if isPrimary := aws.ToBool(address.IsPrimary); isPrimary || len(ng.NatGatewayAddresses) == 1 {
			d.Set("allocation_id", address.AllocationId)
			d.Set(names.AttrAssociationID, address.AssociationId)
			d.Set(names.AttrNetworkInterfaceID, address.NetworkInterfaceId)
			d.Set("private_ip", address.PrivateIp)
			d.Set("public_ip", address.PublicIp)
		} else if !isPrimary {
			if allocationID := aws.ToString(address.AllocationId); allocationID != "" {
				secondaryAllocationIDs = append(secondaryAllocationIDs, allocationID)
			}
			if privateIP := aws.ToString(address.PrivateIp); privateIP != "" {
				secondaryPrivateIPAddresses = append(secondaryPrivateIPAddresses, privateIP)
			}
		}
	}

	d.Set("connectivity_type", ng.ConnectivityType)
	d.Set("secondary_allocation_ids", secondaryAllocationIDs)
	d.Set("secondary_private_ip_address_count", len(secondaryPrivateIPAddresses))
	d.Set("secondary_private_ip_addresses", secondaryPrivateIPAddresses)
	d.Set(names.AttrSubnetID, ng.SubnetId)

	setTagsOut(ctx, ng.Tags)

	return diags
}

func resourceNATGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	switch d.Get("connectivity_type").(string) {
	case string(awstypes.ConnectivityTypePrivate):
		if d.HasChanges("secondary_private_ip_addresses") {
			oRaw, nRaw := d.GetChange("secondary_private_ip_addresses")
			o, n := oRaw.(*schema.Set), nRaw.(*schema.Set)

			if add := n.Difference(o); add.Len() > 0 {
				input := &ec2.AssignPrivateNatGatewayAddressInput{
					NatGatewayId:       aws.String(d.Id()),
					PrivateIpAddresses: flex.ExpandStringValueSet(add),
				}

				_, err := conn.AssignPrivateNatGatewayAddress(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "assigning EC2 NAT Gateway (%s) private IP addresses: %s", d.Id(), err)
				}

				for _, privateIP := range flex.ExpandStringValueSet(add) {
					if _, err := waitNATGatewayAddressAssigned(ctx, conn, d.Id(), privateIP, d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for EC2 NAT Gateway (%s) private IP address (%s) assign: %s", d.Id(), privateIP, err)
					}
				}
			}

			if del := o.Difference(n); del.Len() > 0 {
				input := &ec2.UnassignPrivateNatGatewayAddressInput{
					NatGatewayId:       aws.String(d.Id()),
					PrivateIpAddresses: flex.ExpandStringValueSet(del),
				}

				_, err := conn.UnassignPrivateNatGatewayAddress(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "unassigning EC2 NAT Gateway (%s) private IP addresses: %s", d.Id(), err)
				}

				for _, privateIP := range flex.ExpandStringValueSet(del) {
					if _, err := waitNATGatewayAddressUnassigned(ctx, conn, d.Id(), privateIP, d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for EC2 NAT Gateway (%s) private IP address (%s) unassign: %s", d.Id(), privateIP, err)
					}
				}
			}
		}

	case string(awstypes.ConnectivityTypePublic):
		if d.HasChanges("secondary_allocation_ids") {
			oRaw, nRaw := d.GetChange("secondary_allocation_ids")
			o, n := oRaw.(*schema.Set), nRaw.(*schema.Set)

			if add := n.Difference(o); add.Len() > 0 {
				input := &ec2.AssociateNatGatewayAddressInput{
					AllocationIds: flex.ExpandStringValueSet(add),
					NatGatewayId:  aws.String(d.Id()),
				}

				if d.HasChanges("secondary_private_ip_addresses") {
					oRaw, nRaw := d.GetChange("secondary_private_ip_addresses")
					o, n := oRaw.(*schema.Set), nRaw.(*schema.Set)

					if add := n.Difference(o); add.Len() > 0 {
						input.PrivateIpAddresses = flex.ExpandStringValueSet(add)
					}
				}

				_, err := conn.AssociateNatGatewayAddress(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "associating EC2 NAT Gateway (%s) allocation IDs: %s", d.Id(), err)
				}

				for _, allocationID := range flex.ExpandStringValueSet(add) {
					if _, err := waitNATGatewayAddressAssociated(ctx, conn, d.Id(), allocationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for EC2 NAT Gateway (%s) allocation ID (%s) associate: %s", d.Id(), allocationID, err)
					}
				}
			}

			if del := o.Difference(n); del.Len() > 0 {
				natGateway, err := findNATGatewayByID(ctx, conn, d.Id())

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading EC2 NAT Gateway (%s): %s", d.Id(), err)
				}

				allocationIDs := flex.ExpandStringValueSet(del)
				var associationIDs []string

				for _, natGatewayAddress := range natGateway.NatGatewayAddresses {
					allocationID := aws.ToString(natGatewayAddress.AllocationId)
					if slices.Contains(allocationIDs, allocationID) {
						associationIDs = append(associationIDs, aws.ToString(natGatewayAddress.AssociationId))
					}
				}

				input := &ec2.DisassociateNatGatewayAddressInput{
					AssociationIds: associationIDs,
					NatGatewayId:   aws.String(d.Id()),
				}

				_, err = conn.DisassociateNatGatewayAddress(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "disassociating EC2 NAT Gateway (%s) allocation IDs: %s", d.Id(), err)
				}

				for _, allocationID := range allocationIDs {
					if _, err := waitNATGatewayAddressDisassociated(ctx, conn, d.Id(), allocationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for EC2 NAT Gateway (%s) allocation ID (%s) disassociate: %s", d.Id(), allocationID, err)
					}
				}
			}
		}
	}

	return append(diags, resourceNATGatewayRead(ctx, d, meta)...)
}

func resourceNATGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 NAT Gateway: %s", d.Id())
	_, err := conn.DeleteNatGateway(ctx, &ec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNatGatewayNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 NAT Gateway (%s): %s", d.Id(), err)
	}

	if _, err := waitNATGatewayDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 NAT Gateway (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceNATGatewayCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	switch connectivityType := diff.Get("connectivity_type").(string); connectivityType {
	case string(awstypes.ConnectivityTypePrivate):
		if _, ok := diff.GetOk("allocation_id"); ok {
			return fmt.Errorf(`allocation_id is not supported with connectivity_type = "%s"`, connectivityType)
		}
		if v, ok := diff.GetOk("secondary_allocation_ids"); ok && v.(*schema.Set).Len() > 0 {
			return fmt.Errorf(`secondary_allocation_ids is not supported with connectivity_type = "%s"`, connectivityType)
		}

	case string(awstypes.ConnectivityTypePublic):
		if v := diff.GetRawConfig().GetAttr("secondary_private_ip_address_count"); v.IsKnown() && !v.IsNull() {
			return fmt.Errorf(`secondary_private_ip_address_count is not supported with connectivity_type = "%s"`, connectivityType)
		}

		if diff.Id() != "" && diff.HasChange("secondary_allocation_ids") {
			if err := diff.SetNewComputed("secondary_private_ip_address_count"); err != nil {
				return fmt.Errorf("setting secondary_private_ip_address_count to computed: %s", err)
			}

			if v := diff.GetRawConfig().GetAttr("secondary_private_ip_addresses"); !v.IsKnown() {
				if err := diff.SetNewComputed("secondary_private_ip_addresses"); err != nil {
					return fmt.Errorf("setting secondary_private_ip_addresses to computed: %s", err)
				}
			}
		}
	}

	return nil
}
