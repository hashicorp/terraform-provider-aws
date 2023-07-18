// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_network_interface", name="Network Interface")
// @Tags(identifierAttribute="id")
func ResourceNetworkInterface() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkInterfaceCreate,
		ReadWithoutTimeout:   resourceNetworkInterfaceRead,
		UpdateWithoutTimeout: resourceNetworkInterfaceUpdate,
		DeleteWithoutTimeout: resourceNetworkInterfaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"interface_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.NetworkInterfaceCreationType_Values(), false),
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
			"owner_id": {
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
			"security_groups": {
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
			"subnet_id": {
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
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	ipv4PrefixesSpecified := false
	ipv6PrefixesSpecified := false

	input := &ec2.CreateNetworkInterfaceInput{
		ClientToken: aws.String(id.UniqueId()),
		SubnetId:    aws.String(d.Get("subnet_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("interface_type"); ok {
		input.InterfaceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipv4_prefixes"); ok && v.(*schema.Set).Len() > 0 {
		ipv4PrefixesSpecified = true
		input.Ipv4Prefixes = expandIPv4PrefixSpecificationRequests(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ipv4_prefix_count"); ok {
		input.Ipv4PrefixCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("ipv6_address_count"); ok {
		input.Ipv6AddressCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("ipv6_addresses"); ok && v.(*schema.Set).Len() > 0 {
		input.Ipv6Addresses = expandInstanceIPv6Addresses(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ipv6_prefixes"); ok && v.(*schema.Set).Len() > 0 {
		ipv6PrefixesSpecified = true
		input.Ipv6Prefixes = expandIPv6PrefixSpecificationRequests(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ipv6_prefix_count"); ok {
		input.Ipv6PrefixCount = aws.Int64(int64(v.(int)))
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
				input.SecondaryPrivateIpAddressCount = aws.Int64(int64(v.(int)))
			}
		}
	}

	if v, ok := d.GetOk("security_groups"); ok && v.(*schema.Set).Len() > 0 {
		input.Groups = flex.ExpandStringSet(v.(*schema.Set))
	}

	// If IPv4 or IPv6 prefixes are specified, tag after create.
	// Otherwise "An error occurred (InternalError) when calling the CreateNetworkInterface operation".
	if !(ipv4PrefixesSpecified || ipv6PrefixesSpecified) {
		input.TagSpecifications = getTagSpecificationsIn(ctx, ec2.ResourceTypeNetworkInterface)
	}

	output, err := conn.CreateNetworkInterfaceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Network Interface: %s", err)
	}

	d.SetId(aws.StringValue(output.NetworkInterface.NetworkInterfaceId))

	if _, err := WaitNetworkInterfaceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
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
						SecondaryPrivateIpAddressCount: aws.Int64(int64(privateIPsCount.(int) + 1 - totalPrivateIPs)),
					}

					_, err := conn.AssignPrivateIpAddressesWithContext(ctx, input)

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
			SourceDestCheck:    &ec2.AttributeBooleanValue{Value: aws.Bool(false)},
		}

		_, err := conn.ModifyNetworkInterfaceAttributeWithContext(ctx, input)

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
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return FindNetworkInterfaceByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network Interface (%s): %s", d.Id(), err)
	}

	eni := outputRaw.(*ec2.NetworkInterface)

	ownerID := aws.StringValue(eni.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("network-interface/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	if eni.Attachment != nil {
		if err := d.Set("attachment", []interface{}{flattenNetworkInterfaceAttachment(eni.Attachment)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting attachment: %s", err)
		}
	} else {
		d.Set("attachment", nil)
	}
	d.Set("description", eni.Description)
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
	d.Set("owner_id", ownerID)
	d.Set("private_dns_name", eni.PrivateDnsName)
	d.Set("private_ip", eni.PrivateIpAddress)
	if err := d.Set("private_ips", FlattenNetworkInterfacePrivateIPAddresses(eni.PrivateIpAddresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting private_ips: %s", err)
	}
	d.Set("private_ips_count", len(eni.PrivateIpAddresses)-1)
	if err := d.Set("private_ip_list", FlattenNetworkInterfacePrivateIPAddresses(eni.PrivateIpAddresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting private_ip_list: %s", err)
	}
	if err := d.Set("security_groups", FlattenGroupIdentifiers(eni.Groups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting security_groups: %s", err)
	}
	d.Set("source_dest_check", eni.SourceDestCheck)
	d.Set("subnet_id", eni.SubnetId)

	setTagsOut(ctx, eni.TagSet)

	return diags
}

func resourceNetworkInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	privateIPsNetChange := 0

	if d.HasChange("attachment") {
		oa, na := d.GetChange("attachment")

		if oa != nil && oa.(*schema.Set).Len() > 0 {
			attachment := oa.(*schema.Set).List()[0].(map[string]interface{})

			err := DetachNetworkInterface(ctx, conn, d.Id(), attachment["attachment_id"].(string), NetworkInterfaceDetachedTimeout)

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		if na != nil && na.(*schema.Set).Len() > 0 {
			attachment := na.(*schema.Set).List()[0].(map[string]interface{})

			_, err := attachNetworkInterface(ctx, conn, d.Id(), attachment["instance"].(string), attachment["device_index"].(int), networkInterfaceAttachedTimeout)

			if err != nil {
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
				PrivateIpAddresses: flex.ExpandStringSet(unassignIPs),
			}

			_, err := conn.UnassignPrivateIpAddressesWithContext(ctx, input)

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
				PrivateIpAddresses: flex.ExpandStringSet(assignIPs),
			}

			_, err := conn.AssignPrivateIpAddressesWithContext(ctx, input)

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
				PrivateIpAddresses: flex.ExpandStringList(privateIPsToUnassign),
			}

			_, err := conn.UnassignPrivateIpAddressesWithContext(ctx, input)

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
				PrivateIpAddresses: flex.ExpandStringList(privateIPToAssign),
			}

			_, err := conn.AssignPrivateIpAddressesWithContext(ctx, input)

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
					SecondaryPrivateIpAddressCount: aws.Int64(int64(diff)),
				}

				_, err := conn.AssignPrivateIpAddressesWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignPrivateIpAddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					PrivateIpAddresses: flex.ExpandStringList(privateIPsFiltered[0:-diff]),
				}

				_, err := conn.UnassignPrivateIpAddressesWithContext(ctx, input)

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
					Ipv4PrefixCount:    aws.Int64(int64(diff)),
				}

				_, err := conn.AssignPrivateIpAddressesWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignPrivateIpAddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv4Prefixes:       flex.ExpandStringList(ipv4Prefixes[0:-diff]),
				}

				_, err := conn.UnassignPrivateIpAddressesWithContext(ctx, input)

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
				Ipv4Prefixes:       flex.ExpandStringSet(unassignPrefixes),
			}

			_, err := conn.UnassignPrivateIpAddressesWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) private IPv4 addresses: %s", d.Id(), err)
			}
		}

		// Assign new IPV4 prefixes,
		assignPrefixes := ns.Difference(os)
		if assignPrefixes.Len() != 0 {
			input := &ec2.AssignPrivateIpAddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv4Prefixes:       flex.ExpandStringSet(assignPrefixes),
			}

			_, err := conn.AssignPrivateIpAddressesWithContext(ctx, input)

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
				Ipv6Addresses:      flex.ExpandStringSet(unassignIPs),
			}

			_, err := conn.UnassignIpv6AddressesWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
			}
		}

		// Assign new IPV6 addresses,
		assignIPs := ns.Difference(os)
		if assignIPs.Len() != 0 {
			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringSet(assignIPs),
			}

			_, err := conn.AssignIpv6AddressesWithContext(ctx, input)

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
					Ipv6AddressCount:   aws.Int64(int64(diff)),
				}

				_, err := conn.AssignIpv6AddressesWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6Addresses:      flex.ExpandStringList(ipv6Addresses[0:-diff]),
				}

				_, err := conn.UnassignIpv6AddressesWithContext(ctx, input)

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
				Ipv6Addresses:      flex.ExpandStringList(unassignIPs),
			}

			_, err := conn.UnassignIpv6AddressesWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) private IPv6 addresses: %s", d.Id(), err)
			}
		}

		// Assign each ip one-by-one in order to retain order
		for _, ip := range n.([]interface{}) {
			privateIPToAssign := []interface{}{ip}

			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Addresses:      flex.ExpandStringList(privateIPToAssign),
			}

			_, err := conn.AssignIpv6AddressesWithContext(ctx, input)

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
				Ipv6Prefixes:       flex.ExpandStringSet(unassignPrefixes),
			}

			_, err := conn.UnassignIpv6AddressesWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
			}
		}

		// Assign new IPV6 prefixes,
		assignPrefixes := ns.Difference(os)
		if assignPrefixes.Len() != 0 {
			input := &ec2.AssignIpv6AddressesInput{
				NetworkInterfaceId: aws.String(d.Id()),
				Ipv6Prefixes:       flex.ExpandStringSet(assignPrefixes),
			}

			_, err := conn.AssignIpv6AddressesWithContext(ctx, input)

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
					Ipv6PrefixCount:    aws.Int64(int64(diff)),
				}

				_, err := conn.AssignIpv6AddressesWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "assigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
				}
			} else if diff < 0 {
				input := &ec2.UnassignIpv6AddressesInput{
					NetworkInterfaceId: aws.String(d.Id()),
					Ipv6Prefixes:       flex.ExpandStringList(ipv6Prefixes[0:-diff]),
				}

				_, err := conn.UnassignIpv6AddressesWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "unassigning EC2 Network Interface (%s) IPv6 addresses: %s", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("source_dest_check") {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			SourceDestCheck:    &ec2.AttributeBooleanValue{Value: aws.Bool(d.Get("source_dest_check").(bool))},
		}

		_, err := conn.ModifyNetworkInterfaceAttributeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 Network Interface (%s) SourceDestCheck: %s", d.Id(), err)
		}
	}

	if d.HasChange("security_groups") {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			Groups:             flex.ExpandStringSet(d.Get("security_groups").(*schema.Set)),
		}

		_, err := conn.ModifyNetworkInterfaceAttributeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 Network Interface (%s) Groups: %s", d.Id(), err)
		}
	}

	if d.HasChange("description") {
		input := &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: aws.String(d.Id()),
			Description:        &ec2.AttributeValue{Value: aws.String(d.Get("description").(string))},
		}

		_, err := conn.ModifyNetworkInterfaceAttributeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 Network Interface (%s) Description: %s", d.Id(), err)
		}
	}

	return append(diags, resourceNetworkInterfaceRead(ctx, d, meta)...)
}

func resourceNetworkInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	if v, ok := d.GetOk("attachment"); ok && v.(*schema.Set).Len() > 0 {
		attachment := v.(*schema.Set).List()[0].(map[string]interface{})

		if err := DetachNetworkInterface(ctx, conn, d.Id(), attachment["attachment_id"].(string), NetworkInterfaceDetachedTimeout); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if err := DeleteNetworkInterface(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	return diags
}

func attachNetworkInterface(ctx context.Context, conn *ec2.EC2, networkInterfaceID, instanceID string, deviceIndex int, timeout time.Duration) (string, error) {
	input := &ec2.AttachNetworkInterfaceInput{
		DeviceIndex:        aws.Int64(int64(deviceIndex)),
		InstanceId:         aws.String(instanceID),
		NetworkInterfaceId: aws.String(networkInterfaceID),
	}

	output, err := conn.AttachNetworkInterfaceWithContext(ctx, input)

	if err != nil {
		return "", fmt.Errorf("attaching EC2 Network Interface (%s/%s): %w", networkInterfaceID, instanceID, err)
	}

	attachmentID := aws.StringValue(output.AttachmentId)

	_, err = WaitNetworkInterfaceAttached(ctx, conn, attachmentID, timeout)

	if err != nil {
		return "", fmt.Errorf("attaching EC2 Network Interface (%s/%s): waiting for completion: %w", networkInterfaceID, instanceID, err)
	}

	return attachmentID, nil
}

func DeleteNetworkInterface(ctx context.Context, conn *ec2.EC2, networkInterfaceID string) error {
	log.Printf("[INFO] Deleting EC2 Network Interface: %s", networkInterfaceID)
	_, err := conn.DeleteNetworkInterfaceWithContext(ctx, &ec2.DeleteNetworkInterfaceInput{
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

func DetachNetworkInterface(ctx context.Context, conn *ec2.EC2, networkInterfaceID, attachmentID string, timeout time.Duration) error {
	log.Printf("[INFO] Detaching EC2 Network Interface: %s", networkInterfaceID)
	_, err := conn.DetachNetworkInterfaceWithContext(ctx, &ec2.DetachNetworkInterfaceInput{
		AttachmentId: aws.String(attachmentID),
		Force:        aws.Bool(true),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAttachmentIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching EC2 Network Interface (%s/%s): %w", networkInterfaceID, attachmentID, err)
	}

	_, err = WaitNetworkInterfaceDetached(ctx, conn, attachmentID, timeout)

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching EC2 Network Interface (%s/%s): waiting for completion: %w", networkInterfaceID, attachmentID, err)
	}

	return nil
}

func flattenNetworkInterfaceAssociation(apiObject *ec2.NetworkInterfaceAssociation) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllocationId; v != nil {
		tfMap["allocation_id"] = aws.StringValue(v)
	}

	if v := apiObject.AssociationId; v != nil {
		tfMap["association_id"] = aws.StringValue(v)
	}

	if v := apiObject.CarrierIp; v != nil {
		tfMap["carrier_ip"] = aws.StringValue(v)
	}

	if v := apiObject.CustomerOwnedIp; v != nil {
		tfMap["customer_owned_ip"] = aws.StringValue(v)
	}

	if v := apiObject.IpOwnerId; v != nil {
		tfMap["ip_owner_id"] = aws.StringValue(v)
	}

	if v := apiObject.PublicDnsName; v != nil {
		tfMap["public_dns_name"] = aws.StringValue(v)
	}

	if v := apiObject.PublicIp; v != nil {
		tfMap["public_ip"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenNetworkInterfaceAttachment(apiObject *ec2.NetworkInterfaceAttachment) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AttachmentId; v != nil {
		tfMap["attachment_id"] = aws.StringValue(v)
	}

	if v := apiObject.DeviceIndex; v != nil {
		tfMap["device_index"] = aws.Int64Value(v)
	}

	if v := apiObject.InstanceId; v != nil {
		tfMap["instance"] = aws.StringValue(v)
	}

	return tfMap
}

func expandPrivateIPAddressSpecification(tfString string) *ec2.PrivateIpAddressSpecification {
	if tfString == "" {
		return nil
	}

	apiObject := &ec2.PrivateIpAddressSpecification{
		PrivateIpAddress: aws.String(tfString),
	}

	return apiObject
}

func expandPrivateIPAddressSpecifications(tfList []interface{}) []*ec2.PrivateIpAddressSpecification {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.PrivateIpAddressSpecification

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

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandInstanceIPv6Address(tfString string) *ec2.InstanceIpv6Address {
	if tfString == "" {
		return nil
	}

	apiObject := &ec2.InstanceIpv6Address{
		Ipv6Address: aws.String(tfString),
	}

	return apiObject
}

func expandInstanceIPv6Addresses(tfList []interface{}) []*ec2.InstanceIpv6Address {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.InstanceIpv6Address

	for _, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := expandInstanceIPv6Address(tfString)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenNetworkInterfacePrivateIPAddress(apiObject *ec2.NetworkInterfacePrivateIpAddress) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.PrivateIpAddress; v != nil {
		tfString = aws.StringValue(v)
	}

	return tfString
}

func FlattenNetworkInterfacePrivateIPAddresses(apiObjects []*ec2.NetworkInterfacePrivateIpAddress) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenNetworkInterfacePrivateIPAddress(apiObject))
	}

	return tfList
}

func flattenNetworkInterfaceIPv6Address(apiObject *ec2.NetworkInterfaceIpv6Address) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.Ipv6Address; v != nil {
		tfString = aws.StringValue(v)
	}

	return tfString
}

func flattenNetworkInterfaceIPv6Addresses(apiObjects []*ec2.NetworkInterfaceIpv6Address) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenNetworkInterfaceIPv6Address(apiObject))
	}

	return tfList
}

