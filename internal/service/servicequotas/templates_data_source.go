// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_servicequotas_templates", name="Templates")
// @Region(overrideEnabled=false)
func newTemplatesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceTemplates{}, nil
}

type dataSourceTemplates struct {
	framework.DataSourceWithModel[templatesDataSourceModel]
}

func (d *dataSourceTemplates) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aws_region": schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrRegion: schema.StringAttribute{
				Optional:           true,
				DeprecationMessage: "region is deprecated. Use aws_region instead.",
			},
			"templates": framework.DataSourceComputedListOfObjectAttribute[serviceQuotaIncreaseRequestInTemplateModel](ctx),
		},
	}
}

func (d *dataSourceTemplates) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data templatesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ServiceQuotasClient(ctx)

	region := fwflex.StringValueFromFramework(ctx, data.AWSRegion)
	if region == "" {
		region = fwflex.StringValueFromFramework(ctx, data.Region)
	}
	input := servicequotas.ListServiceQuotaIncreaseRequestsInTemplateInput{
		AwsRegion: aws.String(region),
	}
	output, err := findTemplates(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError("reading Service Quotas Templates", err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.Templates)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	data.ID = fwflex.StringValueToFramework(ctx, region)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (d *dataSourceTemplates) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("aws_region"),
			path.MatchRoot(names.AttrRegion),
		),
	}
}

func findTemplates(ctx context.Context, conn *servicequotas.Client, input *servicequotas.ListServiceQuotaIncreaseRequestsInTemplateInput) ([]awstypes.ServiceQuotaIncreaseRequestInTemplate, error) {
	var output []awstypes.ServiceQuotaIncreaseRequestInTemplate

	pages := servicequotas.NewListServiceQuotaIncreaseRequestsInTemplatePaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.ServiceQuotaIncreaseRequestInTemplateList...)
	}

	return output, nil
}

type templatesDataSourceModel struct {
	AWSRegion types.String                                                                `tfsdk:"aws_region"`
	ID        types.String                                                                `tfsdk:"id"`
	Region    types.String                                                                `tfsdk:"region"`
	Templates fwtypes.ListNestedObjectValueOf[serviceQuotaIncreaseRequestInTemplateModel] `tfsdk:"templates"`
}

type serviceQuotaIncreaseRequestInTemplateModel struct {
	AWSRegion    types.String  `tfsdk:"region"`
	DesiredValue types.Float64 `tfsdk:"value"`
	GlobalQuota  types.Bool    `tfsdk:"global_quota"`
	QuotaCode    types.String  `tfsdk:"quota_code"`
	QuotaName    types.String  `tfsdk:"quota_name"`
	ServiceCode  types.String  `tfsdk:"service_code"`
	ServiceName  types.String  `tfsdk:"service_name"`
	Unit         types.String  `tfsdk:"unit"`
}
