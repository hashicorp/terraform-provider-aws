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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53recoveryreadiness_readiness_check", name="Readiness Check")
// @Tags(identifierAttribute="arn")
func ResourceReadinessCheck() *schema.Resource {
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReadinessCheckCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	name := d.Get("readiness_check_name").(string)
	input := &route53recoveryreadiness.CreateReadinessCheckInput{
		ReadinessCheckName: aws.String(name),
		ResourceSetName:    aws.String(d.Get("resource_set_name").(string)),
	}

	output, err := conn.CreateReadinessCheckWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Readiness Readiness Check (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ReadinessCheckName))

	if err := createTags(ctx, conn, aws.StringValue(output.ReadinessCheckArn), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Recovery Readiness Readiness Check (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceReadinessCheckRead(ctx, d, meta)...)
}

func resourceReadinessCheckRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	input := &route53recoveryreadiness.GetReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Id()),
	}

	resp, err := conn.GetReadinessCheckWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53 Recovery Readiness Readiness Check (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Recovery Readiness Readiness Check (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, resp.ReadinessCheckArn)
	d.Set("readiness_check_name", resp.ReadinessCheckName)
	d.Set("resource_set_name", resp.ResourceSet)

	return diags
}

func resourceReadinessCheckUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &route53recoveryreadiness.UpdateReadinessCheckInput{
			ReadinessCheckName: aws.String(d.Get("readiness_check_name").(string)),
			ResourceSetName:    aws.String(d.Get("resource_set_name").(string)),
		}

		_, err := conn.UpdateReadinessCheckWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Readiness Readiness Check (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceReadinessCheckRead(ctx, d, meta)...)
}

func resourceReadinessCheckDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn(ctx)

	log.Printf("[DEBUG] Deleting Route53 Recovery Readiness Readiness Check: %s", d.Id())
	_, err := conn.DeleteReadinessCheckWithContext(ctx, &route53recoveryreadiness.DeleteReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Readiness Readiness Check (%s): %s", d.Id(), err)
	}

	gcinput := &route53recoveryreadiness.GetReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Id()),
	}
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		_, err := conn.GetReadinessCheckWithContext(ctx, gcinput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
				return nil
			}
			return retry.NonRetryableError(err)
		}
		return retry.RetryableError(fmt.Errorf("Route 53 Recovery Readiness ReadinessCheck (%s) still exists", d.Id()))
	})

	if tfresource.TimedOut(err) {
		_, err = conn.GetReadinessCheckWithContext(ctx, gcinput)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Recovery Readiness ReadinessCheck (%s) deletion: %s", d.Id(), err)
	}

	return diags
}
