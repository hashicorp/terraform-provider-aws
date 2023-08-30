// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_ram_resource_association")
func ResourceResourceAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceAssociationCreate,
		ReadWithoutTimeout:   resourceResourceAssociationRead,
		DeleteWithoutTimeout: resourceResourceAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_share_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceResourceAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)
	resourceARN := d.Get("resource_arn").(string)
	resourceShareARN := d.Get("resource_share_arn").(string)

	input := &ram.AssociateResourceShareInput{
		ClientToken:      aws.String(id.UniqueId()),
		ResourceArns:     aws.StringSlice([]string{resourceARN}),
		ResourceShareArn: aws.String(resourceShareARN),
	}

	log.Printf("[DEBUG] Associating RAM Resource Share: %s", input)
	_, err := conn.AssociateResourceShareWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating RAM Resource Share: %s", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", resourceShareARN, resourceARN))

	if err := waitForResourceShareResourceAssociation(ctx, conn, resourceShareARN, resourceARN); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM Resource Share (%s) Resource Association (%s): %s", resourceShareARN, resourceARN, err)
	}

	return append(diags, resourceResourceAssociationRead(ctx, d, meta)...)
}

func resourceResourceAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	resourceShareARN, resourceARN, err := DecodeResourceAssociationID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Share Resource Association (%s): %s", d.Id(), err)
	}

	resourceShareAssociation, err := GetResourceShareAssociation(ctx, conn, resourceShareARN, resourceARN)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RAM Resource Share (%s) Resource Association (%s) not found, removing from state", resourceShareARN, resourceARN)
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Share Resource Association (%s): %s", d.Id(), err)
	}

	if !d.IsNewResource() && aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusAssociated {
		log.Printf("[WARN] RAM Resource Share (%s) Resource Association (%s) not associated, removing from state", resourceShareARN, resourceARN)
		d.SetId("")
		return diags
	}

	d.Set("resource_arn", resourceARN)
	d.Set("resource_share_arn", resourceShareARN)

	return diags
}

func resourceResourceAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	resourceShareARN, resourceARN, err := DecodeResourceAssociationID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RAM Resource Share Resource Association (%s): %s", d.Id(), err)
	}

	input := &ram.DisassociateResourceShareInput{
		ResourceArns:     aws.StringSlice([]string{resourceARN}),
		ResourceShareArn: aws.String(resourceShareARN),
	}

	log.Printf("[DEBUG] Disassociating RAM Resource Share: %s", input)
	_, err = conn.DisassociateResourceShareWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RAM Resource Share Resource Association (%s): %s", d.Id(), err)
	}

	if err := WaitForResourceShareResourceDisassociation(ctx, conn, resourceShareARN, resourceARN); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RAM Resource Share Resource Association (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func DecodeResourceAssociationID(id string) (string, string, error) {
	idFormatErr := fmt.Errorf("unexpected format of ID (%s), expected SHARE,RESOURCE", id)

	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", idFormatErr
	}

	return parts[0], parts[1], nil
}

func GetResourceShareAssociation(ctx context.Context, conn *ram.RAM, resourceShareARN, resourceARN string) (*ram.ResourceShareAssociation, error) {
	input := &ram.GetResourceShareAssociationsInput{
		AssociationType:   aws.String(ram.ResourceShareAssociationTypeResource),
		ResourceArn:       aws.String(resourceARN),
		ResourceShareArns: aws.StringSlice([]string{resourceShareARN}),
	}

	output, err := conn.GetResourceShareAssociationsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	switch count := len(output.ResourceShareAssociations); count {
	case 0:
		return nil, tfresource.NewEmptyResultError(input)
	case 1:
		return output.ResourceShareAssociations[0], nil
	default:
		return nil, tfresource.NewTooManyResultsError(count, input)
	}
}

func resourceAssociationStateRefreshFunc(ctx context.Context, conn *ram.RAM, resourceShareARN, resourceARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resourceShareAssociation, err := GetResourceShareAssociation(ctx, conn, resourceShareARN, resourceARN)
		if tfresource.NotFound(err) {
			return nil, ram.ResourceShareAssociationStatusDisassociated, nil
		}
		if err != nil {
			return nil, "", err
		}

		if aws.StringValue(resourceShareAssociation.Status) == ram.ResourceShareAssociationStatusFailed {
			extendedErr := fmt.Errorf("association status message: %s", aws.StringValue(resourceShareAssociation.StatusMessage))
			return resourceShareAssociation, aws.StringValue(resourceShareAssociation.Status), extendedErr
		}

		return resourceShareAssociation, aws.StringValue(resourceShareAssociation.Status), nil
	}
}

func waitForResourceShareResourceAssociation(ctx context.Context, conn *ram.RAM, resourceShareARN, resourceARN string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociating},
		Target:  []string{ram.ResourceShareAssociationStatusAssociated},
		Refresh: resourceAssociationStateRefreshFunc(ctx, conn, resourceShareARN, resourceARN),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitForResourceShareResourceDisassociation(ctx context.Context, conn *ram.RAM, resourceShareARN, resourceARN string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated, ram.ResourceShareAssociationStatusDisassociating},
		Target:  []string{ram.ResourceShareAssociationStatusDisassociated},
		Refresh: resourceAssociationStateRefreshFunc(ctx, conn, resourceShareARN, resourceARN),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
