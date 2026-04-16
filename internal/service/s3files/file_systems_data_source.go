// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3files"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_s3files_file_systems", name="File Systems")
func newFileSystemsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &fileSystemsDataSource{}, nil
}

type fileSystemsDataSource struct {
	framework.DataSourceWithModel[fileSystemsDataSourceModel]
}

func (d *fileSystemsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
		Blocks: map[string]schema.Block{
			"file_systems": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[fileSystemsDataSourceFileSystemModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN: schema.StringAttribute{
							Computed:    true,
							Description: "File system ARN",
						},
						names.AttrBucket: schema.StringAttribute{
							Computed:    true,
							Description: "S3 bucket ARN",
						},
						names.AttrCreationTime: schema.StringAttribute{
							CustomType:  timetypes.RFC3339Type{},
							Computed:    true,
							Description: "Creation time",
						},
						names.AttrID: schema.StringAttribute{
							Computed:    true,
							Description: "File system ID",
						},
						names.AttrKMSKeyID: schema.StringAttribute{
							CustomType:  fwtypes.ARNType,
							Computed:    true,
							Description: "KMS key ID for encryption",
						},
						names.AttrName: schema.StringAttribute{
							Computed:    true,
							Description: "File system name",
						},
						names.AttrOwnerID: schema.StringAttribute{
							Computed:    true,
							Description: "AWS account ID of the owner",
						},
						names.AttrRoleARN: schema.StringAttribute{
							Computed:    true,
							Description: "IAM role ARN",
						},
						names.AttrStatus: schema.StringAttribute{
							Computed:    true,
							Description: "File system status",
						},
						names.AttrStatusMessage: schema.StringAttribute{
							Computed:    true,
							Description: "Status message",
						},
					},
				},
			},
		},
	}
}

func (d *fileSystemsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data fileSystemsDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().S3FilesClient(ctx)

	input := s3files.ListFileSystemsInput{}
	output, err := conn.ListFileSystems(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output.FileSystems, &data.FileSystems))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type fileSystemsDataSourceModel struct {
	framework.WithRegionModel
	FileSystems fwtypes.ListNestedObjectValueOf[fileSystemsDataSourceFileSystemModel] `tfsdk:"file_systems"`
}

type fileSystemsDataSourceFileSystemModel struct {
	ARN           types.String      `tfsdk:"arn"`
	Bucket        types.String      `tfsdk:"bucket"`
	CreationTime  timetypes.RFC3339 `tfsdk:"creation_time"`
	ID            types.String      `tfsdk:"id"`
	KmsKeyId      fwtypes.ARN       `tfsdk:"kms_key_id"`
	Name          types.String      `tfsdk:"name"`
	OwnerID       types.String      `tfsdk:"owner_id"`
	RoleArn       types.String      `tfsdk:"role_arn"`
	Status        types.String      `tfsdk:"status"`
	StatusMessage types.String      `tfsdk:"status_message"`
}
