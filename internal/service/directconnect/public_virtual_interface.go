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

// @SDKResource("aws_dx_public_virtual_interface", name="Public Virtual Interface")
// @Tags(identifierAttribute="arn")
func resourcePublicVirtualInterface() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePublicVirtualInterfaceCreate,
		ReadWithoutTimeout:   resourcePublicVirtualInterfaceRead,
		UpdateWithoutTimeout: resourcePublicVirtualInterfaceUpdate,
		DeleteWithoutTimeout: resourcePublicVirtualInterfaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourcePublicVirtualInterfaceImport,
		},

		CustomizeDiff: resourcePublicVirtualInterfaceCustomizeDiff,

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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route_filter_prefixes": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MinItems: 1,
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
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourcePublicVirtualInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	input := &directconnect.CreatePublicVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get(names.AttrConnectionID).(string)),
		NewPublicVirtualInterface: &awstypes.NewPublicVirtualInterface{
			AddressFamily:        awstypes.AddressFamily(d.Get("address_family").(string)),
			Asn:                  int32(d.Get("bgp_asn").(int)),
			Tags:                 getTagsIn(ctx),
			VirtualInterfaceName: aws.String(d.Get(names.AttrName).(string)),
			Vlan:                 int32(d.Get("vlan").(int)),
		},
	}

	if v, ok := d.GetOk("amazon_address"); ok {
		input.NewPublicVirtualInterface.AmazonAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("bgp_auth_key"); ok {
		input.NewPublicVirtualInterface.AuthKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("customer_address"); ok {
		input.NewPublicVirtualInterface.CustomerAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("route_filter_prefixes"); ok {
		input.NewPublicVirtualInterface.RouteFilterPrefixes = expandRouteFilterPrefixes(v.(*schema.Set).List())
	}

	output, err := conn.CreatePublicVirtualInterface(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Public Virtual Interface: %s", err)
	}

	d.SetId(aws.ToString(output.VirtualInterfaceId))

	if _, err := waitPublicVirtualInterfaceAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Public Virtual Interface (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourcePublicVirtualInterfaceRead(ctx, d, meta)...)
}

func resourcePublicVirtualInterfaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vif, err := findVirtualInterfaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Public Virtual Interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Public Virtual Interface (%s): %s", d.Id(), err)
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
	d.Set("customer_address", vif.CustomerAddress)
	d.Set(names.AttrConnectionID, vif.ConnectionId)
	d.Set(names.AttrName, vif.VirtualInterfaceName)
	if err := d.Set("route_filter_prefixes", flattenRouteFilterPrefixes(vif.RouteFilterPrefixes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting route_filter_prefixes: %s", err)
	}
	d.Set("vlan", vif.Vlan)

	return diags
}

func resourcePublicVirtualInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	diags = append(diags, virtualInterfaceUpdate(ctx, d, meta)...)
	if diags.HasError() {
		return diags
	}

	return append(diags, resourcePublicVirtualInterfaceRead(ctx, d, meta)...)
}

func resourcePublicVirtualInterfaceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return virtualInterfaceDelete(ctx, d, meta)
}

func resourcePublicVirtualInterfaceImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vif, err := findVirtualInterfaceByID(ctx, conn, d.Id())

	if err != nil {
		return nil, err
	}

	if vifType := aws.ToString(vif.VirtualInterfaceType); vifType != "public" {
		return nil, fmt.Errorf("virtual interface (%s) has incorrect type: %s", d.Id(), vifType)
	}

	return []*schema.ResourceData{d}, nil
}

func resourcePublicVirtualInterfaceCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	if diff.Id() == "" {
		// New resource.
		if addressFamily := diff.Get("address_family").(string); addressFamily == string(awstypes.AddressFamilyIPv4) {
			if _, ok := diff.GetOk("customer_address"); !ok {
				return fmt.Errorf("'customer_address' must be set when 'address_family' is '%s'", addressFamily)
			}
			if _, ok := diff.GetOk("amazon_address"); !ok {
				return fmt.Errorf("'amazon_address' must be set when 'address_family' is '%s'", addressFamily)
			}
		}
	}

	return nil
}

func waitPublicVirtualInterfaceAvailable(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.VirtualInterface, error) {
	return waitVirtualInterfaceAvailable(
		ctx,
		conn,
		id,
		enum.Slice(awstypes.VirtualInterfaceStatePending),
		enum.Slice(awstypes.VirtualInterfaceStateAvailable, awstypes.VirtualInterfaceStateDown, awstypes.VirtualInterfaceStateVerifying),
		timeout,
	)
}
