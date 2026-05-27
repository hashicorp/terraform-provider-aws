// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_s3files_access_point", name="Access Point")
func newAccessPointDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &accessPointDataSource{}, nil
}

type accessPointDataSource struct {
	framework.DataSourceWithModel[accessPointDataSourceModel]
}

func (d *accessPointDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrFileSystemID: schema.StringAttribute{
				Computed:    true,
				Description: "File system ID",
			},
			names.AttrID: schema.StringAttribute{
				Required:    true,
				Description: "Access point ID",
			},
			names.AttrName: schema.StringAttribute{
				Computed:    true,
				Description: "Access point name",
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed:    true,
				Description: "AWS account ID of the owner",
			},
			names.AttrStatus: schema.StringAttribute{
				Computed:    true,
				Description: "Access point status",
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"posix_user": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[posixUserModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"gid": schema.Int64Attribute{
							Computed:    true,
							Description: "POSIX group ID",
						},
						"secondary_gids": schema.SetAttribute{
							ElementType: types.Int64Type,
							Computed:    true,
							Description: "Secondary POSIX group IDs",
						},
						"uid": schema.Int64Attribute{
							Computed:    true,
							Description: "POSIX user ID",
						},
					},
				},
			},
			"root_directory": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[rootDirectoryModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrPath: schema.StringAttribute{
							Computed:    true,
							Description: "Root directory path",
						},
					},
					Blocks: map[string]schema.Block{
						"creation_permissions": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[creationPermissionsModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"owner_gid": schema.Int64Attribute{
										Computed:    true,
										Description: "Owner group ID",
									},
									"owner_uid": schema.Int64Attribute{
										Computed:    true,
										Description: "Owner user ID",
									},
									names.AttrPermissions: schema.StringAttribute{
										Computed:    true,
										Description: "POSIX permissions",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *accessPointDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data accessPointDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().S3FilesClient(ctx)

	output, err := findAccessPointByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	data.ARN = types.StringPointerValue(output.AccessPointArn)
	data.ID = types.StringPointerValue(output.AccessPointId)

	setTagsOut(ctx, output.Tags)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type accessPointDataSourceModel struct {
	framework.WithRegionModel
	ARN           types.String                                        `tfsdk:"arn"`
	FileSystemID  types.String                                        `tfsdk:"file_system_id"`
	ID            types.String                                        `tfsdk:"id"`
	Name          types.String                                        `tfsdk:"name"`
	OwnerID       types.String                                        `tfsdk:"owner_id"`
	PosixUser     fwtypes.ListNestedObjectValueOf[posixUserModel]     `tfsdk:"posix_user"`
	RootDirectory fwtypes.ListNestedObjectValueOf[rootDirectoryModel] `tfsdk:"root_directory"`
	Status        types.String                                        `tfsdk:"status"`
	Tags          tftags.Map                                          `tfsdk:"tags"`
}
