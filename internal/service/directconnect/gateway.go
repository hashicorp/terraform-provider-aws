// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_gateway")
func ResourceGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGatewayCreate,
		ReadWithoutTimeout:   resourceGatewayRead,
		UpdateWithoutTimeout: resourceGatewayUpdate,
		DeleteWithoutTimeout: resourceGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAmazonSideASN,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &directconnect.CreateDirectConnectGatewayInput{
		DirectConnectGatewayName: aws.String(name),
	}

	if v, ok := d.Get("amazon_side_asn").(string); ok && v != "" {
		v, _ := strconv.ParseInt(v, 10, 64)
		input.AmazonSideAsn = aws.Int64(v)
	}

	output, err := conn.CreateDirectConnectGatewayWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Gateway (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.DirectConnectGateway.DirectConnectGatewayId))

	if _, err := waitGatewayCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGatewayRead(ctx, d, meta)...)
}

func resourceGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	output, err := FindGatewayByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Gateway (%s): %s", d.Id(), err)
	}

	d.Set("amazon_side_asn", strconv.FormatInt(aws.Int64Value(output.AmazonSideAsn), 10))
	d.Set(names.AttrName, output.DirectConnectGatewayName)
	d.Set(names.AttrOwnerAccountID, output.OwnerAccount)

	return diags
}

func resourceGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	if d.HasChange(names.AttrName) {
		input := &directconnect.UpdateDirectConnectGatewayInput{
			DirectConnectGatewayId:      aws.String(d.Id()),
			NewDirectConnectGatewayName: aws.String(d.Get(names.AttrName).(string)),
		}

		_, err := conn.UpdateDirectConnectGatewayWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Direct Connect Gateway (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGatewayRead(ctx, d, meta)...)
}

func resourceGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	log.Printf("[DEBUG] Deleting Direct Connect Gateway: %s", d.Id())
	_, err := conn.DeleteDirectConnectGatewayWithContext(ctx, &directconnect.DeleteDirectConnectGatewayInput{
		DirectConnectGatewayId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect Gateway (%s): %s", d.Id(), err)
	}

	if _, err := waitGatewayDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Gateway (%s) delete: %s", d.Id(), err)
	}

	return diags
}
