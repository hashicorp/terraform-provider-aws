// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecrpublic

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
func newDataSourceImages(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceImages{}, nil
}

type dataSourceImages struct {
	framework.DataSourceWithModel[dataSourceImagesModel]
}

func (d *dataSourceImages) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
				Description: "AWS account ID associated with the public registry that contains the repository. If not specified, the default public registry is assumed.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9]{12}$`), "must be a 12-digit AWS account ID"),
				},
			},
			"images": framework.DataSourceComputedListOfObjectAttribute[imageItemModel](ctx),
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

	var input ecrpublic.DescribeImagesInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, &data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var images []awstypes.ImageDetail

	paginator := ecrpublic.NewDescribeImagesPaginator(conn, &input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.RepositoryName.String())
			return
		}

		images = append(images, output.ImageDetails...)
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, images, &data.Images)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceImagesModel struct {
	framework.WithRegionModel
	RepositoryName types.String                                    `tfsdk:"repository_name"`
	RegistryId     types.String                                    `tfsdk:"registry_id"`
	ImageIds       fwtypes.ListNestedObjectValueOf[imagesIDsModel] `tfsdk:"image_ids"`
	Images         fwtypes.ListNestedObjectValueOf[imageItemModel] `tfsdk:"images"`
}

type imagesIDsModel struct {
	ImageTag    types.String `tfsdk:"image_tag"`
	ImageDigest types.String `tfsdk:"image_digest"`
}

type imageItemModel struct {
	ArtifactMediaType      types.String                      `tfsdk:"artifact_media_type"`
	ImageDigest            types.String                      `tfsdk:"image_digest"`
	ImageManifestMediaType types.String                      `tfsdk:"image_manifest_media_type"`
	ImagePushedAt          timetypes.RFC3339                 `tfsdk:"image_pushed_at"`
	ImageSizeInBytes       types.Int64                       `tfsdk:"image_size_in_bytes"`
	ImageTags              fwtypes.ListValueOf[types.String] `tfsdk:"image_tags"`
	RegistryId             types.String                      `tfsdk:"registry_id"`
	RepositoryName         types.String                      `tfsdk:"repository_name"`
}
