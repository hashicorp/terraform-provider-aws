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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ram_resource_association", name="Resource Association")
func resourceResourceAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceAssociationCreate,
		ReadWithoutTimeout:   resourceResourceAssociationRead,
		DeleteWithoutTimeout: resourceResourceAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrResourceARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
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
	resourceAssociationResourceIDPartCount = 2
)

func resourceResourceAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	resourceShareARN, resourceARN := d.Get("resource_share_arn").(string), d.Get(names.AttrResourceARN).(string)
	id, err := flex.FlattenResourceId([]string{resourceShareARN, resourceARN}, resourceAssociationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findResourceAssociationByTwoPartKey(ctx, conn, resourceShareARN, resourceARN)

	switch {
	case err == nil:
		return sdkdiag.AppendFromErr(diags, fmt.Errorf("RAM Resource Association (%s) already exists", id))
	case retry.NotFound(err):
		break
	default:
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Association: %s", err)
	}

	if err := createResourceShareResourceAssociation(ctx, conn, resourceShareARN, resourceARN); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(id)

	return append(diags, resourceResourceAssociationRead(ctx, d, meta)...)
}

func resourceResourceAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), resourceAssociationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	resourceShareARN, resourceARN := parts[0], parts[1]

	resourceAssociation, err := findResourceAssociationByTwoPartKey(ctx, conn, resourceShareARN, resourceARN)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] RAM Resource Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrResourceARN, resourceAssociation.AssociatedEntity)
	d.Set("resource_share_arn", resourceAssociation.ResourceShareArn)

	return diags
}

func resourceResourceAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), resourceAssociationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	resourceShareARN, resourceARN := parts[0], parts[1]

	log.Printf("[DEBUG] Deleting RAM Resource Association: %s", d.Id())
	if err := deleteResourceShareResourceAssociation(ctx, conn, resourceShareARN, resourceARN); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func createResourceShareResourceAssociation(ctx context.Context, conn *ram.Client, resourceShareARN, resourceARN string) error {
	input := ram.AssociateResourceShareInput{
		ClientToken:      aws.String(sdkid.UniqueId()),
		ResourceArns:     []string{resourceARN},
		ResourceShareArn: aws.String(resourceShareARN),
	}
	_, err := conn.AssociateResourceShare(ctx, &input)

	if err != nil {
		return fmt.Errorf("creating RAM Resource Share (%s) Resource (%s) Association: %w", resourceShareARN, resourceARN, err)
	}

	if _, err := waitResourceAssociationCreated(ctx, conn, resourceShareARN, resourceARN); err != nil {
		return fmt.Errorf("waiting for RAM Resource Share (%s) Resource (%s) Association create: %w", resourceShareARN, resourceARN, err)
	}

	return nil
}

func deleteResourceShareResourceAssociation(ctx context.Context, conn *ram.Client, resourceShareARN, resourceARN string) error {
	input := ram.DisassociateResourceShareInput{
		ClientToken:      aws.String(sdkid.UniqueId()),
		ResourceArns:     []string{resourceARN},
		ResourceShareArn: aws.String(resourceShareARN),
	}
	_, err := conn.DisassociateResourceShare(ctx, &input)

	if errs.IsA[*awstypes.UnknownResourceException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting RAM Resource Share (%s) Resource (%s) Association: %w", resourceShareARN, resourceARN, err)
	}

	if _, err := waitResourceAssociationDeleted(ctx, conn, resourceShareARN, resourceARN); err != nil {
		return fmt.Errorf("waiting for RAM Resource Share (%s) Resource (%s) Association delete: %w", resourceShareARN, resourceARN, err)
	}

	return nil
}

func findResourceAssociationByTwoPartKey(ctx context.Context, conn *ram.Client, resourceShareARN, resourceARN string) (*awstypes.ResourceShareAssociation, error) {
	input := ram.GetResourceShareAssociationsInput{
		AssociationType:   awstypes.ResourceShareAssociationTypeResource,
		ResourceArn:       aws.String(resourceARN),
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

func findResourceShareAssociation(ctx context.Context, conn *ram.Client, input *ram.GetResourceShareAssociationsInput) (*awstypes.ResourceShareAssociation, error) {
	output, err := findResourceShareAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findResourceShareAssociations(ctx context.Context, conn *ram.Client, input *ram.GetResourceShareAssociationsInput) ([]awstypes.ResourceShareAssociation, error) {
	var output []awstypes.ResourceShareAssociation

	pages := ram.NewGetResourceShareAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceArnNotFoundException](err) || errs.IsA[*awstypes.UnknownResourceException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ResourceShareAssociations...)
	}

	return output, nil
}

func statusResourceAssociation(conn *ram.Client, resourceShareARN, resourceARN string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findResourceAssociationByTwoPartKey(ctx, conn, resourceShareARN, resourceARN)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitResourceAssociationCreated(ctx context.Context, conn *ram.Client, resourceShareARN, resourceARN string) (*awstypes.ResourceShareAssociation, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceShareAssociationStatusAssociating),
		Target:  enum.Slice(awstypes.ResourceShareAssociationStatusAssociated),
		Refresh: statusResourceAssociation(conn, resourceShareARN, resourceARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceShareAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitResourceAssociationDeleted(ctx context.Context, conn *ram.Client, resourceShareARN, resourceARN string) (*awstypes.ResourceShareAssociation, error) { //nolint:unparam
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceShareAssociationStatusAssociated, awstypes.ResourceShareAssociationStatusDisassociating),
		Target:  []string{},
		Refresh: statusResourceAssociation(conn, resourceShareARN, resourceARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceShareAssociation); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
