// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
			"auto_provision_zones": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_ips": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AvailabilityMode](),
			},
			"availability_zone_address": {
				Type:     schema.TypeSet,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					o, n := d.GetChange("availability_zone_address")
					os, ns := o.(*schema.Set), n.(*schema.Set)
					return EqualityFuncNATGatewayAvailabilityZoneAddressSet(os, ns)
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"availability_zone_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
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
			"regional_nat_gateway_address": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrAssociationID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrNetworkInterfaceID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			// internal attribute to trigger recreation when mode changes between auto and manual
			"regional_nat_gateway_auto_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secondary_allocation_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"secondary_private_ip_address_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
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
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: resourceNATGatewayCustomizeDiff,
	}
}

// Compares old and new sets of availability_zone_address configurations for equality.
// Two sets are considered equal if they contain the same availability_zone or availability_zone_id
// and the same allocation IDs, even when either availability_zone or availability_zone_id
// is absent in the new set.
func EqualityFuncNATGatewayAvailabilityZoneAddressSet(os, ns *schema.Set) bool {
	if os.Len() != ns.Len() {
		return false
	}
	osl, nsl := os.List(), ns.List()
	for _, oss := range osl {
		var oaz, oazid string
		var oo map[string]any
		var oldAllocationIDs *schema.Set
		if v, ok := oss.(map[string]any); ok {
			oo = v
		} else {
			return false
		}
		if v, ok := oo[names.AttrAvailabilityZone]; ok {
			oaz = v.(string)
		}
		if v, ok := oo["availability_zone_id"]; ok {
			oazid = v.(string)
		}
		if v, ok := oo["allocation_ids"]; ok {
			oldAllocationIDs = v.(*schema.Set)
		}
		found := false
		for _, nss := range nsl {
			var naz, nazid string
			var nn map[string]any
			var newAllocationIDs *schema.Set
			if v, ok := nss.(map[string]any); ok {
				nn = v
			} else {
				return false
			}
			if v, ok := nn[names.AttrAvailabilityZone]; ok {
				naz = v.(string)
			}
			if v, ok := nn["availability_zone_id"]; ok {
				nazid = v.(string)
			}
			if oaz != naz && oazid != nazid {
				continue
			}
			found = true
			if v, ok := nn["allocation_ids"]; ok {
				newAllocationIDs = v.(*schema.Set)
			}
			if !newAllocationIDs.Equal(oldAllocationIDs) {
				return false
			}

			if !((oaz == naz && nazid == "") || (oazid == nazid && naz == "")) {
				return false
			}
			break
		}
		if !found {
			return false
		}
	}
	return true
}

func resourceNATGatewayCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateNatGatewayInput{
		ClientToken:       aws.String(sdkid.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeNatgateway),
	}

	if v, ok := d.GetOk("allocation_id"); ok {
		input.AllocationId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("availability_mode"); ok {
		input.AvailabilityMode = awstypes.AvailabilityMode(v.(string))
	}

	if v, ok := d.GetOk("availability_zone_address"); ok {
		input.AvailabilityZoneAddresses = expandNATGatewayAvailabilityZoneAddresses(v.(*schema.Set).List())
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

	if v, ok := d.GetOk(names.AttrVPCID); ok {
		input.VpcId = aws.String(v.(string))
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

func resourceNATGatewayRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	natGateway, err := findNATGatewayByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EC2 NAT Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 NAT Gateway (%s): %s", d.Id(), err)
	}

	d.Set("availability_mode", natGateway.AvailabilityMode)
	d.Set("connectivity_type", natGateway.ConnectivityType)
	d.Set(names.AttrVPCID, natGateway.VpcId)

	switch natGateway.AvailabilityMode {
	case awstypes.AvailabilityModeZonal:
		var secondaryAllocationIDs, secondaryPrivateIPAddresses []string

		for _, natGatewayAddress := range natGateway.NatGatewayAddresses {
			// Length check guarantees the attributes are always set (#30865).
			if isPrimary := aws.ToBool(natGatewayAddress.IsPrimary); isPrimary || len(natGateway.NatGatewayAddresses) == 1 {
				d.Set("allocation_id", natGatewayAddress.AllocationId)
				d.Set(names.AttrAssociationID, natGatewayAddress.AssociationId)
				d.Set(names.AttrNetworkInterfaceID, natGatewayAddress.NetworkInterfaceId)
				d.Set("private_ip", natGatewayAddress.PrivateIp)
				d.Set("public_ip", natGatewayAddress.PublicIp)
			} else if !isPrimary {
				if allocationID := aws.ToString(natGatewayAddress.AllocationId); allocationID != "" {
					secondaryAllocationIDs = append(secondaryAllocationIDs, allocationID)
				}
				if privateIP := aws.ToString(natGatewayAddress.PrivateIp); privateIP != "" {
					secondaryPrivateIPAddresses = append(secondaryPrivateIPAddresses, privateIP)
				}
			}
		}
		d.Set("secondary_allocation_ids", secondaryAllocationIDs)
		d.Set("secondary_private_ip_address_count", len(secondaryPrivateIPAddresses))
		d.Set("secondary_private_ip_addresses", secondaryPrivateIPAddresses)
		d.Set(names.AttrSubnetID, natGateway.SubnetId)

	case awstypes.AvailabilityModeRegional:
		d.Set("auto_provision_zones", natGateway.AutoProvisionZones)
		d.Set("auto_scaling_ips", natGateway.AutoScalingIps)
		if natGateway.AutoProvisionZones == awstypes.AutoProvisionZonesStateEnabled {
			d.Set("availability_zone_address", nil)
		} else if err := d.Set("availability_zone_address", flattenNATGatewayAvailabilityZoneAddresses(natGateway.NatGatewayAddresses)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting availability_zone_address: %s", err)
		}

		if err := d.Set("regional_nat_gateway_address", flattenRegionalNATGatewayAddress(natGateway.NatGatewayAddresses)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting regional_nat_gateway_address: %s", err)
		}
		d.Set("regional_nat_gateway_auto_mode", natGateway.AutoProvisionZones)
		d.Set("route_table_id", natGateway.RouteTableId)
	}

	setTagsOut(ctx, natGateway.Tags)

	return diags
}

func resourceNATGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	var availabilityMode awstypes.AvailabilityMode
	if v, ok := d.Get("availability_mode").(string); ok {
		availabilityMode = awstypes.AvailabilityMode(v)
	} else {
		availabilityMode = awstypes.AvailabilityModeZonal
	}

	switch availabilityMode {
	case awstypes.AvailabilityModeZonal:
		switch awstypes.ConnectivityType(d.Get("connectivity_type").(string)) {
		case awstypes.ConnectivityTypePrivate:
			if d.HasChanges("secondary_private_ip_addresses") {
				o, n := d.GetChange("secondary_private_ip_addresses")
				os, ns := o.(*schema.Set), n.(*schema.Set)

				if add := ns.Difference(os); add.Len() > 0 {
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

				if del := os.Difference(ns); del.Len() > 0 {
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
		case awstypes.ConnectivityTypePublic:
			if !d.GetRawConfig().GetAttr("secondary_allocation_ids").IsNull() && d.HasChanges("secondary_allocation_ids") {
				o, n := d.GetChange("secondary_allocation_ids")
				os, ns := o.(*schema.Set), n.(*schema.Set)

				if add := ns.Difference(os); add.Len() > 0 {
					input := &ec2.AssociateNatGatewayAddressInput{
						AllocationIds: flex.ExpandStringValueSet(add),
						NatGatewayId:  aws.String(d.Id()),
					}

					if d.HasChanges("secondary_private_ip_addresses") {
						o, n := d.GetChange("secondary_private_ip_addresses")
						os, ns := o.(*schema.Set), n.(*schema.Set)

						if add := ns.Difference(os); add.Len() > 0 {
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

				if del := os.Difference(ns); del.Len() > 0 {
					natGateway, err := findNATGatewayByID(ctx, conn, d.Id())

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "reading EC2 NAT Gateway (%s): %s", d.Id(), err)
					}

					allocationIDs := flex.ExpandStringValueSet(del)
					var associationIDs []string

					for _, natGatewayAddress := range natGateway.NatGatewayAddresses {
						if allocationID := aws.ToString(natGatewayAddress.AllocationId); slices.Contains(allocationIDs, allocationID) {
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
	case awstypes.AvailabilityModeRegional:
		if d.HasChanges("availability_zone_address") {
			o, n := d.GetChange("availability_zone_address")
			os, ns := o.(*schema.Set), n.(*schema.Set)

			// Pre-fetch AZ ID→Name map once if needed to avoid redundant API calls
			var azIDtoNameMap map[string]string
			if needsAZIDtoNameMap(os) || needsAZIDtoNameMap(ns) {
				var err error
				azIDtoNameMap, err = makeAZIDtoNameMap(ctx, conn)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "retrieving availability zone ID to name map: %s", err)
				}
			}

			oldMap, err := processAZAddressSet(os, azIDtoNameMap)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "processing old availability zone address set: %s", err)
			}
			newMap, err := processAZAddressSet(ns, azIDtoNameMap)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "processing new availability zone address set: %s", err)
			}

			// Collect all unique AZ keys
			allKeys := make(map[string]bool)
			for az := range oldMap {
				allKeys[az] = true
			}
			for az := range newMap {
				allKeys[az] = true
			}

			var removedAllAZ []string
			for az := range allKeys {
				var oldSet, newSet *schema.Set
				if v, ok := oldMap[az]; ok {
					oldSet = v
				} else {
					oldSet = schema.NewSet(schema.HashString, []any{})
				}
				if v, ok := newMap[az]; ok {
					newSet = v
				} else {
					newSet = schema.NewSet(schema.HashString, []any{})
				}

				added := newSet.Difference(oldSet)
				removed := oldSet.Difference(newSet)

				if added.Len() > 0 {
					if err := associateRegionalNATGatewayAddresses(ctx, conn, d, az, added); err != nil {
						return append(diags, err...)
					}
				}

				if removed.Len() > 0 {
					removedAllAZ = append(removedAllAZ, flex.ExpandStringValueSet(removed)...)
				}
			}
			if len(removedAllAZ) > 0 {
				if err := disassociateRegionalNATGatewayAddresses(ctx, conn, d, removedAllAZ); err != nil {
					return append(diags, err...)
				}
			}
		}
	}

	return append(diags, resourceNATGatewayRead(ctx, d, meta)...)
}

func resourceNATGatewayDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// Eventual consistency wait for any attached appliances to be detached before deleting the NAT Gateway
	if _, err := waitNATGatewayAttachedAppliancesDetached(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 NAT Gateway (%s) attached appliances to detach: %s", d.Id(), err)
	}

	log.Printf("[INFO] Deleting EC2 NAT Gateway: %s", d.Id())
	input := ec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(d.Id()),
	}
	_, err := conn.DeleteNatGateway(ctx, &input)

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

func resourceNATGatewayCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta any) error {
	switch connectivityType := awstypes.ConnectivityType(diff.Get("connectivity_type").(string)); connectivityType {
	case awstypes.ConnectivityTypePrivate:
		if _, ok := diff.GetOk("allocation_id"); ok {
			return fmt.Errorf(`allocation_id is not supported with connectivity_type = "%s"`, connectivityType)
		}

		if v, ok := diff.GetOk("secondary_allocation_ids"); ok && v.(*schema.Set).Len() > 0 {
			return fmt.Errorf(`secondary_allocation_ids is not supported with connectivity_type = "%s"`, connectivityType)
		}

		if diff.Id() != "" && diff.HasChange("secondary_private_ip_address_count") {
			if v := diff.GetRawConfig().GetAttr("secondary_private_ip_address_count"); v.IsKnown() && !v.IsNull() {
				if err := diff.ForceNew("secondary_private_ip_address_count"); err != nil {
					return fmt.Errorf("setting secondary_private_ip_address_count to ForceNew: %w", err)
				}
			}
		}

		if diff.Id() != "" && diff.HasChange("secondary_private_ip_addresses") {
			if err := diff.SetNewComputed("secondary_private_ip_address_count"); err != nil {
				return fmt.Errorf("setting secondary_private_ip_address_count to Computed: %w", err)
			}
		}
	case awstypes.ConnectivityTypePublic:
		if v := diff.GetRawConfig().GetAttr("secondary_private_ip_address_count"); v.IsKnown() && !v.IsNull() {
			return fmt.Errorf(`secondary_private_ip_address_count is not supported with connectivity_type = "%s"`, connectivityType)
		}

		if diff.Id() != "" {
			if v := diff.GetRawConfig().GetAttr("secondary_allocation_ids"); diff.HasChange("secondary_allocation_ids") || !v.IsWhollyKnown() {
				if err := diff.SetNewComputed("secondary_private_ip_address_count"); err != nil {
					return fmt.Errorf("setting secondary_private_ip_address_count to Computed: %w", err)
				}

				if v := diff.GetRawConfig().GetAttr("secondary_private_ip_addresses"); !v.IsKnown() || v.IsNull() {
					if err := diff.SetNewComputed("secondary_private_ip_addresses"); err != nil {
						return fmt.Errorf("setting secondary_private_ip_addresses to Computed: %w", err)
					}
				}
			}
		}
	}

	switch availabilityMode := awstypes.AvailabilityMode(diff.Get("availability_mode").(string)); availabilityMode {
	case awstypes.AvailabilityModeRegional:
		if err := validateRegionalNATGatewayAvailabilityZoneAddress(diff); err != nil {
			return err
		}
	}

	return nil
}

func validateRegionalNATGatewayAvailabilityZoneAddress(diff *schema.ResourceDiff) error {
	if diff.Id() != "" && diff.HasChange("availability_zone_address") {
		o, n := diff.GetChange("availability_zone_address")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		if (os.Len() > 0 && ns.Len() == 0) || (ns.Len() > 0 && os.Len() == 0) {
			if err := diff.SetNewComputed("regional_nat_gateway_auto_mode"); err != nil {
				return fmt.Errorf("setting regional_nat_gateway_auto_mode to Computed: %w", err)
			}
			if err := diff.ForceNew("regional_nat_gateway_auto_mode"); err != nil {
				return fmt.Errorf("setting regional_nat_gateway_auto_mode to ForceNew: %w", err)
			}
		}
		// regional_nat_gateway_address should recompute when AZ addresses actually change.
		if !EqualityFuncNATGatewayAvailabilityZoneAddressSet(os, ns) {
			if err := diff.SetNewComputed("regional_nat_gateway_address"); err != nil {
				return fmt.Errorf("setting regional_nat_gateway_address to NewComputed: %w", err)
			}
		}
	}
	n := diff.Get("availability_zone_address").(*schema.Set)
	for _, v := range n.List() {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		var az, azid string
		if v, ok := m[names.AttrAvailabilityZone]; ok {
			az = v.(string)
		}
		if v, ok := m["availability_zone_id"]; ok {
			azid = v.(string)
		}
		// Allow both if they're both set (AWS returns both), but require at least one
		if az == "" && azid == "" {
			return fmt.Errorf("At least one of availability_zone or availability_zone_id must be specified")
		}
	}
	return nil
}

func expandNATGatewayAvailabilityZoneAddresses(vs []any) []awstypes.AvailabilityZoneAddress {
	if len(vs) == 0 {
		return nil
	}

	var addresses []awstypes.AvailabilityZoneAddress

	for _, v := range vs {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		address := awstypes.AvailabilityZoneAddress{}

		if v, ok := m["allocation_ids"]; ok {
			if s := v.(*schema.Set); s.Len() > 0 {
				address.AllocationIds = flex.ExpandStringValueSet(s)
			}
		}

		if v, ok := m[names.AttrAvailabilityZone]; ok && v.(string) != "" {
			address.AvailabilityZone = aws.String(v.(string))
		} else if v, ok := m["availability_zone_id"]; ok && v.(string) != "" {
			address.AvailabilityZoneId = aws.String(v.(string))
		}

		addresses = append(addresses, address)
	}

	return addresses
}

func flattenNATGatewayAvailabilityZoneAddresses(addresses []awstypes.NatGatewayAddress) []map[string]any {
	var result []map[string]any

	type azAddress struct {
		allocationIDs []string
	}
	mmap := make(map[string]azAddress)

	for _, addr := range addresses {
		if addr.Status != awstypes.NatGatewayAddressStatusSucceeded {
			continue
		}

		azKey := aws.ToString(addr.AvailabilityZone) + ":" + aws.ToString(addr.AvailabilityZoneId)

		azAddr := mmap[azKey]
		azAddr.allocationIDs = append(azAddr.allocationIDs, aws.ToString(addr.AllocationId))
		mmap[azKey] = azAddr
	}

	// Sort by AZ name for deterministic ordering
	var azKeys []string
	for azKey := range mmap {
		azKeys = append(azKeys, azKey)
	}
	slices.Sort(azKeys)

	for _, azKey := range azKeys {
		azAddr := mmap[azKey]
		m := make(map[string]any)
		parts := strings.Split(azKey, ":")
		m[names.AttrAvailabilityZone], m["availability_zone_id"] = parts[0], parts[1]
		m["allocation_ids"] = flex.FlattenStringValueSet(azAddr.allocationIDs)
		result = append(result, m)
	}

	return result
}

func flattenRegionalNATGatewayAddress(addresses []awstypes.NatGatewayAddress) []map[string]any {
	var result []map[string]any

	for _, addr := range addresses {
		m := make(map[string]any)
		m["allocation_id"] = aws.ToString(addr.AllocationId)
		m[names.AttrAssociationID] = aws.ToString(addr.AssociationId)
		m[names.AttrAvailabilityZone] = aws.ToString(addr.AvailabilityZone)
		m["availability_zone_id"] = aws.ToString(addr.AvailabilityZoneId)
		m[names.AttrNetworkInterfaceID] = aws.ToString(addr.NetworkInterfaceId)
		m["public_ip"] = aws.ToString(addr.PublicIp)
		m[names.AttrStatus] = addr.Status
		result = append(result, m)
	}

	return result
}

func makeAZIDtoNameMap(ctx context.Context, conn *ec2.Client) (map[string]string, error) {
	input := &ec2.DescribeAvailabilityZonesInput{
		Filters: newAttributeFilterList(map[string]string{
			"zone-type": "availability-zone",
		}),
	}

	output, err := conn.DescribeAvailabilityZones(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("describing availability zones: %w", err)
	}

	azIDtoNameMap := make(map[string]string)
	for _, az := range output.AvailabilityZones {
		azIDtoNameMap[aws.ToString(az.ZoneId)] = aws.ToString(az.ZoneName)
	}

	return azIDtoNameMap, nil
}

func needsAZIDtoNameMap(s *schema.Set) bool {
	for _, addr := range s.List() {
		item, ok := addr.(map[string]any)
		if !ok {
			continue
		}
		var az, azID string
		if v, ok := item[names.AttrAvailabilityZone]; ok {
			az = v.(string)
		}
		if v, ok := item["availability_zone_id"]; ok {
			azID = v.(string)
		}
		// If AZ ID is specified but AZ name is not, we need the map
		if az == "" && azID != "" {
			return true
		}
	}
	return false
}

func processAZAddressSet(s *schema.Set, azIDtoNameMap map[string]string) (map[string]*schema.Set, error) {
	sl := s.List()
	result := make(map[string]*schema.Set)
	for _, addr := range sl {
		var az string
		var azID string
		item, ok := addr.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected map[string]any in availability_zone_address set, got: %T", addr)
		}
		if v, ok := item[names.AttrAvailabilityZone]; ok {
			az = v.(string)
		}
		if v, ok := item["availability_zone_id"]; ok {
			azID = v.(string)
		}

		if az == "" && azID != "" {
			if azIDtoNameMap == nil {
				return nil, fmt.Errorf("availability zone ID to name map required but not provided")
			}
			var exists bool
			az, exists = azIDtoNameMap[azID]
			if !exists {
				return nil, fmt.Errorf("availability zone ID %q not found", azID)
			}
		}

		if az == "" {
			return nil, fmt.Errorf("either availability_zone or availability_zone_id must be specified")
		}

		if v, ok := item["allocation_ids"]; ok {
			result[az] = v.(*schema.Set)
		}
	}
	return result, nil
}

