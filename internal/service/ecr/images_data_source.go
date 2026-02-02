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
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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
			"describe_images": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to call DescribeImages API to get detailed image information",
			},
			"image_details": framework.DataSourceComputedListOfObjectAttribute[imageDetailsModel](ctx),
			"image_ids":     framework.DataSourceComputedListOfObjectAttribute[imagesIDsModel](ctx),
			"max_results": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of images to return",
			},
			"registry_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the registry (AWS account ID)",
			},
			names.AttrRepositoryName: schema.StringAttribute{
				Required:    true,
				Description: "Name of the repository",
			},
			"tag_status": schema.StringAttribute{
				Optional:    true,
				Description: "Filter images by tag status. Valid values: TAGGED, UNTAGGED, ANY",
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

	// Set tag status filter if provided
	if !data.TagStatus.IsNull() && !data.TagStatus.IsUnknown() {
		tagStatus := awstypes.TagStatus(data.TagStatus.ValueString())
		input.Filter = &awstypes.ListImagesFilter{
			TagStatus: tagStatus,
		}
	}

	// Set max results if provided
	if !data.MaxResults.IsNull() && !data.MaxResults.IsUnknown() {
		maxResults := int32(data.MaxResults.ValueInt64())
		input.MaxResults = &maxResults
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

	// If describe_images is true, call DescribeImages API
	if !data.DescribeImages.IsNull() && data.DescribeImages.ValueBool() {
		registryID := ""
		if !data.RegistryID.IsNull() {
			registryID = data.RegistryID.ValueString()
		}

		imageDetails, err := findImagesDetails(ctx, conn, data.RepositoryName.ValueString(), registryID, output)
		if err != nil {
			resp.Diagnostics.AddError("describing ECR Images", err.Error())
			return
		}

		resp.Diagnostics.Append(fwflex.Flatten(ctx, imageDetails, &data.ImageDetails)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findImagesDetails(ctx context.Context, conn *ecr.Client, repositoryName, registryID string, imageIds []awstypes.ImageIdentifier) ([]awstypes.ImageDetail, error) {
	var output []awstypes.ImageDetail

	// DescribeImages has a limit of 100 images per request
	const batchSize = 100
	for i := 0; i < len(imageIds); i += batchSize {
		end := min(i+batchSize, len(imageIds))

		input := &ecr.DescribeImagesInput{
			RepositoryName: &repositoryName,
			ImageIds:       imageIds[i:end],
		}
		if registryID != "" {
			input.RegistryId = &registryID
		}

		result, err := conn.DescribeImages(ctx, input)
		if err != nil {
			return nil, err
		}

		output = append(output, result.ImageDetails...)
	}

	return output, nil
}

func findImages(ctx context.Context, conn *ecr.Client, input *ecr.ListImagesInput) ([]awstypes.ImageIdentifier, error) {
	var output []awstypes.ImageIdentifier

	paginator := ecr.NewListImagesPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if errs.IsA[*awstypes.RepositoryNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: input,
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
	DescribeImages types.Bool                                         `tfsdk:"describe_images"`
	ImageDetails   fwtypes.ListNestedObjectValueOf[imageDetailsModel] `tfsdk:"image_details"`
	ImageIDs       fwtypes.ListNestedObjectValueOf[imagesIDsModel]    `tfsdk:"image_ids"`
	MaxResults     types.Int64                                        `tfsdk:"max_results"`
	RegistryID     types.String                                       `tfsdk:"registry_id"`
	RepositoryName types.String                                       `tfsdk:"repository_name"`
	TagStatus      types.String                                       `tfsdk:"tag_status"`
}

type imageDetailsModel struct {
	ImageDigest      types.String                      `tfsdk:"image_digest"`
	ImagePushedAt    types.String                      `tfsdk:"image_pushed_at"`
	ImageSizeInBytes types.Int64                       `tfsdk:"image_size_in_bytes"`
	ImageTags        fwtypes.ListValueOf[types.String] `tfsdk:"image_tags"`
	RegistryID       types.String                      `tfsdk:"registry_id"`
	RepositoryName   types.String                      `tfsdk:"repository_name"`
}

type imagesIDsModel struct {
	ImageDigest types.String `tfsdk:"image_digest"`
	ImageTag    types.String `tfsdk:"image_tag"`
}
