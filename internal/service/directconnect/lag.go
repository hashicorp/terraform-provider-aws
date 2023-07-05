// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_lag", name="LAG")
// @Tags(identifierAttribute="arn")
func ResourceLag() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLagCreate,
		ReadWithoutTimeout:   resourceLagRead,
		UpdateWithoutTimeout: resourceLagUpdate,
		DeleteWithoutTimeout: resourceLagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"connections_bandwidth": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validConnectionBandWidth(),
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"has_logical_redundancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	name := d.Get("name").(string)
	input := &directconnect.CreateLagInput{
		ConnectionsBandwidth: aws.String(d.Get("connections_bandwidth").(string)),
		LagName:              aws.String(name),
		Location:             aws.String(d.Get("location").(string)),
		Tags:                 getTagsIn(ctx),
	}

	var connectionIDSpecified bool
	if v, ok := d.GetOk("connection_id"); ok {
		connectionIDSpecified = true
		input.ConnectionId = aws.String(v.(string))
		input.NumberOfConnections = aws.Int64(1)
	} else {
		input.NumberOfConnections = aws.Int64(1)
	}

	if v, ok := d.GetOk("provider_name"); ok {
		input.ProviderName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Direct Connect LAG: %s", input)
	output, err := conn.CreateLagWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect LAG (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.LagId))

	// Delete unmanaged connection.
	if !connectionIDSpecified {
		if err := deleteConnection(ctx, conn, aws.StringValue(output.Connections[0].ConnectionId), waitConnectionDeleted); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceLagRead(ctx, d, meta)...)
}

func resourceLagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	lag, err := FindLagByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect LAG (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect LAG (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    aws.StringValue(lag.Region),
		Service:   "directconnect",
		AccountID: aws.StringValue(lag.OwnerAccount),
		Resource:  fmt.Sprintf("dxlag/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("connections_bandwidth", lag.ConnectionsBandwidth)
	d.Set("has_logical_redundancy", lag.HasLogicalRedundancy)
	d.Set("jumbo_frame_capable", lag.JumboFrameCapable)
	d.Set("location", lag.Location)
	d.Set("name", lag.LagName)
	d.Set("owner_account_id", lag.OwnerAccount)
	d.Set("provider_name", lag.ProviderName)

	return diags
}

func resourceLagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	if d.HasChange("name") {
		input := &directconnect.UpdateLagInput{
			LagId:   aws.String(d.Id()),
			LagName: aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating Direct Connect LAG: %s", input)
		_, err := conn.UpdateLagWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Direct Connect LAG (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLagRead(ctx, d, meta)...)
}

func resourceLagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	if d.Get("force_destroy").(bool) {
		lag, err := FindLagByID(ctx, conn, d.Id())

		if tfresource.NotFound(err) {
			return diags
		}

		for _, connection := range lag.Connections {
			if err := deleteConnection(ctx, conn, aws.StringValue(connection.ConnectionId), waitConnectionDeleted); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	} else if v, ok := d.GetOk("connection_id"); ok {
		if err := deleteConnectionLAGAssociation(ctx, conn, v.(string), d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[DEBUG] Deleting Direct Connect LAG: %s", d.Id())
	_, err := conn.DeleteLagWithContext(ctx, &directconnect.DeleteLagInput{
		LagId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Lag with ID") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect LAG (%s): %s", d.Id(), err)
	}

	_, err = waitLagDeleted(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect LAG (%s) delete: %s", d.Id(), err)
	}

	return diags
}
