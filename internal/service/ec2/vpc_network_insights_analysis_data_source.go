// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_network_insights_analysis", name="Network Insights Analysis")
// @Tags
// @Testing(tagsTest=false)
func dataSourceNetworkInsightsAnalysis() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceNetworkInsightsAnalysisRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"alternate_path_hints": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"component_arn": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"component_id": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"explanations":   networkInsightsAnalysisExplanationsSchema(),
				names.AttrFilter: customFiltersSchema(),
				"filter_in_arns": {
					Type:     schema.TypeList,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"forward_path_components": networkInsightsAnalysisPathComponentsSchema(),
				"network_insights_analysis_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"network_insights_path_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"path_found": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				"return_path_components": networkInsightsAnalysisPathComponentsSchema(),
				"start_date": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrStatus: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrStatusMessage: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags: tftags.TagsSchemaComputed(),
				"warning_message": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceNetworkInsightsAnalysisRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeNetworkInsightsAnalysesInput{}

	if v, ok := d.GetOk("network_insights_analysis_id"); ok {
		input.NetworkInsightsAnalysisIds = []string{v.(string)}
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	output, err := findNetworkInsightsAnalysis(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Network Insights Analysis", err))
	}

	networkInsightsAnalysisID := aws.ToString(output.NetworkInsightsAnalysisId)
	d.SetId(networkInsightsAnalysisID)
	if err := d.Set("alternate_path_hints", flattenAlternatePathHints(output.AlternatePathHints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting alternate_path_hints: %s", err)
	}
	d.Set(names.AttrARN, output.NetworkInsightsAnalysisArn)
	if err := d.Set("explanations", flattenExplanations(output.Explanations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting explanations: %s", err)
	}
	d.Set("filter_in_arns", output.FilterInArns)
	if err := d.Set("forward_path_components", flattenPathComponents(output.ForwardPathComponents)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting forward_path_components: %s", err)
	}
	d.Set("network_insights_analysis_id", networkInsightsAnalysisID)
	d.Set("network_insights_path_id", output.NetworkInsightsPathId)
	d.Set("path_found", output.NetworkPathFound)
	if err := d.Set("return_path_components", flattenPathComponents(output.ReturnPathComponents)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting return_path_components: %s", err)
	}
	d.Set("start_date", output.StartDate.Format(time.RFC3339))
	d.Set(names.AttrStatus, output.Status)
	d.Set(names.AttrStatusMessage, output.StatusMessage)
	d.Set("warning_message", output.WarningMessage)

	setTagsOut(ctx, output.Tags)

	return diags
}
