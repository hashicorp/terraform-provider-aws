// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_resiliencehubv2_service", name="Service")
// @Tags(identifierAttribute="arn")
func newServiceDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &serviceDataSource{}, nil
}

type serviceDataSource struct {
	framework.DataSourceWithModel[serviceDataSourceModel]
}

func (d *serviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"associated_system": framework.DataSourceComputedListOfObjectAttribute[associatedSystemModel](ctx),
			names.AttrDescription: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrKMSKeyID: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrName: fwschema.StringAttribute{
				Computed: true,
			},
			"permission_model": framework.DataSourceComputedListOfObjectAttribute[permissionModelModel](ctx),
			"policy_arn": fwschema.StringAttribute{
				Computed: true,
			},
			"regions": fwschema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *serviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serviceDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ResilienceHubV2Client(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ServiceARN)
	svc, err := findServiceByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, svc, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, svc.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type serviceDataSourceModel struct {
	framework.WithRegionModel
	AssociatedSystems fwtypes.ListNestedObjectValueOf[associatedSystemModel] `tfsdk:"associated_system"`
	Description       types.String                                           `tfsdk:"description"`
	KMSKeyID          types.String                                           `tfsdk:"kms_key_id"`
	Name              types.String                                           `tfsdk:"name"`
	PermissionModel   fwtypes.ListNestedObjectValueOf[permissionModelModel]  `tfsdk:"permission_model"`
	PolicyARN         types.String                                           `tfsdk:"policy_arn"`
	Regions           fwtypes.ListOfString                                   `tfsdk:"regions"`
	ServiceARN        fwtypes.ARN                                            `tfsdk:"arn"`
	Tags              tftags.Map                                             `tfsdk:"tags"`
}
