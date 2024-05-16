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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"resource_arn": {
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

func resourceResourceAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	resourceShareARN, resourceARN := d.Get("resource_share_arn").(string), d.Get("resource_arn").(string)
	id := errs.Must(flex.FlattenResourceId([]string{resourceShareARN, resourceARN}, resourceAssociationResourceIDPartCount, false))
	_, err := findResourceAssociationByTwoPartKey(ctx, conn, resourceShareARN, resourceARN)

	switch {
	case err == nil:
		return sdkdiag.AppendFromErr(diags, fmt.Errorf("RAM Resource Association (%s) already exists", id))
	case tfresource.NotFound(err):
		break
	default:
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Association: %s", err)
	}

	input := &ram.AssociateResourceShareInput{
		ClientToken:      aws.String(sdkid.UniqueId()),
		ResourceArns:     aws.StringSlice([]string{resourceARN}),
		ResourceShareArn: aws.String(resourceShareARN),
	}

	_, err = conn.AssociateResourceShareWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RAM Resource Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitResourceAssociationCreated(ctx, conn, resourceShareARN, resourceARN); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM Resource Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceResourceAssociationRead(ctx, d, meta)...)
}

func resourceResourceAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), resourceAssociationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	resourceShareARN, resourceARN := parts[0], parts[1]

	resourceAssociation, err := findResourceAssociationByTwoPartKey(ctx, conn, resourceShareARN, resourceARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RAM Resource Association %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Association (%s): %s", d.Id(), err)
	}

	d.Set("resource_arn", resourceAssociation.AssociatedEntity)
	d.Set("resource_share_arn", resourceAssociation.ResourceShareArn)

	return diags
}

func resourceResourceAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), resourceAssociationResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	resourceShareARN, resourceARN := parts[0], parts[1]

	log.Printf("[DEBUG] Deleting RAM Resource Association: %s", d.Id())
	_, err = conn.DisassociateResourceShareWithContext(ctx, &ram.DisassociateResourceShareInput{
		ResourceArns:     aws.StringSlice([]string{resourceARN}),
		ResourceShareArn: aws.String(resourceShareARN),
	})

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RAM Resource Association (%s): %s", d.Id(), err)
	}

	if _, err := waitResourceAssociationDeleted(ctx, conn, resourceShareARN, resourceARN); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM Resource Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findResourceAssociationByTwoPartKey(ctx context.Context, conn *ram.RAM, resourceShareARN, resourceARN string) (*ram.ResourceShareAssociation, error) {
	input := &ram.GetResourceShareAssociationsInput{
		AssociationType:   aws.String(ram.ResourceShareAssociationTypeResource),
		ResourceArn:       aws.String(resourceARN),
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

func findResourceShareAssociation(ctx context.Context, conn *ram.RAM, input *ram.GetResourceShareAssociationsInput) (*ram.ResourceShareAssociation, error) {
	output, err := findResourceShareAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findResourceShareAssociations(ctx context.Context, conn *ram.RAM, input *ram.GetResourceShareAssociationsInput) ([]*ram.ResourceShareAssociation, error) {
	var output []*ram.ResourceShareAssociation

	err := conn.GetResourceShareAssociationsPagesWithContext(ctx, input, func(page *ram.GetResourceShareAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResourceShareAssociations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeResourceArnNotFoundException, ram.ErrCodeUnknownResourceException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusResourceAssociation(ctx context.Context, conn *ram.RAM, resourceShareARN, resourceARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findResourceAssociationByTwoPartKey(ctx, conn, resourceShareARN, resourceARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitResourceAssociationCreated(ctx context.Context, conn *ram.RAM, resourceShareARN, resourceARN string) (*ram.ResourceShareAssociation, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociating},
		Target:  []string{ram.ResourceShareAssociationStatusAssociated},
		Refresh: statusResourceAssociation(ctx, conn, resourceShareARN, resourceARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ram.ResourceShareAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitResourceAssociationDeleted(ctx context.Context, conn *ram.RAM, resourceShareARN, resourceARN string) (*ram.ResourceShareAssociation, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated, ram.ResourceShareAssociationStatusDisassociating},
		Target:  []string{},
		Refresh: statusResourceAssociation(ctx, conn, resourceShareARN, resourceARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ram.ResourceShareAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
