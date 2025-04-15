// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_hosted_transit_virtual_interface", name="Hosted Transit Virtual Interface")
func resourceHostedTransitVirtualInterface() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostedTransitVirtualInterfaceCreate,
		ReadWithoutTimeout:   resourceHostedTransitVirtualInterfaceRead,
		DeleteWithoutTimeout: resourceHostedTransitVirtualInterfaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHostedTransitVirtualInterfaceImport,
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
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"mtu": {
				Type:         schema.TypeInt,
				Default:      1500,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntInSlice([]int{1500, 8500}),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrOwnerAccountID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"vlan": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceHostedTransitVirtualInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	input := &directconnect.AllocateTransitVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get(names.AttrConnectionID).(string)),
		NewTransitVirtualInterfaceAllocation: &awstypes.NewTransitVirtualInterfaceAllocation{
			AddressFamily:        awstypes.AddressFamily(d.Get("address_family").(string)),
			Asn:                  int32(d.Get("bgp_asn").(int)),
			Mtu:                  aws.Int32(int32(d.Get("mtu").(int))),
			VirtualInterfaceName: aws.String(d.Get(names.AttrName).(string)),
			Vlan:                 int32(d.Get("vlan").(int)),
		},
		OwnerAccount: aws.String(d.Get(names.AttrOwnerAccountID).(string)),
	}

	if v, ok := d.GetOk("amazon_address"); ok {
		input.NewTransitVirtualInterfaceAllocation.AmazonAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("bgp_auth_key"); ok {
		input.NewTransitVirtualInterfaceAllocation.AuthKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("customer_address"); ok {
		input.NewTransitVirtualInterfaceAllocation.CustomerAddress = aws.String(v.(string))
	}

	output, err := conn.AllocateTransitVirtualInterface(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Hosted Transit Virtual Interface: %s", err)
	}

	d.SetId(aws.ToString(output.VirtualInterface.VirtualInterfaceId))

	if _, err := waitHostedTransitVirtualInterfaceAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Hosted Transit Virtual Interface (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceHostedTransitVirtualInterfaceRead(ctx, d, meta)...)
}

func resourceHostedTransitVirtualInterfaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vif, err := findVirtualInterfaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Hosted Transit Virtual Interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Hosted Transit Virtual Interface (%s): %s", d.Id(), err)
	}

	d.Set("address_family", vif.AddressFamily)
	d.Set("amazon_address", vif.AmazonAddress)
	d.Set("amazon_side_asn", flex.Int64ToStringValue(vif.AmazonSideAsn))
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
	d.Set("jumbo_frame_capable", vif.JumboFrameCapable)
	d.Set("mtu", vif.Mtu)
	d.Set(names.AttrName, vif.VirtualInterfaceName)
	d.Set(names.AttrOwnerAccountID, vif.OwnerAccount)
	d.Set("vlan", vif.Vlan)

	return diags
}

func resourceHostedTransitVirtualInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return virtualInterfaceDelete(ctx, d, meta)
}

func resourceHostedTransitVirtualInterfaceImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
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

func waitHostedTransitVirtualInterfaceAvailable(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.VirtualInterface, error) {
	return waitVirtualInterfaceAvailable(
		ctx,
		conn,
		id,
		enum.Slice(awstypes.VirtualInterfaceStatePending),
		enum.Slice(awstypes.VirtualInterfaceStateAvailable, awstypes.VirtualInterfaceStateConfirming, awstypes.VirtualInterfaceStateDown),
		timeout,
	)
}
