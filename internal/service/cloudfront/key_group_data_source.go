// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

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
	// using the services/cloudfront/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	/// awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	/// fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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
// @FrameworkDataSource(name="Key Group")
func newDataSourceKeyGroup(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceKeyGroup{}, nil
}

const (
	DSNameKeyGroup = "Key Group Data Source"
)

type dataSourceKeyGroup struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceKeyGroup) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_cloudfront_key_group"
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
func (d *dataSourceKeyGroup) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName(names.AttrID),
						path.MatchRelative().AtParent().AtName(names.AttrName),
					),
				},
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"items": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrComment: schema.StringAttribute{
				Optional: true,
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

// TIP: ==== ASSIGN CRUD METHODS ====
// Data sources only have a read method.
func (d *dataSourceKeyGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// TIP: ==== DATA SOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the config
	// 3. Get information about a resource from AWS
	// 4. Set the ID, arguments, and attributes
	// 5. Set the tags
	// 6. Set the state
	// TIP: -- 1. Get a client connection to the relevant service
	conn := d.Meta().CloudFrontClient(ctx)

	// TIP: -- 2. Fetch the config
	var data dataSourceKeyGroupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var keyGroupID string

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		keyGroupID = data.ID.ValueString()
	} else {
		name := data.Name.ValueString()
		input := &cloudfront.ListKeyGroupsInput{}

		err := listKeyGroupsPages(ctx, conn, input, func(page *cloudfront.ListKeyGroupsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, policySummary := range page.KeyGroupList.Items {
				if keyGroup := policySummary.KeyGroup; aws.ToString(keyGroup.KeyGroupConfig.Name) == name {
					keyGroupID = aws.ToString(keyGroup.Id)

					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameKeyGroup, data.Name.String(), err),
				err.Error(),
			)
			return
		}

		if keyGroupID == "" {
			err := fmt.Errorf("no matching CloudFront Key Group (%s)", name)
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameKeyGroup, data.Name.String(), err),
				err.Error(),
			)
			return
		}
	}

	// TIP: -- 3. Get information about a resource from AWS
	out, err := findKeyGroupByID(ctx, conn, keyGroupID)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameKeyGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 4. Set the ID, arguments, and attributes
	// Using a field name prefix allows mapping fields such as `KeyGroupId` to `ID`
	resp.Diagnostics.Append(flex.Flatten(ctx, out.KeyGroup.KeyGroupConfig, &data, flex.WithFieldNamePrefix("KeyGroup"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 5. Set the tags
	data.ID = flex.StringToFramework(ctx, out.KeyGroup.Id)
	data.Name = flex.StringToFramework(ctx, out.KeyGroup.KeyGroupConfig.Name)
	data.Comment = flex.StringToFramework(ctx, out.KeyGroup.KeyGroupConfig.Comment)
	data.LastModifiedTime = flex.TimeToFramework(ctx, out.KeyGroup.LastModifiedTime)

	data.Etag = flex.StringToFramework(ctx, out.ETag)

	// TIP: -- 6. Set the state
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
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
type dataSourceKeyGroupModel struct {
	Etag             types.String      `tfsdk:"etag"`
	Items            types.List        `tfsdk:"items"`
	Comment          types.String      `tfsdk:"comment"`
	ID               types.String      `tfsdk:"id"`
	Name             types.String      `tfsdk:"name"`
	LastModifiedTime timetypes.RFC3339 `tfsdk:"last_modified_time"`
}

type complexArgumentModel struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}
