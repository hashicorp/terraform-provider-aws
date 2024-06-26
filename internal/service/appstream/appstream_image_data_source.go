// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/appstream/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"time"
	"sort"
	//"errors"

	//"regexp"

	//"github.com/aws/aws-sdk-go-v2/aws"
	//"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	//"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	//"github.com/hashicorp/aws-sdk-go-base/v2/validation"
	//"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	//"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	//"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	//"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	//"github.com/hashicorp/terraform-plugin-framework/internal/fwtype"
	//"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	//"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"

	//"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"

	//fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: ==== FILE STRUCTURE ====
// All data sources should follow this basic outline. Improve this data source's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main data source struct with schema method
// 4. Read method
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource(name="Appstream Image")
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
	resp.TypeName = "aws_appstream_appstream_image"
}

// TIP: ==== SCHEMA ====
// In the schema, add each of the arguments and attributes in snake
// case (e.g., delete_automated_backups).
// * Alphabetize arguments to make them easier to find.
// * Do not add a blank line between arguments/attributes.
//
// Users can configure argument values while attribute values cannot be
// configured and are used as output. Arguments have either:
// Required: true,
// Optional: true,
//
// All attributes will be computed and some arguments. If users will
// want to read updated information or detect drift for an argument,
// it should be computed:
// Computed: true,
//
// You will typically find arguments in the input struct
// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
// they are only in the input struct (e.g., ModifyDBInstanceInput) for
// the modify operation.
//
// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
func (d *dataSourceAppstreamImage) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional: true, 
				Computed: true,
			},
			names.AttrMostRecent : schema.BoolAttribute {
				Optional: true,
			},
			"applications": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"app_block_arn": schema.StringAttribute{
							Computed: true,
						},
						"arn": schema.StringAttribute{
							Computed: true,
						},
						"created_time": schema.Float64Attribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"display_name": schema.StringAttribute{
							Computed: true,
						},
						"enabled": schema.StringAttribute{
							Computed: true,
						},
						"icon_s3_location": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"s3_bucket": schema.StringAttribute{
										Computed: true,
									},
									"s3_key": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"icon_url": schema.StringAttribute{
							Computed: true,
						},
						"launch_parameters": schema.StringAttribute{
							Computed: true,
						},
						"launch_path": schema.StringAttribute{
							Computed: true,
						},
						"metadata": schema.MapAttribute {
							ElementType: types.StringType,
							CustomType:  fwtypes.NewMapTypeOf[types.String](ctx),
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},						
						"platforms": schema.ListAttribute{ // help
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Computed:    true,
						},
						"working_directory": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"appstream_agent_version": schema.StringAttribute{
				Optional: true,
			},
			"base_image_arn": schema.StringAttribute{
				Computed: true,
			},                                
			"created_time": schema.Float64Attribute{
				Computed: true,
			},        
			"description": schema.StringAttribute{
				Computed: true,
			},        
			"display_name": schema.StringAttribute{
				Computed: true,
			},        
			"image_builder_name": schema.StringAttribute{
				Computed: true,
			},                
			"image_builder_supported": schema.StringAttribute{
				Computed: true,
			},
			"image_errors": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"allow_fleet": schema.StringAttribute{
							Computed: true,
						},
						"allow_image_builder": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},

			"image_permissions": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"allow_fleet": schema.StringAttribute{
							Computed: true,
						},
						"allow_image_builder": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"platform": schema.StringAttribute{
				Computed: true,
			},    
			"public_base_image_released_date": schema.Float64Attribute{
				Computed: true,
			},                            
			"state": schema.StringAttribute{
				Computed: true,
			},
			"state_change_reason": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							Computed: true,
						},
						"message/": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},

			"type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.VisibilityType](),
				Optional: true,
			},
		},
	}
}

