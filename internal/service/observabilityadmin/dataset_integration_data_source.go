// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_observabilityadmin_dataset_integration", name="Dataset Integration")
// @Tags(identifierAttribute="arn")
// @Testing(preCheck="testAccDatasetIntegrationPreCheck", serialize=true)
func newDatasetIntegrationDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &datasetIntegrationDataSource{}, nil
}

type datasetIntegrationDataSource struct {
	framework.DataSourceWithModel[datasetIntegrationDataSourceModel]
}

func (d *datasetIntegrationDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (d *datasetIntegrationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data datasetIntegrationDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ObservabilityAdminClient(ctx)

	var input observabilityadmin.ListDatasetIntegrationsInput
	summaries, err := findDatasetIntegrations(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	// At most one dataset integration exists per account and Region.
	summary, err := tfresource.AssertSingleValueResult(summaries)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, summary, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type datasetIntegrationDataSourceModel struct {
	framework.WithRegionModel
	ARN       types.String      `tfsdk:"arn"`
	CreatedAt timetypes.RFC3339 `tfsdk:"created_at"`
	RoleARN   fwtypes.ARN       `tfsdk:"role_arn"`
	Tags      tftags.Map        `tfsdk:"tags"`
	UpdatedAt timetypes.RFC3339 `tfsdk:"updated_at"`
}
