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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("{{ .ProviderResourceName }}")
func new{{ .ListResource }}ResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResource{{ .ListResource }}{}
	l.SetResourceSchema(resource{{ .ListResource }}())
	return &l
}

var _ list.ListResource = &listResource{{ .ListResource }}{}

type listResource{{ .ListResource }} struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResource{{ .ListResource }}) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
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
	conn := l.Meta().{{ .Service }}Client(ctx)
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
	tflog.Info(ctx, "Listing {{ .HumanFriendlyServiceShort }} {{ .HumanListResourceName }}")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input {{ .SDKPackage }}.List{{ .ListResource }}sInput
		for item, err := range list{{ .ListResource }}s(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.{{ .ListResource }}Arn)
	        {{- if .IncludeComments }}
			// TIP: -- 4. Set identifying attributes for logging
			// Set one or more logging fields with attributes that will identify the resource.
			// Typically, these will be the attributes used in the Resource Identity
	        {{- end }}
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), arn)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(arn)

			tflog.Info(ctx, "Reading {{ .HumanFriendlyServiceShort }} {{ .HumanListResourceName }}")
	        {{- if .IncludeComments }}
			// TIP: -- 5. Populate the resource
			// In many cases, the value in item will have sufficient information to populate the resource data.
			// If this is the case, factor out the resource flattening from resource{{ .ListResource }}Read.
			// See, for example, the implementation of `vpc_subnet_list`.
			//
			// If resource{{ .ListResource }}Read makes additional API calls to populate the resource data
			// they should only be made if request.IncludeResource is true.
	        {{- end }}
			diags := resource{{ .ListResource }}Read(ctx, rd, l.Meta())
			if diags.HasError() {
				tflog.Error(ctx, "Reading {{ .HumanFriendlyServiceShort }} {{ .HumanListResourceName }}", map[string]any{
					"diags": sdkdiag.DiagnosticsString(diags),
				})
				continue
			}
			if rd.Id() == "" {
				tflog.Warn(ctx, "Resource disappeared during listing, skipping")
				continue
			}

			result.DisplayName = aws.ToString(item.{{ .ListResource }}Name)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

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
