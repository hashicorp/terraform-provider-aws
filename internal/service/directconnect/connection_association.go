// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_dx_connection_association")
func ResourceConnectionAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionAssociationCreate,
		ReadWithoutTimeout:   resourceConnectionAssociationRead,
		DeleteWithoutTimeout: resourceConnectionAssociationDelete,

		Schema: map[string]*schema.Schema{
			"connection_id": {
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

func resourceConnectionAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	connectionID := d.Get("connection_id").(string)
	lagID := d.Get("lag_id").(string)
	input := &directconnect.AssociateConnectionWithLagInput{
		ConnectionId: aws.String(connectionID),
		LagId:        aws.String(lagID),
	}

	log.Printf("[DEBUG] Creating Direct Connect Connection LAG Association: %#v", input)
	output, err := conn.AssociateConnectionWithLag(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect Connection (%s) LAG (%s) Association: %s", connectionID, lagID, err)
	}

	d.SetId(aws.ToString(output.ConnectionId))

	return diags
}

func resourceConnectionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	lagID := d.Get("lag_id").(string)
	err := FindConnectionAssociationExists(ctx, conn, d.Id(), lagID)

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

func resourceConnectionAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	_, err := tfresource.RetryWhen(ctx, connectionDisassociatedTimeout,
		func() (interface{}, error) {
			return conn.DisassociateConnectionFromLag(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Connection does not exist") ||
				errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Lag does not exist") {
				return false, nil
			}

			if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "is in a transitioning state") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return fmt.Errorf("deleting Direct Connect Connection (%s) LAG (%s) Association: %w", connectionID, lagID, err)
	}

	return err
}
