// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_connection_association", name="Connection LAG Association")
func resourceConnectionAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionAssociationCreate,
		ReadWithoutTimeout:   resourceConnectionAssociationRead,
		DeleteWithoutTimeout: resourceConnectionAssociationDelete,

		Schema: map[string]*schema.Schema{
			names.AttrConnectionID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"lag_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceConnectionAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	connectionID := d.Get(names.AttrConnectionID).(string)
	lagID := d.Get("lag_id").(string)
	input := &directconnect.AssociateConnectionWithLagInput{
		ConnectionId: aws.String(connectionID),
		LagId:        aws.String(lagID),
	}

	output, err := conn.AssociateConnectionWithLag(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Connection (%s) LAG (%s) Association: %s", connectionID, lagID, err)
	}

	d.SetId(aws.ToString(output.ConnectionId))

	return append(diags, resourceConnectionAssociationRead(ctx, d, meta)...)
}

func resourceConnectionAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	lagID := d.Get("lag_id").(string)
	err := findConnectionLAGAssociation(ctx, conn, d.Id(), lagID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Connection (%s) LAG (%s) Association not found, removing from state", d.Id(), lagID)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Connection (%s) LAG (%s) Association: %s", d.Id(), lagID, err)
	}

	return diags
}

func resourceConnectionAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	if err := deleteConnectionLAGAssociation(ctx, conn, d.Id(), d.Get("lag_id").(string)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func deleteConnectionLAGAssociation(ctx context.Context, conn *directconnect.Client, connectionID, lagID string) error {
	input := &directconnect.DisassociateConnectionFromLagInput{
		ConnectionId: aws.String(connectionID),
		LagId:        aws.String(lagID),
	}

	const (
		timeout = 1 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.DirectConnectClientException](ctx, timeout,
		func() (any, error) {
			return conn.DisassociateConnectionFromLag(ctx, input)
		}, "is in a transitioning state")

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Connection does not exist") ||
		errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Lag does not exist") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Direct Connect Connection (%s) LAG (%s) Association: %w", connectionID, lagID, err)
	}

	return nil
}

func findConnectionLAGAssociation(ctx context.Context, conn *directconnect.Client, connectionID, lagID string) error {
	connection, err := findConnectionByID(ctx, conn, connectionID)

	if err != nil {
		return err
	}

	if lagID != aws.ToString(connection.LagId) {
		return &retry.NotFoundError{}
	}

	return nil
}
