// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_auditmanager_control", name="Control")
// @Tags
func newControlDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &controlDataSource{}, nil
}

type controlDataSource struct {
	framework.DataSourceWithModel[controlDataSourceModel]
}

func (d *controlDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"action_plan_instructions": schema.StringAttribute{
				Computed: true,
			},
			"action_plan_title": schema.StringAttribute{
				Computed: true,
			},
			names.AttrARN:             framework.ARNAttributeComputedOnly(),
			"control_mapping_sources": framework.DataSourceComputedListOfObjectAttribute[controlMappingSourceModel](ctx),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"testing_information": schema.StringAttribute{
				Computed: true,
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ControlType](),
				Required:   true,
			},
		},
	}
}

func (d *controlDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data controlDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().AuditManagerClient(ctx)

	controlMetadata, err := findControlByTwoPartKey(ctx, conn, data.Name.ValueString(), data.Type.ValueEnum())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Control (%s)", data.Name.ValueString()), err.Error())

		return
	}

	// Control metadata from the ListControls API does not contain all information available
	// about a control. Use control ID to get complete information.
	id := aws.ToString(controlMetadata.Id)
	control, err := findControlByID(ctx, conn, id)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Control (%s)", id), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, control, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, id)
	setTagsOut(ctx, control.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findControlByTwoPartKey(ctx context.Context, conn *auditmanager.Client, name string, controlType awstypes.ControlType) (*awstypes.ControlMetadata, error) {
	input := auditmanager.ListControlsInput{
		ControlType: controlType,
	}

	return findControl(ctx, conn, &input, func(v *awstypes.ControlMetadata) bool {
		return aws.ToString(v.Name) == name
	})
}

func findControl(ctx context.Context, conn *auditmanager.Client, input *auditmanager.ListControlsInput, filter tfslices.Predicate[*awstypes.ControlMetadata]) (*awstypes.ControlMetadata, error) {
	output, err := findControls(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findControls(ctx context.Context, conn *auditmanager.Client, input *auditmanager.ListControlsInput, filter tfslices.Predicate[*awstypes.ControlMetadata]) ([]awstypes.ControlMetadata, error) {
	var output []awstypes.ControlMetadata

	pages := auditmanager.NewListControlsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ControlMetadataList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

type controlDataSourceModel struct {
	framework.WithRegionModel
	ActionPlanInstructions types.String                                               `tfsdk:"action_plan_instructions"`
	ActionPlanTitle        types.String                                               `tfsdk:"action_plan_title"`
	ARN                    types.String                                               `tfsdk:"arn"`
	ControlMappingSources  fwtypes.ListNestedObjectValueOf[controlMappingSourceModel] `tfsdk:"control_mapping_sources"`
	Description            types.String                                               `tfsdk:"description"`
	ID                     types.String                                               `tfsdk:"id"`
	Name                   types.String                                               `tfsdk:"name"`
	Tags                   tftags.Map                                                 `tfsdk:"tags"`
	TestingInformation     types.String                                               `tfsdk:"testing_information"`
	Type                   fwtypes.StringEnum[awstypes.ControlType]                   `tfsdk:"type"`
}
