// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			names.AttrPrincipal: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringIsNotEmpty,
					validation.Any(
						verify.ValidAccountID,
						verify.ValidARN,
					),
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
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	resourceShareARN, principal := d.Get("resource_share_arn").(string), d.Get(names.AttrPrincipal).(string)
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
		Principals:       []string{principal},
		ResourceShareArn: aws.String(resourceShareARN),
	}

	_, err = conn.AssociateResourceShare(ctx, input)

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
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

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

	d.Set(names.AttrPrincipal, principalAssociation.AssociatedEntity)
	d.Set("resource_share_arn", principalAssociation.ResourceShareArn)

	return diags
}

func resourcePrincipalAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), principalAssociationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	resourceShareARN, principal := parts[0], parts[1]

	log.Printf("[DEBUG] Deleting RAM Principal Association: %s", d.Id())
	_, err = conn.DisassociateResourceShare(ctx, &ram.DisassociateResourceShareInput{
		Principals:       []string{principal},
		ResourceShareArn: aws.String(resourceShareARN),
	})

	if errs.IsA[*awstypes.UnknownResourceException](err) {
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

func findPrincipalAssociationByTwoPartKey(ctx context.Context, conn *ram.Client, resourceShareARN, principal string) (*awstypes.ResourceShareAssociation, error) {
	input := &ram.GetResourceShareAssociationsInput{
		AssociationType:   awstypes.ResourceShareAssociationTypePrincipal,
		Principal:         aws.String(principal),
		ResourceShareArns: []string{resourceShareARN},
	}

	output, err := findResourceShareAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.ResourceShareAssociationStatusDisassociated {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, err
}

func statusPrincipalAssociation(ctx context.Context, conn *ram.Client, resourceShareARN, principal string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findPrincipalAssociationByTwoPartKey(ctx, conn, resourceShareARN, principal)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitPrincipalAssociationCreated(ctx context.Context, conn *ram.Client, resourceShareARN, principal string) (*awstypes.ResourceShareAssociation, error) {
	const (
		timeout = 3 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.ResourceShareAssociationStatusAssociating),
		Target:         enum.Slice(awstypes.ResourceShareAssociationStatusAssociated),
		Refresh:        statusPrincipalAssociation(ctx, conn, resourceShareARN, principal),
		Timeout:        timeout,
		NotFoundChecks: 20,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceShareAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitPrincipalAssociationDeleted(ctx context.Context, conn *ram.Client, resourceShareARN, principal string) (*awstypes.ResourceShareAssociation, error) {
	const (
		timeout = 3 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceShareAssociationStatusAssociated, awstypes.ResourceShareAssociationStatusDisassociating),
		Target:  []string{},
		Refresh: statusPrincipalAssociation(ctx, conn, resourceShareARN, principal),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceShareAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
