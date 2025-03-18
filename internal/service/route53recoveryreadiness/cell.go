// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoveryreadiness

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53recoveryreadiness/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoveryreadiness_cell", name="Cell")
// @Tags(identifierAttribute="arn")
func resourceCell() *schema.Resource {
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
	}
}

func resourceCellCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	name := d.Get("cell_name").(string)
	input := &route53recoveryreadiness.CreateCellInput{
		CellName: aws.String(name),
		Cells:    flex.ExpandStringValueList(d.Get("cells").([]any)),
	}

	output, err := conn.CreateCell(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Readiness Cell (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.CellName))

	if err := createTags(ctx, conn, aws.ToString(output.CellArn), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Recovery Readiness Cell (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceCellRead(ctx, d, meta)...)
}

func resourceCellRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	output, err := findCellByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Recovery Readiness Cell (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Recovery Readiness Cell (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.CellArn)
	d.Set("cell_name", output.CellName)
	d.Set("cells", output.Cells)
	d.Set("parent_readiness_scopes", output.ParentReadinessScopes)

	return diags
}

func resourceCellUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &route53recoveryreadiness.UpdateCellInput{
			CellName: aws.String(d.Id()),
			Cells:    flex.ExpandStringValueList(d.Get("cells").([]any)),
		}

		_, err := conn.UpdateCell(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Readiness Cell (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCellRead(ctx, d, meta)...)
}

func resourceCellDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Recovery Readiness Cell: %s", d.Id())
	_, err := conn.DeleteCell(ctx, &route53recoveryreadiness.DeleteCellInput{
		CellName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Readiness Cell (%s): %s", d.Id(), err)
	}

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := findCellByName(ctx, conn, d.Id())
		if err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("Route 53 Recovery Readiness Cell (%s) still exists", d.Id()))
	})
	if tfresource.TimedOut(err) {
		_, err = findCellByName(ctx, conn, d.Id())
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Recovery Readiness Cell (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func findCellByName(ctx context.Context, conn *route53recoveryreadiness.Client, name string) (*route53recoveryreadiness.GetCellOutput, error) {
	input := &route53recoveryreadiness.GetCellInput{
		CellName: aws.String(name),
	}

	output, err := conn.GetCell(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
