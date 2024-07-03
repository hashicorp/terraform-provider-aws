// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_hosted_connection")
func ResourceHostedConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostedConnectionCreate,
		ReadWithoutTimeout:   resourceHostedConnectionRead,
		DeleteWithoutTimeout: resourceHostedConnectionDelete,

		Schema: map[string]*schema.Schema{
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validConnectionBandWidth(),
			},
			names.AttrConnectionID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"has_logical_redundancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"lag_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"loa_issue_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLocation: {
				Type:     schema.TypeString,
				Computed: true,
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
			"partner_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrProviderName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRegion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vlan": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
			},
		},
	}
}

func resourceHostedConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &directconnect.AllocateHostedConnectionInput{
		Bandwidth:      aws.String(d.Get("bandwidth").(string)),
		ConnectionId:   aws.String(d.Get(names.AttrConnectionID).(string)),
		ConnectionName: aws.String(name),
		OwnerAccount:   aws.String(d.Get(names.AttrOwnerAccountID).(string)),
		Vlan:           aws.Int64(int64(d.Get("vlan").(int))),
	}

	log.Printf("[DEBUG] Creating Direct Connect Hosted Connection: %s", input)
	output, err := conn.AllocateHostedConnectionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Hosted Connection (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ConnectionId))

	return append(diags, resourceHostedConnectionRead(ctx, d, meta)...)
}

func resourceHostedConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	connection, err := FindHostedConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Hosted Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Hosted Connection (%s): %s", d.Id(), err)
	}

	// Cannot set the following attributes from the response:
	// - connection_id: conn.ConnectionId is this resource's ID, not the ID of the interconnect or LAG
	// - tags: conn.Tags seems to always come back empty and DescribeTags needs to be called from the owner account
	d.Set("aws_device", connection.AwsDeviceV2)
	d.Set("has_logical_redundancy", connection.HasLogicalRedundancy)
	d.Set("jumbo_frame_capable", connection.JumboFrameCapable)
	d.Set("lag_id", connection.LagId)
	d.Set("loa_issue_time", aws.TimeValue(connection.LoaIssueTime).Format(time.RFC3339))
	d.Set(names.AttrLocation, connection.Location)
	d.Set(names.AttrName, connection.ConnectionName)
	d.Set(names.AttrOwnerAccountID, connection.OwnerAccount)
	d.Set("partner_name", connection.PartnerName)
	d.Set(names.AttrProviderName, connection.ProviderName)
	d.Set(names.AttrRegion, connection.Region)
	d.Set(names.AttrState, connection.ConnectionState)
	d.Set("vlan", connection.Vlan)

	return diags
}

func resourceHostedConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	if err := deleteConnection(ctx, conn, d.Id(), waitHostedConnectionDeleted); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}
