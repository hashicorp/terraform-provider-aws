// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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
				Required: true,
			},
			names.AttrName: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: fwschema.StringAttribute{
				Computed: true,
			},
			"policy_arn": fwschema.StringAttribute{
				Computed: true,
			},
			"regions": fwschema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]fwschema.Block{
			"permission_model": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[permissionModelModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"invoker_role_name": fwschema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
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

	svc, err := findServiceByARN(ctx, conn, data.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, svc, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	data.ARN = types.StringValue(aws.ToString(svc.ServiceArn))

	tags, err := listTags(ctx, conn, data.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}
	setTagsOut(ctx, tags.Map())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type serviceDataSourceModel struct {
	framework.WithRegionModel
	ARN             types.String                                          `tfsdk:"arn"`
	Description     types.String                                          `tfsdk:"description"`
	Name            types.String                                          `tfsdk:"name"`
	PermissionModel fwtypes.ListNestedObjectValueOf[permissionModelModel] `tfsdk:"permission_model"`
	PolicyArn       types.String                                          `tfsdk:"policy_arn"`
	Regions         fwtypes.ListOfString                                  `tfsdk:"regions"`
	Tags            tftags.Map                                            `tfsdk:"tags"`
}
