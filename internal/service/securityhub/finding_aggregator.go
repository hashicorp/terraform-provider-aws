// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	allRegions                = "ALL_REGIONS"
	allRegionsExceptSpecified = "ALL_REGIONS_EXCEPT_SPECIFIED"
	specifiedRegions          = "SPECIFIED_REGIONS"
)

// @SDKResource("aws_securityhub_finding_aggregator")
func ResourceFindingAggregator() *schema.Resource {
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
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					allRegions,
					allRegionsExceptSpecified,
					specifiedRegions,
				}, false),
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

	req := &securityhub.CreateFindingAggregatorInput{
		RegionLinkingMode: &linkingMode,
	}

	if v, ok := d.GetOk("specified_regions"); ok && (linkingMode == allRegionsExceptSpecified || linkingMode == specifiedRegions) {
		req.Regions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Creating Security Hub finding aggregator")

	resp, err := conn.CreateFindingAggregator(ctx, req)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating finding aggregator for Security Hub: %s", err)
	}

	d.SetId(aws.ToString(resp.FindingAggregatorArn))

	return append(diags, resourceFindingAggregatorRead(ctx, d, meta)...)
}

func resourceFindingAggregatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	output, err := FindFindingAggregatorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Finding aggregator (%s) not found, removing from state", d.Id())
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

	aggregatorArn := d.Id()

	linkingMode := d.Get("linking_mode").(string)

	req := &securityhub.UpdateFindingAggregatorInput{
		FindingAggregatorArn: &aggregatorArn,
		RegionLinkingMode:    &linkingMode,
	}

	if v, ok := d.GetOk("specified_regions"); ok && (linkingMode == allRegionsExceptSpecified || linkingMode == specifiedRegions) {
		req.Regions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	resp, err := conn.UpdateFindingAggregator(ctx, req)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub finding aggregator (%s): %s", aggregatorArn, err)
	}

	d.SetId(aws.ToString(resp.FindingAggregatorArn))

	return append(diags, resourceFindingAggregatorRead(ctx, d, meta)...)
}

func resourceFindingAggregatorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	aggregatorArn := d.Id()

	log.Printf("[DEBUG] Disabling Security Hub finding aggregator %s", aggregatorArn)

	_, err := conn.DeleteFindingAggregator(ctx, &securityhub.DeleteFindingAggregatorInput{
		FindingAggregatorArn: &aggregatorArn,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Security Hub finding aggregator %s: %s", aggregatorArn, err)
	}

	return diags
}
