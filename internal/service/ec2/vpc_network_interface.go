// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

// @SDKResource("aws_network_interface", name="Network Interface")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceNetworkInterface() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkInterfaceCreate,
		ReadWithoutTimeout:   resourceNetworkInterfaceRead,
		UpdateWithoutTimeout: resourceNetworkInterfaceUpdate,
		DeleteWithoutTimeout: resourceNetworkInterfaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachment": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attachment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"instance": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"interface_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.NetworkInterfaceCreationType](),
			},
			"ipv4_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
				},
				ConflictsWith: []string{"ipv4_prefix_count"},
			},
			"ipv4_prefix_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ipv4_prefixes"},
			},
			"ipv6_address_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ipv6_addresses", "ipv6_address_list"},
			},
			"ipv6_address_list": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"ipv6_addresses", "ipv6_address_count"},
			},
			"ipv6_address_list_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ipv6_addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv6Address,
				},
				ConflictsWith: []string{"ipv6_address_count", "ipv6_address_list"},
			},
			"ipv6_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidIPv6CIDRNetworkAddress,
				},
				ConflictsWith: []string{"ipv6_prefix_count"},
			},
			"ipv6_prefix_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"ipv6_prefixes"},
			},
			"mac_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"private_ips": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"private_ip_list"},
			},
			"private_ips_count": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"private_ip_list"},
			},
			"private_ip_list": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"private_ips", "private_ips_count"},
			},
			"private_ip_list_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"source_dest_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
			customdiff.ForceNewIf("private_ips", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				privateIPListEnabled := d.Get("private_ip_list_enabled").(bool)
				if privateIPListEnabled {
					return false
				}
				_, new := d.GetChange("private_ips")
				if new != nil {
					oldPrimaryIP := ""
					if v, ok := d.GetOk("private_ip_list"); ok {
						for _, ip := range v.([]interface{}) {
							oldPrimaryIP = ip.(string)
							break
						}
					}
					for _, ip := range new.(*schema.Set).List() {
						// no need for new resource if we still have the primary ip
						if oldPrimaryIP == ip.(string) {
							return false
						}
					}
					// new primary ip requires a new resource
					return true
				} else {
					return false
				}
			}),
			customdiff.ForceNewIf("private_ip_list", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				privateIPListEnabled := d.Get("private_ip_list_enabled").(bool)
				if !privateIPListEnabled {
					return false
				}
				old, new := d.GetChange("private_ip_list")
				if old != nil && new != nil {
					oldPrimaryIP := ""
					newPrimaryIP := ""
					for _, ip := range old.([]interface{}) {
						oldPrimaryIP = ip.(string)
						break
					}
					for _, ip := range new.([]interface{}) {
						newPrimaryIP = ip.(string)
						break
					}

					// change in primary private ip requires a new resource
					return oldPrimaryIP != newPrimaryIP
				} else {
					return false
				}
			}),
			customdiff.ComputedIf("private_ips", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("private_ip_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("private_ips") {
						return false
					} else {
						return diff.HasChange("private_ips_count")
					}
				} else {
					return diff.HasChange("private_ip_list")
				}
			}),
			customdiff.ComputedIf("private_ips_count", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("private_ip_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("private_ips_count") {
						return false
					} else {
						// compute the new count if private_ips change
						return diff.HasChange("private_ips")
					}
				} else {
					// compute the new count if private_ip_list changes
					return diff.HasChange("private_ip_list")
				}
			}),
			customdiff.ComputedIf("private_ip_list", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if diff.Get("private_ip_list_enabled").(bool) {
					// if the list is controlling it does not need to be computed
					return false
				} else {
					// list is not controlling so compute new list if private_ips or private_ips_count changes
					return diff.HasChange("private_ips") || diff.HasChange("private_ips_count") || diff.HasChange("private_ip_list")
				}
			}),
			customdiff.ComputedIf("ipv6_addresses", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("ipv6_address_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("private_ips") {
						return false
					} else {
						return diff.HasChange("ipv6_address_count")
					}
				} else {
					return diff.HasChange("ipv6_address_list")
				}
			}),
			customdiff.ComputedIf("ipv6_address_count", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if !diff.Get("ipv6_address_list_enabled").(bool) {
					// it is not computed if we are actively updating it
					if diff.HasChange("ipv6_address_count") {
						return false
					} else {
						// compute the new count if ipv6_addresses change
						return diff.HasChange("ipv6_addresses")
					}
				} else {
					// compute the new count if ipv6_address_list changes
					return diff.HasChange("ipv6_address_list")
				}
			}),
			customdiff.ComputedIf("ipv6_address_list", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				if diff.Get("ipv6_address_list_enabled").(bool) {
					// if the list is controlling it does not need to be computed
					return false
				} else {
					// list is not controlling so compute new list if anything changes
					return diff.HasChange("ipv6_addresses") || diff.HasChange("ipv6_address_count") || diff.HasChange("ipv6_address_list")
				}
			}),
		),
	}
}

func resourceNetworkInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	ipv4PrefixesSpecified := false
	ipv6PrefixesSpecified := false

	input := &ec2.CreateNetworkInterfaceInput{
		ClientToken: aws.String(id.UniqueId()),
		SubnetId:    aws.String(d.Get(names.AttrSubnetID).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("interface_type"); ok {
		input.InterfaceType = types.NetworkInterfaceCreationType(v.(string))
	}

	if v, ok := d.GetOk("ipv4_prefixes"); ok && v.(*schema.Set).Len() > 0 {
		ipv4PrefixesSpecified = true
		input.Ipv4Prefixes = expandIPv4PrefixSpecificationRequests(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ipv4_prefix_count"); ok {
		input.Ipv4PrefixCount = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("ipv6_address_count"); ok {
		input.Ipv6AddressCount = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("ipv6_addresses"); ok && v.(*schema.Set).Len() > 0 {
		input.Ipv6Addresses = expandInstanceIPv6Addresses(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ipv6_prefixes"); ok && v.(*schema.Set).Len() > 0 {
		ipv6PrefixesSpecified = true
		input.Ipv6Prefixes = expandIPv6PrefixSpecificationRequests(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ipv6_prefix_count"); ok {
		input.Ipv6PrefixCount = aws.Int32(int32(v.(int)))
	}

	if d.Get("private_ip_list_enabled").(bool) {
		if v, ok := d.GetOk("private_ip_list"); ok && len(v.([]interface{})) > 0 {
			input.PrivateIpAddresses = expandPrivateIPAddressSpecifications(v.([]interface{}))
		}
	} else {
		if v, ok := d.GetOk("private_ips"); ok && v.(*schema.Set).Len() > 0 {
			privateIPs := v.(*schema.Set).List()
			// total includes the primary
			totalPrivateIPs := len(privateIPs)
			// private_ips_count is for secondaries
			if v, ok := d.GetOk("private_ips_count"); ok {
				// reduce total count if necessary
				if v.(int)+1 < totalPrivateIPs {
					totalPrivateIPs = v.(int) + 1
				}
			}
			// truncate the list
			countLimitedIPs := make([]interface{}, totalPrivateIPs)
			for i, ip := range privateIPs {
				countLimitedIPs[i] = ip.(string)
				if i == totalPrivateIPs-1 {
					break
				}
			}
			input.PrivateIpAddresses = expandPrivateIPAddressSpecifications(countLimitedIPs)
		} else {
			if v, ok := d.GetOk("private_ips_count"); ok {
				input.SecondaryPrivateIpAddressCount = aws.Int32(int32(v.(int)))
			}
		}
	}

	if v, ok := d.GetOk(names.AttrSecurityGroups); ok && v.(*schema.Set).Len() > 0 {
		input.Groups = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	// If IPv4 or IPv6 prefixes are specified, tag after create.
	// Otherwise "An error occurred (InternalError) when calling the CreateNetworkInterface operation".
	if !(ipv4PrefixesSpecified || ipv6PrefixesSpecified) {
		input.TagSpecifications = getTagSpecificationsIn(ctx, types.ResourceTypeNetworkInterface)
	}

	output, err := conn.CreateNetworkInterface(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Network Interface: %s", err)
	}

	d.SetId(aws.ToString(output.NetworkInterface.NetworkInterfaceId))

	if _, err := waitNetworkInterfaceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Network Interface (%s) create: %s", d.Id(), err)
	}

	if !d.Get("private_ip_list_enabled").(bool) {
		// add more ips to match the count
		if v, ok := d.GetOk("private_ips"); ok && v.(*schema.Set).Len() > 0 {
			totalPrivateIPs := v.(*schema.Set).Len()
			if privateIPsCount, ok := d.GetOk("private_ips_count"); ok {
				if privateIPsCount.(int)+1 > totalPrivateIPs {
					input := &ec2.AssignPrivateIpAddressesInput{
						NetworkInterfaceId:             aws.String(d.Id()),
						SecondaryPrivateIpAddressCount: aws.Int32(int32(privateIPsCount.(int) + 1 - totalPrivateIPs)),
					}

					_, err := conn.AssignPrivateIpAddresses(ctx, input)

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
					}
				}
			}
		}
	}

	if ipv4PrefixesSpecified || ipv6PrefixesSpecified {
		if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting EC2 Network Interface (%s) tags: %s", d.Id(), err)
		}
	}

	// Default value is enabled.
	if !d.Get("source_dest_check").(bool) {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			SourceDestCheck:    &types.AttributeBooleanValue{Value: aws.Bool(false)},
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 Network Interface (%s) SourceDestCheck: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("attachment"); ok && v.(*schema.Set).Len() > 0 {
		attachment := v.(*schema.Set).List()[0].(map[string]interface{})

		_, err := attachNetworkInterface(ctx, conn, d.Id(), attachment["instance"].(string), attachment["device_index"].(int), networkInterfaceAttachedTimeout)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceNetworkInterfaceRead(ctx, d, meta)...)
}

func resourceNetworkInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findNetworkInterfaceByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network Interface (%s): %s", d.Id(), err)
	}

	eni := outputRaw.(*types.NetworkInterface)

	ownerID := aws.ToString(eni.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ec2",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  "network-interface/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	if eni.Attachment != nil {
		if err := d.Set("attachment", []interface{}{flattenNetworkInterfaceAttachment(eni.Attachment)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting attachment: %s", err)
		}
	} else {
		d.Set("attachment", nil)
	}
	d.Set(names.AttrDescription, eni.Description)
	d.Set("interface_type", eni.InterfaceType)
	if err := d.Set("ipv4_prefixes", flattenIPv4PrefixSpecifications(eni.Ipv4Prefixes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ipv4_prefixes: %s", err)
	}
	d.Set("ipv4_prefix_count", len(eni.Ipv4Prefixes))
	d.Set("ipv6_address_count", len(eni.Ipv6Addresses))
	if err := d.Set("ipv6_address_list", flattenNetworkInterfaceIPv6Addresses(eni.Ipv6Addresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ipv6 address list: %s", err)
	}
	if err := d.Set("ipv6_addresses", flattenNetworkInterfaceIPv6Addresses(eni.Ipv6Addresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ipv6_addresses: %s", err)
	}
	if err := d.Set("ipv6_prefixes", flattenIPv6PrefixSpecifications(eni.Ipv6Prefixes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ipv6_prefixes: %s", err)
	}
	d.Set("ipv6_prefix_count", len(eni.Ipv6Prefixes))
	d.Set("mac_address", eni.MacAddress)
	d.Set("outpost_arn", eni.OutpostArn)
	d.Set(names.AttrOwnerID, ownerID)
	d.Set("private_dns_name", eni.PrivateDnsName)
	d.Set("private_ip", eni.PrivateIpAddress)
	if err := d.Set("private_ips", flattenNetworkInterfacePrivateIPAddresses(eni.PrivateIpAddresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting private_ips: %s", err)
	}
	d.Set("private_ips_count", len(eni.PrivateIpAddresses)-1)
	if err := d.Set("private_ip_list", flattenNetworkInterfacePrivateIPAddresses(eni.PrivateIpAddresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting private_ip_list: %s", err)
	}
	if err := d.Set(names.AttrSecurityGroups, flattenGroupIdentifiers(eni.Groups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting security_groups: %s", err)
	}
	d.Set("source_dest_check", eni.SourceDestCheck)
	d.Set(names.AttrSubnetID, eni.SubnetId)

	setTagsOut(ctx, eni.TagSet)

	return diags
}

func resourceNetworkInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	privateIPsNetChange := 0

	if d.HasChange("attachment") {
		oa, na := d.GetChange("attachment")

		if oa != nil && oa.(*schema.Set).Len() > 0 {
			attachment := oa.(*schema.Set).List()[0].(map[string]interface{})

			if err := detachNetworkInterface(ctx, conn, d.Id(), attachment["attachment_id"].(string), networkInterfaceDetachedTimeout); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		if na != nil && na.(*schema.Set).Len() > 0 {
			attachment := na.(*schema.Set).List()[0].(map[string]interface{})

			if _, err := attachNetworkInterface(ctx, conn, d.Id(), attachment["instance"].(string), attachment["device_index"].(int), networkInterfaceAttachedTimeout); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChange("private_ips") && !d.Get("private_ip_list_enabled").(bool) {
		o, n := d.GetChange("private_ips")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IP addresses.
		unassignIPs := os.Difference(ns)
		if unassignIPs.Len() != 0 {
			input := &ec2.UnassignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringValueSet(unassignIPs),
			}

			_, err := conn.UnassignPrivateIpAddresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
			}

			privateIPsNetChange -= unassignIPs.Len()
		}

		// Assign new IP addresses.
		assignIPs := ns.Difference(os)
		if assignIPs.Len() != 0 {
			input := &ec2.AssignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringValueSet(assignIPs),
			}

			_, err := conn.AssignPrivateIpAddresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
			}
			privateIPsNetChange += assignIPs.Len()
		}
	}

	if d.HasChange("private_ip_list") && d.Get("private_ip_list_enabled").(bool) {
		o, n := d.GetChange("private_ip_list")
		if o == nil {
			o = make([]string, 0)
		}
		if n == nil {
			n = make([]string, 0)
		}
		if len(o.([]interface{}))-1 > 0 {
			privateIPsToUnassign := make([]interface{}, len(o.([]interface{}))-1)
			idx := 0
			for i, ip := range o.([]interface{}) {
				// skip primary private ip address
				if i == 0 {
					continue
				}
				privateIPsToUnassign[idx] = ip
				idx += 1
			}

			// Unassign the secondary IP addresses
			input := &ec2.UnassignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringValueList(privateIPsToUnassign),
			}

			_, err := conn.UnassignPrivateIpAddresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
			}
		}

		// Assign each ip one-by-one in order to retain order
		for i, ip := range n.([]interface{}) {
			// skip primary private ip address
			if i == 0 {
				continue
			}
			privateIPToAssign := []interface{}{ip}

			input := &ec2.AssignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				PrivateIpAddresses: flex.ExpandStringValueList(privateIPToAssign),
			}

			_, err := conn.AssignPrivateIpAddresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("private_ips_count") && !d.Get("private_ip_list_enabled").(bool) {
		o, n := d.GetChange("private_ips_count")
		privateIPs := d.Get("private_ips").(*schema.Set).List()
		privateIPsFiltered := privateIPs[:0]
		primaryIP := d.Get("private_ip")

		for _, ip := range privateIPs {
			if ip != primaryIP {
				privateIPsFiltered = append(privateIPsFiltered, ip)
			}
		}

		if o != nil && n != nil && n != len(privateIPsFiltered) {
			if diff := n.(int) - o.(int) - privateIPsNetChange; diff > 0 {
				input := &ec2.AssignPrivateIpAddressesInput{
					NetworkInterfaceId:             aws.String(d.Id()),
					SecondaryPrivateIpAddressCount: aws.Int32(int32(diff)),
				}

				_, err := conn.AssignPrivateIpAddresses(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignPrivateIpAddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					PrivateIpAddresses: flex.ExpandStringValueList(privateIPsFiltered[0:-diff]),
				}

				_, err := conn.UnassignPrivateIpAddresses(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("ipv4_prefix_count") {
		o, n := d.GetChange("ipv4_prefix_count")
		ipv4Prefixes := d.Get("ipv4_prefixes").(*schema.Set).List()

		if o, n := o.(int), n.(int); n != len(ipv4Prefixes) {
			if diff := n - o; diff > 0 {
				input := &ec2.AssignPrivateIpAddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv4PrefixCount:    aws.Int32(int32(diff)),
				}

				_, err := conn.AssignPrivateIpAddresses(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignPrivateIpAddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv4Prefixes:       flex.ExpandStringValueList(ipv4Prefixes[0:-diff]),
				}

				_, err := conn.UnassignPrivateIpAddresses(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("ipv4_prefixes") {
		o, n := d.GetChange("ipv4_prefixes")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IPV4 prefixes.
		unassignPrefixes := os.Difference(ns)
		if unassignPrefixes.Len() != 0 {
			input := &ec2.UnassignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv4Prefixes:       flex.ExpandStringValueSet(unassignPrefixes),
			}

			_, err := conn.UnassignPrivateIpAddresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
			}
		}

		// Assign new IPV4 prefixes,
		assignPrefixes := ns.Difference(os)
		if assignPrefixes.Len() != 0 {
			input := &ec2.AssignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv4Prefixes:       flex.ExpandStringValueSet(assignPrefixes),
			}

			_, err := conn.AssignPrivateIpAddresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("ipv6_addresses") && !d.Get("ipv6_address_list_enabled").(bool) {
		o, n := d.GetChange("ipv6_addresses")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IPV6 addresses.
		unassignIPs := os.Difference(ns)
		if unassignIPs.Len() != 0 {
			input := &ec2.UnassignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringValueSet(unassignIPs),
			}

			_, err := conn.UnassignIpv6Addresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
			}
		}

		// Assign new IPV6 addresses,
		assignIPs := ns.Difference(os)
		if assignIPs.Len() != 0 {
			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringValueSet(assignIPs),
			}

			_, err := conn.AssignIpv6Addresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("ipv6_address_count") && !d.Get("ipv6_address_list_enabled").(bool) {
		o, n := d.GetChange("ipv6_address_count")
		ipv6Addresses := d.Get("ipv6_addresses").(*schema.Set).List()

		if o != nil && n != nil && n != len(ipv6Addresses) {
			if diff := n.(int) - o.(int); diff > 0 {
				input := &ec2.AssignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6AddressCount:   aws.Int32(int32(diff)),
				}

				_, err := conn.AssignIpv6Addresses(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6Addresses:      flex.ExpandStringValueList(ipv6Addresses[0:-diff]),
				}

				_, err := conn.UnassignIpv6Addresses(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("ipv6_address_list") && d.Get("ipv6_address_list_enabled").(bool) {
		o, n := d.GetChange("ipv6_address_list")
		if o == nil {
			o = make([]string, 0)
		}
		if n == nil {
			n = make([]string, 0)
		}

		// Unassign old IPV6 addresses
		if len(o.([]interface{})) > 0 {
			unassignIPs := make([]interface{}, len(o.([]interface{})))
			copy(unassignIPs, o.([]interface{}))

			input := &ec2.UnassignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringValueList(unassignIPs),
			}

			_, err := conn.UnassignIpv6Addresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) private IPv6 addresses: %s", d.Id(), err)
			}
		}

		// Assign each ip one-by-one in order to retain order
		for _, ip := range n.([]interface{}) {
			privateIPToAssign := []interface{}{ip}

			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringValueList(privateIPToAssign),
			}

			_, err := conn.AssignIpv6Addresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) private IPv6 addresses: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("ipv6_prefixes") {
		o, n := d.GetChange("ipv6_prefixes")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// Unassign old IPV6 prefixes.
		unassignPrefixes := os.Difference(ns)
		if unassignPrefixes.Len() != 0 {
			input := &ec2.UnassignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Prefixes:       flex.ExpandStringValueSet(unassignPrefixes),
			}

			_, err := conn.UnassignIpv6Addresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
			}
		}

		// Assign new IPV6 prefixes,
		assignPrefixes := ns.Difference(os)
		if assignPrefixes.Len() != 0 {
			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Prefixes:       flex.ExpandStringValueSet(assignPrefixes),
			}

			_, err := conn.AssignIpv6Addresses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("ipv6_prefix_count") {
		o, n := d.GetChange("ipv6_prefix_count")
		ipv6Prefixes := d.Get("ipv6_prefixes").(*schema.Set).List()

		if o, n := o.(int), n.(int); n != len(ipv6Prefixes) {
			if diff := n - o; diff > 0 {
				input := &ec2.AssignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6PrefixCount:    aws.Int32(int32(diff)),
				}

				_, err := conn.AssignIpv6Addresses(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6Prefixes:       flex.ExpandStringValueList(ipv6Prefixes[0:-diff]),
				}

				_, err := conn.UnassignIpv6Addresses(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("source_dest_check") {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			SourceDestCheck:    &types.AttributeBooleanValue{Value: aws.Bool(d.Get("source_dest_check").(bool))},
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 Network Interface (%s) SourceDestCheck: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrSecurityGroups) {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			Groups:             flex.ExpandStringValueSet(d.Get(names.AttrSecurityGroups).(*schema.Set)),
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 Network Interface (%s) Groups: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrDescription) {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			Description:        &types.AttributeValue{Value: aws.String(d.Get(names.AttrDescription).(string))},
		}

		_, err := conn.ModifyNetworkInterfaceAttribute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 Network Interface (%s) Description: %s", d.Id(), err)
		}
	}

	return append(diags, resourceNetworkInterfaceRead(ctx, d, meta)...)
}

func resourceNetworkInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if v, ok := d.GetOk("attachment"); ok && v.(*schema.Set).Len() > 0 {
		attachment := v.(*schema.Set).List()[0].(map[string]interface{})

		if err := detachNetworkInterface(ctx, conn, d.Id(), attachment["attachment_id"].(string), networkInterfaceDetachedTimeout); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if err := deleteNetworkInterface(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	return diags
}

func attachNetworkInterface(ctx context.Context, conn *ec2.Client, networkInterfaceID, instanceID string, deviceIndex int, timeout time.Duration) (string, error) {
	input := &ec2.AttachNetworkInterfaceInput{
		DeviceIndex:        aws.Int32(int32(deviceIndex)),
		InstanceId:         aws.String(instanceID),
		NetworkInterfaceId: aws.String(networkInterfaceID),
	}

	output, err := conn.AttachNetworkInterface(ctx, input)

	if err != nil {
		return "", fmt.Errorf("attaching EC2 Network Interface (%s/%s): %w", networkInterfaceID, instanceID, err)
	}

	attachmentID := aws.ToString(output.AttachmentId)

	if _, err := waitNetworkInterfaceAttached(ctx, conn, attachmentID, timeout); err != nil {
		return "", fmt.Errorf("waiting for EC2 Network Interface (%s/%s) attach: %w", networkInterfaceID, instanceID, err)
	}

	return attachmentID, nil
}

func deleteNetworkInterface(ctx context.Context, conn *ec2.Client, networkInterfaceID string) error {
	log.Printf("[INFO] Deleting EC2 Network Interface: %s", networkInterfaceID)
	_, err := conn.DeleteNetworkInterface(ctx, &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: aws.String(networkInterfaceID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInterfaceIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Network Interface (%s): %w", networkInterfaceID, err)
	}

	return nil
}

func detachNetworkInterface(ctx context.Context, conn *ec2.Client, networkInterfaceID, attachmentID string, timeout time.Duration) error {
	log.Printf("[INFO] Detaching EC2 Network Interface: %s", networkInterfaceID)
	_, err := conn.DetachNetworkInterface(ctx, &ec2.DetachNetworkInterfaceInput{
		AttachmentId: aws.String(attachmentID),
		Force:        aws.Bool(true),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAttachmentIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching EC2 Network Interface (%s/%s): %w", networkInterfaceID, attachmentID, err)
	}

	_, err = waitNetworkInterfaceDetached(ctx, conn, attachmentID, timeout)

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("waiting for EC2 Network Interface (%s/%s) detach: %w", networkInterfaceID, attachmentID, err)
	}

	return nil
}

func flattenNetworkInterfaceAssociation(apiObject *types.NetworkInterfaceAssociation) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllocationId; v != nil {
		tfMap["allocation_id"] = aws.ToString(v)
	}

	if v := apiObject.AssociationId; v != nil {
		tfMap[names.AttrAssociationID] = aws.ToString(v)
	}

	if v := apiObject.CarrierIp; v != nil {
		tfMap["carrier_ip"] = aws.ToString(v)
	}

	if v := apiObject.CustomerOwnedIp; v != nil {
		tfMap["customer_owned_ip"] = aws.ToString(v)
	}

	if v := apiObject.IpOwnerId; v != nil {
		tfMap["ip_owner_id"] = aws.ToString(v)
	}

	if v := apiObject.PublicDnsName; v != nil {
		tfMap["public_dns_name"] = aws.ToString(v)
	}

	if v := apiObject.PublicIp; v != nil {
		tfMap["public_ip"] = aws.ToString(v)
	}

	return tfMap
}

func flattenNetworkInterfaceAttachment(apiObject *types.NetworkInterfaceAttachment) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AttachmentId; v != nil {
		tfMap["attachment_id"] = aws.ToString(v)
	}

	if v := apiObject.DeviceIndex; v != nil {
		tfMap["device_index"] = aws.ToInt32(v)
	}

	if v := apiObject.InstanceId; v != nil {
		tfMap["instance"] = aws.ToString(v)
	}

	return tfMap
}

func expandPrivateIPAddressSpecification(tfString string) *types.PrivateIpAddressSpecification {
	if tfString == "" {
		return nil
	}

	apiObject := &types.PrivateIpAddressSpecification{
		PrivateIpAddress: aws.String(tfString),
	}

	return apiObject
}

func expandPrivateIPAddressSpecifications(tfList []interface{}) []types.PrivateIpAddressSpecification {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.PrivateIpAddressSpecification

	for i, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := expandPrivateIPAddressSpecification(tfString)

		if apiObject == nil {
			continue
		}

		if i == 0 {
			apiObject.Primary = aws.Bool(true)
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandInstanceIPv6Address(tfString string) *types.InstanceIpv6Address {
	if tfString == "" {
		return nil
	}

	apiObject := &types.InstanceIpv6Address{
		Ipv6Address: aws.String(tfString),
	}

	return apiObject
}

func expandInstanceIPv6Addresses(tfList []interface{}) []types.InstanceIpv6Address {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.InstanceIpv6Address

	for _, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := expandInstanceIPv6Address(tfString)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenNetworkInterfacePrivateIPAddress(apiObject *types.NetworkInterfacePrivateIpAddress) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.PrivateIpAddress; v != nil {
		tfString = aws.ToString(v)
	}

	return tfString
}

func flattenNetworkInterfacePrivateIPAddresses(apiObjects []types.NetworkInterfacePrivateIpAddress) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenNetworkInterfacePrivateIPAddress(&apiObject))
	}

	return tfList
}

func flattenNetworkInterfaceIPv6Address(apiObject *types.NetworkInterfaceIpv6Address) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.Ipv6Address; v != nil {
		tfString = aws.ToString(v)
	}

	return tfString
}

func flattenNetworkInterfaceIPv6Addresses(apiObjects []types.NetworkInterfaceIpv6Address) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenNetworkInterfaceIPv6Address(&apiObject))
	}

	return tfList
}

func expandIPv4PrefixSpecificationRequest(tfString string) *types.Ipv4PrefixSpecificationRequest {
	if tfString == "" {
		return nil
	}

	apiObject := &types.Ipv4PrefixSpecificationRequest{
		Ipv4Prefix: aws.String(tfString),
	}

	return apiObject
}

func expandIPv4PrefixSpecificationRequests(tfList []interface{}) []types.Ipv4PrefixSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.Ipv4PrefixSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := expandIPv4PrefixSpecificationRequest(tfString)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandIPv6PrefixSpecificationRequest(tfString string) *types.Ipv6PrefixSpecificationRequest {
	if tfString == "" {
		return nil
	}

	apiObject := &types.Ipv6PrefixSpecificationRequest{
		Ipv6Prefix: aws.String(tfString),
	}

	return apiObject
}

func expandIPv6PrefixSpecificationRequests(tfList []interface{}) []types.Ipv6PrefixSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.Ipv6PrefixSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := expandIPv6PrefixSpecificationRequest(tfString)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenIPv4PrefixSpecification(apiObject *types.Ipv4PrefixSpecification) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.Ipv4Prefix; v != nil {
		tfString = aws.ToString(v)
	}

	return tfString
}

func flattenIPv4PrefixSpecifications(apiObjects []types.Ipv4PrefixSpecification) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenIPv4PrefixSpecification(&apiObject))
	}

	return tfList
}

func flattenIPv6PrefixSpecification(apiObject *types.Ipv6PrefixSpecification) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.Ipv6Prefix; v != nil {
		tfString = aws.ToString(v)
	}

	return tfString
}

func flattenIPv6PrefixSpecifications(apiObjects []types.Ipv6PrefixSpecification) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenIPv6PrefixSpecification(&apiObject))
	}

	return tfList
}

// Some AWS services creates ENIs behind the scenes and keeps these around for a while
// which can prevent security groups and subnets attached to such ENIs from being destroyed
func deleteLingeringENIs(ctx context.Context, conn *ec2.Client, filterName, resourceId string, timeout time.Duration) error {
	var g multierror.Group

	tflog.Trace(ctx, "Checking for lingering ENIs")

	enis, err := findNetworkInterfaces(ctx, conn, &ec2.DescribeNetworkInterfacesInput{
		Filters: newAttributeFilterList(map[string]string{
			filterName: resourceId,
		}),
	})
	if err != nil {
		return fmt.Errorf("listing EC2 Network Interfaces: %w", err)
	}

	for _, eni := range enis {
		eni := &eni

		if found := deleteLingeringLambdaENI(ctx, &g, conn, eni, timeout); found {
			continue
		}

		if found := deleteLingeringComprehendENI(ctx, &g, conn, eni, timeout); found {
			continue
		}

		deleteLingeringDMSENI(ctx, &g, conn, eni, timeout)
	}

	return g.Wait().ErrorOrNil()
}

func deleteLingeringLambdaENI(ctx context.Context, g *multierror.Group, conn *ec2.Client, eni *types.NetworkInterface, timeout time.Duration) bool {
	// AWS Lambda service team confirms P99 deletion time of ~35 minutes. Buffer for safety.
	if minimumTimeout := 45 * time.Minute; timeout < minimumTimeout {
		timeout = minimumTimeout
	}

	if !strings.HasPrefix(aws.ToString(eni.Description), "AWS Lambda VPC ENI") {
		return false
	}

	g.Go(func() error {
		networkInterfaceID := aws.ToString(eni.NetworkInterfaceId)

		if eni.Attachment != nil && aws.ToString(eni.Attachment.InstanceOwnerId) == "amazon-aws" {
			networkInterface, err := waitNetworkInterfaceAvailableAfterUse(ctx, conn, networkInterfaceID, timeout)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("waiting for Lambda ENI (%s) to become available for detachment: %w", networkInterfaceID, err)
			}

			eni = networkInterface
		}

		if eni.Attachment != nil {
			if err := detachNetworkInterface(ctx, conn, networkInterfaceID, aws.ToString(eni.Attachment.AttachmentId), timeout); err != nil {
				return fmt.Errorf("detaching Lambda ENI (%s): %w", networkInterfaceID, err)
			}
		}

		if err := deleteNetworkInterface(ctx, conn, networkInterfaceID); err != nil {
			return fmt.Errorf("deleting Lambda ENI (%s): %w", networkInterfaceID, err)
		}

		return nil
	})

	return true
}

