// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_verifiedpermissions_policy_store", name="Policy Store")
// @Tags(identifierAttribute="arn")
func newPolicyStoreDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &policyStoreDataSource{}, nil
}

const (
	DSNamePolicyStore = "Policy Store Data Source"
)

type policyStoreDataSource struct {
	framework.DataSourceWithModel[policyStoreDataSourceModel]
}

func (d *policyStoreDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			names.AttrLastUpdatedDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrTags:        tftags.TagsAttributeComputedOnly(),
			"validation_settings": framework.DataSourceComputedListOfObjectAttribute[validationSettingsModel](ctx),
		},
	}
}
func (d *policyStoreDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data policyStoreDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().VerifiedPermissionsClient(ctx)

	output, err := findPolicyStoreByID(ctx, conn, data.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionReading, DSNamePolicyStore, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type policyStoreDataSourceModel struct {
	framework.WithRegionModel
	ARN                types.String                                             `tfsdk:"arn"`
	CreatedDate        timetypes.RFC3339                                        `tfsdk:"created_date"`
	Description        types.String                                             `tfsdk:"description"`
	ID                 types.String                                             `tfsdk:"id"`
	LastUpdatedDate    timetypes.RFC3339                                        `tfsdk:"last_updated_date"`
	Tags               tftags.Map                                               `tfsdk:"tags"`
	ValidationSettings fwtypes.ListNestedObjectValueOf[validationSettingsModel] `tfsdk:"validation_settings"`
}

type validationSettingsModel struct {
	Mode fwtypes.StringEnum[awstypes.ValidationMode] `tfsdk:"mode"`
}
