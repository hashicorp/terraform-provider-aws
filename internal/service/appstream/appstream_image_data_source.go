// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/hashicorp/terraform-provider-aws/internal/create"

	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"

	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func newDataSourceAppstreamImage(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAppstreamImage{}, nil
}

const (
	DSNameAppstreamImage = "Appstream Image Data Source"
)

type dataSourceAppstreamImage struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceAppstreamImage) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_appstream_image"
}

func (d *dataSourceAppstreamImage) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{

			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				Optional:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("name"),
					}...),
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("name_regex"),
					}...),
				},
			},
			"applications": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dsApplications](ctx),
				Computed:   true,
			},
			"appstream_agent_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrMostRecent: schema.BoolAttribute{
				Optional: true,
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
			"image_errors": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dsImageErrors](ctx),
				Computed:   true,
			},
			"image_permissions": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dsImagePermissions](ctx),
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot(names.AttrARN),
					}...),
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("name_regex"),
					}...),
				},
			},
			"name_regex": schema.StringAttribute{
				CustomType: fwtypes.RegexpType,
				Optional:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot(names.AttrName),
					}...),
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot(names.AttrARN),
					}...),
				},
			},

			"platform": schema.StringAttribute{
				Computed: true,
			},
			"public_base_image_released_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"state_change_reason": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dsStateChange](ctx),
				Computed:   true,
			},
			"type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VisibilityType](),
				Optional:   true,
			},
		},
	}
}

func (d *dataSourceAppstreamImage) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().AppStreamClient(ctx)

	var data dsImage
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var describeImagesInput appstream.DescribeImagesInput // camel case
	if !data.Name.IsNull() {
		describeImagesInput.Names = []string{data.Name.ValueString()}
	}
	if !data.Arn.IsNull() {
		describeImagesInput.Arns = []string{data.Arn.ValueString()}
	}

	images, findImagesError := findImages(ctx, conn, &describeImagesInput)
	if findImagesError != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppStream, create.ErrActionReading, DSNameAppstreamImage, data.Arn.String(), findImagesError),
			findImagesError.Error(),
		)
		return
	}
	if len(images) < 1 {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppStream, create.ErrActionReading, DSNameAppstreamImage, data.Arn.String(), findImagesError),
			"Your query returned no results. Please change your search criteria and try again.",
		)
		return
	}
	if len(images) > 1 {
		if !data.MostRecent.ValueBool() {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AppStream, create.ErrActionReading, DSNameAppstreamImage, data.Arn.String(), findImagesError),
				"Your query returned more than one result. Please try a more specific search criteria, or set `most_recent` attribute to true.",
			)
			return
		}
		sort.Slice(images, func(i, j int) bool {
			itime, _ := time.Parse(time.RFC3339, images[i].CreatedTime.Month().String())
			jtime, _ := time.Parse(time.RFC3339, images[j].CreatedTime.Month().String())
			return itime.Unix() > jtime.Unix()
		})
	}
	image := images[0]

	// TIP: -- 6. Set the state
	data.Type = fwtypes.StringEnum[awstypes.VisibilityType]{}.StringEnumValue(string(image.Visibility))
	resp.Diagnostics.Append(flex.Flatten(ctx, &image, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if image.PublicBaseImageReleasedDate != nil {
		data.PubilcBaseImageReleasedDate = timetypes.NewRFC3339TimeValue(*image.PublicBaseImageReleasedDate)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dsApplications struct {
	AppBlockArn      types.String                              `tfsdk:"app_block_arn"`
	Arn              fwtypes.ARN                               `tfsdk:"arn"`
	CreatedTime      timetypes.RFC3339                         `tfsdk:"created_time"`
	Description      types.String                              `tfsdk:"description"`
	DisplayName      types.String                              `tfsdk:"display_name"`
	Enabled          types.Bool                                `tfsdk:"enabled"`
	IconS3Location   fwtypes.ListNestedObjectValueOf[dsIconS3] `tfsdk:"icon_s3_location"`
	IconUrl          types.String                              `tfsdk:"icon_url"`
	InstanceFamilies fwtypes.ListValueOf[types.String]         `tfsdk:"instance_families"`
	LaunchParameters types.String                              `tfsdk:"launch_parameters"`
	LaunchPath       types.String                              `tfsdk:"launch_path"`
	Metadata         fwtypes.MapValueOf[types.String]          `tfsdk:"metadata"`
	Name             types.String                              `tfsdk:"name"`
	Platforms        fwtypes.ListValueOf[types.String]         `tfsdk:"platforms"`
	WorkingDirectory types.String                              `tfsdk:"working_directory"`
}

type dsIconS3 struct {
	S3Bucket types.String `tfsdk:"s3_bucket"`
	S3Key    types.String `tfsdk:"s3_key"`
}

type dsImageErrors struct {
	ErrorCode      types.String      `tfsdk:"error_code"`
	ErrorMessage   types.String      `tfsdk:"error_message"`
	ErrorTimestamp timetypes.RFC3339 `tfsdk:"error_timestamp"`
}

type dsStateChange struct {
	Code    types.String `tfsdk:"code"`
	Message types.String `tfsdk:"message"`
}

type dsImage struct {
	Applications                fwtypes.ListNestedObjectValueOf[dsApplications]     `tfsdk:"applications"`
	AppStreamAgentVersion       types.String                                        `tfsdk:"appstream_agent_version"`
	Arn                         fwtypes.ARN                                         `tfsdk:"arn"`
	BaseImageArn                types.String                                        `tfsdk:"base_image_arn"`
	CreatedTime                 timetypes.RFC3339                                   `tfsdk:"created_time"`
	Description                 types.String                                        `tfsdk:"description"`
	DisplayName                 types.String                                        `tfsdk:"display_name"`
	ImageBuilderName            types.String                                        `tfsdk:"image_builder_name"`
	ImageBuilderSupported       types.Bool                                          `tfsdk:"image_builder_supported"`
	ImageErrors                 fwtypes.ListNestedObjectValueOf[dsImageErrors]      `tfsdk:"image_errors"`
	ImagePermissions            fwtypes.ListNestedObjectValueOf[dsImagePermissions] `tfsdk:"image_permissions"`
	MostRecent                  types.Bool                                          `tfsdk:"most_recent"`
	Name                        types.String                                        `tfsdk:"name"`
	NameRegex                   fwtypes.Regexp                                      `tfsdk:"name_regex"`
	Platform                    types.String                                        `tfsdk:"platform"`
	PubilcBaseImageReleasedDate timetypes.RFC3339                                   `tfsdk:"public_base_image_released_date"`
	State                       types.String                                        `tfsdk:"state"`
	StateChangeReason           fwtypes.ListNestedObjectValueOf[dsStateChange]      `tfsdk:"state_change_reason"`
	Type                        fwtypes.StringEnum[awstypes.VisibilityType]         `tfsdk:"type"`
}

type dsImagePermissions struct {
	AllowFleet        types.Bool `tfsdk:"allow_fleet"`
	AllowImageBuilder types.Bool `tfsdk:"allow_image_builder"`
}
