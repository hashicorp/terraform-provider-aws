// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_usage_limit", name="Usage Limit")
// @Tags(identifierAttribute="arn")
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
			"amount": {
				Type:     schema.TypeInt,
				Required: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"breach_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      redshift.UsageLimitBreachActionLog,
				ValidateFunc: validation.StringInSlice(redshift.UsageLimitBreachAction_Values(), false),
			},
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"feature_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(redshift.UsageLimitFeatureType_Values(), false),
			},
			"limit_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(redshift.UsageLimitLimitType_Values(), false),
			},
			"period": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      redshift.UsageLimitPeriodMonthly,
				ValidateFunc: validation.StringInSlice(redshift.UsageLimitPeriod_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceUsageLimitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	clusterId := d.Get(names.AttrClusterIdentifier).(string)
	input := redshift.CreateUsageLimitInput{
		Amount:            aws.Int64(int64(d.Get("amount").(int))),
		ClusterIdentifier: aws.String(clusterId),
		FeatureType:       aws.String(d.Get("feature_type").(string)),
		LimitType:         aws.String(d.Get("limit_type").(string)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("breach_action"); ok {
		input.BreachAction = aws.String(v.(string))
	}

	if v, ok := d.GetOk("period"); ok {
		input.Period = aws.String(v.(string))
	}

	out, err := conn.CreateUsageLimitWithContext(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Usage Limit (%s): %s", clusterId, err)
	}

	d.SetId(aws.StringValue(out.UsageLimitId))

	return append(diags, resourceUsageLimitRead(ctx, d, meta)...)
}

func resourceUsageLimitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	out, err := findUsageLimitByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Usage Limit (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Usage Limit (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("usagelimit:%s", d.Id()),
	}.String()

	d.Set(names.AttrARN, arn)
	d.Set("amount", out.Amount)
	d.Set("period", out.Period)
	d.Set("limit_type", out.LimitType)
	d.Set("feature_type", out.FeatureType)
	d.Set("breach_action", out.BreachAction)
	d.Set(names.AttrClusterIdentifier, out.ClusterIdentifier)

	setTagsOut(ctx, out.Tags)

	return diags
}

func resourceUsageLimitUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &redshift.ModifyUsageLimitInput{
			UsageLimitId: aws.String(d.Id()),
		}

		if d.HasChange("amount") {
			input.Amount = aws.Int64(int64(d.Get("amount").(int)))
		}

		if d.HasChange("breach_action") {
			input.BreachAction = aws.String(d.Get("breach_action").(string))
		}

		_, err := conn.ModifyUsageLimitWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Usage Limit (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUsageLimitRead(ctx, d, meta)...)
}

func resourceUsageLimitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	deleteInput := redshift.DeleteUsageLimitInput{
		UsageLimitId: aws.String(d.Id()),
	}

	_, err := conn.DeleteUsageLimitWithContext(ctx, &deleteInput)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeUsageLimitNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Usage Limit (%s): %s", d.Id(), err)
	}

	return diags
}