func expandIPv4PrefixSpecificationRequest(tfString string) *ec2.Ipv4PrefixSpecificationRequest {
	if tfString == "" {
		return nil
	}

	apiObject := &ec2.Ipv4PrefixSpecificationRequest{
		Ipv4Prefix: aws.String(tfString),
	}

	return apiObject
}

func expandIPv4PrefixSpecificationRequests(tfList []interface{}) []*ec2.Ipv4PrefixSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.Ipv4PrefixSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := expandIPv4PrefixSpecificationRequest(tfString)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandIPv6PrefixSpecificationRequest(tfString string) *ec2.Ipv6PrefixSpecificationRequest {
	if tfString == "" {
		return nil
	}

	apiObject := &ec2.Ipv6PrefixSpecificationRequest{
		Ipv6Prefix: aws.String(tfString),
	}

	return apiObject
}

func expandIPv6PrefixSpecificationRequests(tfList []interface{}) []*ec2.Ipv6PrefixSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.Ipv6PrefixSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObject := expandIPv6PrefixSpecificationRequest(tfString)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenIPv4PrefixSpecification(apiObject *ec2.Ipv4PrefixSpecification) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.Ipv4Prefix; v != nil {
		tfString = aws.StringValue(v)
	}

	return tfString
}

func flattenIPv4PrefixSpecifications(apiObjects []*ec2.Ipv4PrefixSpecification) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenIPv4PrefixSpecification(apiObject))
	}

	return tfList
}

