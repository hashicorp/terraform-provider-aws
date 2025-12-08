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

// @FrameworkDataSource("aws_auditmanager_framework", name="Framework")
// @Tags
func newFrameworkDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &frameworkDataSource{}, nil
}

type frameworkDataSource struct {
	framework.DataSourceWithModel[frameworkDataSourceModel]
}

func (d *frameworkDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"compliance_type": schema.StringAttribute{
				Computed: true,
			},
			"control_sets": framework.DataSourceComputedListOfObjectAttribute[controlSetModel](ctx),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"framework_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.FrameworkType](),
				Required:   true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *frameworkDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data frameworkDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().AuditManagerClient(ctx)

	frameworkMetadata, err := findFrameworkByTwoPartKey(ctx, conn, data.Name.ValueString(), data.Type.ValueEnum())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Framework (%s)", data.Name.ValueString()), err.Error())

		return
	}

	// Framework metadata from the ListFrameworks API does not contain all information available
	// about a framework. Use framework ID to get complete information.
	id := aws.ToString(frameworkMetadata.Id)
	framework, err := findFrameworkByID(ctx, conn, id)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Framework (%s)", id), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, framework, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, id)
	setTagsOut(ctx, framework.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findFrameworkByTwoPartKey(ctx context.Context, conn *auditmanager.Client, name string, frameworkType awstypes.FrameworkType) (*awstypes.AssessmentFrameworkMetadata, error) {
	input := auditmanager.ListAssessmentFrameworksInput{
		FrameworkType: frameworkType,
	}

	return findFramework(ctx, conn, &input, func(v *awstypes.AssessmentFrameworkMetadata) bool {
		return aws.ToString(v.Name) == name
	})
}

func findFramework(ctx context.Context, conn *auditmanager.Client, input *auditmanager.ListAssessmentFrameworksInput, filter tfslices.Predicate[*awstypes.AssessmentFrameworkMetadata]) (*awstypes.AssessmentFrameworkMetadata, error) {
	output, err := findFrameworks(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFrameworks(ctx context.Context, conn *auditmanager.Client, input *auditmanager.ListAssessmentFrameworksInput, filter tfslices.Predicate[*awstypes.AssessmentFrameworkMetadata]) ([]awstypes.AssessmentFrameworkMetadata, error) {
	var output []awstypes.AssessmentFrameworkMetadata

	pages := auditmanager.NewListAssessmentFrameworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.FrameworkMetadataList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

type frameworkDataSourceModel struct {
	framework.WithRegionModel
	ARN            types.String                                     `tfsdk:"arn"`
	ComplianceType types.String                                     `tfsdk:"compliance_type"`
	ControlSets    fwtypes.ListNestedObjectValueOf[controlSetModel] `tfsdk:"control_sets"`
	Description    types.String                                     `tfsdk:"description"`
	ID             types.String                                     `tfsdk:"id"`
	Name           types.String                                     `tfsdk:"name"`
	Tags           tftags.Map                                       `tfsdk:"tags"`
	Type           fwtypes.StringEnum[awstypes.FrameworkType]       `tfsdk:"framework_type"`
}
