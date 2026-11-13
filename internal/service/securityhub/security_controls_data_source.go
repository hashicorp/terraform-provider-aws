// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package securityhub

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkDataSource("aws_securityhub_security_controls", name="Security Controls")
func newSecurityControlsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &securityControlsDataSource{}

	return d, nil
}

type securityControlsDataSource struct {
	framework.DataSourceWithModel[securityControlsDataSourceModel]
}

func (d *securityControlsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"security_control_definitions": framework.DataSourceComputedListOfObjectAttribute[securityControlDefinitionModel](ctx),
			"standards_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
		},
	}
}

func (d *securityControlsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data securityControlsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().SecurityHubClient(ctx)

	input := securityhub.ListSecurityControlDefinitionsInput{
		StandardsArn: fwflex.StringFromFramework(ctx, data.StandardsARN),
	}

	out, err := findSecurityControls(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError("reading Security Hub Security Controls", err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data.SecurityControlDefinitions)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type securityControlsDataSourceModel struct {
	framework.WithRegionModel
	SecurityControlDefinitions fwtypes.ListNestedObjectValueOf[securityControlDefinitionModel] `tfsdk:"security_control_definitions"`
	StandardsARN               fwtypes.ARN                                                     `tfsdk:"standards_arn"`
}

type securityControlDefinitionModel struct {
	CurrentRegionAvailability fwtypes.StringEnum[awstypes.RegionAvailabilityStatus]      `tfsdk:"current_region_availability"`
	CustomizableProperties    fwtypes.ListOfStringEnum[awstypes.SecurityControlProperty] `tfsdk:"customizable_properties"`
	Description               types.String                                               `tfsdk:"description"`
	RemediationURL            types.String                                               `tfsdk:"remediation_url"`
	SecurityControlID         types.String                                               `tfsdk:"security_control_id"`
	SeverityRating            fwtypes.StringEnum[awstypes.SeverityRating]                `tfsdk:"severity_rating"`
	Title                     types.String                                               `tfsdk:"title"`
}

func findSecurityControls(ctx context.Context, conn *securityhub.Client, input *securityhub.ListSecurityControlDefinitionsInput) ([]awstypes.SecurityControlDefinition, error) {
	var output []awstypes.SecurityControlDefinition

	pages := securityhub.NewListSecurityControlDefinitionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("listing Security Control Definitions: %w", err)
		}

		output = append(output, page.SecurityControlDefinitions...)
	}

	return output, nil
}
