// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafregional_size_constraint_set", name="Size Constraint Set")
func resourceSizeConstraintSet() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get(names.AttrName).(string)

	log.Printf("[INFO] Creating WAF Regional SizeConstraintSet: %s", name)

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &wafregional.CreateSizeConstraintSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateSizeConstraintSet(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional SizeConstraintSet: %s", err)
	}
	resp := out.(*wafregional.CreateSizeConstraintSetOutput)

	d.SetId(aws.ToString(resp.SizeConstraintSet.SizeConstraintSetId))

	return append(diags, resourceSizeConstraintSetUpdate(ctx, d, meta)...)
}

func resourceSizeConstraintSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)

	log.Printf("[INFO] Reading WAF Regional SizeConstraintSet: %s", d.Get(names.AttrName).(string))
	params := &wafregional.GetSizeConstraintSetInput{
		SizeConstraintSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetSizeConstraintSet(ctx, params)
	if !d.IsNewResource() && errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		log.Printf("[WARN] WAF Regional SizeConstraintSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting WAF Regional Size Constraint Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, resp.SizeConstraintSet.Name)
	d.Set("size_constraints", FlattenSizeConstraints(resp.SizeConstraintSet.SizeConstraints))

	return diags
}

func resourceSizeConstraintSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient)

	if d.HasChange("size_constraints") {
		o, n := d.GetChange("size_constraints")
		oldConstraints, newConstraints := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateRegionalSizeConstraintSetResource(ctx, d.Id(), oldConstraints, newConstraints, client.WAFRegionalClient(ctx), client.Region)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional SizeConstraintSet(%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSizeConstraintSetRead(ctx, d, meta)...)
}

func resourceSizeConstraintSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	oldConstraints := d.Get("size_constraints").(*schema.Set).List()

	if len(oldConstraints) > 0 {
		noConstraints := []interface{}{}
		err := updateRegionalSizeConstraintSetResource(ctx, d.Id(), oldConstraints, noConstraints, conn, region)
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "deleting WAF Regional SizeConstraintSet(%s): %s", d.Id(), err)
			}
		}
	}

	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &wafregional.DeleteSizeConstraintSetInput{
			ChangeToken:         token,
			SizeConstraintSetId: aws.String(d.Id()),
		}
		return conn.DeleteSizeConstraintSet(ctx, req)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Size Constraint Set (%s): %s", d.Id(), err)
	}

	return diags
}

func updateRegionalSizeConstraintSetResource(ctx context.Context, id string, oldConstraints, newConstraints []interface{}, conn *wafregional.Client, region string) error {
	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &wafregional.UpdateSizeConstraintSetInput{
			ChangeToken:         token,
			SizeConstraintSetId: aws.String(id),
			Updates:             DiffSizeConstraints(oldConstraints, newConstraints),
		}

		log.Printf("[INFO] Updating WAF Regional SizeConstraintSet: %s", id)
		return conn.UpdateSizeConstraintSet(ctx, req)
	})

	return err
}
