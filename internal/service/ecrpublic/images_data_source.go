// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ecrpublic_images", name="Images")
func newDataSourceImages(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceImages{}, nil
}

const (
	DSNameImages = "Images Data Source"
)

type dataSourceImages struct {
	framework.DataSourceWithModel[dataSourceImagesModel]
}

func (d *dataSourceImages) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides details about AWS ECR Public Images in a public repository.",
		Attributes: map[string]schema.Attribute{
			names.AttrRepositoryName: schema.StringAttribute{
				Description: "Name of the public repository.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 205),
				},
			},
			"registry_id": schema.StringAttribute{
				Description: "The AWS account ID associated with the public registry that contains the repository. If not specified, the default public registry is assumed.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9]{12}$`), "must be a 12-digit AWS account ID"),
				},
			},
			"images":     framework.DataSourceComputedListOfObjectAttribute[imageItemModel](ctx),
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"image_ids": schema.ListNestedBlock{
				Description: "List of image IDs to filter. Each image ID can use either a tag or digest.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[imagesIDsModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"image_tag": schema.StringAttribute{
							Description: "Image tag.",
							Optional:    true,
						},
						"image_digest": schema.StringAttribute{
							Description: "Image digest.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceImages) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceImagesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ECRPublicClient(ctx)

	input := &ecrpublic.DescribeImagesInput{
		RepositoryName: data.RepositoryName.ValueStringPointer(),
	}

	if !data.RegistryID.IsNull() {
		input.RegistryId = data.RegistryID.ValueStringPointer()
	}

	if !data.ImageIDs.IsNull() {
		imageIDs, diags := data.ImageIDs.ToSlice(ctx)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(imageIDs) > 0 {
			imageIds := make([]awstypes.ImageIdentifier, 0, len(imageIDs))

			for _, id := range imageIDs {
				identifier := awstypes.ImageIdentifier{}

				if !id.ImageTag.IsNull() {
					identifier.ImageTag = id.ImageTag.ValueStringPointer()
				}

				if !id.ImageDigest.IsNull() {
					identifier.ImageDigest = id.ImageDigest.ValueStringPointer()
				}

				if identifier.ImageTag != nil || identifier.ImageDigest != nil {
					imageIds = append(imageIds, identifier)
				}
			}

			if len(imageIds) > 0 {
				input.ImageIds = imageIds
			}
		}
	}

	images := make([]imageItemModel, 0)

	paginator := ecrpublic.NewDescribeImagesPaginator(conn, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.RepositoryName.String())
			return
		}

		for _, img := range output.ImageDetails {
			item := imageItemModel{
				ArtifactMediaType:      fwflex.StringToFramework(ctx, img.ArtifactMediaType),
				ImageDigest:            fwflex.StringToFramework(ctx, img.ImageDigest),
				ImageManifestMediaType: fwflex.StringToFramework(ctx, img.ImageManifestMediaType),
				ImageSizeInBytes:       fwflex.Int64ToFramework(ctx, img.ImageSizeInBytes),
				RegistryId:             fwflex.StringToFramework(ctx, img.RegistryId),
				RepositoryName:         fwflex.StringToFramework(ctx, img.RepositoryName),
			}

			if img.ImagePushedAt != nil {
				item.ImagePushedAt = types.StringValue(img.ImagePushedAt.Format("2006-01-02T15:04:05Z"))
			} else {
				item.ImagePushedAt = types.StringNull()
			}

			// Convert tags to ListValueOf
			item.ImageTags = fwflex.FlattenFrameworkStringValueListOfString(ctx, img.ImageTags)

			images = append(images, item)
		}
	}

	data.ID = types.StringValue(data.RepositoryName.ValueString())
	data.Images = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, images)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceImagesModel struct {
	framework.WithRegionModel
	ID             types.String                                    `tfsdk:"id"`
	RepositoryName types.String                                    `tfsdk:"repository_name"`
	RegistryID     types.String                                    `tfsdk:"registry_id"`
	ImageIDs       fwtypes.ListNestedObjectValueOf[imagesIDsModel] `tfsdk:"image_ids"`
	Images         fwtypes.ListNestedObjectValueOf[imageItemModel] `tfsdk:"images"`
}

type imagesIDsModel struct {
	ImageTag    types.String `tfsdk:"image_tag"`
	ImageDigest types.String `tfsdk:"image_digest"`
}

type imageItemModel struct {
	ArtifactMediaType      types.String                      `tfsdk:"artifact_media_type"`
	ImageDigest            types.String                      `tfsdk:"digest"`
	ImageManifestMediaType types.String                      `tfsdk:"image_manifest_media_type"`
	ImagePushedAt          types.String                      `tfsdk:"pushed_at"`
	ImageSizeInBytes       types.Int64                       `tfsdk:"size_in_bytes"`
	ImageTags              fwtypes.ListValueOf[types.String] `tfsdk:"tags"`
	RegistryId             types.String                      `tfsdk:"registry_id"`
	RepositoryName         types.String                      `tfsdk:"repository_name"`
}
