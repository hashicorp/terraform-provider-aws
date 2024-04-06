// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ram_principal_association", name="Principal Association")
func resourcePrincipalAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePrincipalAssociationCreate,
		ReadWithoutTimeout:   resourcePrincipalAssociationRead,
		DeleteWithoutTimeout: resourcePrincipalAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					verify.ValidAccountID,
					verify.ValidARN,
				),
			},
			"resource_share_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

const (
	principalAssociationResourceIDPartCount = 2
)

func resourcePrincipalAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	resourceShareARN, principal := d.Get("resource_share_arn").(string), d.Get("principal").(string)
	id := errs.Must(flex.FlattenResourceId([]string{resourceShareARN, principal}, principalAssociationResourceIDPartCount, false))
	_, err := findPrincipalAssociationByTwoPartKey(ctx, conn, resourceShareARN, principal)

	switch {
	case err == nil:
		return sdkdiag.AppendFromErr(diags, fmt.Errorf("RAM Principal Association (%s) already exists", id))
	case tfresource.NotFound(err):
		break
	default:
		return sdkdiag.AppendErrorf(diags, "reading RAM Principal Association: %s", err)
	}

	input := &ram.AssociateResourceShareInput{
		ClientToken:      aws.String(sdkid.UniqueId()),
		Principals:       []*string{aws.String(principal)},
		ResourceShareArn: aws.String(resourceShareARN),
	}

	_, err = conn.AssociateResourceShareWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RAM Principal Association (%s): %s", id, err)
	}

	d.SetId(id)

	// AWS Account ID principals need to be accepted to become ASSOCIATED.
	if itypes.IsAWSAccountID(principal) {
		return append(diags, resourcePrincipalAssociationRead(ctx, d, meta)...)
	}

	if _, err := waitPrincipalAssociationCreated(ctx, conn, resourceShareARN, principal); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM Principal Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourcePrincipalAssociationRead(ctx, d, meta)...)
}

func resourcePrincipalAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), principalAssociationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	resourceShareARN, principal := parts[0], parts[1]

	principalAssociation, err := findPrincipalAssociationByTwoPartKey(ctx, conn, resourceShareARN, principal)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RAM Resource Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Association (%s): %s", d.Id(), err)
	}

	d.Set("principal", principalAssociation.AssociatedEntity)
	d.Set("resource_share_arn", principalAssociation.ResourceShareArn)

	return diags
}

func resourcePrincipalAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), principalAssociationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	resourceShareARN, principal := parts[0], parts[1]

	log.Printf("[DEBUG] Deleting RAM Principal Association: %s", d.Id())
	_, err = conn.DisassociateResourceShareWithContext(ctx, &ram.DisassociateResourceShareInput{
		Principals:       []*string{aws.String(principal)},
		ResourceShareArn: aws.String(resourceShareARN),
	})

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RAM Principal Association (%s): %s", d.Id(), err)
	}

	if _, err := waitPrincipalAssociationDeleted(ctx, conn, resourceShareARN, principal); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM Principal Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findPrincipalAssociationByTwoPartKey(ctx context.Context, conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	input := &ram.GetResourceShareAssociationsInput{
		AssociationType:   aws.String(ram.ResourceShareAssociationTypePrincipal),
		Principal:         aws.String(principal),
		ResourceShareArns: aws.StringSlice([]string{resourceShareARN}),
	}

	output, err := findResourceShareAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status); status == ram.ResourceShareAssociationStatusDisassociated {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, err
}

func statusPrincipalAssociation(ctx context.Context, conn *ram.RAM, resourceShareARN, principal string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findPrincipalAssociationByTwoPartKey(ctx, conn, resourceShareARN, principal)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitPrincipalAssociationCreated(ctx context.Context, conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	const (
		timeout = 3 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:        []string{ram.ResourceShareAssociationStatusAssociating},
		Target:         []string{ram.ResourceShareAssociationStatusAssociated},
		Refresh:        statusPrincipalAssociation(ctx, conn, resourceShareARN, principal),
		Timeout:        timeout,
		NotFoundChecks: 20,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ram.ResourceShareAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitPrincipalAssociationDeleted(ctx context.Context, conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	const (
		timeout = 3 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated, ram.ResourceShareAssociationStatusDisassociating},
		Target:  []string{},
		Refresh: statusPrincipalAssociation(ctx, conn, resourceShareARN, principal),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ram.ResourceShareAssociation); ok {
		return v, err
	}

	return nil, err
}