func flattenIPv6PrefixSpecification(apiObject *ec2.Ipv6PrefixSpecification) string {
	if apiObject == nil {
		return ""
	}

	tfString := ""

	if v := apiObject.Ipv6Prefix; v != nil {
		tfString = aws.StringValue(v)
	}

	return tfString
}

func flattenIPv6PrefixSpecifications(apiObjects []*ec2.Ipv6PrefixSpecification) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenIPv6PrefixSpecification(apiObject))
	}

	return tfList
}

// Some AWS services creates ENIs behind the scenes and keeps these around for a while
// which can prevent security groups and subnets attached to such ENIs from being destroyed
func deleteLingeringENIs(ctx context.Context, conn *ec2.EC2, filterName, resourceId string, timeout time.Duration) error {
	var g multierror.Group

	err := multierror.Append(nil, deleteLingeringLambdaENIs(ctx, &g, conn, filterName, resourceId, timeout))

	err = multierror.Append(err, deleteLingeringComprehendENIs(ctx, &g, conn, filterName, resourceId, timeout))

	return multierror.Append(err, g.Wait()).ErrorOrNil()
}

func deleteLingeringLambdaENIs(ctx context.Context, g *multierror.Group, conn *ec2.EC2, filterName, resourceId string, timeout time.Duration) error {
	// AWS Lambda service team confirms P99 deletion time of ~35 minutes. Buffer for safety.
	if minimumTimeout := 45 * time.Minute; timeout < minimumTimeout {
		timeout = minimumTimeout
	}

	networkInterfaces, err := FindNetworkInterfaces(ctx, conn, &ec2.DescribeNetworkInterfacesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			filterName:    resourceId,
			"description": "AWS Lambda VPC ENI*",
		}),
	})

	if err != nil {
		return fmt.Errorf("listing EC2 Network Interfaces: %w", err)
	}

	for _, v := range networkInterfaces {
		v := v
		g.Go(func() error {
			networkInterfaceID := aws.StringValue(v.NetworkInterfaceId)

			if v.Attachment != nil && aws.StringValue(v.Attachment.InstanceOwnerId) == "amazon-aws" {
				networkInterface, err := WaitNetworkInterfaceAvailableAfterUse(ctx, conn, networkInterfaceID, timeout)

				if tfresource.NotFound(err) {
					return nil
				}

				if err != nil {
					return fmt.Errorf("waiting for Lambda ENI (%s) to become available for detachment: %w", networkInterfaceID, err)
				}

				v = networkInterface
			}

			if v.Attachment != nil {
				err = DetachNetworkInterface(ctx, conn, networkInterfaceID, aws.StringValue(v.Attachment.AttachmentId), timeout)

				if err != nil {
					return fmt.Errorf("detaching Lambda ENI (%s): %w", networkInterfaceID, err)
				}
			}

			err = DeleteNetworkInterface(ctx, conn, networkInterfaceID)

			if err != nil {
				return fmt.Errorf("deleting Lambda ENI (%s): %w", networkInterfaceID, err)
			}

			return nil
		})
	}

	return nil
}

