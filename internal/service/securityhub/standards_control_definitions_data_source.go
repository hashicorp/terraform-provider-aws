// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_securityhub_standards_control_definitions", name="Standards Control Definitions")
func newStandardsControlDefinitionsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &standardsControlDefinitionsDataSource{}

	return d, nil
}

type standardsControlDefinitionsDataSource struct {
	framework.DataSourceWithModel[standardsControlDefinitionsDataSourceModel]
}

func (d *standardsControlDefinitionsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"current_region_availability": schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"severity_rating": schema.StringAttribute{
				Optional: true,
			},
			"standards_arn": schema.StringAttribute{
				Optional: true,
			},
			"control_definitions": framework.DataSourceComputedListOfObjectAttribute[securityControlDefinitionModel](ctx),
		},
	}
}

func (d *standardsControlDefinitionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data standardsControlDefinitionsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().SecurityHubClient(ctx)

	input := &securityhub.ListSecurityControlDefinitionsInput{}

	if !data.StandardsARN.IsNull() {
		input.StandardsArn = data.StandardsARN.ValueStringPointer()
	}

	filter := func(v *awstypes.SecurityControlDefinition) bool {
		if !data.CurrentRegionAvailability.IsNull() {
			if string(v.CurrentRegionAvailability) != data.CurrentRegionAvailability.ValueString() {
				return false
			}
		}

		if !data.SeverityRating.IsNull() {
			if string(v.SeverityRating) != data.SeverityRating.ValueString() {
				return false
			}
		}

		return true
	}

	out, err := findSecurityControlDefinitions(ctx, conn, input, filter)

	if err != nil {
		response.Diagnostics.AddError("reading SecurityHub Standards Control Definitions", err.Error())
		return
	}

	data.ID = types.StringValue(d.Meta().Region(ctx))
	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data.ControlDefinitions)...)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type standardsControlDefinitionsDataSourceModel struct {
	framework.WithRegionModel
	ID                        types.String                                                    `tfsdk:"id"`
	CurrentRegionAvailability types.String                                                    `tfsdk:"current_region_availability"`
	SeverityRating            types.String                                                    `tfsdk:"severity_rating"`
	StandardsARN              types.String                                                    `tfsdk:"standards_arn"`
	ControlDefinitions        fwtypes.ListNestedObjectValueOf[securityControlDefinitionModel] `tfsdk:"control_definitions"`
}

type securityControlDefinitionModel struct {
	ControlID                 types.String                                                        `tfsdk:"control_id"`
	CurrentRegionAvailability fwtypes.StringEnum[awstypes.RegionAvailabilityStatus]               `tfsdk:"current_region_availability"`
	CustomizableProperties    fwtypes.ListOfString                                                `tfsdk:"customizable_properties"`
	Description               types.String                                                        `tfsdk:"description"`
	ParameterDefinitions      fwtypes.MapValueOf[fwtypes.ObjectValueOf[parameterDefinitionModel]] `tfsdk:"parameter_definitions"`
	RemediationURL            types.String                                                        `tfsdk:"remediation_url"`
	SeverityRating            fwtypes.StringEnum[awstypes.SeverityRating]                         `tfsdk:"severity_rating"`
	Title                     types.String                                                        `tfsdk:"title"`
}

type parameterDefinitionModel struct {
	Description          types.String                                     `tfsdk:"description"`
	ConfigurationOptions fwtypes.ObjectValueOf[configurationOptionsModel] `tfsdk:"configuration_options"`
}

type configurationOptionsModel struct {
	Boolean     types.Bool                              `tfsdk:"boolean"`
	Double      types.Float64                           `tfsdk:"double"`
	Enum        fwtypes.ObjectValueOf[enumOptionsModel] `tfsdk:"enum"`
	EnumList    fwtypes.ObjectValueOf[enumOptionsModel] `tfsdk:"enum_list"`
	Integer     types.Int64                             `tfsdk:"integer"`
	IntegerList fwtypes.ListOfInt64                     `tfsdk:"integer_list"`
	String      types.String                            `tfsdk:"string"`
	StringList  fwtypes.ListOfString                    `tfsdk:"string_list"`
}

type enumOptionsModel struct {
	AllowedValues fwtypes.ListOfString `tfsdk:"allowed_values"`
	DefaultValue  types.String         `tfsdk:"default_value"`
	MaxItems      types.Int64          `tfsdk:"max_items"`
	MinItems      types.Int64          `tfsdk:"min_items"`
}

func findSecurityControlDefinitions(ctx context.Context, conn *securityhub.Client, input *securityhub.ListSecurityControlDefinitionsInput, filter tfslices.Predicate[*awstypes.SecurityControlDefinition]) ([]awstypes.SecurityControlDefinition, error) {
	var output []awstypes.SecurityControlDefinition

	pages := securityhub.NewListSecurityControlDefinitionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("listing Security Control Definitions: %w", err)
		}

		for _, v := range page.SecurityControlDefinitions {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
