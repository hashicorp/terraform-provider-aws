// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_appstream_fleet_stack_association")
func ResourceFleetStackAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFleetStackAssociationCreate,
		ReadWithoutTimeout:   resourceFleetStackAssociationRead,
		DeleteWithoutTimeout: resourceFleetStackAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"fleet_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stack_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFleetStackAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)
	input := &appstream.AssociateFleetInput{
		FleetName: aws.String(d.Get("fleet_name").(string)),
		StackName: aws.String(d.Get("stack_name").(string)),
	}

	err := retry.RetryContext(ctx, fleetOperationTimeout, func() *retry.RetryError {
		_, err := conn.AssociateFleet(ctx, input)
		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.AssociateFleet(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream Fleet Stack Association (%s): %s", d.Id(), err)
	}

	d.SetId(EncodeStackFleetID(d.Get("fleet_name").(string), d.Get("stack_name").(string)))

	return append(diags, resourceFleetStackAssociationRead(ctx, d, meta)...)
}

func resourceFleetStackAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	fleetName, stackName, err := DecodeStackFleetID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding AppStream Fleet Stack Association ID (%s): %s", d.Id(), err)
	}

	err = FindFleetStackAssociation(ctx, conn, fleetName, stackName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppStream Fleet Stack Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream Fleet Stack Association (%s): %s", d.Id(), err)
	}

	d.Set("fleet_name", fleetName)
	d.Set("stack_name", stackName)

	return diags
}

func resourceFleetStackAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	fleetName, stackName, err := DecodeStackFleetID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding AppStream Fleet Stack Association ID (%s): %s", d.Id(), err)
	}

	_, err = conn.DisassociateFleet(ctx, &appstream.DisassociateFleetInput{
		StackName: aws.String(stackName),
		FleetName: aws.String(fleetName),
	})

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting AppStream Fleet Stack Association (%s): %s", d.Id(), err)
	}
	return diags
}

func EncodeStackFleetID(fleetName, stackName string) string {
	return fmt.Sprintf("%s/%s", fleetName, stackName)
}

func DecodeStackFleetID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format FleetName/StackName, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
