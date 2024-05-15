// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

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
	// using the services/shield/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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
// @FrameworkDataSource(name="Protection")
func newDataSourceProtection(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceProtection{}, nil
}

const (
	DSNameProtection = "Protection Data Source"
)

type dataSourceProtection struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceProtection) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_shield_protection"
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
func (d *dataSourceProtection) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"resource_arn": schema.StringAttribute{
				Optional: true,
			},
			"protection_id": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
		},
	}
}

// TIP: ==== ASSIGN CRUD METHODS ====
// Data sources only have a read method.
func (d *dataSourceProtection) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
	conn := d.Meta().ShieldClient(ctx)

	// TIP: -- 2. Fetch the config
	var data dataSourceProtectionData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &shield.DescribeProtectionInput{
		ProtectionId: data.ProtectionId.ValueStringPointer(),
		ResourceArn:  data.ResourceArn.ValueStringPointer(),
	}

	// TIP: -- 3. Get information about a resource from AWS
	out, err := findProtection(ctx, conn, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionReading, DSNameProtection, fmt.Sprintf("%s%s", data.ID.String(), data.ResourceArn.String()), err),
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
	data.ARN = flex.StringToFramework(ctx, out.ProtectionArn)
	data.ResourceArn = flex.StringToFramework(ctx, out.ResourceArn)
	data.ID = flex.StringToFramework(ctx, out.Id)
	data.ProtectionId = flex.StringToFramework(ctx, out.Id)

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	spew.Dump(data)
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
type dataSourceProtectionData struct {
	ARN          types.String `tfsdk:"arn"`
	ResourceArn  types.String `tfsdk:"resource_arn"`
	ProtectionId types.String `tfsdk:"protection_id"`
	ID           types.String `tfsdk:"id"`
}
