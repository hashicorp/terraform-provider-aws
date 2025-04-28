// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_appstream_image", name="Image")
func newImageDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &imageDataSource{}, nil
}

type imageDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *imageDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"applications": framework.DataSourceComputedListOfObjectAttribute[applicationModel](ctx),
			"appstream_agent_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				Optional:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot(names.AttrName), path.MatchRoot("name_regex")),
				},
			},
			"base_image_arn": schema.StringAttribute{
				Computed: true,
			},
			names.AttrCreatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrDisplayName: schema.StringAttribute{
				Computed: true,
			},
			"image_builder_name": schema.StringAttribute{
				Computed: true,
			},
			"image_builder_supported": schema.BoolAttribute{
				Computed: true,
			},
			"image_permissions": framework.DataSourceComputedListOfObjectAttribute[imagePermissionsModel](ctx),
			names.AttrMostRecent: schema.BoolAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot(names.AttrARN), path.MatchRoot("name_regex")),
				},
			},
			"name_regex": schema.StringAttribute{
				CustomType: fwtypes.RegexpType,
				Optional:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot(names.AttrARN), path.MatchRoot(names.AttrName)),
				},
			},
			"platform": schema.StringAttribute{
				Computed: true,
			},
			"public_base_image_released_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			"state_change_reason": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[imageStateChangeReasonModel](ctx),
				Computed:   true,
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VisibilityType](),
				Optional:   true,
			},
		},
	}
}

func (d *imageDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data imageDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().AppStreamClient(ctx)

	var input appstream.DescribeImagesInput
	if !data.ARN.IsNull() {
		input.Arns = []string{data.ARN.ValueString()}
	}
	if !data.Name.IsNull() {
		input.Names = []string{data.Name.ValueString()}
	}

	images, err := findImages(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError("reading AppStream Images", err.Error())

		return
	}

	if !data.NameRegex.IsNull() {
		r := data.NameRegex.ValueRegexp()
		images = tfslices.Filter(images, func(v awstypes.Image) bool {
			name := aws.ToString(v.Name)
			// Check for a very rare case where the response would include no
			// image name. No name means nothing to attempt a match against,
			// therefore we are skipping such image.
			return name != "" && r.MatchString(name)
		})
	}

	switch l := len(images); l {
	case 0:
		err = tfresource.NewEmptyResultError(input)
	case 1:
		// OK
	default:
		if data.MostRecent.ValueBool() {
			slices.SortFunc(images, func(a, b awstypes.Image) int {
				if aws.ToTime(a.CreatedTime).After(aws.ToTime(b.CreatedTime)) {
					return -1
				}
				if aws.ToTime(a.CreatedTime).Before(aws.ToTime(b.CreatedTime)) {
					return 1
				}
				return 0
			})
		} else {
			err = tfresource.NewTooManyResultsError(l, input)
		}
	}

	if err != nil {
		response.Diagnostics.AddError("reading AppStream Images", tfresource.SingularDataSourceFindError("AppStream Image", err).Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, &images[0], &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findImages(ctx context.Context, conn *appstream.Client, input *appstream.DescribeImagesInput) ([]awstypes.Image, error) {
	var output []awstypes.Image

	pages := appstream.NewDescribeImagesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Images...)
	}

	return output, nil
}

type imageDataSourceModel struct {
	Applications                fwtypes.ListNestedObjectValueOf[applicationModel]            `tfsdk:"applications"`
	AppStreamAgentVersion       types.String                                                 `tfsdk:"appstream_agent_version"`
	ARN                         fwtypes.ARN                                                  `tfsdk:"arn"`
	BaseImageARN                types.String                                                 `tfsdk:"base_image_arn"`
	CreatedTime                 timetypes.RFC3339                                            `tfsdk:"created_time"`
	Description                 types.String                                                 `tfsdk:"description"`
	DisplayName                 types.String                                                 `tfsdk:"display_name"`
	ImageBuilderName            types.String                                                 `tfsdk:"image_builder_name"`
	ImageBuilderSupported       types.Bool                                                   `tfsdk:"image_builder_supported"`
	ImagePermissions            fwtypes.ListNestedObjectValueOf[imagePermissionsModel]       `tfsdk:"image_permissions"`
	MostRecent                  types.Bool                                                   `tfsdk:"most_recent"`
	Name                        types.String                                                 `tfsdk:"name"`
	NameRegex                   fwtypes.Regexp                                               `tfsdk:"name_regex"`
	Platform                    types.String                                                 `tfsdk:"platform"`
	PublicBaseImageReleasedDate timetypes.RFC3339                                            `tfsdk:"public_base_image_released_date"`
	State                       types.String                                                 `tfsdk:"state"`
	StateChangeReason           fwtypes.ListNestedObjectValueOf[imageStateChangeReasonModel] `tfsdk:"state_change_reason"`
	VisibilityType              fwtypes.StringEnum[awstypes.VisibilityType]                  `tfsdk:"type"`
}

type applicationModel struct {
	AppBlockARN      types.String                                     `tfsdk:"app_block_arn"`
	ARN              fwtypes.ARN                                      `tfsdk:"arn"`
	CreatedTime      timetypes.RFC3339                                `tfsdk:"created_time"`
	Description      types.String                                     `tfsdk:"description"`
	DisplayName      types.String                                     `tfsdk:"display_name"`
	Enabled          types.Bool                                       `tfsdk:"enabled"`
	IconS3Location   fwtypes.ListNestedObjectValueOf[s3LocationModel] `tfsdk:"icon_s3_location"`
	IconURL          types.String                                     `tfsdk:"icon_url"`
	InstanceFamilies fwtypes.ListOfString                             `tfsdk:"instance_families"`
	LaunchParameters types.String                                     `tfsdk:"launch_parameters"`
	LaunchPath       types.String                                     `tfsdk:"launch_path"`
	Metadata         fwtypes.MapOfString                              `tfsdk:"metadata"`
	Name             types.String                                     `tfsdk:"name"`
	Platforms        fwtypes.ListOfString                             `tfsdk:"platforms"`
	WorkingDirectory types.String                                     `tfsdk:"working_directory"`
}

type s3LocationModel struct {
	S3Bucket types.String `tfsdk:"s3_bucket"`
	S3Key    types.String `tfsdk:"s3_key"`
}

type imageStateChangeReasonModel struct {
	Code    types.String `tfsdk:"code"`
	Message types.String `tfsdk:"message"`
}

type imagePermissionsModel struct {
	AllowFleet        types.Bool `tfsdk:"allow_fleet"`
	AllowImageBuilder types.Bool `tfsdk:"allow_image_builder"`
}
