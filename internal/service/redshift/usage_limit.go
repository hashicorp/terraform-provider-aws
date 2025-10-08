// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.UsageLimitBreachActionLog,
				ValidateDiagFunc: enum.Validate[awstypes.UsageLimitBreachAction](),
			},
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"feature_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.UsageLimitFeatureType](),
			},
			"limit_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.UsageLimitLimitType](),
			},
			"period": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.UsageLimitPeriodMonthly,
				ValidateDiagFunc: enum.Validate[awstypes.UsageLimitPeriod](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceUsageLimitCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	clusterId := d.Get(names.AttrClusterIdentifier).(string)
	input := redshift.CreateUsageLimitInput{
		Amount:            aws.Int64(int64(d.Get("amount").(int))),
		ClusterIdentifier: aws.String(clusterId),
		FeatureType:       awstypes.UsageLimitFeatureType(d.Get("feature_type").(string)),
		LimitType:         awstypes.UsageLimitLimitType(d.Get("limit_type").(string)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("breach_action"); ok {
		input.BreachAction = awstypes.UsageLimitBreachAction(v.(string))
	}

	if v, ok := d.GetOk("period"); ok {
		input.Period = awstypes.UsageLimitPeriod(v.(string))
	}

	out, err := conn.CreateUsageLimit(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Usage Limit (%s): %s", clusterId, err)
	}

	d.SetId(aws.ToString(out.UsageLimitId))

	return append(diags, resourceUsageLimitRead(ctx, d, meta)...)
}

func resourceUsageLimitRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

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
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   names.Redshift,
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
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

func resourceUsageLimitUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &redshift.ModifyUsageLimitInput{
			UsageLimitId: aws.String(d.Id()),
		}

		if d.HasChange("amount") {
			input.Amount = aws.Int64(int64(d.Get("amount").(int)))
		}

		if d.HasChange("breach_action") {
			input.BreachAction = awstypes.UsageLimitBreachAction(d.Get("breach_action").(string))
		}

		_, err := conn.ModifyUsageLimit(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Usage Limit (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUsageLimitRead(ctx, d, meta)...)
}

func resourceUsageLimitDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	deleteInput := redshift.DeleteUsageLimitInput{
		UsageLimitId: aws.String(d.Id()),
	}

	_, err := conn.DeleteUsageLimit(ctx, &deleteInput)

	if err != nil {
		if errs.IsA[*awstypes.UsageLimitNotFoundFault](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Usage Limit (%s): %s", d.Id(), err)
	}

	return diags
}
