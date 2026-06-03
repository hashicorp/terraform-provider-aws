// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package securityhub

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	securityhubschema "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub/internal/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_securityhub_insight", name="Insight")
// @ArnIdentity(identityDuplicateAttributes="id")
// @Testing(serialize=true)
// @Testing(preIdentityVersion="v6.42.0")
func resourceInsight() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInsightCreate,
		ReadWithoutTimeout:   resourceInsightRead,
		UpdateWithoutTimeout: resourceInsightUpdate,
		DeleteWithoutTimeout: resourceInsightDelete,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"filters": securityhubschema.InsightFiltersSchema(),
				"group_by_attribute": {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
			}
		},
	}
}

func resourceInsightCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := securityhub.CreateInsightInput{
		GroupByAttribute: aws.String(d.Get("group_by_attribute").(string)),
		Name:             aws.String(name),
	}

	if v, ok := d.GetOk("filters"); ok {
		input.Filters = securityhubschema.ExpandSecurityFindingFilters(v.([]any))
	}

	output, err := conn.CreateInsight(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Insight (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.InsightArn))

	return append(diags, resourceInsightRead(ctx, d, meta)...)
}

func resourceInsightRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	insight, err := findInsightByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Security Hub Insight (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Insight (%s): %s", d.Id(), err)
	}

	if err := resourceInsightFlatten(ctx, insight, d); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceInsightUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := securityhub.UpdateInsightInput{
		InsightArn: aws.String(d.Id()),
	}

	if d.HasChange("filters") {
		input.Filters = securityhubschema.ExpandSecurityFindingFilters(d.Get("filters").([]any))
	}

	if d.HasChange("group_by_attribute") {
		input.GroupByAttribute = aws.String(d.Get("group_by_attribute").(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	_, err := conn.UpdateInsight(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Insight (%s): %s", d.Id(), err)
	}

	return append(diags, resourceInsightRead(ctx, d, meta)...)
}

func resourceInsightDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	log.Printf("[DEBUG] Deleting Security Hub Insight: %s", d.Id())
	input := securityhub.DeleteInsightInput{
		InsightArn: aws.String(d.Id()),
	}
	_, err := conn.DeleteInsight(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Insight (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceInsightFlatten(_ context.Context, insight *types.Insight, d *schema.ResourceData) error {
	d.Set(names.AttrARN, insight.InsightArn)
	if err := d.Set("filters", securityhubschema.FlattenSecurityFindingFilters(insight.Filters)); err != nil {
		return fmt.Errorf("setting filters: %w", err)
	}
	d.Set("group_by_attribute", insight.GroupByAttribute)
	d.Set(names.AttrName, insight.Name)

	return nil
}

func findInsightByARN(ctx context.Context, conn *securityhub.Client, arn string) (*types.Insight, error) {
	input := securityhub.GetInsightsInput{
		InsightArns: []string{arn},
	}

	return findInsight(ctx, conn, &input)
}

func findInsight(ctx context.Context, conn *securityhub.Client, input *securityhub.GetInsightsInput) (*types.Insight, error) {
	output, err := findInsights(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInsights(ctx context.Context, conn *securityhub.Client, input *securityhub.GetInsightsInput) ([]types.Insight, error) {
	var output []types.Insight

	pages := securityhub.NewGetInsightsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Insights...)
	}

	return output, nil
}
