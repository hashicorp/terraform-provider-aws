// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoveryreadiness

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoveryreadiness_cell", name="Cell")
// @Tags(identifierAttribute="arn")
func ResourceCell() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCellCreate,
		ReadWithoutTimeout:   resourceCellRead,
		UpdateWithoutTimeout: resourceCellUpdate,
		DeleteWithoutTimeout: resourceCellDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cell_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cells": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"parent_readiness_scopes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCellCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	name := d.Get("cell_name").(string)
	input := &route53recoveryreadiness.CreateCellInput{
		CellName: aws.String(name),
		Cells:    flex.ExpandStringList(d.Get("cells").([]interface{})),
	}

	output, err := conn.CreateCellWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Readiness Cell (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.CellName))

	if err := createTags(ctx, conn, aws.StringValue(output.CellArn), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Recovery Readiness Cell (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceCellRead(ctx, d, meta)...)
}

func resourceCellRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	input := &route53recoveryreadiness.GetCellInput{
		CellName: aws.String(d.Id()),
	}

	resp, err := conn.GetCellWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53 Recovery Readiness Cell (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Recovery Readiness Cell (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, resp.CellArn)
	d.Set("cell_name", resp.CellName)
	d.Set("cells", resp.Cells)
	d.Set("parent_readiness_scopes", resp.ParentReadinessScopes)

	return diags
}

func resourceCellUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &route53recoveryreadiness.UpdateCellInput{
			CellName: aws.String(d.Id()),
			Cells:    flex.ExpandStringList(d.Get("cells").([]interface{})),
		}

		_, err := conn.UpdateCellWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Readiness Cell (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCellRead(ctx, d, meta)...)
}

func resourceCellDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	log.Printf("[DEBUG] Deleting Route53 Recovery Readiness Cell: %s", d.Id())
	_, err := conn.DeleteCellWithContext(ctx, &route53recoveryreadiness.DeleteCellInput{
		CellName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Readiness Cell (%s): %s", d.Id(), err)
	}

	gcinput := &route53recoveryreadiness.GetCellInput{
		CellName: aws.String(d.Id()),
	}
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := conn.GetCellWithContext(ctx, gcinput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("Route 53 Recovery Readiness Cell (%s) still exists", d.Id()))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.GetCellWithContext(ctx, gcinput)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Recovery Readiness Cell (%s) deletion: %s", d.Id(), err)
	}

	return diags
}
