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

// @SDKResource("aws_dx_private_virtual_interface", name="Private Virtual Interface")
// @Tags(identifierAttribute="arn")
func ResourcePrivateVirtualInterface() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePrivateVirtualInterfaceCreate,
		ReadWithoutTimeout:   resourcePrivateVirtualInterfaceRead,
		UpdateWithoutTimeout: resourcePrivateVirtualInterfaceUpdate,
		DeleteWithoutTimeout: resourcePrivateVirtualInterfaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePrivateVirtualInterfaceImport,
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
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"vpn_gateway_id"},
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"mtu": {
				Type:         schema.TypeInt,
				Default:      1500,
				Optional:     true,
				ValidateFunc: validation.IntInSlice([]int{1500, 9001}),
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
			"vpn_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"dx_gateway_id"},
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

func resourcePrivateVirtualInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	vgwIdRaw, vgwOk := d.GetOk("vpn_gateway_id")
	dxgwIdRaw, dxgwOk := d.GetOk("dx_gateway_id")
	if vgwOk == dxgwOk {
		return sdkdiag.AppendErrorf(diags, "One of ['vpn_gateway_id', 'dx_gateway_id'] must be set to create a Direct Connect private virtual interface")
	}

	req := &directconnect.CreatePrivateVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get(names.AttrConnectionID).(string)),
		NewPrivateVirtualInterface: &directconnect.NewPrivateVirtualInterface{
			AddressFamily:        aws.String(d.Get("address_family").(string)),
			Asn:                  aws.Int64(int64(d.Get("bgp_asn").(int))),
			EnableSiteLink:       aws.Bool(d.Get("sitelink_enabled").(bool)),
			Mtu:                  aws.Int64(int64(d.Get("mtu").(int))),
			Tags:                 getTagsIn(ctx),
			VirtualInterfaceName: aws.String(d.Get(names.AttrName).(string)),
			Vlan:                 aws.Int64(int64(d.Get("vlan").(int))),
		},
	}
	if vgwOk && vgwIdRaw.(string) != "" {
		req.NewPrivateVirtualInterface.VirtualGatewayId = aws.String(vgwIdRaw.(string))
	}
	if dxgwOk && dxgwIdRaw.(string) != "" {
		req.NewPrivateVirtualInterface.DirectConnectGatewayId = aws.String(dxgwIdRaw.(string))
	}
	if v, ok := d.GetOk("amazon_address"); ok {
		req.NewPrivateVirtualInterface.AmazonAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("bgp_auth_key"); ok {
		req.NewPrivateVirtualInterface.AuthKey = aws.String(v.(string))
	}
	if v, ok := d.GetOk("customer_address"); ok {
		req.NewPrivateVirtualInterface.CustomerAddress = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Direct Connect private virtual interface: %s", req)
	resp, err := conn.CreatePrivateVirtualInterfaceWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect private virtual interface: %s", err)
	}

	d.SetId(aws.StringValue(resp.VirtualInterfaceId))

	if err := privateVirtualInterfaceWaitUntilAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourcePrivateVirtualInterfaceRead(ctx, d, meta)...)
}

func resourcePrivateVirtualInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	vif, err := virtualInterfaceRead(ctx, d.Id(), conn)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	if vif == nil {
		log.Printf("[WARN] Direct Connect private virtual interface (%s) not found, removing from state", d.Id())
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
	d.Set("vpn_gateway_id", vif.VirtualGatewayId)

	return diags
}

func resourcePrivateVirtualInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	diags = append(diags, virtualInterfaceUpdate(ctx, d, meta)...)
	if diags.HasError() {
		return diags
	}

	if err := privateVirtualInterfaceWaitUntilAvailable(ctx, meta.(*conns.AWSClient).DirectConnectConn(ctx), d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourcePrivateVirtualInterfaceRead(ctx, d, meta)...)
}

func resourcePrivateVirtualInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return virtualInterfaceDelete(ctx, d, meta)
}

func resourcePrivateVirtualInterfaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	vif, err := virtualInterfaceRead(ctx, d.Id(), conn)
	if err != nil {
		return nil, err
	}
	if vif == nil {
		return nil, fmt.Errorf("virtual interface (%s) not found", d.Id())
	}

	if vifType := aws.StringValue(vif.VirtualInterfaceType); vifType != "private" {
		return nil, fmt.Errorf("virtual interface (%s) has incorrect type: %s", d.Id(), vifType)
	}

	return []*schema.ResourceData{d}, nil
}

func privateVirtualInterfaceWaitUntilAvailable(ctx context.Context, conn *directconnect.DirectConnect, vifId string, timeout time.Duration) error {
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