func deleteLingeringComprehendENI(ctx context.Context, g *multierror.Group, conn *ec2.Client, eni *types.NetworkInterface, timeout time.Duration) bool {
	// Deletion appears to take approximately 5 minutes
	if minimumTimeout := 10 * time.Minute; timeout < minimumTimeout {
		timeout = minimumTimeout
	}

	if !strings.HasSuffix(aws.ToString(eni.RequesterId), ":Comprehend") {
		return false
	}

	g.Go(func() error {
		networkInterfaceID := aws.ToString(eni.NetworkInterfaceId)

		if eni.Attachment != nil {
			if err := detachNetworkInterface(ctx, conn, networkInterfaceID, aws.ToString(eni.Attachment.AttachmentId), timeout); err != nil {
				return fmt.Errorf("detaching Comprehend ENI (%s): %w", networkInterfaceID, err)
			}
		}

		if err := deleteNetworkInterface(ctx, conn, networkInterfaceID); err != nil {
			return fmt.Errorf("deleting Comprehend ENI (%s): %w", networkInterfaceID, err)
		}

		return nil
	})

	return true
}

func deleteLingeringDMSENI(ctx context.Context, g *multierror.Group, conn *ec2.Client, v *types.NetworkInterface, timeout time.Duration) bool {
	// Deletion appears to take approximately 5 minutes
	if minimumTimeout := 10 * time.Minute; timeout < minimumTimeout {
		timeout = minimumTimeout
	}

	if aws.ToString(v.Description) != "DMSNetworkInterface" {
		return false
	}

	g.Go(func() error {
		networkInterfaceID := aws.ToString(v.NetworkInterfaceId)

		if v.Attachment != nil {
			if err := detachNetworkInterface(ctx, conn, networkInterfaceID, aws.ToString(v.Attachment.AttachmentId), timeout); err != nil {
				return fmt.Errorf("detaching DMS ENI (%s): %w", networkInterfaceID, err)
			}
		}

		if err := deleteNetworkInterface(ctx, conn, networkInterfaceID); err != nil {
			return fmt.Errorf("deleting DMS ENI (%s): %w", networkInterfaceID, err)
		}

		return nil
	})

	return true
}

// Flattens security group identifiers into a []string, where the elements returned are the GroupIDs
func flattenGroupIdentifiers(dtos []types.GroupIdentifier) []string {
	ids := make([]string, 0, len(dtos))
	for _, v := range dtos {
		group_id := aws.ToString(v.GroupId)
		ids = append(ids, group_id)
	}
	return ids
}
