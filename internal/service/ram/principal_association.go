// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
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

func resourcePrincipalAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	resourceShareARN, principal := d.Get("resource_share_arn").(string), d.Get(names.AttrPrincipal).(string)
	id, err := flex.FlattenResourceId([]string{resourceShareARN, principal}, principalAssociationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findPrincipalAssociationByTwoPartKey(ctx, conn, resourceShareARN, principal)

	switch {
	case err == nil:
		return sdkdiag.AppendFromErr(diags, fmt.Errorf("RAM Principal Association (%s) already exists", id))
	case retry.NotFound(err):
		break
	default:
		return sdkdiag.AppendErrorf(diags, "reading RAM Principal Association: %s", err)
	}

	if err := createResourceSharePrincipalAssociation(ctx, conn, resourceShareARN, principal); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(id)

	return append(diags, resourcePrincipalAssociationRead(ctx, d, meta)...)
}

func resourcePrincipalAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), principalAssociationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	resourceShareARN, principal := parts[0], parts[1]

	principalAssociation, err := findPrincipalAssociationByTwoPartKey(ctx, conn, resourceShareARN, principal)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] RAM Principal Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Principal Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrPrincipal, principalAssociation.AssociatedEntity)
	d.Set("resource_share_arn", principalAssociation.ResourceShareArn)

	return diags
}

func resourcePrincipalAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), principalAssociationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	resourceShareARN, principal := parts[0], parts[1]

	log.Printf("[DEBUG] Deleting RAM Principal Association: %s", d.Id())
	if err := deleteResourceSharePrincipalAssociation(ctx, conn, resourceShareARN, principal); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func createResourceSharePrincipalAssociation(ctx context.Context, conn *ram.Client, resourceShareARN, principal string, sources ...string) error {
	input := ram.AssociateResourceShareInput{
		ClientToken:      aws.String(sdkid.UniqueId()),
		Principals:       []string{principal},
		ResourceShareArn: aws.String(resourceShareARN),
	}
	if len(sources) > 0 {
		input.Sources = sources
	}
	_, err := conn.AssociateResourceShare(ctx, &input)

	if err != nil {
		return fmt.Errorf("creating RAM Resource Share (%s) Principal (%s) Association: %w", resourceShareARN, principal, err)
	}

	// AWS Account ID principals need to be accepted to become ASSOCIATED.
	if inttypes.IsAWSAccountID(principal) {
		return nil
	}

	if _, err := waitPrincipalAssociationCreated(ctx, conn, resourceShareARN, principal); err != nil {
		return fmt.Errorf("waiting for RAM Resource Share (%s) Principal (%s) Association create: %w", resourceShareARN, principal, err)
	}

	return nil
}

func deleteResourceSharePrincipalAssociation(ctx context.Context, conn *ram.Client, resourceShareARN, principal string, sources ...string) error {
	input := ram.DisassociateResourceShareInput{
		ClientToken:      aws.String(sdkid.UniqueId()),
		Principals:       []string{principal},
		ResourceShareArn: aws.String(resourceShareARN),
	}
	if len(sources) > 0 {
		input.Sources = sources
	}
	_, err := conn.DisassociateResourceShare(ctx, &input)

	if errs.IsA[*awstypes.UnknownResourceException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting RAM Resource Share (%s) Principal (%s) Association: %w", resourceShareARN, principal, err)
	}

	if _, err := waitPrincipalAssociationDeleted(ctx, conn, resourceShareARN, principal); err != nil {
		return fmt.Errorf("waiting for RAM Resource Share (%s) Principal (%s) Association delete: %w", resourceShareARN, principal, err)
	}

	return nil
}

func findPrincipalAssociationByTwoPartKey(ctx context.Context, conn *ram.Client, resourceShareARN, principal string) (*awstypes.ResourceShareAssociation, error) {
	input := ram.GetResourceShareAssociationsInput{
		AssociationType:   awstypes.ResourceShareAssociationTypePrincipal,
		Principal:         aws.String(principal),
		ResourceShareArns: []string{resourceShareARN},
	}

	output, err := findResourceShareAssociation(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.ResourceShareAssociationStatusDisassociated {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, err
}

func statusPrincipalAssociation(conn *ram.Client, resourceShareARN, principal string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findPrincipalAssociationByTwoPartKey(ctx, conn, resourceShareARN, principal)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitPrincipalAssociationCreated(ctx context.Context, conn *ram.Client, resourceShareARN, principal string) (*awstypes.ResourceShareAssociation, error) { //nolint:unparam
	const (
		timeout = 3 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.ResourceShareAssociationStatusAssociating),
		Target:         enum.Slice(awstypes.ResourceShareAssociationStatusAssociated),
		Refresh:        statusPrincipalAssociation(conn, resourceShareARN, principal),
		Timeout:        timeout,
		NotFoundChecks: 20,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceShareAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitPrincipalAssociationDeleted(ctx context.Context, conn *ram.Client, resourceShareARN, principal string) (*awstypes.ResourceShareAssociation, error) { //nolint:unparam
	const (
		timeout = 3 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceShareAssociationStatusAssociated, awstypes.ResourceShareAssociationStatusDisassociating),
		Target:  []string{},
		Refresh: statusPrincipalAssociation(conn, resourceShareARN, principal),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceShareAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
