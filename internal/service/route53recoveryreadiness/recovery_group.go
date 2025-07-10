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

// @SDKResource("aws_route53recoveryreadiness_recovery_group", name="Recovery Group")
// @Tags(identifierAttribute="arn")
func resourceRecoveryGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRecoveryGroupCreate,
		ReadWithoutTimeout:   resourceRecoveryGroupRead,
		UpdateWithoutTimeout: resourceRecoveryGroupUpdate,
		DeleteWithoutTimeout: resourceRecoveryGroupDelete,
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
			"cells": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"recovery_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceRecoveryGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	name := d.Get("recovery_group_name").(string)
	input := &route53recoveryreadiness.CreateRecoveryGroupInput{
		Cells:             flex.ExpandStringValueList(d.Get("cells").([]any)),
		RecoveryGroupName: aws.String(name),
	}

	output, err := conn.CreateRecoveryGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Readiness Recovery Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.RecoveryGroupName))

	if err := createTags(ctx, conn, aws.ToString(output.RecoveryGroupArn), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Recovery Readiness Recovery Group (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceRecoveryGroupRead(ctx, d, meta)...)
}

func resourceRecoveryGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	output, err := findRecoveryGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Recovery Readiness Recovery Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Recovery Readiness Recovery Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.RecoveryGroupArn)
	d.Set("recovery_group_name", output.RecoveryGroupName)
	d.Set("cells", output.Cells)

	return diags
}

func resourceRecoveryGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &route53recoveryreadiness.UpdateRecoveryGroupInput{
			RecoveryGroupName: aws.String(d.Id()),
			Cells:             flex.ExpandStringValueList(d.Get("cells").([]any)),
		}

		_, err := conn.UpdateRecoveryGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route 53 Recovery Readiness Recovery Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRecoveryGroupRead(ctx, d, meta)...)
}

func resourceRecoveryGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Recovery Readiness Recovery Group: %s", d.Id())
	_, err := conn.DeleteRecoveryGroup(ctx, &route53recoveryreadiness.DeleteRecoveryGroupInput{
		RecoveryGroupName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Readiness Recovery Group (%s): %s", d.Id(), err)
	}

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := findRecoveryGroupByName(ctx, conn, d.Id())
		if err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("Route53 Recovery Readiness Recovery Group (%s) still exists", d.Id()))
	})

	if tfresource.TimedOut(err) {
		_, err = findRecoveryGroupByName(ctx, conn, d.Id())
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route53 Recovery Readiness Recovery Group (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func findRecoveryGroupByName(ctx context.Context, conn *route53recoveryreadiness.Client, name string) (*route53recoveryreadiness.GetRecoveryGroupOutput, error) {
	input := &route53recoveryreadiness.GetRecoveryGroupInput{
		RecoveryGroupName: aws.String(name),
	}

	output, err := conn.GetRecoveryGroup(ctx, input)
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
