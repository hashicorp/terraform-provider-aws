// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_macie2_account", name="Account")
func resourceAccount() *schema.Resource {
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FindingPublishingFrequency](),
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MacieStatus](),
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

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := &macie2.EnableMacieInput{
		ClientToken: aws.String(id.UniqueId()),
	}

	if v, ok := d.GetOk("finding_publishing_frequency"); ok {
		input.FindingPublishingFrequency = awstypes.FindingPublishingFrequency(v.(string))
	}
	if v, ok := d.GetOk(names.AttrStatus); ok {
		input.Status = awstypes.MacieStatus(v.(string))
	}

	err := retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		_, err := conn.EnableMacie(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, string(awstypes.ErrorCodeClientError)) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.EnableMacie(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling Macie Account: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID(ctx))

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := &macie2.GetMacieSessionInput{}

	resp, err := conn.GetMacieSession(ctx, input)

	if !d.IsNewResource() && (errs.IsA[*awstypes.ResourceNotFoundException](err) ||
		errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled")) {
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
	d.Set(names.AttrCreatedAt, aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))

	return diags
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := &macie2.UpdateMacieSessionInput{}

	if d.HasChange("finding_publishing_frequency") {
		input.FindingPublishingFrequency = awstypes.FindingPublishingFrequency(d.Get("finding_publishing_frequency").(string))
	}

	if d.HasChange(names.AttrStatus) {
		input.Status = awstypes.MacieStatus(d.Get(names.AttrStatus).(string))
	}

	_, err := conn.UpdateMacieSession(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Macie Account (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAccountRead(ctx, d, meta)...)
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := &macie2.DisableMacieInput{}

	err := retry.RetryContext(ctx, 4*time.Minute, func() *retry.RetryError {
		_, err := conn.DisableMacie(ctx, input)

		if errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "Cannot disable Macie while associated with an administrator account") {
			return retry.RetryableError(err)
		}

		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
				errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled") {
				return nil
			}
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DisableMacie(ctx, input)
	}

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
			errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "disabling Macie Account (%s): %s", d.Id(), err)
	}

	return diags
}
