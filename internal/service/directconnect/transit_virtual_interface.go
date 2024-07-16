// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_transit_virtual_interface", name="Transit Virtual Interface")
// @Tags(identifierAttribute="arn")
func ResourceTransitVirtualInterface() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitVirtualInterfaceCreate,
		ReadWithoutTimeout:   resourceTransitVirtualInterfaceRead,
		UpdateWithoutTimeout: resourceTransitVirtualInterfaceUpdate,
		DeleteWithoutTimeout: resourceTransitVirtualInterfaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTransitVirtualInterfaceImport,
		},

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					directconnect.AddressFamilyIpv4,
					directconnect.AddressFamilyIpv6,
				}, false),
			},
			"amazon_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"amazon_side_asn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_asn": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"bgp_auth_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrConnectionID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"customer_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"mtu": {
				Type:         schema.TypeInt,
				Default:      1500,
				Optional:     true,
				ValidateFunc: validation.IntInSlice([]int{1500, 8500}),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"sitelink_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vlan": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTransitVirtualInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	req := &directconnect.CreateTransitVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get(names.AttrConnectionID).(string)),
		NewTransitVirtualInterface: &directconnect.NewTransitVirtualInterface{
			AddressFamily:          aws.String(d.Get("address_family").(string)),
			Asn:                    aws.Int64(int64(d.Get("bgp_asn").(int))),
			DirectConnectGatewayId: aws.String(d.Get("dx_gateway_id").(string)),
			EnableSiteLink:         aws.Bool(d.Get("sitelink_enabled").(bool)),
			Mtu:                    aws.Int64(int64(d.Get("mtu").(int))),
			Tags:                   getTagsIn(ctx),
			VirtualInterfaceName:   aws.String(d.Get(names.AttrName).(string)),
			Vlan:                   aws.Int64(int64(d.Get("vlan").(int))),
		},
	}
	if v, ok := d.GetOk("amazon_address"); ok {
		req.NewTransitVirtualInterface.AmazonAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("bgp_auth_key"); ok {
		req.NewTransitVirtualInterface.AuthKey = aws.String(v.(string))
	}
	if v, ok := d.GetOk("customer_address"); ok {
		req.NewTransitVirtualInterface.CustomerAddress = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Direct Connect transit virtual interface: %s", req)
	resp, err := conn.CreateTransitVirtualInterfaceWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect transit virtual interface: %s", err)
	}

	d.SetId(aws.StringValue(resp.VirtualInterface.VirtualInterfaceId))

	if err := transitVirtualInterfaceWaitUntilAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceTransitVirtualInterfaceRead(ctx, d, meta)...)
}

func resourceTransitVirtualInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	vif, err := virtualInterfaceRead(ctx, d.Id(), conn)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	if vif == nil {
		log.Printf("[WARN] Direct Connect transit virtual interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("address_family", vif.AddressFamily)
	d.Set("amazon_address", vif.AmazonAddress)
	d.Set("amazon_side_asn", strconv.FormatInt(aws.Int64Value(vif.AmazonSideAsn), 10))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "directconnect",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("aws_device", vif.AwsDeviceV2)
	d.Set("bgp_asn", vif.Asn)
	d.Set("bgp_auth_key", vif.AuthKey)
	d.Set(names.AttrConnectionID, vif.ConnectionId)
	d.Set("customer_address", vif.CustomerAddress)
	d.Set("dx_gateway_id", vif.DirectConnectGatewayId)
	d.Set("jumbo_frame_capable", vif.JumboFrameCapable)
	d.Set("mtu", vif.Mtu)
	d.Set(names.AttrName, vif.VirtualInterfaceName)
	d.Set("sitelink_enabled", vif.SiteLinkEnabled)
	d.Set("vlan", vif.Vlan)

	return diags
}

func resourceTransitVirtualInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	diags = append(diags, virtualInterfaceUpdate(ctx, d, meta)...)
	if diags.HasError() {
		return diags
	}

	if err := transitVirtualInterfaceWaitUntilAvailable(ctx, meta.(*conns.AWSClient).DirectConnectConn(ctx), d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceTransitVirtualInterfaceRead(ctx, d, meta)...)
}

func resourceTransitVirtualInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return virtualInterfaceDelete(ctx, d, meta)
}

func resourceTransitVirtualInterfaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	vif, err := virtualInterfaceRead(ctx, d.Id(), conn)
	if err != nil {
		return nil, err
	}
	if vif == nil {
		return nil, fmt.Errorf("virtual interface (%s) not found", d.Id())
	}

	if vifType := aws.StringValue(vif.VirtualInterfaceType); vifType != "transit" {
		return nil, fmt.Errorf("virtual interface (%s) has incorrect type: %s", d.Id(), vifType)
	}

	return []*schema.ResourceData{d}, nil
}

func transitVirtualInterfaceWaitUntilAvailable(ctx context.Context, conn *directconnect.DirectConnect, vifId string, timeout time.Duration) error {
	return virtualInterfaceWaitUntilAvailable(ctx, conn,
		vifId,
		timeout,
		[]string{
			directconnect.VirtualInterfaceStatePending,
		},
		[]string{
			directconnect.VirtualInterfaceStateAvailable,
			directconnect.VirtualInterfaceStateDown,
		})
}