// TIP: ==== ASSIGN CRUD METHODS ====
// Data sources only have a read method.
func (d *dataSourceAppstreamImage) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// TIP: ==== DATA SOURCE READ ====
	// Generally, the Read function  do the following things. Make
	// sure there is a good reason if you should don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the config
	// 3. Get information about a resource from AWS
	// 4. Set the ID, arguments, and attributes
	// 5. Set the tags
	// 6. Set the state
	// TIP: -- 1. Get a client connection to the relevant service
	conn := d.Meta().AppStreamClient(ctx)
	
	// TIP: -- 2. Fetch the config
	var data dsImage
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var describeImagesInput appstream.DescribeImagesInput // camel case 
	resp.Diagnostics.Append(flex.Expand(ctx,&data,&describeImagesInput)...)
	
	/*
	// TIP: -- 3. Get information about a resource from AWS
	//out, err := appstream.DescribeImagesAPIClient.DescribeImages(j, ctx, &h)
	_,err := conn.DescribeImages(ctx,&describeImagesInput)
	//out,err := conn.DescribeImages(ctx, conn, data.name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppStream, create.ErrActionReading, DSNameAppstreamImage, data.Arn.String(), err),
			err.Error(),
		)
		return
	}
	// TIP: -- 4. Set the ID, arguments, and attributes
	//
	// For simple data types (i.e., schema.StringAttribute, schema.BoolAttribute,
	// schema.Int64Attribute, and schema.Float64Attribue), simply setting the  
	// appropriate data struct field is sufficient. The flex package implements
	// helpers for converting between Go and Plugin-Framework types seamlessly. No 
	// error or nil checking is necessary.
	//
	// However, there are some situations where more handling is needed such as
	// complex data types (e.g., schema.ListAttribute, schema.SetAttribute). In 
	// these cases the flatten function may have a diagnostics return value, which
	// should be appended to resp.Diagnostics.
	*/

	
	
	images, findImagesError := findImages(ctx,conn,&describeImagesInput)
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
		if names.AttrMostRecent == "false" {
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
	resp.Diagnostics.Append(flex.Flatten(ctx,image,&data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}



// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name. 
//
// Nested objects are represented in their own data struct. These will 
// also have a corresponding attribute type mapping for use inside flex
// functions.
// timetypes.RFC3339
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values

type dsApplications struct {
	AppBlockArn types.String `tfsdk:"app_block_arn"`
	Arn types.String `tfsdk:"arn"`
	CreatedTime timetypes.RFC3339 `tfsdk:"created_time"`
	Description types.String `tfsdk:"description"`
	DisplayName types.String `tfsdk:"display_name"`
	Enabled types.Bool `tfsdk:"enabled"`
	IconS3Location dsIconS3 `tfsdk:"icon_s3_location"`
	IconUrl types.String `tfsdk:"icon_url"`
	InstanceFamilies fwtypes.ListValueOf[types.String] `tfsdk:"instance_families"`
	LaunchParameters types.String `tfsdk:"launch_parameters"`
	LaunchPath types.String `tfsdk:"launch_path"`
	Metadata fwtypes.MapValueOf[types.String] `tfsdk:"metadata"`
	Name types.String `tfsdk:"name"`
	Platforms fwtypes.ListValueOf[types.String] `tfsdk:"platforms"`
	WorkingDirectory types.String `tfsdk:"working_directory"`
 }
 
 type dsIconS3 struct {
	S3Bucket types.String `tfsdk:"s3_bucket"`
	S3Key types.String `tfsdk:"s3_key"`
 }

 type dsImageErrors struct {
	ErrorCode types.String `tfsdk:"error_code"`
	ErrorMessage types.String `tfsdk:"error_message"`
	ErrorTimestamp timetypes.RFC3339 `tfsdk:"error_timestamp"`
 }

 type dsStateChange struct {
	Code types.String `tfsdk:"code"`
	Message types.String `tfsdk:"message"`
 }
 
 type dsImage struct {
	Name types.String `tfsdk:"name"`
	Applications fwtypes.ListNestedObjectValueOf[dsApplications] `tfsdk:"applications"`
	AppStreamAgentVersion types.String `tfsdk:"app_stream_agent_version"`
	Arn types.String `tfsdk:"arn"`
	BaseImageArn types.String `tfsdk:"base_image_arn"`
	CreatedTime timetypes.RFC3339 `tfsdk:"created_time"`
	Description types.String `tfsdk:"description"`
	DisplayName types.String `tfsdk:"display_name"`
	ImageBuilderName types.String `tfsdk:"image_builder_name"`
	ImageBuilderSupported types.Bool `tfsdk:"image_builder_supported"`
	ImageErrors fwtypes.ListNestedObjectValueOf[dsImageErrors] `tfsdk:"image_errors"`
	ImagePermissions fwtypes.ListNestedObjectValueOf[dsImagePermissions] `tfsdk:"image_permissions"`
	Platform types.String `tfsdk:"platform"`
	PubilcBaseImageReleasedDate timetypes.RFC3339 `tfsdk:"public_base_image_released_date"`
	State types.String `tfsdk:"state"`
	StateChangeReason fwtypes.ListNestedObjectValueOf[dsStateChange] `tfsdk:"state_change_reason"`
	Type types.String `tfsdk:"type"`
 }

 
type dsImagePermissions struct {
	AllowFleet types.Bool `tfsdk:"allow_fleet"`
	AllowImageBuilder types.Bool `tfsdk:"allow_image_builder"`
 }
