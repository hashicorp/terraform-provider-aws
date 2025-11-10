// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshiftserverless_usage_limit", name="Usage Limit")
func resourceUsageLimit() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUsageLimitCreate,
		ReadWithoutTimeout:   resourceUsageLimitRead,
		UpdateWithoutTimeout: resourceUsageLimitUpdate,
		DeleteWithoutTimeout: resourceUsageLimitDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"amount": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"breach_action": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.UsageLimitBreachActionLog,
				ValidateDiagFunc: enum.Validate[awstypes.UsageLimitBreachAction](),
			},
			"period": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.UsageLimitPeriodMonthly,
				ValidateDiagFunc: enum.Validate[awstypes.UsageLimitPeriod](),
			},
			names.AttrResourceARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"usage_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.UsageLimitUsageType](),
			},
		},
	}
}

func resourceUsageLimitCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	input := redshiftserverless.CreateUsageLimitInput{
		Amount:      aws.Int64(int64(d.Get("amount").(int))),
		ResourceArn: aws.String(d.Get(names.AttrResourceARN).(string)),
		UsageType:   awstypes.UsageLimitUsageType(d.Get("usage_type").(string)),
	}

	if v, ok := d.GetOk("period"); ok {
		input.Period = awstypes.UsageLimitPeriod(v.(string))
	}

	if v, ok := d.GetOk("breach_action"); ok {
		input.BreachAction = awstypes.UsageLimitBreachAction(v.(string))
	}

	out, err := conn.CreateUsageLimit(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Serverless Usage Limit : %s", err)
	}

	d.SetId(aws.ToString(out.UsageLimit.UsageLimitId))

	return append(diags, resourceUsageLimitRead(ctx, d, meta)...)
}

func resourceUsageLimitRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	out, err := findUsageLimitByName(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless UsageLimit (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Serverless Usage Limit (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.UsageLimitArn)
	d.Set("breach_action", out.BreachAction)
	d.Set("period", out.Period)
	d.Set("usage_type", out.UsageType)
	d.Set(names.AttrResourceARN, out.ResourceArn)
	d.Set("amount", out.Amount)

	return diags
}

func resourceUsageLimitUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	input := &redshiftserverless.UpdateUsageLimitInput{
		UsageLimitId: aws.String(d.Id()),
	}

	if d.HasChange("amount") {
		input.Amount = aws.Int64(int64(d.Get("amount").(int)))
	}

	if d.HasChange("breach_action") {
		input.BreachAction = awstypes.UsageLimitBreachAction(d.Get("breach_action").(string))
	}

	_, err := conn.UpdateUsageLimit(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Redshift Serverless Usage Limit (%s): %s", d.Id(), err)
	}

	return append(diags, resourceUsageLimitRead(ctx, d, meta)...)
}

func resourceUsageLimitDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift Serverless Usage Limit: %s", d.Id())
	_, err := conn.DeleteUsageLimit(ctx, &redshiftserverless.DeleteUsageLimitInput{
		UsageLimitId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Serverless Usage Limit (%s): %s", d.Id(), err)
	}

	return diags
}

func findUsageLimitByName(ctx context.Context, conn *redshiftserverless.Client, id string) (*awstypes.UsageLimit, error) {
	input := &redshiftserverless.GetUsageLimitInput{
		UsageLimitId: aws.String(id),
	}

	output, err := conn.GetUsageLimit(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "does not exist") {
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

	return output.UsageLimit, nil
}
