// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	linkingModeAllRegions                = "ALL_REGIONS"
	linkingModeAllRegionsExceptSpecified = "ALL_REGIONS_EXCEPT_SPECIFIED"
	linkingModeSpecifiedRegions          = "SPECIFIED_REGIONS"
)

func linkingMode_Values() []string {
	return []string{
		linkingModeAllRegions,
		linkingModeAllRegionsExceptSpecified,
		linkingModeSpecifiedRegions,
	}
}

// @SDKResource("aws_securityhub_finding_aggregator", name="Finding Aggregator")
func resourceFindingAggregator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFindingAggregatorCreate,
		ReadWithoutTimeout:   resourceFindingAggregatorRead,
		UpdateWithoutTimeout: resourceFindingAggregatorUpdate,
		DeleteWithoutTimeout: resourceFindingAggregatorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"linking_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(linkingMode_Values(), false),
			},
			"specified_regions": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceFindingAggregatorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	linkingMode := d.Get("linking_mode").(string)
	input := &securityhub.CreateFindingAggregatorInput{
		RegionLinkingMode: aws.String(linkingMode),
	}

	if v, ok := d.GetOk("specified_regions"); ok && v.(*schema.Set).Len() > 0 && (linkingMode == linkingModeAllRegionsExceptSpecified || linkingMode == linkingModeSpecifiedRegions) {
		input.Regions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateFindingAggregator(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Finding Aggregator: %s", err)
	}

	d.SetId(aws.ToString(output.FindingAggregatorArn))

	return append(diags, resourceFindingAggregatorRead(ctx, d, meta)...)
}

func resourceFindingAggregatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	output, err := findFindingAggregatorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Finding Aggregator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub finding aggregator to find %s: %s", d.Id(), err)
	}

	d.Set("linking_mode", output.RegionLinkingMode)
	if len(output.Regions) > 0 {
		d.Set("specified_regions", flex.FlattenStringValueList(output.Regions))
	}

	return diags
}

func resourceFindingAggregatorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	linkingMode := d.Get("linking_mode").(string)
	input := &securityhub.UpdateFindingAggregatorInput{
		FindingAggregatorArn: aws.String(d.Id()),
		RegionLinkingMode:    aws.String(linkingMode),
	}

	if v, ok := d.GetOk("specified_regions"); ok && v.(*schema.Set).Len() > 0 && (linkingMode == linkingModeAllRegionsExceptSpecified || linkingMode == linkingModeSpecifiedRegions) {
		input.Regions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	_, err := conn.UpdateFindingAggregator(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Finding Aggregator (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFindingAggregatorRead(ctx, d, meta)...)
}

func resourceFindingAggregatorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	log.Printf("[DEBUG] Deleting Security Hub Finding Aggregator: %s", d.Id())
	_, err := conn.DeleteFindingAggregator(ctx, &securityhub.DeleteFindingAggregatorInput{
		FindingAggregatorArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Finding Aggregator (%s): %s", d.Id(), err)
	}

	return diags
}

func findFindingAggregatorByARN(ctx context.Context, conn *securityhub.Client, arn string) (*securityhub.GetFindingAggregatorOutput, error) {
	input := &securityhub.GetFindingAggregatorInput{
		FindingAggregatorArn: aws.String(arn),
	}

	return findFindingAggregator(ctx, conn, input)
}

func findFindingAggregator(ctx context.Context, conn *securityhub.Client, input *securityhub.GetFindingAggregatorInput) (*securityhub.GetFindingAggregatorOutput, error) {
	output, err := conn.GetFindingAggregator(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
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
