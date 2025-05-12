// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ecr_images", name="Images")
func newImagesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &imagesDataSource{}, nil
}

type imagesDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *imagesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrRepositoryName: schema.StringAttribute{
				Required:    true,
				Description: "Name of the repository",
			},
			"registry_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the registry (AWS account ID)",
			},
			"image_ids": schema.ListAttribute{
				Computed:    true,
				Description: "List of image IDs in the repository",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"image_digest": types.StringType,
						"image_tag":    types.StringType,
					},
				},
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

	input := &ecr.ListImagesInput{
		RepositoryName: data.RepositoryName.ValueStringPointer(),
	}

	if !data.RegistryID.IsNull() && !data.RegistryID.IsUnknown() {
		input.RegistryId = data.RegistryID.ValueStringPointer()
	}

	output, err := findImages(ctx, conn, input)

	if err != nil {
		resp.Diagnostics.AddError("reading ECR Images", err.Error())
		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, data.RepositoryName.ValueString())

	imageIDs := make([]attr.Value, 0, len(output))
	for _, imageID := range output {
		imageIDAttrs := map[string]attr.Value{
			"image_digest": types.StringNull(),
			"image_tag":    types.StringNull(),
		}

		if imageID.ImageDigest != nil {
			imageIDAttrs["image_digest"] = fwflex.StringValueToFramework(ctx, *imageID.ImageDigest)
		}

		if imageID.ImageTag != nil {
			imageIDAttrs["image_tag"] = fwflex.StringValueToFramework(ctx, *imageID.ImageTag)
		}

		imageIDObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"image_digest": types.StringType,
				"image_tag":    types.StringType,
			},
			imageIDAttrs,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		imageIDs = append(imageIDs, imageIDObj)
	}

	imageIDsList, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"image_digest": types.StringType,
				"image_tag":    types.StringType,
			},
		},
		imageIDs,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ImageIDs = imageIDsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findImages(ctx context.Context, conn *ecr.Client, input *ecr.ListImagesInput) ([]awstypes.ImageIdentifier, error) {
	var output []awstypes.ImageIdentifier

	pages := ecr.NewListImagesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.RepositoryNotFoundException](err) {
			return nil, &retry.NotFoundError{
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
	ID             types.String `tfsdk:"id"`
	RepositoryName types.String `tfsdk:"repository_name"`
	RegistryID     types.String `tfsdk:"registry_id"`
	ImageIDs       types.List   `tfsdk:"image_ids"`
}