func deleteLingeringComprehendENIs(ctx context.Context, g *multierror.Group, conn *ec2.EC2, filterName, resourceId string, timeout time.Duration) error {
	// Deletion appears to take approximately 5 minutes
	if minimumTimeout := 10 * time.Minute; timeout < minimumTimeout {
		timeout = minimumTimeout
	}

	enis, err := FindNetworkInterfaces(ctx, conn, &ec2.DescribeNetworkInterfacesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			filterName: resourceId,
		}),
	})
	if err != nil {
		return fmt.Errorf("listing EC2 Network Interfaces: %w", err)
	}

	networkInterfaces := make([]*ec2.NetworkInterface, 0, len(enis))
	for _, v := range enis {
		if strings.HasSuffix(aws.StringValue(v.RequesterId), ":Comprehend") {
			networkInterfaces = append(networkInterfaces, v)
		}
	}

	for _, v := range networkInterfaces {
		v := v
		g.Go(func() error {
			networkInterfaceID := aws.StringValue(v.NetworkInterfaceId)

			if v.Attachment != nil {
				err = DetachNetworkInterface(ctx, conn, networkInterfaceID, aws.StringValue(v.Attachment.AttachmentId), timeout)

				if err != nil {
					return fmt.Errorf("detaching Comprehend ENI (%s): %w", networkInterfaceID, err)
				}
			}

			err := DeleteNetworkInterface(ctx, conn, networkInterfaceID)

			if err != nil {
				return fmt.Errorf("deleting Comprehend ENI (%s): %w", networkInterfaceID, err)
			}

			return nil
		})
	}

	return nil
}

// Flattens security group identifiers into a []string, where the elements returned are the GroupIDs
func FlattenGroupIdentifiers(dtos []*ec2.GroupIdentifier) []string {
	ids := make([]string, 0, len(dtos))
	for _, v := range dtos {
		group_id := aws.StringValue(v.GroupId)
		ids = append(ids, group_id)
	}
	return ids
}
