// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ecr_images", name="Images")
func newImagesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &imagesDataSource{}, nil
}

type imagesDataSource struct {
	framework.DataSourceWithModel[imagesDataSourceModel]
}

func (d *imagesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"image_ids": framework.DataSourceComputedListOfObjectAttribute[imagesIDsModel](ctx),
			"registry_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the registry (AWS account ID)",
			},
			names.AttrRepositoryName: schema.StringAttribute{
				Required:    true,
				Description: "Name of the repository",
			},
		},
	}
}

func (d *imagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data imagesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ECRClient(ctx)

	var input ecr.ListImagesInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, &data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := findImages(ctx, conn, &input)
	if err != nil {
		resp.Diagnostics.AddError("reading ECR Images", err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.ImageIDs)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findImages(ctx context.Context, conn *ecr.Client, input *ecr.ListImagesInput) ([]awstypes.ImageIdentifier, error) {
	var output []awstypes.ImageIdentifier

	paginator := ecr.NewListImagesPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if errs.IsA[*awstypes.RepositoryNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ImageIds...)
	}

	return output, nil
}

type imagesDataSourceModel struct {
	framework.WithRegionModel
	ImageIDs       fwtypes.ListNestedObjectValueOf[imagesIDsModel] `tfsdk:"image_ids"`
	RegistryID     types.String                                    `tfsdk:"registry_id"`
	RepositoryName types.String                                    `tfsdk:"repository_name"`
}

type imagesIDsModel struct {
	ImageDigest types.String `tfsdk:"image_digest"`
	ImageTag    types.String `tfsdk:"image_tag"`
}
