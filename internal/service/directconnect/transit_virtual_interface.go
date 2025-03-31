// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_transit_virtual_interface", name="Transit Virtual Interface")
// @Tags(identifierAttribute="arn")
func resourceTransitVirtualInterface() *schema.Resource {
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AddressFamily](),
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
	}
}

func resourceTransitVirtualInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	input := &directconnect.CreateTransitVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get(names.AttrConnectionID).(string)),
		NewTransitVirtualInterface: &awstypes.NewTransitVirtualInterface{
			AddressFamily:          awstypes.AddressFamily(d.Get("address_family").(string)),
			Asn:                    int32(d.Get("bgp_asn").(int)),
			DirectConnectGatewayId: aws.String(d.Get("dx_gateway_id").(string)),
			EnableSiteLink:         aws.Bool(d.Get("sitelink_enabled").(bool)),
			Mtu:                    aws.Int32(int32(d.Get("mtu").(int))),
			Tags:                   getTagsIn(ctx),
			VirtualInterfaceName:   aws.String(d.Get(names.AttrName).(string)),
			Vlan:                   int32(d.Get("vlan").(int)),
		},
	}

	if v, ok := d.GetOk("amazon_address"); ok {
		input.NewTransitVirtualInterface.AmazonAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("bgp_auth_key"); ok {
		input.NewTransitVirtualInterface.AuthKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("customer_address"); ok {
		input.NewTransitVirtualInterface.CustomerAddress = aws.String(v.(string))
	}

	output, err := conn.CreateTransitVirtualInterface(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Transit Virtual Interface: %s", err)
	}

	d.SetId(aws.ToString(output.VirtualInterface.VirtualInterfaceId))

	if _, err := waitTransitVirtualInterfaceAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Transit Virtual Interface (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTransitVirtualInterfaceRead(ctx, d, meta)...)
}

func resourceTransitVirtualInterfaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vif, err := findVirtualInterfaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Transit Virtual Interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Transit Virtual Interface (%s): %s", d.Id(), err)
	}

	d.Set("address_family", vif.AddressFamily)
	d.Set("amazon_address", vif.AmazonAddress)
	d.Set("amazon_side_asn", strconv.FormatInt(aws.ToInt64(vif.AmazonSideAsn), 10))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    meta.(*conns.AWSClient).Region(ctx),
		Service:   "directconnect",
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
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

func resourceTransitVirtualInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	diags = append(diags, virtualInterfaceUpdate(ctx, d, meta)...)
	if diags.HasError() {
		return diags
	}

	if _, err := waitTransitVirtualInterfaceAvailable(ctx, meta.(*conns.AWSClient).DirectConnectClient(ctx), d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Transit Virtual Interface (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceTransitVirtualInterfaceRead(ctx, d, meta)...)
}

func resourceTransitVirtualInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return virtualInterfaceDelete(ctx, d, meta)
}

func resourceTransitVirtualInterfaceImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vif, err := findVirtualInterfaceByID(ctx, conn, d.Id())

	if err != nil {
		return nil, err
	}

	if vifType := aws.ToString(vif.VirtualInterfaceType); vifType != "transit" {
		return nil, fmt.Errorf("virtual interface (%s) has incorrect type: %s", d.Id(), vifType)
	}

	return []*schema.ResourceData{d}, nil
}

func waitTransitVirtualInterfaceAvailable(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.VirtualInterface, error) {
	return waitVirtualInterfaceAvailable(
		ctx,
		conn,
		id,
		enum.Slice(awstypes.VirtualInterfaceStatePending),
		enum.Slice(awstypes.VirtualInterfaceStateAvailable, awstypes.VirtualInterfaceStateDown),
		timeout,
	)
}
