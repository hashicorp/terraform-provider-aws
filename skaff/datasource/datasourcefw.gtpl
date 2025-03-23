// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package {{ .ServicePackage }}

{{- if .IncludeComments }}
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
// it.{{- end }}

import (
{{- if .IncludeComments }}
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/{{ .SDKPackage }}/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
{{- end }}
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}"
	awstypes "github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/{{ .SDKPackage }}"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
{{- if .IncludeTags }}
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
{{- end }}
	"github.com/hashicorp/terraform-provider-aws/names"
)
{{ if .IncludeComments }}
// TIP: ==== FILE STRUCTURE ====
// All data sources should follow this basic outline. Improve this data source's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main data source struct with schema method
// 4. Read method
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)
{{- end }}

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("{{ .ProviderResourceName }}", name="{{ .HumanDataSourceName }}")
func newDataSource{{ .DataSource }}(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSource{{ .DataSource }}{}, nil
}

const (
	DSName{{ .DataSource }} = "{{ .HumanDataSourceName }} Data Source"
)

type dataSource{{ .DataSource }} struct {
	framework.DataSourceWithConfigure
}

{{ if .IncludeComments }}
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
{{- end }}
func (d *dataSource{{ .DataSource }}) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			{{- if .IncludeTags }}
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			{{- end }}
			names.AttrType: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"complex_argument": schema.ListNestedBlock{
				{{- if .IncludeComments }}
				// TIP: ==== CUSTOM TYPES ====
				// Use a custom type to identify the model type of the tested object
				{{- end }}
				CustomType: fwtypes.NewListNestedObjectTypeOf[complexArgumentModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						{{- if .IncludeComments }}
						// TIP: Attributes that are required on a corresponding resource will be 
						// computed on the data source (unless required as part of the search criteria).
						{{- end }}
						"nested_required": schema.StringAttribute{
							Computed: true,
						},
						"nested_computed": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

{{- if .IncludeComments }}
// TIP: ==== ASSIGN CRUD METHODS ====
// Data sources only have a read method.
{{- end }}
func (d *dataSource{{ .DataSource }}) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	{{- if .IncludeComments }}
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
	{{- end }}

	{{- if .IncludeComments }}
	// TIP: -- 1. Get a client connection to the relevant service
	{{- end }}
	conn := d.Meta().{{ .Service }}Client(ctx)
	{{ if .IncludeComments }}
	// TIP: -- 2. Fetch the config
	{{- end }}
	var data dataSource{{ .DataSource }}Model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	{{ if .IncludeComments }}
	// TIP: -- 3. Get information about a resource from AWS
	{{- end }}
	out, err := find{{ .DataSource }}ByName(ctx, conn, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.{{ .Service }}, create.ErrActionReading, DSName{{ .DataSource }}, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	{{ if .IncludeComments -}}
	// TIP: -- 4. Set the ID, arguments, and attributes
	// Using a field name prefix allows mapping fields such as `{{ .DataSource }}Id` to `ID`
	{{- end }}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("{{ .DataSource }}"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	{{ if .IncludeComments -}}
	// TIP: -- 5. Set the tags
	{{- end }}
	{{- if .IncludeTags }}
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)
	tags := KeyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	data.Tags = tftags.FlattenStringValueMap(ctx, tags.Map())
	{{- end }}

	{{ if .IncludeComments -}}
	// TIP: -- 6. Set the state
	{{- end }}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

{{ if .IncludeComments }}
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
{{- end }}
type dataSource{{ .DataSource }}Model struct {
	ARN             types.String                                          `tfsdk:"arn"`
	ComplexArgument fwtypes.ListNestedObjectValueOf[complexArgumentModel] `tfsdk:"complex_argument"`
	Description     types.String                                          `tfsdk:"description"`
	ID              types.String                                          `tfsdk:"id"`
	Name            types.String                                          `tfsdk:"name"`
	{{- if .IncludeTags }}
	Tags            tftags.Map                                            `tfsdk:"tags"`
	{{- end }}
	Type            types.String                                          `tfsdk:"type"`
}

type complexArgumentModel struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}
