// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic

import (
	"context"
	"fmt"
	"regexp"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ecrpublic_image", name="Image")
func newDataSourceImage(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceImage{}, nil
}

const (
	DSNameImage = "Image Data Source"
)

type dataSourceImage struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceImage) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"image_digest": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"image_pushed_at": schema.Int64Attribute{
				Computed: true,
			},
			"image_size_in_bytes": schema.Int64Attribute{
				Computed: true,
			},
			"image_tag": schema.StringAttribute{
				Optional: true,
			},
			"image_tags": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"image_uri": schema.StringAttribute{
				Computed: true,
			},
			names.AttrMostRecent: schema.BoolAttribute{
				Optional: true,
			},
			"registry_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile("[0-9]{12}"), "must satisfy regular expression [0-9]{12}"),
				},
			},
			names.AttrRepositoryName: schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *dataSourceImage) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ECRPublicClient(ctx)

	var data dataSourceImageModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := ecrpublic.DescribeImagesInput{
		RepositoryName: data.RepositoryName.ValueStringPointer(),
	}
	if !data.RegistryId.IsNull() {
		input.RegistryId = data.RegistryId.ValueStringPointer()
	}
	if !data.ImageDigest.IsNull() {
		input.ImageIds = []awstypes.ImageIdentifier{
			{ImageDigest: data.ImageDigest.ValueStringPointer()},
		}
	}
	if !data.ImageTag.IsNull() {
		if input.ImageIds == nil {
			input.ImageIds = []awstypes.ImageIdentifier{
				{ImageTag: data.ImageTag.ValueStringPointer()},
			}
		} else {
			input.ImageIds[0].ImageTag = data.ImageTag.ValueStringPointer()
		}
	}

	imageDetails, err := findImageDetails(ctx, conn, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ECRPublic, create.ErrActionReading, DSNameImage, "", err),
			err.Error(),
		)
		return
	}
	if len(imageDetails) == 0 {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ECRPublic, create.ErrActionReading, DSNameImage, "", err),
			"Your query returned no results. Please change your search criteria and try again.")
		return
	} else if len(imageDetails) > 1 && data.MostRecent.ValueBool() == false {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ECRPublic, create.ErrActionReading, DSNameImage, "", err),
			"Your query returned more than one result. Please try a more specific search criteria, or set `most_recent` attribute to true.")
		return
	} else if len(imageDetails) > 1 && data.MostRecent.ValueBool() == true {
		sort.Slice(imageDetails, func(i, j int) bool {
			return imageDetails[i].ImagePushedAt.After(*imageDetails[j].ImagePushedAt)
		})
	}
	imageDetail := imageDetails[0]

	repository, err := findRepositoryDetail(ctx, conn, data.RepositoryName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ECRPublic, create.ErrActionReading, DSNameImage, "", err),
			err.Error(),
		)
		return
	}

	data.Id = types.StringPointerValue(imageDetail.ImageDigest)
	data.ImagePushedAt = types.Int64Value(imageDetail.ImagePushedAt.Unix())
	data.ImageURI = types.StringValue(fmt.Sprintf("%s@%s", *repository.RepositoryUri, *imageDetail.ImageDigest))

	resp.Diagnostics.Append(flex.Flatten(ctx, imageDetail, &data, flex.WithIgnoredFieldNames([]string{"ImagePushedAt"}))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *dataSourceImage) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("image_digest"),
			path.MatchRoot("image_tag"),
			path.MatchRoot(names.AttrMostRecent),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot("image_tag"),
			path.MatchRoot(names.AttrMostRecent),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot("image_digest"),
			path.MatchRoot(names.AttrMostRecent),
		),
	}
}

func findImageDetails(ctx context.Context, conn *ecrpublic.Client, input *ecrpublic.DescribeImagesInput) ([]awstypes.ImageDetail, error) {
	var output []awstypes.ImageDetail

	pages := ecrpublic.NewDescribeImagesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		output = append(output, page.ImageDetails...)
	}

	return output, nil
}

func findRepositoryDetail(ctx context.Context, conn *ecrpublic.Client, repositoryName string) (awstypes.Repository, error) {
	var output awstypes.Repository

	repositoryDetails, err := conn.DescribeRepositories(ctx, &ecrpublic.DescribeRepositoriesInput{
		RepositoryNames: []string{repositoryName},
	})
	if err != nil {
		return output, err
	}
	output = repositoryDetails.Repositories[0]
	return output, nil
}

type dataSourceImageModel struct {
	RepositoryName   types.String `tfsdk:"repository_name"`
	RegistryId       types.String `tfsdk:"registry_id"`
	Id               types.String `tfsdk:"id"`
	MostRecent       types.Bool   `tfsdk:"most_recent"`
	ImageDigest      types.String `tfsdk:"image_digest"`
	ImageTag         types.String `tfsdk:"image_tag"`
	ImagePushedAt    types.Int64  `tfsdk:"image_pushed_at"`
	ImageSizeInBytes types.Int64  `tfsdk:"image_size_in_bytes"`
	ImageTags        types.List   `tfsdk:"image_tags"`
	ImageURI         types.String `tfsdk:"image_uri"`
}
