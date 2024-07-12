// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_macie2_account")
func ResourceAccount() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountCreate,
		ReadWithoutTimeout:   resourceAccountRead,
		UpdateWithoutTimeout: resourceAccountUpdate,
		DeleteWithoutTimeout: resourceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"finding_publishing_frequency": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(macie2.FindingPublishingFrequency_Values(), false),
			},
			names.AttrStatus: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(macie2.MacieStatus_Values(), false),
			},
			names.AttrServiceRole: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.EnableMacieInput{
		ClientToken: aws.String(id.UniqueId()),
	}

	if v, ok := d.GetOk("finding_publishing_frequency"); ok {
		input.FindingPublishingFrequency = aws.String(v.(string))
	}
	if v, ok := d.GetOk(names.AttrStatus); ok {
		input.Status = aws.String(v.(string))
	}

	err := retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		_, err := conn.EnableMacieWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.EnableMacieWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling Macie Account: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.GetMacieSessionInput{}

	resp, err := conn.GetMacieSessionWithContext(ctx, input)

	if !d.IsNewResource() && (tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled")) {
		log.Printf("[WARN] Macie not enabled for AWS account (%s), removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie Account (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrStatus, resp.Status)
	d.Set("finding_publishing_frequency", resp.FindingPublishingFrequency)
	d.Set(names.AttrServiceRole, resp.ServiceRole)
	d.Set(names.AttrCreatedAt, aws.TimeValue(resp.CreatedAt).Format(time.RFC3339))
	d.Set("updated_at", aws.TimeValue(resp.UpdatedAt).Format(time.RFC3339))

	return diags
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.UpdateMacieSessionInput{}

	if d.HasChange("finding_publishing_frequency") {
		input.FindingPublishingFrequency = aws.String(d.Get("finding_publishing_frequency").(string))
	}

	if d.HasChange(names.AttrStatus) {
		input.Status = aws.String(d.Get(names.AttrStatus).(string))
	}

	_, err := conn.UpdateMacieSessionWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Macie Account (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Conn(ctx)

	input := &macie2.DisableMacieInput{}

	err := retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		_, err := conn.DisableMacieWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, macie2.ErrCodeConflictException, "Cannot disable Macie while associated with an administrator account") {
			return retry.RetryableError(err)
		}

		if err != nil {
			if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
				tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
				return nil
			}
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DisableMacieWithContext(ctx, input)
	}

	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "disabling Macie Account (%s): %s", d.Id(), err)
	}

	return diags
}
