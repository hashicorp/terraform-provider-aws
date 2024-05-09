// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_size_constraint_set")
func ResourceSizeConstraintSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSizeConstraintSetCreate,
		ReadWithoutTimeout:   resourceSizeConstraintSetRead,
		UpdateWithoutTimeout: resourceSizeConstraintSetUpdate,
		DeleteWithoutTimeout: resourceSizeConstraintSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: SizeConstraintSetSchema(),
	}
}

func resourceSizeConstraintSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &waf.CreateSizeConstraintSetInput{
		Name: aws.String(name),
	}

	wr := NewRetryer(conn)
	outputRaw, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input.ChangeToken = token

		return conn.CreateSizeConstraintSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Size Constraint Set (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*waf.CreateSizeConstraintSetOutput).SizeConstraintSet.SizeConstraintSetId))

	return append(diags, resourceSizeConstraintSetUpdate(ctx, d, meta)...)
}

func resourceSizeConstraintSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	sizeConstraintSet, err := FindSizeConstraintSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Size Constraint Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Size Constraint Set (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("sizeconstraintset/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())
	d.Set(names.AttrName, sizeConstraintSet.Name)
	if err := d.Set("size_constraints", FlattenSizeConstraints(sizeConstraintSet.SizeConstraints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting size_constraints: %s", err)
	}

	return diags
}

func resourceSizeConstraintSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("size_constraints") {
		o, n := d.GetChange("size_constraints")
		oldConstraints, newConstraints := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateSizeConstraintSetResource(ctx, conn, d.Id(), oldConstraints, newConstraints)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Size Constraint Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSizeConstraintSetRead(ctx, d, meta)...)
}

func resourceSizeConstraintSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	oldConstraints := d.Get("size_constraints").(*schema.Set).List()

	if len(oldConstraints) > 0 {
		err := updateSizeConstraintSetResource(ctx, conn, d.Id(), oldConstraints, []interface{}{})
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "updating WAF Size Constraint Set (%s): %s", d.Id(), err)
			}
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		return conn.DeleteSizeConstraintSet(ctx, &waf.DeleteSizeConstraintSetInput{
			ChangeToken:         token,
			SizeConstraintSetId: aws.String(d.Id()),
		})
	})
	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Size Constraint Set (%s): %s", d.Id(), err)
	}

	return diags
}

func updateSizeConstraintSetResource(ctx context.Context, conn *waf.Client, id string, oldS, newS []interface{}) error {
	input := &waf.UpdateSizeConstraintSetInput{
		SizeConstraintSetId: aws.String(id),
		Updates:             DiffSizeConstraints(oldS, newS),
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input.ChangeToken = token

		return conn.UpdateSizeConstraintSet(ctx, input)
	})

	return err
}

func FindSizeConstraintSetByID(ctx context.Context, conn *waf.Client, id string) (*awstypes.SizeConstraintSet, error) {
	input := &waf.GetSizeConstraintSetInput{
		SizeConstraintSetId: aws.String(id),
	}

	output, err := conn.GetSizeConstraintSet(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SizeConstraintSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SizeConstraintSet, nil
}
