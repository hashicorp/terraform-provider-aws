package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceNetworkInsightsAnalysis() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceNetworkInsightsAnalysisRead,

		Schema: map[string]*schema.Schema{
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"explanations": networkInsightsAnalysisExplanationsSchema,
			"filter":       CustomFiltersSchema(),
			"filter_in_arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"forward_path_components": networkInsightsAnalysisPathComponentsSchema,
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
			"return_path_components": networkInsightsAnalysisPathComponentsSchema,
			"start_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"warning_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceNetworkInsightsAnalysisRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeNetworkInsightsAnalysesInput{}

	if v, ok := d.GetOk("network_insights_analysis_id"); ok {
		input.NetworkInsightsAnalysisIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	output, err := FindNetworkInsightsAnalysis(ctx, conn, input)

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("EC2 Network Insights Analysis", err))
	}

	networkInsightsAnalysisID := aws.StringValue(output.NetworkInsightsAnalysisId)
	d.SetId(networkInsightsAnalysisID)
	if err := d.Set("alternate_path_hints", flattenAlternatePathHints(output.AlternatePathHints)); err != nil {
		return diag.Errorf("setting alternate_path_hints: %s", err)
	}
	d.Set("arn", output.NetworkInsightsAnalysisArn)
	if err := d.Set("explanations", flattenExplanations(output.Explanations)); err != nil {
		return diag.Errorf("setting explanations: %s", err)
	}
	d.Set("filter_in_arns", aws.StringValueSlice(output.FilterInArns))
	if err := d.Set("forward_path_components", flattenPathComponents(output.ForwardPathComponents)); err != nil {
		return diag.Errorf("setting forward_path_components: %s", err)
	}
	d.Set("network_insights_analysis_id", networkInsightsAnalysisID)
	d.Set("network_insights_path_id", output.NetworkInsightsPathId)
	d.Set("path_found", output.NetworkPathFound)
	if err := d.Set("return_path_components", flattenPathComponents(output.ReturnPathComponents)); err != nil {
		return diag.Errorf("setting return_path_components: %s", err)
	}
	d.Set("start_date", output.StartDate.Format(time.RFC3339))
	d.Set("status", output.Status)
	d.Set("status_message", output.StatusMessage)
	d.Set("warning_message", output.WarningMessage)

	if err := d.Set("tags", KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
