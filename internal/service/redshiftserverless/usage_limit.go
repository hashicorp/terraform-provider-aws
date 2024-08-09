// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
				Type:         schema.TypeString,
				Optional:     true,
				Default:      redshiftserverless.UsageLimitBreachActionLog,
				ValidateFunc: validation.StringInSlice(redshiftserverless.UsageLimitBreachAction_Values(), false),
			},
			"period": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      redshiftserverless.UsageLimitPeriodMonthly,
				ValidateFunc: validation.StringInSlice(redshiftserverless.UsageLimitPeriod_Values(), false),
			},
			names.AttrResourceARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"usage_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(redshiftserverless.UsageLimitUsageType_Values(), false),
			},
		},
	}
}

func resourceUsageLimitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	input := redshiftserverless.CreateUsageLimitInput{
		Amount:      aws.Int64(int64(d.Get("amount").(int))),
		ResourceArn: aws.String(d.Get(names.AttrResourceARN).(string)),
		UsageType:   aws.String(d.Get("usage_type").(string)),
	}

	if v, ok := d.GetOk("period"); ok {
		input.Period = aws.String(v.(string))
	}

	if v, ok := d.GetOk("breach_action"); ok {
		input.BreachAction = aws.String(v.(string))
	}

	out, err := conn.CreateUsageLimitWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Serverless Usage Limit : %s", err)
	}

	d.SetId(aws.StringValue(out.UsageLimit.UsageLimitId))

	return append(diags, resourceUsageLimitRead(ctx, d, meta)...)
}

func resourceUsageLimitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

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

func resourceUsageLimitUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	input := &redshiftserverless.UpdateUsageLimitInput{
		UsageLimitId: aws.String(d.Id()),
	}

	if d.HasChange("amount") {
		input.Amount = aws.Int64(int64(d.Get("amount").(int)))
	}

	if d.HasChange("breach_action") {
		input.BreachAction = aws.String(d.Get("breach_action").(string))
	}

	_, err := conn.UpdateUsageLimitWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Redshift Serverless Usage Limit (%s): %s", d.Id(), err)
	}

	return append(diags, resourceUsageLimitRead(ctx, d, meta)...)
}

func resourceUsageLimitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn(ctx)

	log.Printf("[DEBUG] Deleting Redshift Serverless Usage Limit: %s", d.Id())
	_, err := conn.DeleteUsageLimitWithContext(ctx, &redshiftserverless.DeleteUsageLimitInput{
		UsageLimitId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Serverless Usage Limit (%s): %s", d.Id(), err)
	}

	return diags
}
