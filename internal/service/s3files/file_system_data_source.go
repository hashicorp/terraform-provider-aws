// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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

// @FrameworkDataSource("aws_s3files_file_system", name="File System")
func newFileSystemDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &fileSystemDataSource{}, nil
}

type fileSystemDataSource struct {
	framework.DataSourceWithModel[fileSystemDataSourceModel]
}

func (d *fileSystemDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
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
				Required:    true,
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
			names.AttrPrefix: schema.StringAttribute{
				Computed:    true,
				Description: "S3 bucket prefix",
			},
			names.AttrRoleARN: schema.StringAttribute{
				Computed:    true,
				Description: "IAM role ARN for S3 access",
			},
			names.AttrStatus: schema.StringAttribute{
				Computed:    true,
				Description: "File system status",
			},
			names.AttrStatusMessage: schema.StringAttribute{
				Computed:    true,
				Description: "Status message",
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *fileSystemDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data fileSystemDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().S3FilesClient(ctx)

	output, err := findFileSystemByID(ctx, conn, data.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	data.ARN = types.StringPointerValue(output.FileSystemArn)
	data.ID = types.StringPointerValue(output.FileSystemId)

	setTagsOut(ctx, output.Tags)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type fileSystemDataSourceModel struct {
	framework.WithRegionModel
	ARN           types.String      `tfsdk:"arn"`
	Bucket        types.String      `tfsdk:"bucket"`
	CreationTime  timetypes.RFC3339 `tfsdk:"creation_time"`
	ID            types.String      `tfsdk:"id"`
	KmsKeyId      fwtypes.ARN       `tfsdk:"kms_key_id"`
	Name          types.String      `tfsdk:"name"`
	OwnerID       types.String      `tfsdk:"owner_id"`
	Prefix        types.String      `tfsdk:"prefix" autoflex:",omitempty"`
	RoleArn       types.String      `tfsdk:"role_arn"`
	Status        types.String      `tfsdk:"status"`
	StatusMessage types.String      `tfsdk:"status_message"`
	Tags          tftags.Map        `tfsdk:"tags"`
}