// Associates allocation IDs to a regional NAT Gateway for a specific availability zone
func associateRegionalNATGatewayAddresses(ctx context.Context, conn *ec2.Client, d *schema.ResourceData, az string, allocationIDs *schema.Set) diag.Diagnostics {
	var diags diag.Diagnostics

	input := &ec2.AssociateNatGatewayAddressInput{
		AllocationIds: flex.ExpandStringValueSet(allocationIDs),
		NatGatewayId:  aws.String(d.Id()),
	}

	if az != "" {
		input.AvailabilityZone = aws.String(az)
	}
	if _, err := conn.AssociateNatGatewayAddress(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "associating EC2 NAT Gateway (%s) allocation IDs for AZ %s: %s", d.Id(), az, err)
	}

	for _, allocationID := range flex.ExpandStringValueSet(allocationIDs) {
		if _, err := waitNATGatewayAddressAssociated(ctx, conn, d.Id(), allocationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 NAT Gateway (%s) allocation ID (%s) associate: %s", d.Id(), allocationID, err)
		}
	}

	return diags
}

// Disassociates allocation IDs from a regional NAT Gateway for a specific availability zone
func disassociateRegionalNATGatewayAddresses(ctx context.Context, conn *ec2.Client, d *schema.ResourceData, allocationIDs []string) diag.Diagnostics {
	var diags diag.Diagnostics

	natGateway, err := findNATGatewayByID(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 NAT Gateway (%s): %s", d.Id(), err)
	}

	var associationIDs []string

	for _, natGatewayAddress := range natGateway.NatGatewayAddresses {
		if allocationID := aws.ToString(natGatewayAddress.AllocationId); slices.Contains(allocationIDs, allocationID) {
			associationIDs = append(associationIDs, aws.ToString(natGatewayAddress.AssociationId))
		}
	}

	if len(associationIDs) == 0 {
		return diags
	}

	input := &ec2.DisassociateNatGatewayAddressInput{
		AssociationIds: associationIDs,
		NatGatewayId:   aws.String(d.Id()),
	}

	if _, err := conn.DisassociateNatGatewayAddress(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating EC2 NAT Gateway (%s) allocation IDs: %s", d.Id(), err)
	}

	for _, allocationID := range allocationIDs {
		if _, err := waitNATGatewayAddressDisassociated(ctx, conn, d.Id(), allocationID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 NAT Gateway (%s) allocation ID (%s) disassociate: %s", d.Id(), allocationID, err)
		}
	}

	return diags
}
