// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_connection_confirmation", name="Connection Confirmation")
func resourceConnectionConfirmation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionConfirmationCreate,
		ReadWithoutTimeout:   resourceConnectionConfirmationRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			names.AttrConnectionID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceConnectionConfirmationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	connectionID := d.Get(names.AttrConnectionID).(string)
	input := &directconnect.ConfirmConnectionInput{
		ConnectionId: aws.String(connectionID),
	}

	_, err := conn.ConfirmConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "confirming Direct Connect Connection (%s): %s", connectionID, err)
	}

	d.SetId(connectionID)

	if _, err := waitConnectionConfirmed(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Connection (%s) confirm: %s", d.Id(), err)
	}

	return diags
}

func resourceConnectionConfirmationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	_, err := findConnectionByID(ctx, conn, d.Id())

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

func waitConnectionConfirmed(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Connection, error) { //nolint:unparam
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionStatePending, awstypes.ConnectionStateOrdering, awstypes.ConnectionStateRequested),
		Target:  enum.Slice(awstypes.ConnectionStateAvailable),
		Refresh: statusConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Connection); ok {
		return output, err
	}

	return nil, err
}
