// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_connection_confirmation")
func ResourceConnectionConfirmation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionConfirmationCreate,
		ReadWithoutTimeout:   resourceConnectionConfirmationRead,
		DeleteWithoutTimeout: resourceConnectionConfirmationDelete,

		Schema: map[string]*schema.Schema{
			names.AttrConnectionID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceConnectionConfirmationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	connectionID := d.Get(names.AttrConnectionID).(string)
	input := &directconnect.ConfirmConnectionInput{
		ConnectionId: aws.String(connectionID),
	}

	log.Printf("[DEBUG] Confirming Direct Connect Connection: %s", input)
	_, err := conn.ConfirmConnectionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "confirming Direct Connection Connection (%s): %s", connectionID, err)
	}

	d.SetId(connectionID)

	if _, err := waitConnectionConfirmed(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connection Connection (%s) confirm: %s", d.Id(), err)
	}

	return diags
}

func resourceConnectionConfirmationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn(ctx)

	_, err := FindConnectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Connection (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceConnectionConfirmationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[WARN] Will not delete Direct Connect connection. Terraform will remove this resource from the state file, however resources may remain.")
	return diags
}
