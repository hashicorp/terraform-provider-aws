// Copyright IBM Corp. 2014, 2026
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
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}"
	awstypes "github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)
{{ if .IncludeComments }}
// TIP: ==== FILE STRUCTURE ====
// All list resources should follow this basic outline. Improve this list resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main data source struct with schema method
// 4. Read method
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)
{{- end }}

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("{{ .ProviderResourceName }}")
func new{{ .ListResource }}ResourceAsListResource() list.ListResourceWithConfigure {
	return &{{ .ListResourceLowerCamel }}ListResource{}
}

var _ list.ListResource = &{{ .ListResourceLowerCamel }}ListResource{}

type {{ .ListResourceLowerCamel }}ListResource struct {
	{{ .ListResourceLowerCamel }}Resource
	framework.WithList
}

func (r *{{ .ListResourceLowerCamel }}ListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	{{- if .IncludeComments }}
	// TIP: ==== LIST RESOURCE LIST ====
	// Generally, the List function should do the following things. Make
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
	conn := r.Meta().{{ .Service }}Client(ctx)
	{{ if .IncludeComments }}
	// TIP: -- 2. Fetch the config
	{{- end }}
	var query list{{ .ListResource }}Model
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}
	{{ if .IncludeComments }}
	// TIP: -- 3. Get information about a resource from AWS
	{{- end }}
	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		var input {{ .SDKPackage }}.List{{ .ListResource }}sInput
		for item, err := range list{{ .ListResource }}s(ctx, conn, &input) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			var data {{ .ResourceLowerCamel }}ResourceModel
	        {{ if .IncludeComments -}}
	        // TIP: -- 4. Set the ID, arguments, and attributes
	        // Using a field name prefix allows mapping fields such as `{{ .ListResource }}Id` to `ID`
	        {{- end }}
			r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
				if diags := fwflex.Flatten(ctx, item, &data, fwflex.WithFieldNamePrefix("{{ .ListResource }}")); diags.HasError() {
					result.Diagnostics.Append(diags...)
					yield(result)
					return
				}

				{{ if .IncludeComments -}}
				// TIP: -- 5. Set the display name
				{{- end }}
				name := aws.ToString(item.{{ .ListResource }}Id)
				data.Name = fwflex.StringValueToFramework(ctx, name)
				result.DisplayName = name
			})

			if !yield(result) {
				return
			}
		}
	}
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
type list{{ .ListResource }}Model struct {
	framework.WithRegionModel
}
{{ if .IncludeComments }}
// TIP: ==== LISTING FUNCTION ====
// This listing function is written using an iterator pattern to handle pagination
{{- end }}
func list{{ .ListResource }}s(ctx context.Context, conn *{{ .SDKPackage }}.Client, input *{{ .SDKPackage }}.List{{ .ListResource }}sInput) iter.Seq2[awstypes.{{ .ListResource }}, error] {
	return func(yield func(awstypes.{{ .ListResource }}, error) bool) {
		pages := {{ .SDKPackage }}.NewList{{ .ListResource }}sPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.{{ .ListResource }}{}, fmt.Errorf("listing {{ .HumanFriendlyServiceShort }} {{ .HumanListResourceName }} resources: %w", err))
				return
			}

			for _, item := range page.{{ .ListResource }}s {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
