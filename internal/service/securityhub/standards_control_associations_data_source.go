// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name=Standards Control Associations)
func newDataSourceStandardsControlAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceStandardsControlAssociations{}

	return d, nil
}

const (
	DSNameStandardsControlAssociations = "Standards Control Associations Data Source"
)

type dataSourceStandardsControlAssociations struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceStandardsControlAssociations) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_securityhub_standards_control_associations"
}

func (d *dataSourceStandardsControlAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"security_control_id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"standards_control_associations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[standardsControlAssociationData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"association_status": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AssociationStatus](),
							Computed:   true,
						},
						"related_requirements": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Computed:    true,
						},
						"security_control_arn": schema.StringAttribute{
							Computed: true,
						},
						"security_control_id": schema.StringAttribute{
							Computed: true,
						},
						"standards_arn": schema.StringAttribute{
							Computed: true,
						},
						"standards_control_description": schema.StringAttribute{
							Computed: true,
						},
						"standards_control_title": schema.StringAttribute{
							Computed: true,
						},
						"updated_at": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						"updated_reason": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceStandardsControlAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceStandardsControlAssociationsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().SecurityHubClient(ctx)

	input := &securityhub.ListStandardsControlAssociationsInput{
		SecurityControlId: data.SecurityControlID.ValueStringPointer(),
	}

	out, err := findStandardsControlAssociations(ctx, conn, input, tfslices.PredicateTrue[*awstypes.StandardsControlAssociationSummary]())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionReading, DSNameStandardsControlAssociations, data.SecurityControlID.String(), err),
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(d.Meta().Region)
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data.StandardsControlAssociations)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceStandardsControlAssociationsData struct {
	ID                           types.String                                                     `tfsdk:"id"`
	SecurityControlID            types.String                                                     `tfsdk:"security_control_id"`
	StandardsControlAssociations fwtypes.ListNestedObjectValueOf[standardsControlAssociationData] `tfsdk:"standards_control_associations"`
}

type standardsControlAssociationData struct {
	AssociationStatus           fwtypes.StringEnum[awstypes.AssociationStatus] `tfsdk:"association_status"`
	RelatedRequirements         fwtypes.ListValueOf[types.String]              `tfsdk:"related_requirements"`
	SecurityControlARN          types.String                                   `tfsdk:"security_control_arn"`
	SecurityControlID           types.String                                   `tfsdk:"security_control_id"`
	StandardsARN                types.String                                   `tfsdk:"standards_arn"`
	StandardsControlDescription types.String                                   `tfsdk:"standards_control_description"`
	StandardsControlTitle       types.String                                   `tfsdk:"standards_control_title"`
	UpdatedAt                   timetypes.RFC3339                              `tfsdk:"updated_at"`
	UpdatedReason               types.String                                   `tfsdk:"updated_reason"`
}
