// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_securityhub_standards_control_associations", name="Standards Control Associations")
func newStandardsControlAssociationsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &standardsControlAssociationsDataSource{}

	return d, nil
}

type standardsControlAssociationsDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *standardsControlAssociationsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"security_control_id": schema.StringAttribute{
				Required: true,
			},
			"standards_control_associations": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[standardsControlAssociationData](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[standardsControlAssociationData](ctx),
				},
			},
		},
	}
}

func (d *standardsControlAssociationsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data standardsControlAssociationsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().SecurityHubClient(ctx)

	input := &securityhub.ListStandardsControlAssociationsInput{
		SecurityControlId: data.SecurityControlID.ValueStringPointer(),
	}

	out, err := findStandardsControlAssociations(ctx, conn, input, tfslices.PredicateTrue[*awstypes.StandardsControlAssociationSummary]())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SecurityHub Standards Control Associations (%s)", data.SecurityControlID.ValueString()), err.Error())

		return
	}

	data.ID = types.StringValue(d.Meta().Region(ctx))
	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data.StandardsControlAssociations)...)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type standardsControlAssociationsDataSourceModel struct {
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
