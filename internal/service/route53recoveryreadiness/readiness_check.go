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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoveryreadiness_readiness_check", name="Readiness Check")
// @Tags(identifierAttribute="arn")
func resourceReadinessCheck() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReadinessCheckCreate,
		ReadWithoutTimeout:   resourceReadinessCheckRead,
		UpdateWithoutTimeout: resourceReadinessCheckUpdate,
		DeleteWithoutTimeout: resourceReadinessCheckDelete,
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
			"readiness_check_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_set_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceReadinessCheckCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	name := d.Get("readiness_check_name").(string)
	input := &route53recoveryreadiness.CreateReadinessCheckInput{
		ReadinessCheckName: aws.String(name),
		ResourceSetName:    aws.String(d.Get("resource_set_name").(string)),
	}

	output, err := conn.CreateReadinessCheck(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Readiness Readiness Check (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ReadinessCheckName))

	if err := createTags(ctx, conn, aws.ToString(output.ReadinessCheckArn), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Recovery Readiness Readiness Check (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceReadinessCheckRead(ctx, d, meta)...)
}

func resourceReadinessCheckRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	output, err := findReadinessCheckByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Recovery Readiness Readiness Check (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Recovery Readiness Readiness Check (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ReadinessCheckArn)
	d.Set("readiness_check_name", output.ReadinessCheckName)
	d.Set("resource_set_name", output.ResourceSet)

	return diags
}

func resourceReadinessCheckUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &route53recoveryreadiness.UpdateReadinessCheckInput{
			ReadinessCheckName: aws.String(d.Get("readiness_check_name").(string)),
			ResourceSetName:    aws.String(d.Get("resource_set_name").(string)),
		}

		_, err := conn.UpdateReadinessCheck(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Readiness Readiness Check (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceReadinessCheckRead(ctx, d, meta)...)
}

func resourceReadinessCheckDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessClient(ctx)

	log.Printf("[DEBUG] Deleting Route53 Recovery Readiness Readiness Check: %s", d.Id())
	_, err := conn.DeleteReadinessCheck(ctx, &route53recoveryreadiness.DeleteReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Readiness Readiness Check (%s): %s", d.Id(), err)
	}

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err = findReadinessCheckByName(ctx, conn, d.Id())
		if err != nil {
			if tfresource.NotFound(err) {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("Route 53 Recovery Readiness Readiness Check (%s) still exists", d.Id()))
	})

	if tfresource.TimedOut(err) {
		_, err = findReadinessCheckByName(ctx, conn, d.Id())
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Recovery Readiness Readiness Check (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func findReadinessCheckByName(ctx context.Context, conn *route53recoveryreadiness.Client, name string) (*route53recoveryreadiness.GetReadinessCheckOutput, error) {
	input := &route53recoveryreadiness.GetReadinessCheckInput{
		ReadinessCheckName: aws.String(name),
	}

	output, err := conn.GetReadinessCheck(ctx, input)
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
