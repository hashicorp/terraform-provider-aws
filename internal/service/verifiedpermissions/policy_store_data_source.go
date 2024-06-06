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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Policy Store")
func newDataSourcePolicyStore(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourcePolicyStore{}, nil
}

const (
	DSNamePolicyStore = "Policy Store Data Source"
)

type dataSourcePolicyStore struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourcePolicyStore) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_verifiedpermissions_policy_store"
}

func (d *dataSourcePolicyStore) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
			"validation_settings": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[validationSettingsDataSource](ctx),
				ElementType: fwtypes.NewObjectTypeOf[validationSettingsDataSource](ctx),
				Computed:    true,
			},
		},
	}
}
func (d *dataSourcePolicyStore) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().VerifiedPermissionsClient(ctx)

	var data dataSourcePolicyStoreData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPolicyStoreByID(ctx, conn, data.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionReading, DSNamePolicyStore, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourcePolicyStoreData struct {
	ARN                types.String                                                  `tfsdk:"arn"`
	CreatedDate        timetypes.RFC3339                                             `tfsdk:"created_date"`
	Description        types.String                                                  `tfsdk:"description"`
	ID                 types.String                                                  `tfsdk:"id"`
	LastUpdatedDate    timetypes.RFC3339                                             `tfsdk:"last_updated_date"`
	ValidationSettings fwtypes.ListNestedObjectValueOf[validationSettingsDataSource] `tfsdk:"validation_settings"`
}

type validationSettingsDataSource struct {
	Mode fwtypes.StringEnum[awstypes.ValidationMode] `tfsdk:"mode"`
}
